// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// LabelsProto translate prometheus.Labels into LabelPair protobuf
func LabelsProto(l prometheus.Labels) []*dto.LabelPair {
	var pair []*dto.LabelPair
	for name, value := range l {
		pair = append(pair, &dto.LabelPair{Name: proto.String(name), Value: proto.String(value)})
	}

	return pair
}
