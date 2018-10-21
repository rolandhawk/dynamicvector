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

// Histogram is a histogram dynamicvector
type Histogram struct {
	Vector
}

// NewHistogram will return a new dynamicvector histogram.
func NewHistogram(opts HistogramOpts) *Histogram {
	vec := &vector{
		Name:      prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:      opts.Help,
		Labels:    NewLabels(opts.ConstLabels),
		Expire:    opts.Expire,
		MaxLength: opts.MaxLength,
	}

	vec.constructor = func(labelValues []string) Metric {
		b := make(map[float64]uint64)
		for _, v := range opts.Buckets {
			b[v] = 0
		}

		return &histogramUnit{
			vec:     vec,
			labels:  labelValues,
			last:    time.Now(),
			buckets: b,
		}
	}
	vec.Reset()

	return &Histogram{vec}
}

// With is a syntatic sugar for Vector.With(labels).(prometheus.Histogram)
func (h *Histogram) With(labels prometheus.Labels) prometheus.Histogram {
	return h.Vector.With(labels).(prometheus.Histogram)
}

type histogramUnit struct {
	sum     float64
	count   uint64
	buckets map[float64]uint64
	vec     *vector
	labels  []string
	last    time.Time

	mtx sync.RWMutex
}

func (u *histogramUnit) Desc() *prometheus.Desc {
	return u.vec.desc
}

func (u *histogramUnit) Write(metric *dto.Metric) error {
	u.mtx.RLock()
	defer u.mtx.RUnlock()

	var buckets []*dto.Bucket
	for bound, count := range u.buckets {
		buckets = append(buckets, &dto.Bucket{CumulativeCount: proto.Uint64(count), UpperBound: proto.Float64(bound)})
	}

	metric.Label = LabelsProto(u.vec.Labels.Generate(u.labels))
	metric.Histogram = &dto.Histogram{SampleCount: proto.Uint64(u.count), SampleSum: proto.Float64(u.sum), Bucket: buckets}

	return nil
}

func (u *histogramUnit) Describe(ch chan<- *prometheus.Desc) {
	ch <- u.vec.desc
}

func (u *histogramUnit) Collect(ch chan<- prometheus.Metric) {
	ch <- u
}

func (u *histogramUnit) Observe(v float64) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	for b := range u.buckets {
		if b > v {
			u.buckets[b]++
		}
	}
	u.count++
	u.sum += v
	u.last = time.Now()
}

func (u *histogramUnit) LastEdit() time.Time {
	return u.last
}
