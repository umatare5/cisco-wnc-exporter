// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the band-keyed air quality module of the AP collector.
package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// apBandDescs holds the descriptors of the band-keyed air quality series. It is nil when the
// spectrum module is off, and the metrics.Spectrum guards in the AP collector's Describe and
// Collect are what keep the series out of a scrape and the nil out of a dereference.
type apBandDescs struct {
	worstAirQuality    *prometheus.Desc
	worstAirQualityMin *prometheus.Desc
	worstInterferers   *prometheus.Desc
	worstChannel       *prometheus.Desc
}

// newAPBandDescs builds the descriptors of the band-keyed air quality series.
//
// The band is the whole identifier. The controller ranks the channels of a band across
// every AP that scans it and keeps one row per band, so neither the AP nor the radio the
// reading was taken on keys the row.
func newAPBandDescs() *apBandDescs {
	bandLabels := []string{labelBand}

	return &apBandDescs{
		worstAirQuality: prometheus.NewDesc(
			"wnc_rrm_worst_channel_air_quality_index_avg",
			"Average CleanAir air quality index of the channel the controller ranks worst in "+
				"this band, over the air quality reporting period. Higher is cleaner, and "+
				"wnc_rrm_worst_channel_number reports which channel it is",
			bandLabels, nil,
		),
		worstAirQualityMin: prometheus.NewDesc(
			"wnc_rrm_worst_channel_air_quality_index_min",
			"Lowest CleanAir air quality index the controller saw on that channel during the "+
				"same reporting period, which the average does not carry",
			bandLabels, nil,
		),
		worstInterferers: prometheus.NewDesc(
			"wnc_rrm_worst_channel_interferers",
			"Interference devices CleanAir counts on the channel the controller ranks worst in "+
				"this band",
			bandLabels, nil,
		),
		worstChannel: prometheus.NewDesc(
			"wnc_rrm_worst_channel_number",
			"Channel the controller ranks worst in this band. It is a reading rather than a "+
				"label, so a change moves the value instead of starting a new series",
			bandLabels, nil,
		),
	}
}

// describe sends every descriptor of the band-keyed air quality series.
func (d *apBandDescs) describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{
		d.worstAirQuality, d.worstAirQualityMin, d.worstInterferers, d.worstChannel,
	} {
		ch <- desc
	}
}

// collect publishes the four readings of every band the controller ranked.
//
// It runs once per scrape rather than inside the per-radio loop. The band label is the
// whole identifier, so emitting from that loop would repeat every series with the same
// label set, which Gather rejects by failing the entire endpoint.
//
// Both guards withhold a row rather than publish part of it. A row whose band-id cannot
// be named is skipped instead of labeled unknown, and a row reporting no channel is a
// slot the controller has not ranked: none of the four leaves can be told apart from a
// reading there, and the worst reading of each is what a zero would publish.
func (d *apBandDescs) collect(ch chan<- prometheus.Metric, rows []rrm.SpectrumAqWorstTable) {
	for i := range rows {
		row := &rows[i]

		band, ok := RRMWorstBand(row)
		if !ok {
			continue
		}

		if row.ChannelNum == 0 {
			continue
		}

		for _, metric := range []Float64Metric{
			{d.worstAirQuality, float64(row.Aqi)},
			{d.worstAirQualityMin, float64(row.MinAqi)},
			{d.worstInterferers, float64(row.TotalIntfDeviceCount)},
			{d.worstChannel, float64(row.ChannelNum)},
		} {
			ch <- prometheus.MustNewConstMetric(metric.Desc, prometheus.GaugeValue, metric.Value, band)
		}
	}
}
