// Copyright (c) 2018 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"github.com/prometheus/client_golang/prometheus"
)

type curry struct {
	*vector

	labels prometheus.Labels
}

func (c *curry) Delete(lbs prometheus.Labels) bool {
	return c.vector.Delete(c.mergeLabels(lbs))
}

func (c *curry) GetMetricWith(lbs prometheus.Labels) (prometheus.Metric, error) {
	return c.vector.GetMetricWith(c.mergeLabels(lbs))
}

func (c *curry) With(lbs prometheus.Labels) prometheus.Metric {
	return c.vector.With(c.mergeLabels(lbs))
}

func (c *curry) CurryWith(lbs prometheus.Labels) (Vector, error) {
	return &curry{
		vector: c.vector,
		labels: c.mergeLabels(lbs),
	}, nil
}

func (c *curry) MustCurryWith(lbs prometheus.Labels) Vector {
	v, err := c.CurryWith(lbs)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *curry) mergeLabels(lbs prometheus.Labels) prometheus.Labels {
	res := make(prometheus.Labels)
	for k, v := range c.labels {
		res[k] = v
	}
	for k, v := range lbs {
		res[k] = v
	}

	return res
}
