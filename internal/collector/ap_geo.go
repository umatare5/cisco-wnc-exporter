// Package collector provides collectors for cisco-wnc-exporter.
// This file holds the AP coordinate module of the AP collector.
package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/geolocation"
)

// The bounds of a WGS 84 coordinate pair. The schema the controller serves types both leaves
// as decimal64 with no range statement, so a value outside these is well formed on the wire
// and this is the only place it is refused.
const (
	longitudeBound = 180
	latitudeBound  = 90
)

// apGeoDescs holds the descriptors of the coordinate module. A nil value means the module is
// disabled, which is what keeps both series out of a default scrape.
type apGeoDescs struct {
	longitude *prometheus.Desc
	latitude  *prometheus.Desc
}

// newAPGeoDescs builds the descriptors of the coordinate module.
//
// The unit is the platform's own: `show ap geolocation summary` heads both columns with
// degrees, and the schema the controller serves declares `units "degree"` on both leaves.
// They are two families rather than one carrying an axis label because they are different
// quantities over different domains, so an average or a sum across such a label denotes
// nothing and a sum() spanning both silently mixes them.
func newAPGeoDescs() *apGeoDescs {
	apLabels := []string{labelMAC}

	return &apGeoDescs{
		longitude: prometheus.NewDesc(
			"wnc_ap_longitude_degrees",
			"WGS 84 longitude the controller reports for this AP. It is the position "+
				"configured on or derived for the AP rather than a measurement, so it "+
				"changes when someone changes it. Published only with its latitude",
			apLabels, nil,
		),
		latitude: prometheus.NewDesc(
			"wnc_ap_latitude_degrees",
			"WGS 84 latitude the controller reports for this AP. It is the position "+
				"configured on or derived for the AP rather than a measurement, so it "+
				"changes when someone changes it. Published only with its longitude",
			apLabels, nil,
		),
	}
}

// describe sends both descriptors of the coordinate module.
func (d *apGeoDescs) describe(ch chan<- *prometheus.Desc) {
	ch <- d.longitude
	ch <- d.latitude
}

// collect publishes the coordinate pair of every AP the controller reports a usable one for.
//
// It runs once per scrape rather than inside the per-radio loop. The AP MAC is the whole label
// set, so emitting from that loop would repeat both series with identical labels on a
// multi-radio AP, and Gather answers that by failing the whole endpoint rather than the series.
//
// The pair is emitted together or not at all. A longitude published without its latitude puts
// the AP on the prime meridian at whatever latitude the query joins in, which reads as a
// location rather than as the gap it is.
func (d *apGeoDescs) collect(ch chan<- prometheus.Metric, records []geolocation.ApGeoLocData) {
	for i := range records {
		record := &records[i]

		// The list key is mandatory in the schema, but this house has measured another
		// list of this controller padded with rows whose key leaf was empty, and ApMAC
		// is a plain string with no omitempty, so an omitted key arrives as one.
		if record.ApMAC == "" {
			continue
		}

		// The schema's choice puts one case on the wire, so the invalid case leaves
		// Ellipse nil and this chain covers it without reading Loc.Invalid.
		if record.Loc == nil || record.Loc.Ellipse == nil || record.Loc.Ellipse.Center == nil {
			continue
		}

		center := record.Loc.Ellipse.Center
		longitude, hasLongitude := parseCoordinate(center.Longitude, longitudeBound)
		latitude, hasLatitude := parseCoordinate(center.Latitude, latitudeBound)
		if !hasLongitude || !hasLatitude {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			d.longitude, prometheus.GaugeValue, longitude, record.ApMAC,
		)
		ch <- prometheus.MustNewConstMetric(
			d.latitude, prometheus.GaugeValue, latitude, record.ApMAC,
		)
	}
}

// parseCoordinate reads one coordinate leaf, reporting false for one the controller omitted or
// sent outside its domain.
//
// RFC 7951 writes the decimal64 the schema declares as a JSON string, which is why the SDK
// types the leaf as *string and this is where it becomes a number. The range test is affirmative
// on purpose: ParseFloat accepts "NaN" and "Inf" with a nil error, and a negated test would pass
// both through to poison every aggregation this AP appears in.
func parseCoordinate(leaf *string, bound float64) (float64, bool) {
	if leaf == nil {
		return 0, false
	}

	value, err := strconv.ParseFloat(*leaf, 64)
	if err != nil {
		return 0, false
	}
	if value >= -bound && value <= bound {
		return value, true
	}

	return 0, false
}
