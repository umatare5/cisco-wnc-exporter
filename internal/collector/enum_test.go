package collector

import (
	"slices"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// enumTables names every table the encoding resolves a reading against, keyed by the
// typedef it transcribes. Go cannot enumerate package-level variables, so this list is
// hand kept: a table missing from here is invisible to every assertion in this file.
var enumTables = map[string]enumTable{
	"ap-discovery-failure-reason":       apDiscoveryFailureReasons,
	"enm-ap-join-failure-reason":        apJoinFailureReasons,
	"enm-ap-config-failure-reason":      apConfigFailureReasons,
	"last-failure-phase":                apFailurePhases,
	"enm-dtls-handshake-failure-reason": apDTLSFailureReasons,
	"spam-ap-reboot-reason":             apRebootReasons,
	"spam-ap-disconnect-reason":         apDisconnectReasons,
	"enum-ap-state":                     apOperationStates,
	"client-co-state":                   clientStates,
	"dot11-client-roam-type":            clientRoamTypes,
	"apf-vap-pmf-policies":              wlanPMFPolicies,
	"ft-dot11r-mode":                    wlanFTModes,
}

// TestEnumTables_NumberEveryMemberOfTheirEnumerationOnce pins each table to the member
// count and the value range the controller's schema declares. Count, distinctness and
// coverage together pin the set of values a table carries, so a dropped or a duplicated
// row fails here. They do not pin which spelling holds which value: a mistyped spelling,
// and two spellings whose values are exchanged, both leave that set intact, and are
// caught only for the rows another assertion names.
func TestEnumTables_NumberEveryMemberOfTheirEnumerationOnce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typedef string
		members int
		lowest  float64
	}{
		{"ap-discovery-failure-reason", 19, 0},
		{"enm-ap-join-failure-reason", 42, 0},
		{"enm-ap-config-failure-reason", 14, 0},
		{"last-failure-phase", 7, 0},
		{"enm-dtls-handshake-failure-reason", 10, 0},
		{"spam-ap-reboot-reason", 59, 0},
		{"spam-ap-disconnect-reason", 41, 0},
		// The one enumeration of the twelve that declares no member at zero, which is
		// why no value is free to stand for a reading the encoding cannot name.
		{"enum-ap-state", 6, 1},
		{"client-co-state", 14, 0},
		{"dot11-client-roam-type", 5, 0},
		{"apf-vap-pmf-policies", 3, 0},
		{"ft-dot11r-mode", 3, 0},
	}

	if len(tests) != len(enumTables) {
		t.Fatalf("%d typedefs are pinned but %d tables are registered, so one is unpinned",
			len(tests), len(enumTables))
	}

	for _, tt := range tests {
		t.Run(tt.typedef, func(t *testing.T) {
			t.Parallel()

			table, ok := enumTables[tt.typedef]
			if !ok {
				t.Fatalf("%s has no table registered", tt.typedef)
			}
			if len(table) != tt.members {
				t.Errorf("%s has %d members, want %d", tt.typedef, len(table), tt.members)
			}
			if _, ok := table[""]; ok {
				t.Errorf("%s numbers the empty spelling, which no leaf carries", tt.typedef)
			}

			owner := make(map[float64]string, len(table))
			for spelling, value := range table {
				if first, taken := owner[value]; taken {
					t.Errorf("%s gives %v to both %q and %q", tt.typedef, value, first, spelling)
					continue
				}
				owner[value] = spelling
			}

			for value := tt.lowest; value < tt.lowest+float64(tt.members); value++ {
				if _, ok := owner[value]; !ok {
					t.Errorf("%s numbers no spelling %v, so its run is not contiguous", tt.typedef, value)
				}
			}
		})
	}
}

// TestEnumTables_SpellingsAreUniqueAcrossEnumerations is what licenses one shared table
// type and one emit helper. A spelling two typedefs share would let a reading resolved
// against the wrong table return the other enumeration's number instead of nothing.
func TestEnumTables_SpellingsAreUniqueAcrossEnumerations(t *testing.T) {
	t.Parallel()

	const wantSpellings = 223

	owner := make(map[string]string, wantSpellings)
	for typedef, table := range enumTables {
		for spelling := range table {
			if first, taken := owner[spelling]; taken {
				t.Errorf("%q belongs to both %s and %s", spelling, first, typedef)
				continue
			}
			owner[spelling] = typedef
		}
	}

	if len(owner) != wantSpellings {
		t.Errorf("the twelve tables carry %d distinct spellings, want %d", len(owner), wantSpellings)
	}
}

// TestEnumTables_PinTheMembersAPublishedStatementNames holds the numbers a HELP string,
// a shipped query or the withhold rule names. Each of these is quoted somewhere outside
// the table, so renumbering one would leave a published statement wrong.
func TestEnumTables_PinTheMembersAPublishedStatementNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typedef  string
		spelling string
		want     float64
	}{
		// Zero is a real state here, not the absence of one.
		{"client-co-state", "client-status-idle", 0},
		// And here it is the enumeration's own unknown member, misspelled by the device.
		{"spam-ap-disconnect-reason", "unkown", 0}, //nolint:misspell
		// The state a client holds while it passes traffic, named in the HELP.
		{"client-co-state", "client-status-run", 11},
		// The four "no failure" members the reason HELP strings name as 0.
		{"ap-discovery-failure-reason", "disc-fail-none", 0},
		{"enm-ap-join-failure-reason", "jf-none", 0},
		{"enm-ap-config-failure-reason", "cf-none", 0},
		{"enm-dtls-handshake-failure-reason", "dtls-hs-success", 0},
		// The phase HELP enumerates the whole run and names both ends.
		{"last-failure-phase", "ap-con-failure-unknown", 0},
		{"last-failure-phase", "ap-con-failure-run", 6},
		// The two AP states the oper-state test pins separately.
		{"enum-ap-state", "registered", 4},
		{"enum-ap-state", "downloading", 5},
		// The middle PMF setting admits an unprotected association, so its number is
		// what a query has to match rather than a threshold.
		{"apf-vap-pmf-policies", "apf-vap-pmf-optional", 1},
		{"apf-vap-pmf-policies", "apf-vap-pmf-required", 2},
		// The FT HELP says to match by equality, which only holds at these numbers.
		{"ft-dot11r-mode", "dot11r-enabled", 1},
		{"ft-dot11r-mode", "dot11r-adaptive-enabled", 2},
		// The oper-state HELP enumerates the whole run; four members are named there and
		// nowhere else.
		{"enum-ap-state", "ap-down", 1},
		{"enum-ap-state", "ap-up", 2},
		{"enum-ap-state", "unregistered", 3},
		{"enum-ap-state", "pre-downloading", 6},
		// The five interior phases the error-phase HELP enumerates.
		{"last-failure-phase", "ap-con-failure-discovery", 1},
		{"last-failure-phase", "ap-con-failure-dtls", 2},
		{"last-failure-phase", "ap-con-failure-join", 3},
		{"last-failure-phase", "ap-con-failure-config", 4},
		{"last-failure-phase", "ap-con-failure-imgdwnld", 5},
		// The four roam types the fixture does not witness.
		{"dot11-client-roam-type", "dot11-roam-type-none", 0},
		{"dot11-client-roam-type", "dot11-roam-type-slow-11i", 1},
		{"dot11-client-roam-type", "dot11-roam-type-cckm", 3},
		{"dot11-client-roam-type", "dot11-roam-type-fast-11r", 4},
		// The PMF setting the fixture does not carry.
		{"apf-vap-pmf-policies", "apf-vap-pmf-disabled", 0},
		// The member the reboot HELP names as 0.
		{"spam-ap-reboot-reason", "ap-reboot-reason-none", 0},
	}

	for _, tt := range tests {
		got, ok := enumTables[tt.typedef][tt.spelling]
		if !ok {
			t.Errorf("%s numbers no %q", tt.typedef, tt.spelling)
			continue
		}
		if got != tt.want {
			t.Errorf("%s[%q] = %v, want %v", tt.typedef, tt.spelling, got, tt.want)
		}
	}

	// The controller declares its own unknown member misspelled and sends that spelling.
	// The corrected spelling must be absent, because the formatter this repository runs
	// with --fix rewrites the misspelling wherever it is unguarded, and a table rewritten
	// that way would never match the wire again.
	if _, ok := apDisconnectReasons["unknown"]; ok {
		t.Error("apDisconnectReasons carries the corrected spelling, which the controller never sends")
	}
}

// enumReadingCollector publishes one reading through emitEnumReading, so the emit is
// observed through a registry. Reading a value off a bare prometheus.Metric would name a
// type from client_model, which is an indirect module of this repository.
type enumReadingCollector struct {
	desc    *prometheus.Desc
	table   enumTable
	reading string
}

func (c enumReadingCollector) Describe(ch chan<- *prometheus.Desc) { ch <- c.desc }

func (c enumReadingCollector) Collect(ch chan<- prometheus.Metric) {
	emitEnumReading(ch, c.desc, c.table, c.reading, fixtureAPMAC)
}

// TestEmitEnumReading_WithholdsWhatItCannotNumber exercises every path of the emit. A
// lookup that discarded its second result would publish the two unnumbered readings as
// zero, and zero is a real member of this enumeration, so only a presence assertion
// separates them.
func TestEmitEnumReading_WithholdsWhatItCannotNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reading string
		want    []float64
	}{
		{"a member of the table", "dtls-hs-peer-alert", []float64{5}},
		{"a member the enumeration numbers zero", "dtls-hs-success", []float64{0}},
		{"a spelling no release declares", fixtureUnnumberedSpelling, nil},
		{"a spelling another enumeration declares", "jf-none", nil},
		{"an absent leaf", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := prometheus.NewRegistry()
			registry.MustRegister(enumReadingCollector{
				desc:    prometheus.NewDesc("wnc_test_enum_reading", "test", []string{labelMAC}, nil),
				table:   apDTLSFailureReasons,
				reading: tt.reading,
			})

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}

			var got []float64
			for _, family := range families {
				for _, metric := range family.GetMetric() {
					got = append(got, metric.GetGauge().GetValue())
				}
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("emitEnumReading(%q) published %v, want %v", tt.reading, got, tt.want)
			}
		})
	}
}

// TestEnumFamilies_HelpDescribesTheValueShape keeps the shipped HELP of the twelve in
// step with what they now publish. Nothing else in this repository reads a HELP string,
// so a descriptor reverted to the label shape would otherwise ship green.
func TestEnumFamilies_HelpDescribesTheValueShape(t *testing.T) {
	t.Parallel()

	names := []string{
		"wnc_ap_last_discovery_failure_reason",
		"wnc_ap_last_join_failure_reason",
		"wnc_ap_last_config_failure_reason",
		"wnc_ap_last_error_phase",
		"wnc_ap_last_dtls_failure_reason",
		"wnc_ap_last_reboot_reason",
		"wnc_ap_last_disconnect_reason",
		"wnc_ap_oper_state",
		"wnc_client_state",
		"wnc_client_roam_type",
		"wnc_wlan_pmf_state",
		"wnc_wlan_ft_state",
	}
	stale := []string{"state label", "always 1"}

	descs := make([]string, 0, 200)
	for _, collector := range fixtureCollectors(t, fullFixtureSnapshot()) {
		ch := make(chan *prometheus.Desc, 200)
		collector.Describe(ch)
		close(ch)
		for desc := range ch {
			descs = append(descs, desc.String())
		}
	}

	for _, name := range names {
		index := slices.IndexFunc(descs, func(desc string) bool {
			return strings.Contains(desc, "fqName: \""+name+"\"")
		})
		if index < 0 {
			t.Errorf("%s is not described, so the clause checks below prove nothing", name)
			continue
		}
		for _, clause := range stale {
			if strings.Contains(descs[index], clause) {
				t.Errorf("%s HELP still says %q, which the label shape left behind", name, clause)
			}
		}
	}
}
