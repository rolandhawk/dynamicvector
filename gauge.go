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

// Gauge is a gauge dynamicvector
type Gauge struct {
	*Vector
}

// NewGauge will return a new dynamicvector gauge.
func NewGauge(opts GaugeOpts) *Gauge {
	return &Gauge{NewVector(opts, NewGaugeUnit)}
}

// With is a syntatic sugar for Vector.GetMetricWith
func (g *Gauge) GetMetricWith(labels prometheus.Labels) (prometheus.Gauge, error) {
	metric, err := g.Vector.GetMetricWith(labels)
	if err != nil {
		return nil, err
	}

	return metric.(prometheus.Gauge), nil
}

// With is a syntatic sugar for Vector.With(labels).(prometheus.Gauge)
func (g *Gauge) With(labels prometheus.Labels) prometheus.Gauge {
	return g.Vector.With(labels).(prometheus.Gauge)
}

// GaugeUnit implement prometheus.Gauge and Metric
type GaugeUnit struct {
	val    float64
	vec    *Vector
	labels []string
	last   time.Time

	mtx sync.RWMutex
}

// NewGaugeUnit will create new counter with specified label values.
func NewGaugeUnit(vec *Vector, labelValues []string) Metric {
	return &GaugeUnit{
		vec:    vec,
		labels: labelValues,
		last:   time.Now(),
	}
}

// Desc implement prometheus.Gauge (prometheus.Metric)
func (u *GaugeUnit) Desc() *prometheus.Desc {
	return u.vec.desc
}

// Write implement prometheus.Gauge (prometheus.Metric)
func (u *GaugeUnit) Write(metric *dto.Metric) error {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	metric.Label = labelsToProto(u.vec.labels.ValuesToPromLabels(u.labels))
	metric.Gauge = &dto.Gauge{Value: proto.Float64(u.val)}

	return nil
}

// Describe implement prometheus.Gauge (prometheus.Collector)
func (u *GaugeUnit) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.vec.desc
}

// Collect implement prometheus.Gauge (prometheus.Collector)
func (u *GaugeUnit) Collect(ch chan<- prometheus.Metric) {
	ch <- u
}

// Set implement prometheus.Gauge
func (u *GaugeUnit) Set(v float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val = v
	u.last = time.Now()
}

// Inc implement prometheus.Gauge
func (u *GaugeUnit) Inc() {
	u.Add(1)
}

// Dec implement prometheus.Gauge
func (u *GaugeUnit) Dec() {
	u.Add(-1)
}

// Add implement prometheus.Gauge
func (u *GaugeUnit) Add(v float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val += v
	u.last = time.Now()
}

// Sub implement prometheus.Gauge
func (u *GaugeUnit) Sub(v float64) {
	u.Add(-v)
}

// SetToCurrentTime implement prometheus.Gauge
func (u *GaugeUnit) SetToCurrentTime() {
	// https://github.com/prometheus/client_golang/blob/master/prometheus/value.go#L85
	u.Set(float64(time.Now().UnixNano()) / 1e9)
}

// LastEdit implement Metric
func (u *GaugeUnit) LastEdit() time.Time {
	return u.last
}
