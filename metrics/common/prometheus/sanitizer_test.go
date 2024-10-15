package prometheus_test

import (
	"testing"

	"github.com/arielsrv/go-metric/metrics/common/prometheus"

	"github.com/stretchr/testify/assert"
)

func TestSanitizePrometheusMetricName(t *testing.T) {
	actual := prometheus.SanitizeMetricName("my-metric-name")
	assert.Equal(t, "my_metric_name", actual)

	actual = prometheus.SanitizeMetricName("0my-metric-name")
	assert.Equal(t, "_0my_metric_name", actual)

	actual = prometheus.SanitizeMetricName("my_metric_name")
	assert.Equal(t, "my_metric_name", actual)
}
