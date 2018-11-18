// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"bytes"
	"sync"

	farmhash "github.com/dgryski/go-farm"
	"github.com/prometheus/client_golang/prometheus"
)

// Labels represent structure data for Metric. It designed to be used internally for dynamicvector.
type Labels struct {
	// Constant is constant label.
	Constant prometheus.Labels

	// Names is ordering for label names.
	Names []string

	// Index is index for name.
	Index map[string]int

	mtx sync.Mutex
}

// NewLabels will create new Labels with initial constant.
func NewLabels(constantLabels prometheus.Labels) *Labels {
	return &Labels{
		Constant: constantLabels,
		Index:    make(map[string]int),
	}
}

// Add will add new label name to Labels
func (l *Labels) Add(name string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	i := len(l.Names)
	l.Names = append(l.Names, name)
	l.Index[name] = i
}

// Hash will hash metric labels.
func (l *Labels) Hash(label prometheus.Labels) uint64 {
	var b bytes.Buffer
	for _, name := range l.Names {
		b.WriteString(label[name])
		b.WriteByte(0)
	}

	// trim so new label name will not change label hash.
	return farmhash.Hash64(bytes.TrimRight(b.Bytes(), "\x00"))
}

// Generate will generate prometheus.Labels given label values and constant labels.
func (l *Labels) Generate(values []string) prometheus.Labels {
	lbl := make(prometheus.Labels)

	for i, name := range l.Names {
		if i >= len(values) {
			lbl[name] = ""
		} else {
			lbl[name] = values[i]
		}
	}

	for name, value := range l.Constant {
		lbl[name] = value
	}

	return lbl
}

// Include will check whether lbl is subset of Labels or not.
func (l *Labels) Include(lbl prometheus.Labels) bool {
	if len(lbl) > len(l.Index) {
		return false
	}

	for name := range lbl {
		if _, found := l.Index[name]; !found {
			return false
		}
	}

	return true
}
