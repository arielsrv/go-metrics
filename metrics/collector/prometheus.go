package collector

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/arielsrv/go-metric/metrics"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"github.com/prometheus/client_golang/prometheus"
)

type collector struct{}

var (
	Prometheus *collector

	counters    = make(map[string]*prometheus.CounterVec)
	countersMtx sync.Mutex

	countersFunc    = make(map[string]prometheus.CounterFunc)
	countersFuncMtx sync.Mutex

	summaries    = make(map[string]*prometheus.SummaryVec)
	summariesMtx sync.Mutex

	gauges    = make(map[string]*prometheus.GaugeVec)
	gaugesMtx sync.Mutex

	gaugesFunc    = make(map[string]prometheus.GaugeFunc)
	gaugesFuncMtx sync.Mutex
)

func (r *collector) IncrementCounter(metricName string, tags ...metrics.Tags) {
	labels := r.buildLabels(tags)
	counterVec, err := r.getOrAddCounterVec(metricName, labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting or creating counter for %s: %v", metricName, err)
		return
	}

	metric, err := counterVec.GetMetricWith(labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Inc()
}

func (r *collector) IncrementCounterFunc(metricName string, counterFunc metrics.CounterFunc) {
	_, found := countersFunc[metricName]
	if !found {
		countersFuncMtx.Lock()
		defer countersFuncMtx.Unlock()
		_, found = countersFunc[metricName]
		if !found {
			counter := prometheus.NewCounterFunc(
				prometheus.CounterOpts{
					Name: r.buildMetricName(metricName),
				}, counterFunc,
			)
			if err := prometheus.Register(counter); err != nil {
				log.Errorf("[metrics-collector]: Error registering counterFunc for %s: %v", metricName, err)
				return
			}
			countersFunc[metricName] = counter
		}
	}
}

func (r *collector) getOrAddCounterVec(metricName string, labels prometheus.Labels) (*prometheus.CounterVec, error) {
	counter, found := counters[metricName]
	if !found {
		countersMtx.Lock()
		defer countersMtx.Unlock()
		counter, found = counters[metricName]
		if !found {
			counterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: r.buildMetricName(metricName),
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(counterVec); err != nil {
				log.Errorf("[metrics-collector]: Error registering counterVec for %s: %v", metricName, err)
				return nil, err
			}
			counters[metricName] = counterVec
			return counterVec, nil
		}
	}

	return counter, nil
}

func (r *collector) getOrAddSummaryVec(metricName string, labels prometheus.Labels) (*prometheus.SummaryVec, error) {
	summary, found := summaries[metricName]
	if !found {
		summariesMtx.Lock()
		defer summariesMtx.Unlock()
		summary, found = summaries[metricName]
		if !found {
			summaryVec := prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: r.buildMetricName(metricName),
					Objectives: map[float64]float64{
						0.5:  0.05,  // Average
						0.95: 0.01,  // P95
						0.99: 0.001, // P99
					},
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(summaryVec); err != nil {
				log.Errorf("[metrics-collector]: Error registering sumaryVec for %s: %v", metricName, err)
				return nil, err
			}
			summaries[metricName] = summaryVec

			return summaryVec, nil
		}
	}

	return summary, nil
}

func (r *collector) RecordExecutionTime(metricName string, duration time.Duration, tags ...metrics.Tags) {
	labels := r.buildLabels(tags)
	summaryVec, err := r.getOrAddSummaryVec(metricName, labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting or creating summary for %s: %v", metricName, err)
		return
	}

	metric, err := summaryVec.GetMetricWith(labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Observe(float64(duration.Milliseconds()))
}

func (r *collector) RecordValue(metricName string, value float64, tags ...metrics.Tags) {
	labels := r.buildLabels(tags)
	gaugeVec, err := r.getOrAddGaugeVec(metricName, labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting or creating gauge for %s: %v", metricName, err)
		return
	}

	metric, err := gaugeVec.GetMetricWith(labels)
	if err != nil {
		log.Errorf("[metrics-collector]: Error getting metric for %s: %v", metricName, err)
		return
	}

	metric.Set(value)
}

func (r *collector) RecordValueFunc(metricName string, valueFunc metrics.RecordValueFunc) {
	_, found := gaugesFunc[metricName]
	if !found {
		gaugesFuncMtx.Lock()
		defer gaugesFuncMtx.Unlock()
		_, found = gaugesFunc[metricName]
		if !found {
			gaugeFunc := prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Name: r.buildMetricName(metricName),
				}, valueFunc,
			)
			if err := prometheus.Register(gaugeFunc); err != nil {
				log.Errorf("[metrics-collector]: Error registering gaugeVec for %s: %v", metricName, err)
				return
			}
			gaugesFunc[metricName] = gaugeFunc
		}
	}
}

func (r *collector) getOrAddGaugeVec(metricName string, labels prometheus.Labels) (*prometheus.GaugeVec, error) {
	gauge, found := gauges[metricName]
	if !found {
		gaugesMtx.Lock()
		defer gaugesMtx.Unlock()
		gauge, found = gauges[metricName]
		if !found {
			gaugeVec := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: r.buildMetricName(metricName),
				}, slices.Collect(maps.Keys(labels)),
			)
			if err := prometheus.Register(gaugeVec); err != nil {
				log.Errorf("[metrics-collector]: Error registering gaugeVec for %s: %v", metricName, err)
				return nil, err
			}
			gauges[metricName] = gaugeVec

			return gaugeVec, nil
		}
	}

	return gauge, nil
}

func (r *collector) buildMetricName(metricName string) string {
	return fmt.Sprintf("__%s", metricName)
}

func (r *collector) buildLabels(tags []metrics.Tags) prometheus.Labels {
	if len(tags) == 0 {
		tags = append(tags, metrics.Tags{})
	}

	return prometheus.Labels(tags[0])
}
