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
	Vector
}

// NewCounter will return a new dynamicvector counter.
func NewCounter(opts CounterOpts) *Counter {
	vec := &vector{
		Name:      prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:      opts.Help,
		Labels:    NewLabels(opts.ConstLabels),
		Expire:    opts.Expire,
		MaxLength: opts.MaxLength,
	}

	vec.constructor = func(labelValues []string) Metric {
		return &counterUnit{
			vec:    vec,
			labels: labelValues,
			last:   time.Now(),
		}
	}
	vec.Reset()

	return &Counter{vec}
}

// With is a syntatic sugar for Vector.With(labels).(prometheus.Counter)
func (c *Counter) With(labels prometheus.Labels) prometheus.Counter {
	return c.Vector.With(labels).(prometheus.Counter)
}

type counterUnit struct {
	val    float64
	vec    *vector
	labels []string
	last   time.Time

	mtx sync.RWMutex
}

func (u *counterUnit) Desc() *prometheus.Desc {
	return u.vec.desc
}

func (u *counterUnit) Write(metric *dto.Metric) error {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	metric.Label = LabelsProto(u.vec.Labels.Generate(u.labels))
	metric.Counter = &dto.Counter{Value: proto.Float64(u.val)}

	return nil
}

func (u *counterUnit) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.vec.desc
}

func (u *counterUnit) Collect(ch chan<- prometheus.Metric) {
	ch <- u
}

func (u *counterUnit) Inc() {
	u.Add(1)
}

func (u *counterUnit) Add(val float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.val += val
	u.last = time.Now()
}

func (u *counterUnit) LastEdit() time.Time {
	return u.last
}
