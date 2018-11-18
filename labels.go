// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"bytes"

	farmhash "github.com/dgryski/go-farm"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Labels represent structure data for Metric. It designed to be used internally for dynamicvector.
type Labels struct {
	// Constant is constant label.
	Constant prometheus.Labels

	// Keys is label keys.
	Keys []string

	index map[string]int // index for label name.
}

// NewLabels will create new Labels with initial constant.
func NewLabels(constantLabels prometheus.Labels) *Labels {
	return &Labels{
		Constant: constantLabels,
		index:    make(map[string]int),
	}
}

// PromLabelsToValues will generate label values from prometheus labels. If there is label key that
// has not registered to Labels yet, it will be added.
func (l *Labels) PromLabelsToValues(lbl prometheus.Labels) []string {
	values := make([]string, len(l.Keys))

	for key, value := range lbl {
		if i, ok := l.index[key]; ok {
			values[i] = value
		} else {
			l.index[key] = len(l.Keys)
			l.Keys = append(l.Keys, key)
			values = append(values, value)
		}
	}

	return values
}

// ValuesToPromLabels will generate prometheus.Labels given label values and constant labels.
func (l *Labels) ValuesToPromLabels(values []string) prometheus.Labels {
	lbl := make(prometheus.Labels)

	for i, name := range l.Keys {
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

// Hash will return hash value from labels.
func (l *Labels) Hash(label prometheus.Labels) uint64 {
	var b bytes.Buffer
	for _, key := range l.Keys {
		b.WriteString(label[key])
		b.WriteByte(0)
	}

	// trim so new label key will not change label hash.
	return farmhash.Hash64(bytes.TrimRight(b.Bytes(), "\x00"))
}

// Include will check whether lbl is subset of Labels or not.
func (l *Labels) Include(lbl prometheus.Labels) bool {
	if len(lbl) > len(l.index) {
		return false
	}

	for name := range lbl {
		if _, found := l.index[name]; !found {
			return false
		}
	}

	return true
}

// LabelsProto translate prometheus.Labels into LabelPair protobuf
func LabelsProto(l prometheus.Labels) []*dto.LabelPair {
	if len(l) == 0 {
		return nil
	}

	pair := make([]*dto.LabelPair, 0, len(l))
	for name, value := range l {
		pair = append(pair, &dto.LabelPair{
			Name:  proto.String(name),
			Value: proto.String(value),
		})
	}

	return pair
}
