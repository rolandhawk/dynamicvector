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
	Vector
}

// NewGauge will return a new dynamicvector gauge.
func NewGauge(opts GaugeOpts) *Gauge {
	vec := &vector{
		Name:      prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:      opts.Help,
		Labels:    NewLabels(opts.ConstLabels),
		Expire:    opts.Expire,
		MaxLength: opts.MaxLength,
	}

	vec.constructor = func(labelValues []string) Metric {
		return &gaugeUnit{
			vec:    vec,
			labels: labelValues,
			last:   time.Now(),
		}
	}
	vec.Reset()

	return &Gauge{vec}
}

// With is a syntatic sugar for Vector.With(labels).(prometheus.Gauge)
func (g *Gauge) With(labels prometheus.Labels) prometheus.Gauge {
	return g.Vector.With(labels).(prometheus.Gauge)
}

type gaugeUnit struct {
	val    float64
	vec    *vector
	labels []string
	last   time.Time

	mtx sync.RWMutex
}

func (u *gaugeUnit) Desc() *prometheus.Desc {
	return u.vec.desc
}

func (u *gaugeUnit) Write(metric *dto.Metric) error {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	metric.Label = LabelsProto(u.vec.Labels.Generate(u.labels))
	metric.Gauge = &dto.Gauge{Value: proto.Float64(u.val)}

	return nil
}

func (u *gaugeUnit) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.vec.desc
}

func (u *gaugeUnit) Collect(ch chan<- prometheus.Metric) {
	ch <- u
}

func (u *gaugeUnit) Set(v float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val = v
	u.last = time.Now()
}

func (u *gaugeUnit) Inc() {
	u.Add(1)
}

func (u *gaugeUnit) Dec() {
	u.Add(-1)
}

func (u *gaugeUnit) Add(v float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val += v
	u.last = time.Now()
}

func (u *gaugeUnit) Sub(v float64) {
	u.Add(-v)
}

func (u *gaugeUnit) SetToCurrentTime() {
	// https://github.com/prometheus/client_golang/blob/master/prometheus/value.go#L85
	u.Set(float64(time.Now().UnixNano()) / 1e9)
}

func (u *gaugeUnit) LastEdit() time.Time {
	return u.last
}
