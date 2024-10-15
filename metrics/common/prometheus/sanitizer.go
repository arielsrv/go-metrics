package prometheus

import (
	"fmt"
	"regexp"
)

var metricNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func SanitizeMetricName(metricName string) string {
	sanitized := metricNameRegex.ReplaceAllString(metricName, "_")
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "_" + sanitized
	}

	return fmt.Sprintf("%s", sanitized)
}
