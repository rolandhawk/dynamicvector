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

func gaugeVector() *dynamicvector.Gauge {
	return dynamicvector.NewGauge(dynamicvector.GaugeOpts{
		Name:        "gauge_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
	})
}

func TestGaugeUnit_Desc(t *testing.T) {
	gv := gaugeVector()

	ch := make(chan *prometheus.Desc, 1)
	gauge := gv.With(prometheus.Labels{"label3": "value3"})
	gauge.Describe(ch)
	close(ch)

	assert.Equal(t, gauge.Desc(), <-ch)
}

func TestGaugeUnit_Collect(t *testing.T) {
	gv := gaugeVector()

	ch := make(chan prometheus.Metric, 1)
	gauge := gv.With(prometheus.Labels{"label3": "value3"})
	gauge.Collect(ch)
	close(ch)

	assert.NotNil(t, <-ch)
}

func TestGaugeUnit_Inc(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	g.Write(m)
	assert.Equal(t, float64(0), *(m.Gauge.Value))

	m = &dto.Metric{}
	g.Inc()
	g.Write(m)
	assert.Equal(t, float64(1), *(m.Gauge.Value))
}

func TestGaugeUnit_Dec(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	g.Write(m)
	assert.Equal(t, float64(0), *(m.Gauge.Value))

	m = &dto.Metric{}
	g.Dec()
	g.Write(m)
	assert.Equal(t, float64(-1), *(m.Gauge.Value))
}

func TestGaugeUnit_Sub(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	g.Write(m)
	assert.Equal(t, float64(0), *(m.Gauge.Value))

	m = &dto.Metric{}
	g.Sub(1)
	g.Write(m)
	assert.Equal(t, float64(-1), *(m.Gauge.Value))
}

func TestGaugeUnit_Set(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	g.Write(m)
	assert.Equal(t, float64(0), *(m.Gauge.Value))

	m = &dto.Metric{}
	g.Set(100)
	g.Write(m)
	assert.Equal(t, float64(100), *(m.Gauge.Value))
}

func TestGaugeUnit_SetToCurrentTime(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	g.Write(m)
	assert.Equal(t, float64(0), *(m.Gauge.Value))

	m = &dto.Metric{}
	g.SetToCurrentTime()
	g.Write(m)
	assert.NotEqual(t, float64(0), *(m.Gauge.Value))
}

func TestGaugeUnit_LastEdit(t *testing.T) {
	gv := gaugeVector()

	g := gv.With(prometheus.Labels{"label3": "value3"})
	last := g.(dynamicvector.Metric).LastEdit()

	g.Inc()
	assert.True(t, last.Before(g.(dynamicvector.Metric).LastEdit()))
}
