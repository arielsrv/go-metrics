package collector_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/arielsrv/go-metric/metrics"

	"github.com/arielsrv/go-metric/metrics/collector"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	addr, ok := listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)

	hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(addr.Port))
	server := &http.Server{Addr: hostPort}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/records", func(w http.ResponseWriter, _ *http.Request) {
		// counters
		collector.Prometheus.IncrementCounter("my_counter", metrics.Tags{"type": "example"})
		collector.Prometheus.IncrementCounter("my_counter", metrics.Tags{"type": "example"})
		collector.Prometheus.IncrementCounter("my_counter", metrics.Tags{"another_type": "example"})
		collector.Prometheus.IncrementCounter("my_counter_empty")
		collector.Prometheus.IncrementCounter("my_counter_empty")
		collector.Prometheus.IncrementCounterFunc("my_counter_func", func() float64 { return 1.0 })

		// time
		collector.Prometheus.RecordExecutionTime("my_execution_time", time.Millisecond*1000)
		collector.Prometheus.RecordExecutionTime("my_execution_time", time.Millisecond*2000)
		collector.Prometheus.RecordExecutionTime("my_execution_time", time.Millisecond*3000)
		collector.Prometheus.RecordExecutionTime("my_execution_time", time.Millisecond*3000, metrics.Tags{"type": "example"})

		// value
		collector.Prometheus.RecordValue("my_value", 100.0)
		collector.Prometheus.RecordValue("my_value", 100.0, metrics.Tags{"type": "example"})
		collector.Prometheus.RecordValueFunc("my_value_func", func() float64 { return 100.0 })

		w.WriteHeader(http.StatusOK)
	})
	go func() {
		time.Sleep(time.Millisecond * 100)

		baseURL := fmt.Sprintf("http://%s", server.Addr)

		response, httpErr := http.Get(fmt.Sprintf("%s/records", baseURL))
		assert.NoError(t, httpErr)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		response, httpErr = http.Get(fmt.Sprintf("%s/metrics", baseURL))
		assert.NoError(t, httpErr)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		body, httpErr := io.ReadAll(response.Body)
		assert.NoError(t, httpErr)

		// assertion
		want := fmt.Sprintf(`__my_counter{type="example"} 2`)
		assert.Contains(t, string(body), want)

		// assertion
		want = fmt.Sprintf(`empty 2`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`__my_counter{another_type="example"} 1`)
		assert.False(t, strings.Contains(string(body), want))

		want = `__my_execution_time{quantile="0.5"} 2000
__my_execution_time{quantile="0.95"} 3000
__my_execution_time{quantile="0.99"} 3000
__my_execution_time_sum 6000
__my_execution_time_count 3`
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`__my_value 100`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`__my_value_func 100`)
		assert.Contains(t, string(body), want)

		want = fmt.Sprintf(`__my_counter_func 1`)
		assert.Contains(t, string(body), want)

		assert.NoError(t, server.Shutdown(context.Background()))
	}()

	err = server.Serve(listener)
	require.Error(t, err)
	require.ErrorIs(t, err, http.ErrServerClosed)
}

func TestCounter_Err(t *testing.T) {
	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "__counter_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(counterVec))
	collector.Prometheus.IncrementCounter("counter_err")
}

func TestCollector_RecordExecutionTime_Err(t *testing.T) {
	summaryVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "__summary_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(summaryVec))
	start := time.Now()
	collector.Prometheus.RecordExecutionTime("summary_err", time.Since(start))
}

func TestCollector_RecordValue_Err(t *testing.T) {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "__gauge_err",
		}, []string{},
	)
	require.NoError(t, prometheus.Register(gaugeVec))
	collector.Prometheus.RecordValue("gauge_err", 1)
}

func TestCollector_RecordValueFunc_Err(t *testing.T) {
	gaugeFunc := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "__gauge_err_func",
		}, func() float64 {
			return 1.0
		},
	)
	require.NoError(t, prometheus.Register(gaugeFunc))
	collector.Prometheus.RecordValueFunc("gauge_err_func", func() float64 {
		return 1.0
	})
}

func TestCollector_IncrementCounterFunc_Err(t *testing.T) {
	counterFunc := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "__counter_err_func",
		}, func() float64 {
			return 1.0
		},
	)
	require.NoError(t, prometheus.Register(counterFunc))
	collector.Prometheus.IncrementCounterFunc("counter_err_func", func() float64 {
		return 1.0
	})
}
