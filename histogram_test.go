// Copyright (h) 2017 Roland Rifandi Utama
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

func createHistogram() *dynamicvector.Histogram {
	return dynamicvector.NewHistogram(dynamicvector.HistogramOpts{
		Name:        "counter_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
		Buckets:     []float64{1, 10, 100},
	})
}

func TestHistogramUnit_Desc(t *testing.T) {
	hv := createHistogram()

	ch := make(chan *prometheus.Desc, 1)
	histogram := hv.With(prometheus.Labels{"label3": "value3"})
	histogram.Describe(ch)
	close(ch)
	assert.Equal(t, histogram.Desc(), <-ch)
}

func TestHistogramUnit_Collect(t *testing.T) {
	hv := createHistogram()

	ch := make(chan prometheus.Metric, 1)
	histogram := hv.With(prometheus.Labels{"label3": "value3"})
	histogram.Collect(ch)
	close(ch)

	assert.NotNil(t, <-ch)
}

func TestHistogramUnit_Observe(t *testing.T) {
	hv := createHistogram()

	h := hv.With(prometheus.Labels{"label3": "value3"})
	m := &dto.Metric{}
	h.Write(m)
	assert.Equal(t, uint64(0), *(m.Histogram.SampleCount))
	assert.Equal(t, float64(0), *(m.Histogram.SampleSum))
	assert.Equal(t, 3, len(m.Histogram.Bucket))
	for _, b := range m.Histogram.Bucket {
		switch *b.UpperBound {
		case float64(1), float64(10), float64(100):
			assert.Equal(t, uint64(0), *(b.CumulativeCount))
		default:
			t.Error("wrong upperbound")
		}
	}

	m = &dto.Metric{}
	h.Observe(5.5)
	h.Write(m)

	h.Write(m)
	assert.Equal(t, uint64(1), *(m.Histogram.SampleCount))
	assert.Equal(t, float64(5.5), *(m.Histogram.SampleSum))
	assert.Equal(t, 3, len(m.Histogram.Bucket))
	for _, b := range m.Histogram.Bucket {
		switch *b.UpperBound {
		case float64(1):
			assert.Equal(t, uint64(0), *(b.CumulativeCount))
		case float64(10), float64(100):
			assert.Equal(t, uint64(1), *(b.CumulativeCount))
		default:
			t.Error("wrong upperbound")
		}
	}
}

func TestHistogramUnit_LastEdit(t *testing.T) {
	hv := createHistogram()

	h := hv.With(prometheus.Labels{"label3": "value3"})
	last := h.(dynamicvector.Metric).LastEdit()

	h.Observe(1)
	assert.True(t, last.Before(h.(dynamicvector.Metric).LastEdit()))
}
