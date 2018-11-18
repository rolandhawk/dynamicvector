// Copyright (c) 2017 Roland Rifandi Utama
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package dynamicvector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Opts is an option for creating metric vector.
type Opts struct {
	// Namespace, Subsystem, and Name are components of the fully-qualified
	// name of the Metric (created by joining these components with
	// "_"). Only Name is mandatory.
	Namespace string
	Subsystem string
	Name      string

	// Help provides information about this metric. Mandatory!
	Help string

	// ConstLabels are used to attach fixed labels to this metric.
	ConstLabels prometheus.Labels

	// Buckets defines the buckets into which observations are counted. Only for Histogram.
	Buckets []float64

	// Expire are used to set how long dynamicvector will keep the metrics. Zero
	// mean never expire.
	Expire time.Duration

	// MaxLength is maximum length that this vector is allowed to have. Zero mean no maximum length.
	MaxLength int
}

// HistogramOpts is an alias for Opts
type HistogramOpts = Opts

// CounterOpts is an alias for Opts
type CounterOpts = Opts

// GaugeOpts is an alias for Opts
type GaugeOpts = Opts
