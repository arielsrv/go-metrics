package metrics

import (
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	prometheus2 "github.com/arielsrv/go-metric/metrics/common/prometheus"

	"github.com/prometheus/client_golang/prometheus"
)

type prometheusCollector struct{}

var (
	counters        = make(map[string]*prometheus.CounterVec)
	countersMtx     sync.Mutex
	countersFunc    = make(map[string]prometheus.CounterFunc)
	countersFuncMtx sync.Mutex
	summaries       = make(map[string]*prometheus.SummaryVec)
	summariesMtx    sync.Mutex
	gauges          = make(map[string]*prometheus.GaugeVec)
	gaugesMtx       sync.Mutex
	gaugesFunc      = make(map[string]prometheus.GaugeFunc)
	gaugesFuncMtx   sync.Mutex
)

func (r *prometheusCollector) IncrementCounter(metricName string, tags ...Tags) {
	labels := r.buildLabels(tags)
	counterVec, err := r.getOrAddCounterVec(metricName, labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting or creating counter for %s: %v", metricName, err)
		return
	}

	metric, err := counterVec.GetMetricWith(labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Inc()
}

func (r *prometheusCollector) IncrementCounterFunc(metricName string, counterFunc CounterFunc) {
	_, found := countersFunc[metricName]
	if !found {
		countersFuncMtx.Lock()
		defer countersFuncMtx.Unlock()
		_, found = countersFunc[metricName]
		if !found {
			counter := prometheus.NewCounterFunc(
				prometheus.CounterOpts{
					Name: prometheus2.SanitizeMetricName(metricName),
				}, counterFunc,
			)
			if err := prometheus.Register(counter); err != nil {
				slog.Error("[metrics-prometheusCollector]: Error registering counterFunc for %s: %v", metricName, err)
				return
			}
			countersFunc[metricName] = counter
		}
	}
}

func (r *prometheusCollector) getOrAddCounterVec(metricName string, labels prometheus.Labels) (*prometheus.CounterVec, error) {
	counter, found := counters[metricName]
	if !found {
		countersMtx.Lock()
		defer countersMtx.Unlock()
		counter, found = counters[metricName]
		if !found {
			counterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: prometheus2.SanitizeMetricName(metricName),
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(counterVec); err != nil {
				slog.Error("[metrics-prometheusCollector]: Error registering counterVec for %s: %v", metricName, err)
				return nil, err
			}
			counters[metricName] = counterVec
			return counterVec, nil
		}
	}

	return counter, nil
}

func (r *prometheusCollector) getOrAddSummaryVec(metricName string, labels prometheus.Labels) (*prometheus.SummaryVec, error) {
	summary, found := summaries[metricName]
	if !found {
		summariesMtx.Lock()
		defer summariesMtx.Unlock()
		summary, found = summaries[metricName]
		if !found {
			summaryVec := prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: prometheus2.SanitizeMetricName(metricName),
					Objectives: map[float64]float64{
						0.5:  0.05,  // Average
						0.95: 0.01,  // P95
						0.99: 0.001, // P99
					},
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(summaryVec); err != nil {
				slog.Error("[metrics-prometheusCollector]: Error registering sumaryVec for %s: %v", metricName, err)
				return nil, err
			}
			summaries[metricName] = summaryVec

			return summaryVec, nil
		}
	}

	return summary, nil
}

func (r *prometheusCollector) RecordExecutionTime(metricName string, duration time.Duration, tags ...Tags) {
	labels := r.buildLabels(tags)
	summaryVec, err := r.getOrAddSummaryVec(metricName, labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting or creating summary for %s: %v", metricName, err)
		return
	}

	metric, err := summaryVec.GetMetricWith(labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Observe(float64(duration.Milliseconds()))
}

func (r *prometheusCollector) RecordValue(metricName string, value float64, tags ...Tags) {
	labels := r.buildLabels(tags)
	gaugeVec, err := r.getOrAddGaugeVec(metricName, labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting or creating gauge for %s: %v", metricName, err)
		return
	}

	metric, err := gaugeVec.GetMetricWith(labels)
	if err != nil {
		slog.Error("[metrics-prometheusCollector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Set(value)
}

func (r *prometheusCollector) RecordValueFunc(metricName string, valueFunc RecordValueFunc) {
	_, found := gaugesFunc[metricName]
	if !found {
		gaugesFuncMtx.Lock()
		defer gaugesFuncMtx.Unlock()
		_, found = gaugesFunc[metricName]
		if !found {
			gaugeFunc := prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: prometheus2.SanitizeMetricName(metricName),
				}, valueFunc,
			)
			if err := prometheus.Register(gaugeFunc); err != nil {
				slog.Error("[metrics-prometheusCollector]: Error registering gaugeVec for %s: %v", metricName, err)
				return
			}
			gaugesFunc[metricName] = gaugeFunc
		}
	}
}

func (r *prometheusCollector) getOrAddGaugeVec(metricName string, labels prometheus.Labels) (*prometheus.GaugeVec, error) {
	gauge, found := gauges[metricName]
	if !found {
		gaugesMtx.Lock()
		defer gaugesMtx.Unlock()
		gauge, found = gauges[metricName]
		if !found {
			gaugeVec := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: prometheus2.SanitizeMetricName(metricName),
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(gaugeVec); err != nil {
				slog.Error("[metrics-prometheusCollector]: Error registering gaugeVec for %s: %v", metricName, err)
				return nil, err
			}
			gauges[metricName] = gaugeVec

			return gaugeVec, nil
		}
	}

	return gauge, nil
}

func (r *prometheusCollector) buildLabels(tags []Tags) prometheus.Labels {
	if len(tags) == 0 {
		tags = append(tags, Tags{})
	}

	return prometheus.Labels(tags[0])
}
