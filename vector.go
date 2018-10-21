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

	// LastEdit indicate last time Metric is edited.
	LastEdit() time.Time
}

// Vector is interface for raw Vector. It include the same behavior as prometheus.*Vec.
type Vector interface {
	prometheus.Collector

	// Delete deletes the metric where the variable labels are the same as those passed in as labels.
	// It returns true if a metric was deleted.
	Delete(prometheus.Labels) bool

	// DeleteLabelValues removes the metric where the variable labels are the same as those passed in
	// as labels (same order as the VariableLabels in Desc). It returns true if a metric was deleted.
	DeleteLabelValues(lvs ...string) bool

	// GetMetricWith returns the Metric for the given Labels map (the label names must match those of
	// the VariableLabels in Desc). If that label map is accessed for the first time, a new Metric is created.
	GetMetricWith(prometheus.Labels) (prometheus.Metric, error)

	// GetMetricWithLabelValues returns the Metric for the given slice of label values
	// (same order as the VariableLabels in Desc). If that combination of label values is accessed for the first time,
	// a new Metric is created.
	GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error)

	// With works as GetMetricWith, but panics where GetMetricWithLabels would have returned an error.
	With(prometheus.Labels) prometheus.Metric

	// WithLabelValues works as GetMetricWithLabelValues, but panics where GetMetricWithLabelValues would have returned an error.
	WithLabelValues(lvs ...string) prometheus.Metric

	// CurryWith returns a vector curried with the provided labels.
	CurryWith(prometheus.Labels) (Vector, error)

	// MustCurryWith works as CurryWith but panics where CurryWith would have returned an error.
	MustCurryWith(prometheus.Labels) Vector

	// Length will return number of metrics in this vector.
	Length() int

	// Reset will delete all metrics in vector.
	Reset()

	// GC will do housekeeping work related to this metrics and return
	// number of metrics that is deleted. Currently there are two things that this method do.
	// First, delete all expired metrics. Second, delete all metrics for vector that exceed MaxLength.
	GC() int
}

// Vector is a dynamicvector that used to keep metrics.
type vector struct {
	// Name is a fully qualified metric name for vector.
	Name string

	// Help is a help for metric
	Help string

	// Labels contain information about metric labels.
	Labels *Labels

	// Expire is a duration to keep metrics. Zero mean no expire.
	Expire time.Duration

	// MaxLength is maximum length that this vector is allowed to have. Zero mean no maximum length.
	MaxLength int

	pseudoLength int                               // it used when resetting vector that already exceed max length.
	metrics      map[uint64]Metric                 // vector metric
	constructor  func(labelValues []string) Metric // constructor to make new metric
	desc         *prometheus.Desc

	mtx sync.RWMutex
}

// Collect implement prometheus.Collector.
func (v *vector) Collect(ch chan<- prometheus.Metric) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	if v.exceedMaxLength() {
		return
	}

	for _, m := range v.metrics {
		if v.Expire == 0 || time.Since(m.LastEdit()) <= v.Expire {
			ch <- m
		}
	}
}

// Describe implement prometheus.Collector.
func (v *vector) Describe(ch chan<- *prometheus.Desc) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	ch <- v.desc
}

func (v *vector) Delete(l prometheus.Labels) bool {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	if !v.Labels.Include(l) {
		return false
	}

	h := v.Labels.Hash(l)
	_, found := v.metrics[h]
	delete(v.metrics, h)

	return found
}

func (v *vector) DeleteLabelValues(lvs ...string) bool {
	if len(lvs) > len(v.Labels.Names) {
		return false
	}

	l := v.Labels.GenerateWithoutConstant(lvs)
	return v.Delete(l)
}

func (v *vector) GetMetricWith(labels prometheus.Labels) (prometheus.Metric, error) {
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
		return nil, fmt.Errorf("vector %s exceed length limit", v.Name)
	}

	return v.create(labels), nil
}

func (v *vector) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	if len(lvs) > len(v.Labels.Names) {
		return nil, fmt.Errorf("label value length exceeds label key length")
	}

	l := v.Labels.GenerateWithoutConstant(lvs)
	return v.GetMetricWith(l)
}

// With will return or create metric with specified labels.
func (v *vector) With(l prometheus.Labels) prometheus.Metric {
	m, err := v.GetMetricWith(l)
	if err != nil {
		panic(err)
	}

	return m
}

func (v *vector) WithLabelValues(lvs ...string) prometheus.Metric {
	m, err := v.GetMetricWithLabelValues(lvs...)
	if err != nil {
		panic(err)
	}

	return m
}

func (v *vector) CurryWith(prometheus.Labels) (Vector, error) {
	return nil, nil
}

func (v *vector) MustCurryWith(prometheus.Labels) Vector {
	return nil
}

// Length will return number of metrics in this vector.
func (v *vector) Length() int {
	if v.pseudoLength > 0 {
		return v.pseudoLength
	} else {
		return len(v.metrics)
	}
}

// Reset will delete all metrics in vector.
func (v *vector) Reset() {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	v.reset()
}

// GC will do housekeeping work related to this metrics and return
// number of metrics that is deleted. Currently there are two things that this method do.
// First, delete all expired metrics. Second, delete all metrics for vector that exceed MaxLength.
func (v *vector) GC() int {
	count := 0
	v.mtx.Lock()
	defer v.mtx.Unlock()

	// delete expired metrics
	for h, m := range v.metrics {
		if v.Expire != 0 && time.Since(m.LastEdit()) > v.Expire {
			delete(v.metrics, h)
			count++
		}
	}

	// delete all metrics for vector that exceed MaxLength
	if v.exceedMaxLength() {
		v.pseudoLength = v.Length()
		v.reset()
		count = count + v.pseudoLength
	}

	return count
}

func (v *vector) get(l prometheus.Labels) prometheus.Metric {
	if !v.Labels.Include(l) {
		return nil
	}

	return v.metrics[v.Labels.Hash(l)]
}

func (v *vector) create(l prometheus.Labels) prometheus.Metric {
	var createDesc bool
	labelValues := make([]string, len(v.Labels.Names))

	for name, value := range l {
		if i, ok := v.Labels.Index[name]; ok {
			labelValues[i] = value
		} else {
			v.Labels.Add(name)
			labelValues = append(labelValues, value)
			createDesc = true
		}
	}

	if createDesc {
		v.desc = prometheus.NewDesc(v.Name, v.Help, v.Labels.Names, v.Labels.Constant)
	}

	metric := v.constructor(labelValues)
	v.metrics[v.Labels.Hash(l)] = metric

	return metric
}

func (v *vector) reset() {
	c := v.Labels.Constant
	v.metrics = make(map[uint64]Metric)
	v.Labels = NewLabels(c)
	v.desc = prometheus.NewDesc(v.Name, v.Help, nil, c)
}

func (v *vector) exceedMaxLength() bool {
	return v.MaxLength > 0 && v.Length() > v.MaxLength
}
