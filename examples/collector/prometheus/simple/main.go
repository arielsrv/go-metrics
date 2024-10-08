package main

import (
	"github.com/arielsrv/go-metric/metrics"
	"github.com/arielsrv/go-metric/metrics/collector"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	router.HandleFunc("/record/{id}", func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		id, found := vars["id"]
		if !found {
			http.Error(writer, "missing id", http.StatusBadRequest)
			return
		}

		collector.Prometheus.IncrementCounter("record", metrics.Tags{"id": id})

		writer.WriteHeader(http.StatusOK)
		length, err := writer.Write([]byte("record created"))
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
