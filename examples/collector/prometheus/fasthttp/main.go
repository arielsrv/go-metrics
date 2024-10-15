package main

import (
	"log/slog"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/arielsrv/go-metric/metrics"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	server := fiber.New(fiber.Config{
		EnablePrintRoutes:     false,
		DisableStartupMessage: true,
	})

	metrics.Collector.Prometheus().IncrementCounter("users_status", metrics.Tags{"status": "success"})
	metrics.Collector.Prometheus().IncrementCounter("users_status", metrics.Tags{"status": "success"})
	metrics.Collector.Prometheus().IncrementCounter("users_created")
	metrics.Collector.Prometheus().IncrementCounter("users_created")
	metrics.Collector.Prometheus().IncrementCounter("users_created")
	metrics.Collector.Prometheus().IncrementCounter("users_created")
	metrics.Collector.Prometheus().IncrementCounter("users_created")

	metrics.Collector.Prometheus().IncrementCounter("order_status", metrics.Tags{"status": "success"}, metrics.Tags{"order_type": "purchase"})

	metrics.Collector.Prometheus().RecordValue("my_metric", 123.45)

	fiberPrometheus := fiberprometheus.NewWithRegistry(prometheus.DefaultRegisterer, "", "", "", nil)
	fiberPrometheus.RegisterAt(server, "/metrics")
	server.Use(fiberPrometheus.Middleware)

	slog.Info("Server started on :3000")
	if err := server.Listen(":3000"); err != nil {
		panic(err)
	}
}
