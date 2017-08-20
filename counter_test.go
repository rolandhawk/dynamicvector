// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/rolandhawk/dynamicvector"
	"github.com/stretchr/testify/assert"
)

func createCounter() *dynamicvector.Counter {
	return dynamicvector.NewCounter(dynamicvector.CounterOpts{
		Name:        "counter_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
	})
}

func TestCounterUnit_Desc(t *testing.T) {
	cv := createCounter()

	ch := make(chan *prometheus.Desc, 1)
	counter := cv.With(prometheus.Labels{"label3": "value3"})
	counter.Describe(ch)
	close(ch)
	assert.Equal(t, counter.Desc(), <-ch)
}

func TestCounterUnit_Collect(t *testing.T) {
	cv := createCounter()

	ch := make(chan prometheus.Metric, 1)
	counter := cv.With(prometheus.Labels{"label3": "value3"})
	counter.Collect(ch)
	close(ch)

	assert.NotNil(t, <-ch)
}

func TestCounterUnit_Inc(t *testing.T) {
	cv := createCounter()

	c := cv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	c.Write(m)
	assert.Equal(t, float64(0), *(m.Counter.Value))

	m = &dto.Metric{}
	c.Inc()
	c.Write(m)
	assert.Equal(t, float64(1), *(m.Counter.Value))
}

func TestCounterUnit_LastEdit(t *testing.T) {
	cv := createCounter()

	c := cv.With(prometheus.Labels{"label3": "value3"})
	last := c.(dynamicvector.Metric).LastEdit()

	c.Inc()
	assert.True(t, last.Before(c.(dynamicvector.Metric).LastEdit()))
}
