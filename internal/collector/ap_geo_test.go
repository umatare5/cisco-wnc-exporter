package collector

import (
	"math"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/geolocation"
)

// TestParseCoordinate_RefusesWhatParseFloatAccepts is why the range test is written in the
// affirmative. strconv.ParseFloat returns a nil error for "NaN", "Inf", "+Inf" and
// "-Infinity", so a guard that only checked the error would put NaN into a gauge and poison
// every aggregation the AP appears in — and a negated range test would pass NaN too, because
// NaN compares false against both bounds.
func TestParseCoordinate_RefusesWhatParseFloatAccepts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		leaf  *string
		bound float64
		want  float64
		ok    bool
	}{
		{"eastern longitude", ptr("139.700000"), longitudeBound, 139.7, true},
		{"western longitude", ptr("-122.084058"), longitudeBound, -122.084058, true},
		{"northern latitude", ptr("35.700000"), latitudeBound, 35.7, true},
		{"southern latitude", ptr("-33.868820"), latitudeBound, -33.86882, true},
		{"null island", ptr("0.0"), latitudeBound, 0, true},
		{"the eastern bound itself", ptr("180.000000"), longitudeBound, 180, true},
		{"the western bound itself", ptr("-180.000000"), longitudeBound, -180, true},
		{"one part in a million past the bound", ptr("180.000001"), longitudeBound, 0, false},
		{"a longitude in a latitude leaf", ptr("139.700000"), latitudeBound, 0, false},

		{"omitted leaf", nil, longitudeBound, 0, false},
		{"empty leaf", ptr(""), longitudeBound, 0, false},
		{"not a number", ptr("north"), longitudeBound, 0, false},
		{"NaN", ptr("NaN"), longitudeBound, 0, false},
		{"Inf", ptr("Inf"), longitudeBound, 0, false},
		{"+Inf", ptr("+Inf"), longitudeBound, 0, false},
		{"-Infinity", ptr("-Infinity"), longitudeBound, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseCoordinate(tt.leaf, tt.bound)
			if ok != tt.ok {
				t.Fatalf("parseCoordinate() ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("parseCoordinate() = %v, want %v", got, tt.want)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("parseCoordinate() returned %v, which no aggregation survives", got)
			}
		})
	}
}

// geoMetricsOnly drives the coordinate module alone, so a case exercises its guards without
// the rest of the AP collector's reads.
type geoMetricsOnly struct {
	descs   *apGeoDescs
	records []geolocation.ApGeoLocData
}

func (g geoMetricsOnly) Describe(ch chan<- *prometheus.Desc) { g.descs.describe(ch) }
func (g geoMetricsOnly) Collect(ch chan<- prometheus.Metric) { g.descs.collect(ch, g.records) }

// geoSeries returns the value of each family the coordinate module published, so a case can
// assert which half of the pair reached the registry.
func geoSeries(t *testing.T, records []geolocation.ApGeoLocData) map[string]float64 {
	t.Helper()

	registry := prometheus.NewRegistry()
	registry.MustRegister(geoMetricsOnly{descs: newAPGeoDescs(), records: records})

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	values := make(map[string]float64, len(families))
	for _, family := range families {
		values[family.GetName()] = family.GetMetric()[0].GetGauge().GetValue()
	}
	return values
}

// TestAPGeoModule_PublishesThePairOrNeither keeps a half-published position out of the
// metrics. A longitude without its latitude puts the AP on the prime meridian at whatever
// latitude the query joins in, which reads as a location rather than as the gap it is.
func TestAPGeoModule_PublishesThePairOrNeither(t *testing.T) {
	t.Parallel()

	point := func(longitude, latitude *string) []geolocation.ApGeoLocData {
		return []geolocation.ApGeoLocData{{
			ApMAC: fixtureAPMAC,
			Loc: &geolocation.GeoLocInfo{
				Ellipse: &geolocation.GeoLocEllipse{
					Center: &geolocation.GeoLocPoint{Longitude: longitude, Latitude: latitude},
				},
			},
		}}
	}

	tests := []struct {
		name    string
		records []geolocation.ApGeoLocData
		want    int
	}{
		{"both coordinates", point(ptr("139.700000"), ptr("35.700000")), 2},
		{"latitude omitted", point(ptr("139.700000"), nil), 0},
		{"longitude omitted", point(nil, ptr("35.700000")), 0},
		{"latitude out of its domain", point(ptr("139.700000"), ptr("95.000000")), 0},
		{"longitude unparsable", point(ptr("east"), ptr("35.700000")), 0},

		{
			// The choice puts one case on the wire, so an invalid position leaves the
			// ellipse absent rather than setting a flag beside it.
			"no ellipse",
			[]geolocation.ApGeoLocData{{ApMAC: fixtureAPMAC, Loc: &geolocation.GeoLocInfo{}}},
			0,
		},
		{
			"no location container",
			[]geolocation.ApGeoLocData{{ApMAC: fixtureAPMAC}},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := len(geoSeries(t, tt.records)); got != tt.want {
				t.Errorf("collect() published %d series, want %d", got, tt.want)
			}
		})
	}
}

// TestAPGeoModule_SkipsARowWithNoKey keeps an empty mac label out of the metrics. The schema
// makes the list key mandatory, but this house has measured another list of this controller
// padded with rows whose key leaf was empty, and the SDK types ApMAC as a plain string, so an
// omitted key is indistinguishable from an empty one here.
func TestAPGeoModule_SkipsARowWithNoKey(t *testing.T) {
	t.Parallel()

	center := &geolocation.GeoLocPoint{Longitude: ptr("139.700000"), Latitude: ptr("35.700000")}
	records := []geolocation.ApGeoLocData{
		{
			ApMAC: "",
			Loc: &geolocation.GeoLocInfo{
				Ellipse: &geolocation.GeoLocEllipse{Center: center},
			},
		},
	}

	if got := len(geoSeries(t, records)); got != 0 {
		t.Errorf("collect() published %d series for a row with no key, want 0", got)
	}

	// Without this the case would pass on a collect that published nothing at all.
	records[0].ApMAC = fixtureAPMAC
	if got := len(geoSeries(t, records)); got != 2 {
		t.Fatalf("collect() published %d series once the row carried its key, want 2", got)
	}
}
