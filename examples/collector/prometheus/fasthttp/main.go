package main

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/arielsrv/go-metric/metrics"
	"github.com/arielsrv/go-metric/metrics/collector"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

func main() {
	server := fiber.New(fiber.Config{
		EnablePrintRoutes:     false,
		DisableStartupMessage: true,
	})

	collector.Prometheus.IncrementCounter("users_status", metrics.Tags{"status": "success"})
	collector.Prometheus.IncrementCounter("users_status", metrics.Tags{"status": "success"})
	collector.Prometheus.IncrementCounter("users_created")
	collector.Prometheus.IncrementCounter("users_created")
	collector.Prometheus.IncrementCounter("users_created")
	collector.Prometheus.IncrementCounter("users_created")
	collector.Prometheus.IncrementCounter("users_created")

	collector.Prometheus.IncrementCounter("order_status", metrics.Tags{"status": "success"}, metrics.Tags{"order_type": "purchase"})

	collector.Prometheus.RecordValue("my_metric", 123.45)

	fiberPrometheus := fiberprometheus.NewWithRegistry(prometheus.DefaultRegisterer, "", "", "", nil)
	fiberPrometheus.RegisterAt(server, "/metrics")
	server.Use(fiberPrometheus.Middleware)

	slog.Info("Server started on :3000")
	if err := server.Listen(":3000"); err != nil {
		panic(err)
	}
}
