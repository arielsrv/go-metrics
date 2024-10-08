# go-metrics

> This package provides a high-level abstract for Prometheus collector

## example prometheus

```go
package main

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/alitto/pond"
	"github.com/pkg/errors"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/arielsrv/go-metric/metrics"
	"github.com/arielsrv/go-metric/metrics/collector"
)

func main() {
	router := mux.NewRouter()

	pool := pond.New(100, 1000)

	httpClient := &http.Client{
		Timeout: time.Duration(10000) * time.Millisecond,
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

		pool.StopAndWait()

		writer.WriteHeader(http.StatusOK)
		length, err := writer.Write([]byte("Record created"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		log.Debugf("[metrics-collector]: Wrote %d bytes to response for record creation", length)
	})

	log.Infof("Server started on :3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}

```

```text
# HELP my_execution_time
# TYPE my_execution_time summary
my_execution_time{status="success",quantile="0.5"} 683
my_execution_time{status="success",quantile="0.95"} 683
my_execution_time{status="success",quantile="0.99"} 683
my_execution_time_sum{status="success"} 683
my_execution_time_count{status="success"} 1
# HELP my_value
# TYPE my_value gauge
my_value 100
# HELP my_value_by_env
# TYPE my_value_by_env gauge
my_value_by_env{env="production"} 100
# HELP order_status
# TYPE order_status counter
order_status{order_type="purchase",status="success"} 1
# HELP users_created
# TYPE users_created counter
users_created 5
# HELP users_status
# TYPE users_status counter
users_status{status="success"} 2
```

## concurrent benchmark

```text
goos: darwin
goarch: arm64
pkg: github.com/arielsrv/go-metric/metrics/collector
cpu: Apple M1 Pro
BenchmarkConcurrentIncrementCounter
BenchmarkConcurrentIncrementCounter-10      5549236        214.5 ns/op
PASS
```
