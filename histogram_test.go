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

func TestHistogram_GetMetricWith_NoError(t *testing.T) {
	v := createHistogram(0)

	_, err := v.GetMetricWith(prometheus.Labels{"label1": "value1"})
	assert.NoError(t, err)
}

func TestHistogram_GetMetricWith_Error(t *testing.T) {
	v := createHistogram(1)

	_, err := v.GetMetricWith(prometheus.Labels{"label1": "value1"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label1": "value2"})
	assert.NoError(t, err)
	_, err = v.GetMetricWith(prometheus.Labels{"label2": "value1"})
	assert.Error(t, err)
}

func TestHistogram_With(t *testing.T) {
	v := createHistogram(0)

	// no assertion, we only test if it panic or not.
	v.With(prometheus.Labels{"label1": "value1"})
}

func TestHistogramUnit_Desc(t *testing.T) {
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan *prometheus.Desc, 1)
	v.Describe(ch)
	close(ch)

	assert.Equal(t, histogram.Desc(), <-ch)
}

func TestHistogramUnit_Write(t *testing.T) {
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})

	var m dto.Metric
	err := histogram.Write(&m)
	assert.NoError(t, err)
	assert.NotNil(t, m.Histogram)
}

func TestHistogramUnit_Describe(t *testing.T) {
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan *prometheus.Desc, 1)
	histogram.Describe(ch)
	close(ch)

	assert.Equal(t, histogram.Desc(), <-ch)
}

func TestHistogramUnit_Collect(t *testing.T) {
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})

	ch := make(chan prometheus.Metric, 1)
	histogram.Collect(ch)
	close(ch)

	assert.Equal(t, histogram, <-ch)
}

func TestHistogramUnit_Observe(t *testing.T) {
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})
	histogram.Observe(1.1)

	var m dto.Metric
	histogram.Write(&m)
	assert.Equal(t, uint64(1), *(m.Histogram.SampleCount))
	assert.Equal(t, float64(1.1), *(m.Histogram.SampleSum))
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
	v := createHistogram(0)
	histogram := v.With(prometheus.Labels{"label1": "value1"})
	last := histogram.(dynamicvector.Metric).LastEdit()

	histogram.Observe(1)
	assert.True(t, last.Before(histogram.(dynamicvector.Metric).LastEdit()))
}

func createHistogram(ml int) *dynamicvector.Histogram {
	return dynamicvector.NewHistogram(dynamicvector.HistogramOpts{
		Name:        "counter_vector",
		Help:        "testing",
		ConstLabels: prometheus.Labels{"label1": "value1", "label2": "value2"},
		Buckets:     []float64{1, 10, 100},
		MaxLength:   ml,
	})
}
