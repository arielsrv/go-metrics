package metrics

import (
	"time"
)

type Metrics interface {
	IncrementCounter(metricName string, tags ...Tags)
	IncrementCounterFunc(metricName string, counterFunc CounterFunc)
	RecordExecutionTime(metricName string, duration time.Duration, tags ...Tags)
	RecordValue(metricName string, value float64, tags ...Tags)
	RecordValueFunc(metricName string, valueFunc RecordValueFunc)
}

type Tags map[string]string

type (
	CounterFunc     func() float64
	RecordValueFunc func() float64
)
