package main

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/alitto/pond"
	"github.com/pkg/errors"

	"github.com/arielsrv/go-metric/metrics"
	"github.com/arielsrv/go-metric/metrics/collector"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	router := mux.NewRouter()

	var config = struct {
		MaxWorkers  int
		MaxCapacity int
	}{
		MaxWorkers:  100,
		MaxCapacity: 1000,
	}

	pool := pond.New(config.MaxWorkers, config.MaxCapacity)

	collector.Prometheus.RecordValue("pool_max_workers", float64(config.MaxWorkers))
	collector.Prometheus.RecordValue("pool_max_capacity", float64(config.MaxCapacity))

	httpClient := &http.Client{
		Timeout: time.Duration(10000) * time.Millisecond,
		Transport: &http.Transport{
			MaxConnsPerHost:     config.MaxWorkers,
			MaxIdleConnsPerHost: config.MaxWorkers / 10,
		},
	}

	collector.Prometheus.RecordValueFunc("pool_workers_running", func() float64 { return float64(pool.RunningWorkers()) })
	collector.Prometheus.RecordValueFunc("pool_workers_idle", func() float64 { return float64(pool.IdleWorkers()) })
	collector.Prometheus.RecordValueFunc("pool_tasks_waiting", func() float64 { return float64(pool.WaitingTasks()) })

	collector.Prometheus.IncrementCounterFunc("pool_tasks_submitted_total", func() float64 { return float64(pool.SubmittedTasks()) })
	collector.Prometheus.IncrementCounterFunc("pool_tasks_successful_total", func() float64 { return float64(pool.SuccessfulTasks()) })
	collector.Prometheus.IncrementCounterFunc("pool_tasks_failed_total", func() float64 { return float64(pool.FailedTasks()) })
	collector.Prometheus.IncrementCounterFunc("pool_tasks_completed_total", func() float64 { return float64(pool.CompletedTasks()) })

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/record/{id}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id, found := vars["id"]
		if !found {
			http.Error(writer, "missing id", http.StatusBadRequest)
			return
		}

		collector.Prometheus.IncrementCounter("record", metrics.Tags{"id": id})

		collector.Prometheus.IncrementCounter("users_status", metrics.Tags{"status": "success"})
		collector.Prometheus.IncrementCounter("users_status", metrics.Tags{"status": "success"})
		collector.Prometheus.IncrementCounter("users_created")
		collector.Prometheus.IncrementCounter("users_created")
		collector.Prometheus.IncrementCounter("users_created")
		collector.Prometheus.IncrementCounter("users_created")
		collector.Prometheus.IncrementCounter("users_created")
		collector.Prometheus.IncrementCounter("order_status", metrics.Tags{"status": "success"}, metrics.Tags{"order_type": "purchase"})
		collector.Prometheus.RecordValue("my_value", 100)
		collector.Prometheus.RecordValue("my_value_by_env", 100, metrics.Tags{"env": "production"})

		for range 1000 {
			pool.Submit(func() {
				start := time.Now()
				apiURL := "https://gorest.co.in/public/v2/users"
				response, httpErr := httpClient.Get(apiURL)
				collector.Prometheus.RecordExecutionTime("http_request_duration_seconds", time.Since(start), metrics.Tags{"URL": apiURL}, metrics.Tags{"method": "GET"}, metrics.Tags{"http_version": "1.1"})
				if httpErr != nil {
					var netError net.Error
					if errors.As(httpErr, &netError) && netError.Timeout() {
						collector.Prometheus.IncrementCounter("httpclient_error", metrics.Tags{"type": "timeout"})
						return
					}
					collector.Prometheus.IncrementCounter("httpclient_error", metrics.Tags{"type": "network"})
					return
				}
				collector.Prometheus.IncrementCounter("httpclient_status", metrics.Tags{"status_code": strconv.Itoa(response.StatusCode)})
			})
		}

		writer.WriteHeader(http.StatusOK)
		length, err := writer.Write([]byte("Record created"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		slog.Debug("[metrics-collector]: Wrote %d bytes to response for record creation", slog.Int("length", length))
	})

	slog.Info("Server started on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}
