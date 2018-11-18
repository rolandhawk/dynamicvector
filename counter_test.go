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

func TestCounter_GetMetricWith(t *testing.T) {
	cv := createCounter()

	_, err := cv.GetMetricWith(prometheus.Labels{"label3": "value3"})
	assert.NoError(t, err)
}

func TestCounter_With(t *testing.T) {
	cv := createCounter()

	// no assertion, we only test if it panic or not.
	cv.With(prometheus.Labels{"label3": "value3"})
}

func TestCounterUnit_Desc(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})

	ch := make(chan *prometheus.Desc, 1)
	cv.Describe(ch)
	close(ch)

	assert.Equal(t, counter.Desc(), <-ch)
}

func TestCounterUnit_Write(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})

	var m dto.Metric
	err := counter.Write(&m)
	assert.NoError(t, err)
	assert.NotNil(t, m.Counter)
}

func TestCounterUnit_Describe(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})

	ch := make(chan *prometheus.Desc, 1)
	counter.Describe(ch)
	close(ch)

	assert.Equal(t, counter.Desc(), <-ch)
}

func TestCounterUnit_Collect(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})

	ch := make(chan prometheus.Metric, 1)
	counter.Collect(ch)
	close(ch)

	assert.Equal(t, counter, <-ch)
}

func TestCounterUnit_Inc(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})
	counter.Inc()

	var m dto.Metric
	counter.Write(&m)
	assert.Equal(t, float64(1), *(m.Counter.Value))
}

func TestCounterUnit_Add(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})
	counter.Add(11.1)

	var m dto.Metric
	counter.Write(&m)
	assert.Equal(t, float64(11.1), *(m.Counter.Value))
}

func TestCounterUnit_LastEdit(t *testing.T) {
	cv := createCounter()
	counter := cv.With(prometheus.Labels{"label3": "value3"})
	last := counter.(dynamicvector.Metric).LastEdit()

	counter.Inc()
	assert.True(t, last.Before(counter.(dynamicvector.Metric).LastEdit()))
}

func createCounter() *dynamicvector.Counter {
	return dynamicvector.NewCounter(dynamicvector.CounterOpts{
		Name:        "counter_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
	})
}
