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

func TestGauge_GetMetricWith_NoError(t *testing.T) {
	v := gaugeVector(0)

	_, err := v.GetMetricWith(prometheus.Labels{"label1": "value1"})
	assert.NoError(t, err)
}

func TestGauge_GetMetricWith_Error(t *testing.T) {
	v := gaugeVector(1)

	_, err := v.GetMetricWith(prometheus.Labels{"label1": "value1"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label1": "value2"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label2": "value1"})
	assert.Error(t, err)
}

func TestGauge_With(t *testing.T) {
	v := gaugeVector(0)

	// no assertion, we only test if it panic or not.
	v.With(prometheus.Labels{"label1": "value1"})
}

func TestGaugeUnit_Desc(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan *prometheus.Desc, 1)
	v.Describe(ch)
	close(ch)

	assert.Equal(t, gauge.Desc(), <-ch)
}

func TestGaugeUnit_Write(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})

	var m dto.Metric
	err := gauge.Write(&m)
	assert.NoError(t, err)
	assert.NotNil(t, m.Gauge)
}

func TestGaugeUnit_Describe(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan *prometheus.Desc, 1)
	gauge.Describe(ch)
	close(ch)

	assert.Equal(t, gauge.Desc(), <-ch)
}

func TestGaugeUnit_Collect(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan prometheus.Metric, 1)
	gauge.Collect(ch)
	close(ch)

	assert.Equal(t, gauge, <-ch)
}

func TestGaugeUnit_Set(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{})
	gauge.Set(2.4)

	var m dto.Metric
	gauge.Write(&m)
	assert.Equal(t, float64(2.4), *(m.Gauge.Value))
}

func TestGaugeUnit_Inc(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{})
	gauge.Inc()

	var m dto.Metric
	gauge.Write(&m)
	assert.Equal(t, float64(1), *(m.Gauge.Value))
}

func TestGaugeUnit_Dec(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{})
	gauge.Dec()

	var m dto.Metric
	gauge.Write(&m)
	assert.Equal(t, float64(-1), *(m.Gauge.Value))
}

func TestGaugeUnit_Add(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})
	gauge.Add(11.1)

	var m dto.Metric
	gauge.Write(&m)
	assert.Equal(t, float64(11.1), *(m.Gauge.Value))
}

func TestGaugeUnit_Sub(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})
	gauge.Sub(11.1)

	var m dto.Metric
	gauge.Write(&m)
	assert.Equal(t, float64(-11.1), *(m.Gauge.Value))
}

func TestGaugeUnit_SetToCurrentTime(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})
	gauge.SetToCurrentTime()

	var m dto.Metric
	gauge.Write(&m)
	assert.NotEqual(t, float64(0), *(m.Gauge.Value))
}

func TestGaugeUnit_LastEdit(t *testing.T) {
	v := gaugeVector(0)
	gauge := v.With(prometheus.Labels{"label1": "value1"})
	last := gauge.(dynamicvector.Metric).LastEdit()

	gauge.Inc()
	assert.True(t, last.Before(gauge.(dynamicvector.Metric).LastEdit()))
}

func gaugeVector(ml int) *dynamicvector.Gauge {
	return dynamicvector.NewGauge(dynamicvector.GaugeOpts{
		Name:        "gauge_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
		MaxLength:   ml,
	})
}
