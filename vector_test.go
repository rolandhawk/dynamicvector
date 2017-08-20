// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/rolandhawk/dynamicvector"
	"github.com/stretchr/testify/assert"
)

func createVector(d time.Duration) *dynamicvector.Vector {
	return dynamicvector.NewCounter(dynamicvector.CounterOpts{
		Name:        "counter_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
		Expire:      d,
	}).Vector
}

func labelEqual(t *testing.T, m *dto.Metric, l map[string]string) {
	lm := make(map[string]string)
	for _, lp := range m.GetLabel() {
		lm[lp.GetName()] = lp.GetValue()
	}

	assert.Equal(t, l, lm)
}

func metricCount(v *dynamicvector.Vector) int {
	ch := make(chan prometheus.Metric, 10)
	v.Collect(ch)
	close(ch)

	index := 0
	for range ch {
		index++
	}

	return index
}

func TestVector_With(t *testing.T) {
	v := createVector(0)

	m1 := v.With(prometheus.Labels{"label3": "value3"})
	m2 := v.With(prometheus.Labels{"label4": "value4"})

	dto1 := &dto.Metric{}
	dto2 := &dto.Metric{}

	m1.Write(dto1)
	m2.Write(dto2)

	assert.Equal(t, 4, len(dto1.GetLabel()))
	assert.Equal(t, 4, len(dto2.GetLabel()))
	labelEqual(t, dto1, map[string]string{"label1": "value1", "label2": "value2", "label3": "value3", "label4": ""})
	labelEqual(t, dto2, map[string]string{"label1": "value1", "label2": "value2", "label3": "", "label4": "value4"})
}

func TestVector_Collect_Normal(t *testing.T) {
	v := createVector(0)

	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	assert.Equal(t, 3, metricCount(v))
}

func TestVector_Collect_Expire(t *testing.T) {
	v := createVector(50 * time.Millisecond)

	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, metricCount(v))
}

func TestVector_Describe(t *testing.T) {
	v := createVector(0)

	ch := make(chan *prometheus.Desc, 10)
	v.Describe(ch)
	d1 := <-ch

	v.With(prometheus.Labels{"label3": "value3"})
	v.Describe(ch)
	d2 := <-ch

	assert.NotEqual(t, d1, d2)

	v.With(prometheus.Labels{"label3": "value4"})
	v.Describe(ch)
	d3 := <-ch

	assert.Equal(t, d3, d2)

	v.With(prometheus.Labels{"label4": "value4"})
	v.Describe(ch)
	d4 := <-ch

	assert.NotEqual(t, d3, d4)
}

func TestVector_Delete(t *testing.T) {
	v := createVector(0)

	v.With(prometheus.Labels{})
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	assert.True(t, v.Delete(prometheus.Labels{"label3": "value4"}))
	assert.False(t, v.Delete(prometheus.Labels{"label5": "value4"}))

	assert.Equal(t, 3, metricCount(v))
}

func TestVector_Reset(t *testing.T) {
	v := createVector(0)

	v.With(prometheus.Labels{})
	v.With(prometheus.Labels{"label3": "value3"})

	v.Reset()

	assert.Equal(t, 0, metricCount(v))
}

func TestVector_GC(t *testing.T) {
	v := createVector(50 * time.Millisecond)

	v.With(prometheus.Labels{})
	v.With(prometheus.Labels{"label3": "value3"})

	assert.Equal(t, 0, v.GC())
	assert.Equal(t, 2, metricCount(v))

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 2, v.GC())
	assert.Equal(t, 0, metricCount(v))
}
