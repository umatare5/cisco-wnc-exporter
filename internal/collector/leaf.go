// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the rule for a leaf the controller omitted from a response.
package collector

import "github.com/prometheus/client_golang/prometheus"

// A leaf the controller omitted arrives as a nil pointer, and neither function below publishes
// anything for one. Zero and false are legitimate readings of every leaf they carry, so
// substituting either for absence makes the two indistinguishable; and for a configuration
// boolean it inverts the setting, because this controller omits the leaf whose default is in
// force and that default is often true.

// appendFlag adds a metric for a boolean leaf the controller sent and none for one it omitted.
func appendFlag(metrics []Float64Metric, desc *prometheus.Desc, leaf *bool) []Float64Metric {
	if leaf == nil {
		return metrics
	}

	return append(metrics, Float64Metric{desc, boolToFloat64(*leaf)})
}

// appendNumber adds a metric for a numeric leaf the controller sent and none for one it omitted.
func appendNumber[T int | int8](
	metrics []Float64Metric, desc *prometheus.Desc, leaf *T,
) []Float64Metric {
	if leaf == nil {
		return metrics
	}

	return append(metrics, Float64Metric{desc, float64(*leaf)})
}
