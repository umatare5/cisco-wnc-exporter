package collector

import (
	"testing"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
)

// enmRadioTypes is every value of enm-radio-type, the typedef of
// radio-oper-data/radio-type. None of them may influence the band, so whether a given
// spelling was ever seen coming back from a controller does not matter here.
var enmRadioTypes = []string{
	"radio-invalid",
	"radio-80211bg",
	"radio-80211a",
	"radio-80211abgn",
	"radio-uwb",
	"radio-remote-lan",
	"radio-80211-6ghz",
	"radio-80211-xor-5-6ghz",
}

func TestAPRadioBand(t *testing.T) {
	t.Parallel()

	// The three band spellings were seen coming back from a controller, and each one
	// agreed with which per-band RRM list the radio appeared in. dot11-invalid-band
	// comes from the model alone.
	tests := []struct {
		name string
		band string
		want string
	}{
		{"2.4 GHz", "dot11-2-dot-4-ghz-band", Band24GHz},
		{"5 GHz", "dot11-5-ghz-band", Band5GHz},
		{"6 GHz", "dot11-6-ghz-band", Band6GHz},
		{"Invalid band", "dot11-invalid-band", BandUnknown},
		{"Leaf absent", "", BandUnknown},
		{"Value from a later release", "dot11-7-ghz-band", BandUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := APRadioBand(&ap.RadioOperData{CurrentActiveBand: tt.band})
			if got != tt.want {
				t.Errorf("APRadioBand(%q) = %q, want %q", tt.band, got, tt.want)
			}
		})
	}
}

// TestAPRadioBand_IgnoresSlotAndRadioType pins the property the defect violated: the
// slot and the radio type must not be able to override the operating band. A dual
// band radio keeps both while it moves between bands.
func TestAPRadioBand_IgnoresSlotAndRadioType(t *testing.T) {
	t.Parallel()

	for _, radioType := range enmRadioTypes {
		t.Run(radioType, func(t *testing.T) {
			t.Parallel()

			for slot := range 3 {
				radio := &ap.RadioOperData{
					RadioSlotID:       slot,
					RadioType:         radioType,
					CurrentActiveBand: "dot11-6-ghz-band",
				}
				if got := APRadioBand(radio); got != Band6GHz {
					t.Errorf("APRadioBand(slot=%d, radio-type=%s) = %q, want %q",
						slot, radioType, got, Band6GHz)
				}
			}
		})
	}
}

// TestAPRadioBand_NeverDefaultsTo24GHz is the regression test for the defect. Every
// radio type on every slot resolved to 2.4 GHz because the failure path shared a
// value with the 2.4 GHz slot constant.
func TestAPRadioBand_NeverDefaultsTo24GHz(t *testing.T) {
	t.Parallel()

	if BandUnknown == Band24GHz {
		t.Fatal("BandUnknown equals Band24GHz, so an unresolved band would read as a measured one")
	}

	for _, radioType := range enmRadioTypes {
		for slot := range 3 {
			radio := &ap.RadioOperData{RadioSlotID: slot, RadioType: radioType}
			if got := APRadioBand(radio); got != BandUnknown {
				t.Errorf("APRadioBand(slot=%d, radio-type=%s) = %q with no operating band, want %q",
					slot, radioType, got, BandUnknown)
			}
		}
	}
}

// TestAPRadioBand_RejectsForeignEnums feeds values from the other two radio-type
// typedefs, and from the strings the previous implementation compared against, into
// the operating band leaf. None may resolve to a band.
func TestAPRadioBand_RejectsForeignEnums(t *testing.T) {
	t.Parallel()

	foreign := []string{
		"client-dot11ac", "client-dot11ax-24ghz-prot", "client-dot11ax-6ghz-prot",
		"dot11-radio-type-a", "dot11-radio-type-bg", "dot11-radio-type-6ghz",
		"dot11bg", "dot11a",
	}

	for _, value := range foreign {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			if got := APRadioBand(&ap.RadioOperData{CurrentActiveBand: value}); got != BandUnknown {
				t.Errorf("APRadioBand(%q) = %q, want %q", value, got, BandUnknown)
			}
		})
	}
}

func TestClientBand(t *testing.T) {
	t.Parallel()

	// Every value of ms-phy-radio-type. The 802.11be entries were added after the
	// others, so a release that predates them simply never sends them.
	//
	// confirmed records whether the spelling was seen coming back from a controller.
	// The rest come from the model alone, and the model is not always right about
	// them, so a wrong spelling there falls through to the unknown band rather than
	// naming the wrong one. A confirmed value that carries a band must resolve to it,
	// which the subtest below enforces.
	tests := []struct {
		phyType   string
		want      string
		confirmed bool
	}{
		{"client-unknown-prot", BandUnknown, false},
		{"client-dot11b", Band24GHz, false},
		{"client-dot11g", Band24GHz, false},
		{"client-dot11a", Band5GHz, false},
		{"client-dot11n-24-ghz-prot", Band24GHz, true},
		{"client-dot11n-5-ghz-prot", Band5GHz, true},
		{"client-dot11ac", Band5GHz, true},
		{"client-phy-type-notappl", BandUnknown, false},
		{"client-ethernet", BandUnknown, false},
		{"client-dot11ax-5ghz-prot", Band5GHz, true},
		{"client-dot11ax-24ghz-prot", Band24GHz, true},
		{"client-802-3", BandUnknown, false},
		{"client-dot11ax-6ghz-prot", Band6GHz, true},
		{"client-dot11be-24ghz-prot", Band24GHz, false},
		{"client-dot11be-5ghz-prot", Band5GHz, false},
		{"client-dot11be-6ghz-prot", Band6GHz, false},
	}

	if len(tests) != 16 {
		t.Fatalf("the table covers %d values, want every value of ms-phy-radio-type", len(tests))
	}

	for _, tt := range tests {
		t.Run(tt.phyType, func(t *testing.T) {
			t.Parallel()

			got := ClientBand(client.CommonOperData{MsRadioType: tt.phyType})
			if got != tt.want {
				t.Errorf("ClientBand(%q) = %q, want %q", tt.phyType, got, tt.want)
			}
			if tt.confirmed && got == BandUnknown {
				t.Errorf("ClientBand(%q) = %q, but a controller does send this value", tt.phyType, got)
			}
		})
	}
}

// TestClientBand_RejectsUnmappable covers the leaf being absent, values the previous
// implementation invented, and values belonging to the other two typedefs.
func TestClientBand_RejectsUnmappable(t *testing.T) {
	t.Parallel()

	unmappable := []string{
		"",
		// Never defined by ms-phy-radio-type, yet mapped by the previous implementation.
		"client-dot11ac-5-ghz-prot", "client-dot11a-5-ghz-prot", "client-dot11bg-24-ghz-prot",
		// Belong to enm-radio-type and to ms-radio-type.
		"radio-80211a", "radio-80211-xor-5-6ghz", "dot11-radio-type-a",
	}

	for _, value := range unmappable {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			got := ClientBand(client.CommonOperData{MsRadioType: value})
			if got != BandUnknown {
				t.Errorf("ClientBand(%q) = %q, want %q", value, got, BandUnknown)
			}
		})
	}
}

func TestCurrentBandInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		radio     *ap.RadioOperData
		wantFound bool
		wantBand  uint8
	}{
		{
			name:      "No record",
			radio:     &ap.RadioOperData{},
			wantFound: false,
		},
		{
			// A single record is used without consulting current-band-id, which is
			// typed as a plain integer whose zero value is also a valid band.
			name: "One record, current-band-id disagrees",
			radio: &ap.RadioOperData{
				CurrentBandID: 2,
				RadioBandInfo: []ap.RadioBandInfo{{BandID: 1}},
			},
			wantFound: true,
			wantBand:  1,
		},
		{
			name: "Two records, the second is operating",
			radio: &ap.RadioOperData{
				CurrentBandID: 2,
				RadioBandInfo: []ap.RadioBandInfo{{BandID: 1}, {BandID: 2}},
			},
			wantFound: true,
			wantBand:  2,
		},
		{
			name: "Two records, the first is operating",
			radio: &ap.RadioOperData{
				CurrentBandID: 0,
				RadioBandInfo: []ap.RadioBandInfo{{BandID: 0}, {BandID: 1}},
			},
			wantFound: true,
			wantBand:  0,
		},
		{
			name: "Two records, neither is operating",
			radio: &ap.RadioOperData{
				CurrentBandID: 9,
				RadioBandInfo: []ap.RadioBandInfo{{BandID: 1}, {BandID: 2}},
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, found := currentBandInfo(tt.radio)
			if found != tt.wantFound {
				t.Fatalf("currentBandInfo() found = %t, want %t", found, tt.wantFound)
			}
			if !found {
				if got != nil {
					t.Error("currentBandInfo() returned a record while reporting none found")
				}
				return
			}
			if got.BandID != tt.wantBand {
				t.Errorf("currentBandInfo() selected band-id %d, want %d", got.BandID, tt.wantBand)
			}
		})
	}
}
