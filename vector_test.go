// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rolandhawk/dynamicvector"
	"github.com/stretchr/testify/assert"
)

func TestVector_GetMetricWith_Normal(t *testing.T) {
	v := createVector(0, 0)

	m1, err := v.GetMetricWith(prometheus.Labels{"label3": "value3"})
	assert.NoError(t, err)
	m2, err := v.GetMetricWith(prometheus.Labels{"label4": "value4"})
	assert.NoError(t, err)
	m3, err := v.GetMetricWith(prometheus.Labels{"label3": "value3"})
	assert.NoError(t, err)

	assert.Equal(t, v, m1.(*metric).v)
	assert.Equal(t, v, m2.(*metric).v)
	assert.Equal(t, m1.(*metric).lbl, []string{"value3"})
	assert.Equal(t, m2.(*metric).lbl, []string{"", "value4"})
	assert.Equal(t, m1, m3)
}

func TestVector_GetMetricWith_Limit(t *testing.T) {
	v := createVector(0, 1)

	_, err := v.GetMetricWith(prometheus.Labels{"label3": "value3"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label3": "value4"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label4": "value4"})
	assert.Error(t, err)
}

func TestVector_With_NoPanic(t *testing.T) {
	v := createVector(0, 0)

	m1, _ := v.GetMetricWith(prometheus.Labels{"label3": "value3"})
	m2 := v.With(prometheus.Labels{"label3": "value3"})

	assert.Equal(t, m1, m2)
}

func TestVector_With_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should be panic")
		}
	}()

	v := createVector(0, 1)
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})
}

func TestVector_Length(t *testing.T) {
	v := createVector(0, 0)
	assert.Equal(t, 0, v.Length())

	v.With(prometheus.Labels{"label3": "value3"})
	assert.Equal(t, 1, v.Length())

	v.With(prometheus.Labels{"label4": "value4"})
	assert.Equal(t, 2, v.Length())

	v.With(prometheus.Labels{"label3": "value3"})
	assert.Equal(t, 2, v.Length())
}

func TestVector_Reset(t *testing.T) {
	v := createVector(0, 0)

	m1 := v.With(prometheus.Labels{"label3": "value3"})
	v.Reset()

	m2 := v.With(prometheus.Labels{"label3": "value3"})
	assert.Equal(t, 1, v.Length())
	if m1 == m2 {
		t.Error("metric should not equal after reset")
	}
}

func TestVector_Delete(t *testing.T) {
	v := createVector(0, 0)

	v.With(prometheus.Labels{})
	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	assert.True(t, v.Delete(prometheus.Labels{"label3": "value4"}))
	assert.False(t, v.Delete(prometheus.Labels{"label5": "value4"}))

	assert.Equal(t, 3, v.Length())
}

func TestVector_Collect_Normal(t *testing.T) {
	v := createVector(0, 0)

	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	assert.Equal(t, 3, len(collect(v)))
}

func TestVector_Collect_Expire(t *testing.T) {
	v := createVector(50*time.Millisecond, 0)

	v.With(prometheus.Labels{"label3": "value3"})
	v.With(prometheus.Labels{"label3": "value4"})
	v.With(prometheus.Labels{"label4": "value4"})

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, len(collect(v)))
}

func TestVector_Collecto_ExceedMaxLen(t *testing.T) {
	v := createVector(0, 1)

	v.GetMetricWith(prometheus.Labels{"label3": "value3"})
	v.GetMetricWith(prometheus.Labels{"label3": "value4"})
	v.GetMetricWith(prometheus.Labels{"label4": "value4"})

	assert.Equal(t, 0, len(collect(v)))
}

func TestVector_Describe(t *testing.T) {
	v := createVector(0, 0)

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

func TestVector_GC_Expire(t *testing.T) {
	v := createVector(50*time.Millisecond, 0)

	v.With(prometheus.Labels{})
	v.With(prometheus.Labels{"label3": "value3"})

	assert.Equal(t, 0, v.GC().Deleted)
	assert.Equal(t, 2, v.Length())

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 2, v.GC().Deleted)
	assert.Equal(t, 0, v.Length())
}

func TestVector_GC_LimitExceeded(t *testing.T) {
	v := createVector(0, 1)

	v.With(prometheus.Labels{})
	stat := v.GC()
	assert.Equal(t, 0, stat.Deleted)
	assert.False(t, stat.LimitExceeded)
	assert.Equal(t, 1, v.Length())

	v.With(prometheus.Labels{"label3": "value3"})
	stat = v.GC()
	assert.Equal(t, 2, stat.Deleted)
	assert.True(t, stat.LimitExceeded)
	assert.Equal(t, 2, v.Length())
}

type metric struct {
	dynamicvector.Metric

	v   *dynamicvector.Vector
	lbl []string
	le  time.Time
}

func newMetric(v *dynamicvector.Vector, lbl []string) dynamicvector.Metric {
	return &metric{v: v, lbl: lbl, le: time.Now()}
}

func (m *metric) LastEdit() time.Time {
	return m.le
}

func createVector(d time.Duration, ml int) *dynamicvector.Vector {
	return dynamicvector.NewVector(dynamicvector.Opts{
		Name:        "vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
		Expire:      d,
		MaxLength:   ml,
	}, newMetric)
}

func collect(v *dynamicvector.Vector) []prometheus.Metric {
	ch := make(chan prometheus.Metric, 10)
	v.Collect(ch)
	close(ch)

	var res []prometheus.Metric
	for m := range ch {
		res = append(res, m)
	}

	return res
}
