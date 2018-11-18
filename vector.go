// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

// Package dynamicvector give flexibility to add any number of labels into
// prometheus Vector. It also give a feature to expire old metrics that
// never been updated.
package dynamicvector

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric is an interface that encapsulate prometheus.Metric interface
type Metric interface {
	prometheus.Metric

	// LastEdit return last time metric is edited
	LastEdit() time.Time
}

// Vector is a dynamicvector that used to keep metrics.
type Vector struct {
	opts        Opts                                           // vector options
	constructor func(vec *Vector, labelValues []string) Metric // constructor to make new metric

	mtx          sync.RWMutex
	labels       *Labels           // Labels contain information about metric labels.
	pseudoLength int               // it used when resetting vector that already exceed max length.
	metrics      map[uint64]Metric // vector metric
	desc         *prometheus.Desc
}

// NewVector will create new vector with specified option and metric constructor.
func NewVector(opts Opts, cons func(v *Vector, labelValues []string) Metric) *Vector {
	vec := &Vector{
		opts:        opts,
		constructor: cons,
	}
	vec.reset()

	return vec
}

// Collect implement prometheus.Collector.
func (v *Vector) Collect(ch chan<- prometheus.Metric) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	if v.exceedMaxLength() {
		return
	}

	for _, m := range v.metrics {
		if !v.isExpire(m.LastEdit()) {
			ch <- m
		}
	}
}

// Describe implement prometheus.Collector.
func (v *Vector) Describe(ch chan<- *prometheus.Desc) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	ch <- v.desc
}

// Delete will delete metric that have exact match labels from vector.
func (v *Vector) Delete(l prometheus.Labels) bool {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	if !v.labels.Include(l) {
		return false
	}

	h := v.labels.Hash(l)
	_, found := v.metrics[h]
	delete(v.metrics, h)

	return found
}

// GetMetricWith returns the Metric for the given Labels map (the label names must match those of
// the VariableLabels in Desc). If that label map is accessed for the first time, a new Metric is created.
// Return error if maxLen is exceeded.
func (v *Vector) GetMetricWith(labels prometheus.Labels) (prometheus.Metric, error) {
	v.mtx.RLock()
	metric := v.get(labels)
	v.mtx.RUnlock()

	if metric != nil {
		return metric, nil
	}

	v.mtx.Lock()
	defer v.mtx.Unlock()

	metric = v.get(labels)
	if metric != nil {
		return metric, nil
	}
	if v.exceedMaxLength() {
		return nil, fmt.Errorf("vector with %s exceed length limit", v.desc.String())
	}

	return v.create(labels), nil
}

// With behave like GetMetricWith except it will panic instead when there is an error.
func (v *Vector) With(l prometheus.Labels) prometheus.Metric {
	m, err := v.GetMetricWith(l)
	if err != nil {
		panic(err)
	}
	return m
}

// Length will return number of metrics in this vector.
func (v *Vector) Length() int {
	if v.pseudoLength > 0 {
		return v.pseudoLength
	} else {
		return len(v.metrics)
	}
}

// Reset will delete all metrics in vector.
func (v *Vector) Reset() {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	v.reset()
}

// GC will do housekeeping work related to this metrics and return
// number of metrics that is deleted. Currently there are two things that this method do.
// First, delete all expired metrics. Second, delete all metrics for vector that exceed MaxLength.
func (v *Vector) GC() GCStat {
	var stat GCStat
	v.mtx.Lock()
	defer v.mtx.Unlock()

	// delete expired metrics
	for h, m := range v.metrics {
		if v.isExpire(m.LastEdit()) {
			delete(v.metrics, h)
			stat.Expire++
		}
	}

	// delete all metrics for vector that exceed MaxLength
	if v.exceedMaxLength() {
		v.pseudoLength = v.Length()
		v.reset()
		stat.Expire = stat.Expire + v.pseudoLength
		stat.LimitExceeded = true
	}

	return stat
}

func (v *Vector) get(l prometheus.Labels) prometheus.Metric {
	if !v.labels.Include(l) {
		return nil
	}

	return v.metrics[v.labels.Hash(l)]
}

func (v *Vector) create(l prometheus.Labels) prometheus.Metric {
	oldLen := len(v.labels.Keys)
	labelValues := v.labels.PromLabelsToValues(l)

	if oldLen != len(v.labels.Keys) {
		v.desc = v.newDesc()
	}

	metric := v.constructor(v, labelValues)
	v.metrics[v.labels.Hash(l)] = metric

	return metric
}

func (v *Vector) reset() {
	v.metrics = make(map[uint64]Metric)
	v.labels = NewLabels(v.opts.ConstLabels)
	v.desc = v.newDesc()
}

func (v *Vector) exceedMaxLength() bool {
	return v.opts.MaxLength > 0 && v.Length() > v.opts.MaxLength
}

func (v *Vector) isExpire(lastEdit time.Time) bool {
	return v.opts.Expire != 0 && time.Since(lastEdit) > v.opts.Expire
}

func (v *Vector) newDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(v.opts.Namespace, v.opts.Subsystem, v.opts.Name),
		v.opts.Help,
		v.labels.Keys,
		v.opts.ConstLabels,
	)
}

// GCStat is status for garbage collector.
type GCStat struct {
	// Number of expired metrics
	Expire int

	// Whether metric exceed limit or not.
	LimitExceeded bool
}
