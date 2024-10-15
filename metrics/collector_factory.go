package metrics

type CollectorFactory interface {
	Prometheus() Metrics
}

type collectorFactory struct{}

func (r *collectorFactory) Prometheus() Metrics {
	return &prometheusCollector{}
}

var Collector *collectorFactory
