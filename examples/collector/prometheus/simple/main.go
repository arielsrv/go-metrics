package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/arielsrv/go-metric/metrics"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	// counters
	metrics.Collector.Prometheus().IncrementCounter("business_counter", metrics.Tags{"type": "example"})
	metrics.Collector.Prometheus().IncrementCounter("business_counter", metrics.Tags{"type": "example"})

	// fixed values
	metrics.Collector.Prometheus().RecordValue("business_value", 100, map[string]string{"section": "pdp"})

	// duration, percentiles
	start := time.Now()
	metrics.Collector.Prometheus().RecordExecutionTime("business_request_duration", time.Since(start), map[string]string{"path": "/api/v1/products"})

	slog.Info("server started, metrics on http://localhost:8081/metrics")
	if err := http.ListenAndServe(":8081", router); err != nil {
		panic(err)
	}
}
