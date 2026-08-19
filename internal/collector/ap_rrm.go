// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the band-keyed RRM run instants of the AP collector.
package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// apRRMDescs holds the descriptors of the band-keyed RRM run instants. A nil value means
// the radio module is disabled, which is what keeps them out of a default scrape.
type apRRMDescs struct {
	lastRFGroupingRunAt *prometheus.Desc
	lastDCARunAt        *prometheus.Desc
}

// newAPRRMDescs builds the descriptors of the band-keyed RRM run instants.
//
// The band is the whole identifier. The controller keeps one record per band rather than
// one per radio, so neither the AP nor the radio an assignment lands on keys the series.
func newAPRRMDescs() *apRRMDescs {
	bandLabels := []string{labelBand}

	return &apRRMDescs{
		lastRFGroupingRunAt: prometheus.NewDesc(
			"wnc_rrm_last_rf_grouping_run_timestamp_seconds",
			"Unix timestamp of the last RF grouping run the controller reports for this band. "+
				"It advances when the algorithm runs rather than when a channel or a transmit "+
				"power level changes, so it can advance with every reading beside it unchanged",
			bandLabels, nil,
		),
		lastDCARunAt: prometheus.NewDesc(
			"wnc_rrm_last_dca_run_timestamp_seconds",
			"Unix timestamp of the last DCA run the controller reports for this band. It "+
				"advances when the algorithm runs rather than when a radio's channel changes, "+
				"so wnc_ap_channel_changes_total can stay flat across a run, and the controller "+
				"schedules it apart from the RF grouping run",
			bandLabels, nil,
		),
	}
}

// describe sends every descriptor of the band-keyed RRM run instants.
func (d *apRRMDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.lastRFGroupingRunAt
	ch <- d.lastDCARunAt
}

// collect publishes the run instants of every band the controller reports one for.
//
// It runs once per scrape rather than inside the per-radio loop. The band label is the
// whole identifier, so emitting from that loop would repeat every series with the same
// label set, which Gather rejects by failing the entire endpoint.
//
// Each guard withholds only what it guards. A record whose PHY type cannot be named
// carries no identity, for the reason RRMWorstBand records. The group container and the
// channel assignment container inside it are both optional on the wire, and reading
// through an absent one ends this collector's scrape where it stands, which costs every
// series after it with nothing but a log line to say so. The channel assignment guard
// comes last so that a band the controller sent a grouping instant for keeps it.
func (d *apRRMDescs) collect(ch chan<- prometheus.Metric, records []rrm.MainData) {
	for i := range records {
		record := &records[i]

		band, ok := RRMMainBand(record)
		if !ok {
			continue
		}

		if record.Grp == nil {
			continue
		}

		emitTimestamp(ch, d.lastRFGroupingRunAt, record.Grp.LastRun, band)

		if record.Grp.DCA == nil {
			continue
		}

		emitTimestamp(ch, d.lastDCARunAt, record.Grp.DCA.DCALastRun, band)
	}
}
