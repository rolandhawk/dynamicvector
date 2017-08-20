// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

// Package dynamicvector give flexibility to add any number of labels into
// prometheus Vector. It also give a feature to expire old metrics that
// never been updated.
package dynamicvector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metric is an interface that encapsulate prometheus.Metric interface
type Metric interface {
	prometheus.Metric

	LastEdit() time.Time
}

// Vector is a dynamicvector that used to keep metrics.
type Vector struct {
	// Name is a fully qualified metric name for vector.
	Name string

	// Help is a help for metric
	Help string

	// Labels contain information about metric labels.
	Labels *Labels

	// Expire is a duration to keep metrics.
	Expire time.Duration

	metrics     map[uint64]Metric
	constructor func(labelValues []string) Metric
	desc        *prometheus.Desc

	mtx sync.RWMutex
}

// Collect will return metrics to channel.
func (v *Vector) Collect(ch chan<- prometheus.Metric) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	for _, m := range v.metrics {
		if v.Expire == 0 || time.Since(m.LastEdit()) <= v.Expire {
			ch <- m
		}
	}
}

// Describe will return prometheus.Desc for this vector.
func (v *Vector) Describe(ch chan<- *prometheus.Desc) {
	v.mtx.RLock()
	defer v.mtx.RUnlock()

	ch <- v.desc
}

// Delete will delete metric that have exact match labels from vector.
func (v *Vector) Delete(l prometheus.Labels) bool {
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

// With will return or create metric with specified labels.
func (v *Vector) With(l prometheus.Labels) prometheus.Metric {
	v.mtx.RLock()
	metric := v.get(l)
	v.mtx.RUnlock()

	if metric != nil {
		return metric
	}

	v.mtx.Lock()
	defer v.mtx.Unlock()
	metric = v.get(l)
	if metric == nil {
		metric = v.create(l)
	}

	return metric
}

// Reset will delete all metrics in vector.
func (v *Vector) Reset() {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	c := v.Labels.Constant
	v.metrics = make(map[uint64]Metric)
	v.Labels = NewLabels(c)
	v.desc = prometheus.NewDesc(v.Name, v.Help, nil, c)
}

// GC will delete all expired metrics and return number of gced metrics.
func (v *Vector) GC() int {
	v.mtx.Lock()
	defer v.mtx.Unlock()

	count := 0
	for h, m := range v.metrics {
		if v.Expire != 0 && time.Since(m.LastEdit()) > v.Expire {
			delete(v.metrics, h)
			count++
		}
	}

	return count
}

func (v *Vector) get(l prometheus.Labels) prometheus.Metric {
	if !v.Labels.Include(l) {
		return nil
	}

	return v.metrics[v.Labels.Hash(l)]
}

func (v *Vector) create(l prometheus.Labels) prometheus.Metric {
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
