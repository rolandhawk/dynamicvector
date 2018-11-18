// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rolandhawk/dynamicvector"
	"github.com/stretchr/testify/assert"
)

func createLabels() *dynamicvector.Labels {
	return dynamicvector.NewLabels(prometheus.Labels{"name": "value"})
}

func TestLabels_PromLabelsToValues(t *testing.T) {
	l := createLabels()

	val := l.PromLabelsToValues(prometheus.Labels{"key1": "value1"})
	assert.Equal(t, val, []string{"value1"})
	assert.Equal(t, l.Keys, []string{"key1"})

	val = l.PromLabelsToValues(prometheus.Labels{"key2": "value2"})
	assert.Equal(t, val, []string{"", "value2"})
	assert.Equal(t, l.Keys, []string{"key1", "key2"})

	val = l.PromLabelsToValues(prometheus.Labels{"key1": "value3"})
	assert.Equal(t, val, []string{"value3", ""})
	assert.Equal(t, l.Keys, []string{"key1", "key2"})
}

func TestLabels_ValuesToPromLabels(t *testing.T) {
	l := createLabels()
	l.PromLabelsToValues(prometheus.Labels{"key1": ""})
	l.PromLabelsToValues(prometheus.Labels{"key2": ""})

	assert.Equal(t, prometheus.Labels{"name": "value", "key1": "value", "key2": ""}, l.ValuesToPromLabels([]string{"value"}))
	assert.Equal(t, prometheus.Labels{"name": "value", "key1": "value", "key2": ""}, l.ValuesToPromLabels([]string{"value", ""}))
	assert.Equal(t, prometheus.Labels{"name": "value", "key1": "", "key2": "value"}, l.ValuesToPromLabels([]string{"", "value"}))
}

func TestLabels_Hash(t *testing.T) {
	l := createLabels()

	lbl1 := prometheus.Labels{"key1": "value"}
	l.PromLabelsToValues(lbl1)
	lbl2 := prometheus.Labels{"key2": "value"}
	l.PromLabelsToValues(lbl2)

	hash1 := l.Hash(lbl1)
	hash2 := l.Hash(lbl2)
	assert.NotEqual(t, hash1, hash2)

	lbl3 := prometheus.Labels{"key3": "value"}
	l.PromLabelsToValues(lbl3)

	assert.Equal(t, hash1, l.Hash(lbl1))
	assert.Equal(t, hash2, l.Hash(lbl2))
	assert.NotEqual(t, l.Hash(lbl1), l.Hash(lbl3))
}

func TestLabels_Include(t *testing.T) {
	l := createLabels()
	l.PromLabelsToValues(prometheus.Labels{"key1": "", "key3": ""})

	assert.True(t, l.Include(prometheus.Labels{"key1": "value"}))
	assert.False(t, l.Include(prometheus.Labels{"key2": "value"}))
	assert.False(t, l.Include(prometheus.Labels{"key2": "value", "key1": "value"}))
}
