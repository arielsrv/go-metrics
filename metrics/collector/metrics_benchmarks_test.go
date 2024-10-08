package collector_test

import (
	"sync"
	"testing"

	"github.com/arielsrv/go-metric/metrics"

	"github.com/arielsrv/go-metric/metrics/collector"
)

func BenchmarkIncrementCounter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		collector.Prometheus.IncrementCounter("my_counter", metrics.Tags{"type": "example"})
	}
}

func BenchmarkConcurrentIncrementCounter(b *testing.B) {
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range b.N / 100 {
				collector.Prometheus.IncrementCounter("my_counter", metrics.Tags{"type": "example"})
			}
		}()
	}
	wg.Wait()
}
