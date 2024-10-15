package metrics_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arielsrv/go-metric/metrics"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	addr, err := rndAddr()
	require.NoError(t, err)

	server := &http.Server{
		Addr: addr.HTTP,
	}

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/records", func(w http.ResponseWriter, _ *http.Request) {
		// counters
		metrics.Collector.Prometheus().IncrementCounter("my_counter", metrics.Tags{"type": "example"})
		metrics.Collector.Prometheus().IncrementCounter("my_counter", metrics.Tags{"type": "example"})
		metrics.Collector.Prometheus().IncrementCounter("my_counter", metrics.Tags{"another_type": "example"})
		metrics.Collector.Prometheus().IncrementCounter("my_counter_empty")
		metrics.Collector.Prometheus().IncrementCounter("my_counter_empty")
		metrics.Collector.Prometheus().IncrementCounterFunc("my_counter_func", func() float64 { return 1.0 })

		// time
		metrics.Collector.Prometheus().RecordExecutionTime("my_execution_time", time.Millisecond*1000)
		metrics.Collector.Prometheus().RecordExecutionTime("my_execution_time", time.Millisecond*2000)
		metrics.Collector.Prometheus().RecordExecutionTime("my_execution_time", time.Millisecond*3000)
		metrics.Collector.Prometheus().RecordExecutionTime("my_execution_time", time.Millisecond*3000, metrics.Tags{"type": "example"})

		// value
		metrics.Collector.Prometheus().RecordValue("my_value", 100.0)
		metrics.Collector.Prometheus().RecordValue("my_value", 100.0, metrics.Tags{"type": "example"})
		metrics.Collector.Prometheus().RecordValueFunc("my_value_func", func() float64 { return 100.0 })

		// sanitized
		metrics.Collector.Prometheus().RecordValue("0my-value", 100.0)
		metrics.Collector.Prometheus().RecordValue("record", 100.0)

		w.WriteHeader(http.StatusOK)
	})

	go func() {
		assert.True(t, ready(5, 500*time.Millisecond, fmt.Sprintf("%s/records", addr.HTTP), t))

		response, httpErr := http.Get(fmt.Sprintf("%s/metrics", addr.HTTP))
		assert.NoError(t, httpErr)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		body, httpErr := io.ReadAll(response.Body)
		assert.NoError(t, httpErr)

		want := fmt.Sprintf(`my_counter{type="example"} 2`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`empty 2`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`my_counter{another_type="example"} 1`)
		assert.False(t, strings.Contains(string(body), want))

		want = `my_execution_time{quantile="0.5"} 2000
my_execution_time{quantile="0.95"} 3000
my_execution_time{quantile="0.99"} 3000
my_execution_time_sum 6000
my_execution_time_count 3`
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`my_value 100`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`my_value_func 100`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`my_counter_func 1`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`_0my_value 100`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`record 100`)
		assert.Contains(t, string(body), want)

		assert.NoError(t, server.Shutdown(context.Background()))
	}()

	err = server.Serve(addr.Listener)
	require.Error(t, err)
	require.ErrorIs(t, err, http.ErrServerClosed)
}

func ready(attempts int, duration time.Duration, endpoint string, t *testing.T) bool {
	var prepared bool
	for range attempts {
		response, httpErr := http.Get(endpoint)
		if httpErr != nil {
			t.Logf("Error preparing metrics: %v", httpErr)
			time.Sleep(duration)
			continue
		}
		if response.StatusCode == http.StatusOK {
			prepared = true
			break
		}
		t.Logf("Failed to prepare metrics, status code: %d", response.StatusCode)
		time.Sleep(duration)
	}
	return prepared
}

func TestCounter_Err(t *testing.T) {
	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(counterVec))
	metrics.Collector.Prometheus().IncrementCounter("counter_err")
}

func TestCollector_RecordExecutionTime_Err(t *testing.T) {
	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "summary_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(summaryVec))
	start := time.Now()
	metrics.Collector.Prometheus().RecordExecutionTime("summary_err", time.Since(start))
}

func TestCollector_RecordValue_Err(t *testing.T) {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gauge_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(gaugeVec))
	metrics.Collector.Prometheus().RecordValue("gauge_err", 1)
}

func TestCollector_RecordValueFunc_Err(t *testing.T) {
	gaugeFunc := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "gauge_err_func",
		}, func() float64 {
			return 1.0
		},
	)
	require.NoError(t, prometheus.Register(gaugeFunc))
	metrics.Collector.Prometheus().RecordValueFunc("gauge_err_func", func() float64 {
		return 1.0
	})
}

func TestCollector_IncrementCounterFunc_Err(t *testing.T) {
	counterFunc := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "counter_err_func",
		}, func() float64 {
			return 1.0
		},
	)
	require.NoError(t, prometheus.Register(counterFunc))
	metrics.Collector.Prometheus().IncrementCounterFunc("counter_err_func", func() float64 {
		return 1.0
	})
}
