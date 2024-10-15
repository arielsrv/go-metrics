package main

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/arielsrv/go-metric/metrics"

	"github.com/alitto/pond"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	router := mux.NewRouter()

	config := struct {
		MaxWorkers  int
		MaxCapacity int
	}{
		MaxWorkers:  100,
		MaxCapacity: 1000,
	}

	pool := pond.New(config.MaxWorkers, config.MaxCapacity)

	metrics.Collector.Prometheus().RecordValue("pool_max_workers", float64(config.MaxWorkers))
	metrics.Collector.Prometheus().RecordValue("pool_max_capacity", float64(config.MaxCapacity))

	httpClient := &http.Client{
		Timeout: time.Duration(10000) * time.Millisecond,
		Transport: &http.Transport{
			MaxConnsPerHost:     config.MaxWorkers,
			MaxIdleConnsPerHost: config.MaxWorkers / 10,
		},
	}

	metrics.Collector.Prometheus().RecordValueFunc("pool_workers_running", func() float64 { return float64(pool.RunningWorkers()) })
	metrics.Collector.Prometheus().RecordValueFunc("pool_workers_idle", func() float64 { return float64(pool.IdleWorkers()) })
	metrics.Collector.Prometheus().RecordValueFunc("pool_tasks_waiting", func() float64 { return float64(pool.WaitingTasks()) })

	metrics.Collector.Prometheus().IncrementCounterFunc("pool_tasks_submitted_total", func() float64 { return float64(pool.SubmittedTasks()) })
	metrics.Collector.Prometheus().IncrementCounterFunc("pool_tasks_successful_total", func() float64 { return float64(pool.SuccessfulTasks()) })
	metrics.Collector.Prometheus().IncrementCounterFunc("pool_tasks_failed_total", func() float64 { return float64(pool.FailedTasks()) })
	metrics.Collector.Prometheus().IncrementCounterFunc("pool_tasks_completed_total", func() float64 { return float64(pool.CompletedTasks()) })

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/record/{id}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id, found := vars["id"]
		if !found {
			http.Error(writer, "missing id", http.StatusBadRequest)
			return
		}

		metrics.Collector.Prometheus().IncrementCounter("record", metrics.Tags{"id": id})

		metrics.Collector.Prometheus().IncrementCounter("users_status", metrics.Tags{"status": "success"})
		metrics.Collector.Prometheus().IncrementCounter("users_status", metrics.Tags{"status": "success"})
		metrics.Collector.Prometheus().IncrementCounter("users_created")
		metrics.Collector.Prometheus().IncrementCounter("users_created")
		metrics.Collector.Prometheus().IncrementCounter("users_created")
		metrics.Collector.Prometheus().IncrementCounter("users_created")
		metrics.Collector.Prometheus().IncrementCounter("users_created")
		metrics.Collector.Prometheus().IncrementCounter("order_status", metrics.Tags{"status": "success"}, metrics.Tags{"order_type": "purchase"})
		metrics.Collector.Prometheus().RecordValue("my_value", 100)
		metrics.Collector.Prometheus().RecordValue("my_value_by_env", 100, metrics.Tags{"env": "production"})

		for range 1000 {
			pool.Submit(func() {
				start := time.Now()
				apiURL := "https://gorest.co.in/public/v2/users"
				response, err := httpClient.Get(apiURL)
				metrics.Collector.Prometheus().RecordExecutionTime("http_request_duration_seconds", time.Since(start), metrics.Tags{"URL": apiURL}, metrics.Tags{"method": "GET"}, metrics.Tags{"http_version": "1.1"})
				if err != nil {
					var netError net.Error
					if errors.As(err, &netError) && netError.Timeout() {
						metrics.Collector.Prometheus().IncrementCounter("httpclient_error", metrics.Tags{"type": "timeout"})
						return
					}
					metrics.Collector.Prometheus().IncrementCounter("httpclient_error", metrics.Tags{"type": "network"})
					return
				}
				metrics.Collector.Prometheus().IncrementCounter("httpclient_status", metrics.Tags{"status_code": strconv.Itoa(response.StatusCode)})
			})
		}

		writer.WriteHeader(http.StatusOK)
		length, err := writer.Write([]byte("Record created"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		slog.Debug("[metrics-collector]: Wrote %d bytes to response for record creation", length)
	})

	slog.Info("Server started on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}
