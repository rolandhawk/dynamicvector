// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Counter is a counter dynamicvector
type Counter struct {
	*Vector
}

// NewCounter will return a new dynamicvector counter.
func NewCounter(opts CounterOpts) *Counter {
	return &Counter{NewVector(opts, NewCounterUnit)}
}

// With is a syntatic sugar for Vector.GetMetricWith
func (c *Counter) GetMetricWith(labels prometheus.Labels) (prometheus.Counter, error) {
	metric, err := c.Vector.GetMetricWith(labels)
	if err != nil {
		return nil, err
	}

	return metric.(prometheus.Counter), nil
}

// With is a syntatic sugar for Vector.With
func (c *Counter) With(labels prometheus.Labels) prometheus.Counter {
	return c.Vector.With(labels).(prometheus.Counter)
}

// CounterUnit implement prometheus.Counter and Metric
type CounterUnit struct {
	val    float64
	vec    *Vector
	labels []string
	last   time.Time

	mtx sync.RWMutex
}

// NewCounterUnit will create new counter with specified label values.
func NewCounterUnit(vec *Vector, labelValues []string) Metric {
	return &CounterUnit{
		vec:    vec,
		labels: labelValues,
		last:   time.Now(),
	}
}

// Desc implement prometheus.Counter (prometheus.Metric)
func (u *CounterUnit) Desc() *prometheus.Desc {
	return u.vec.desc
}

// Write implement prometheus.Counter (prometheus.Metric)
func (u *CounterUnit) Write(metric *dto.Metric) error {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	metric.Label = LabelsProto(u.vec.labels.ValuesToPromLabels(u.labels))
	metric.Counter = &dto.Counter{Value: proto.Float64(u.val)}

	return nil
}

// Describe implement prometheus.Counter (prometheus.Collector)
func (u *CounterUnit) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.vec.desc
}

// Collect implement prometheus.Counter (prometheus.Collector)
func (u *CounterUnit) Collect(ch chan<- prometheus.Metric) {
	ch <- u
}

// Inc implement prometheus.Counter
func (u *CounterUnit) Inc() {
	u.Add(1)
}

// Add implement prometheus.Counter
func (u *CounterUnit) Add(val float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val += val
	u.last = time.Now()
}

// LastEdit implement Metric
func (u *CounterUnit) LastEdit() time.Time {
	return u.last
}
