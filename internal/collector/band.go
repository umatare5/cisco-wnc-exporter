// Package collector provides utilities for WNC collectors.
// This file derives the radio band, and the per-band record it belongs to, from the
// leaves the controller actually reports the band in.
//
// Three different YANG typedefs describe a radio type in this API, and their value
// domains do not overlap:
//
//   - enm-radio-type for access-point-oper-data/radio-oper-data/radio-type
//   - ms-phy-radio-type for client-oper-data/common-oper-data/ms-radio-type
//   - ms-radio-type for client-oper-data/dot11-oper-data/radio-type
//
// The functions below take the SDK record rather than a bare string so that passing
// one subject's value to the other subject's mapping is a compile error. A shared
// string parameter is what allowed the three domains to be conflated in one switch.
package collector

import (
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// Band values published in the band label.
const (
	BandUnknown = "unknown"
	Band24GHz   = "2.4"
	Band5GHz    = "5"
	Band6GHz    = "6"
)

// band-id values of spectrum-aq-worst-table. The leaf is one-based, and every other
// value is left unnamed on purpose, so that a band a later release adds is skipped
// rather than named wrongly.
const (
	rrmWorstBandID24GHz = 1
	rrmWorstBandID5GHz  = 2
	rrmWorstBandID6GHz  = 3
)

// RRMWorstBand returns the band a worst air quality row was ranked for, and reports
// whether the identifier could be named at all.
//
// The source is spectrum-aq-worst-table/band-id, a plain integer carrying neither a
// typedef nor a range, so the mapping is measured rather than modeled. It is one-based,
// which is why it cannot share a table with the band-id of radio-band-info, where zero
// is a band rather than the absence of one.
//
// Unlike the two functions below, this one reports absence instead of naming the band
// unknown. The band is the whole identifier of the series it keys, so a row that cannot
// be named carries nothing to act on, and two such rows would collide on one label set
// and fail the whole scrape rather than the row.
func RRMWorstBand(row *rrm.SpectrumAqWorstTable) (string, bool) {
	switch row.BandID {
	case rrmWorstBandID24GHz:
		return Band24GHz, true
	case rrmWorstBandID5GHz:
		return Band5GHz, true
	case rrmWorstBandID6GHz:
		return Band6GHz, true
	default:
		return "", false
	}
}

// RRMMainBand returns the band an RRM run record reports for, and reports whether the
// PHY type could be named at all.
//
// The source is main-data/phy-type, typed enm-ewlc-dot11-radio-band, which types
// radio-oper-data/current-active-band as well: one value domain, so the three spellings
// below are the three APRadioBand names. Folding the two into one mapping would take the
// bare string the package comment above records as the defect, and their absence contracts
// differ anyway — this one reports absence, for the reason RRMWorstBand records, where an
// info label can carry a band named unknown.
func RRMMainBand(record *rrm.MainData) (string, bool) {
	switch record.PhyType {
	case "dot11-2-dot-4-ghz-band":
		return Band24GHz, true
	case "dot11-5-ghz-band":
		return Band5GHz, true
	case "dot11-6-ghz-band":
		return Band6GHz, true
	default:
		return "", false
	}
}

// APRadioBand returns the band the radio is currently operating on.
//
// The source is radio-oper-data/current-active-band, typed enm-ewlc-dot11-radio-band.
// Neither radio-type nor the slot identifies the band: a dual band radio keeps both
// while it moves between bands, and enm-radio-type has a value for exactly that kind
// of radio. The leaf's own description also lists fewer values than its typedef
// defines, so the description cannot be treated as the value domain.
func APRadioBand(radio *ap.RadioOperData) string {
	switch radio.CurrentActiveBand {
	case "dot11-2-dot-4-ghz-band":
		return Band24GHz
	case "dot11-5-ghz-band":
		return Band5GHz
	case "dot11-6-ghz-band":
		return Band6GHz
	default:
		// dot11-invalid-band, the leaf absent on a remote-LAN radio, and any value a
		// later release adds.
		return BandUnknown
	}
}

// ClientBand returns the band the client associated on.
//
// The source is common-oper-data/ms-radio-type, typed ms-phy-radio-type. Every value
// that typedef defines is listed below, including the ones that describe no band, so
// that reaching the default means the controller reported a value this exporter has
// never been told about rather than one it declined to map.
//
// The channel number cannot be used instead: 6 GHz channel numbering restarts at 1
// and collides with 2.4 GHz. is-11g-client cannot be used either, as the controller
// sets it on clients whose PHY type is not 802.11g.
//
// Not every spelling below has been seen coming back from a controller; the rest come
// from the model, which is not always right about them. That is safe here only because
// a wrong spelling falls through to the unknown band instead of naming another one.
// The test records which values are confirmed.
func ClientBand(data client.CommonOperData) string {
	switch data.MsRadioType {
	case "client-dot11b", "client-dot11g",
		"client-dot11n-24-ghz-prot", "client-dot11ax-24ghz-prot", "client-dot11be-24ghz-prot":
		return Band24GHz
	case "client-dot11a", "client-dot11ac",
		"client-dot11n-5-ghz-prot", "client-dot11ax-5ghz-prot", "client-dot11be-5ghz-prot":
		return Band5GHz
	case "client-dot11ax-6ghz-prot", "client-dot11be-6ghz-prot":
		return Band6GHz
	case "client-unknown-prot", "client-phy-type-notappl", "client-ethernet", "client-802-3":
		// Wired or unclassified clients have no band. Listed so that the default
		// below carries only one meaning.
		return BandUnknown
	default:
		return BandUnknown
	}
}

// currentBandInfo returns the radio-band-info record for the band the radio is
// operating on, and reports whether one was found.
//
// radio-band-info is a list keyed by band-id holding one record per band the radio
// supports, so a dual band radio also carries a record for the band it is not using.
// Reading a fixed index therefore reports another band's regulatory limits and power
// table as though they were the operating band's.
//
// The record is picked by comparing band-id with current-band-id rather than by
// translating either into a band name. Neither leaf documents its value domain, so a
// translation table would pin undocumented numbers in two places at once, while a
// comparison between the two keeps holding if a further band is added.
func currentBandInfo(radio *ap.RadioOperData) (*ap.RadioBandInfo, bool) {
	if len(radio.RadioBandInfo) == 0 {
		return nil, false
	}
	if len(radio.RadioBandInfo) == 1 {
		// One record cannot be ambiguous, so current-band-id is not consulted. That
		// matters because it is typed as a plain integer whose zero value is also a
		// valid band, making an absent leaf indistinguishable from the first band.
		return &radio.RadioBandInfo[0], true
	}

	for i := range radio.RadioBandInfo {
		if int(radio.RadioBandInfo[i].BandID) == radio.CurrentBandID {
			return &radio.RadioBandInfo[i], true
		}
	}
	return nil, false
}
