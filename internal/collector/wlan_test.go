package collector

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/wnc"
)

func TestNewWLANCollector(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		metrics WLANMetrics
		wantNil bool
	}{
		{
			"All modules disabled",
			WLANMetrics{},
			false,
		},
		{
			"General module enabled",
			WLANMetrics{General: true},
			false,
		},
		{
			"Traffic module enabled",
			WLANMetrics{Traffic: true},
			false,
		},
		{
			"Config module enabled",
			WLANMetrics{Config: true},
			false,
		},
		{
			"Info module enabled",
			WLANMetrics{Info: true},
			false,
		},
		{
			"All modules enabled",
			WLANMetrics{
				General: true,
				Traffic: true,
				Config:  true,
				Info:    true,
			},
			false,
		},
		{
			"Info module with labels",
			WLANMetrics{
				Info:       true,
				InfoLabels: []string{"name"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewWLANCollector(nil, nil, tt.metrics)

			if tt.wantNil {
				if collector != nil {
					t.Errorf("NewWLANCollector() = %v, want nil", collector)
				}
			} else {
				if collector == nil {
					t.Errorf("NewWLANCollector() = nil, want non-nil")
					return
				}
				if collector.metrics.General != tt.metrics.General {
					t.Errorf(
						"NewWLANCollector() General = %v, want %v",
						collector.metrics.General,
						tt.metrics.General,
					)
				}
				if collector.metrics.Info != tt.metrics.Info {
					t.Errorf("NewWLANCollector() Info = %v, want %v", collector.metrics.Info, tt.metrics.Info)
				}
			}
		})
	}
}

func TestWLANCollector_isAnyMetricFlagEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metrics  WLANMetrics
		expected bool
	}{
		{
			"All disabled",
			WLANMetrics{},
			false,
		},
		{
			"General enabled",
			WLANMetrics{General: true},
			true,
		},
		{
			"Traffic enabled",
			WLANMetrics{Traffic: true},
			true,
		},
		{
			"Config enabled",
			WLANMetrics{Config: true},
			true,
		},
		{
			"Info enabled",
			WLANMetrics{Info: true},
			true,
		},
		{
			"Multiple enabled",
			WLANMetrics{General: true, Traffic: true},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &WLANCollector{metrics: tt.metrics}
			if got := collector.isAnyMetricFlagEnabled(); got != tt.expected {
				t.Errorf("isAnyMetricFlagEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWLANCollector_Describe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		metrics     WLANMetrics
		expectDescs int
	}{
		{
			"No modules enabled",
			WLANMetrics{},
			0,
		},
		{
			"General module only",
			WLANMetrics{General: true},
			1, // enabled
		},
		{
			"Traffic module only",
			WLANMetrics{Traffic: true},
			3, // client_count, onboarding_clients, data_usage
		},
		{
			"Config module only",
			WLANMetrics{Config: true},
			// auth_psk, auth_dot1x, auth_dot1x_sha256, wpa2, wpa3, session_timeout,
			// load_balance, 11k_neighbor_list, steering, central_switching, central_auth,
			// central_dhcp, central_assoc, policy_enabled, pmf_state, ft_state,
			// policy_binding
			17,
		},
		{
			"Info module only",
			WLANMetrics{Info: true},
			1, // info metric
		},
		{
			"All modules enabled",
			WLANMetrics{
				General: true,
				Traffic: true,
				Config:  true,
				Info:    true,
			},
			22, // 1+3+17+1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewWLANCollector(nil, nil, tt.metrics)

			ch := make(chan *prometheus.Desc, 50)
			collector.Describe(ch)
			close(ch)

			count := 0
			for range ch {
				count++
			}

			if count != tt.expectDescs {
				t.Errorf("Describe() sent %d descriptors, want %d", count, tt.expectDescs)
			}
		})
	}
}

func TestBuildWLANToPolicyMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		policyListEntries []wlan.PolicyListEntry
		wlanPolicies      []wlan.WlanPolicy
		expectedMappings  map[string]string
	}{
		{
			"Empty inputs",
			[]wlan.PolicyListEntry{},
			[]wlan.WlanPolicy{},
			map[string]string{},
		},
		{
			"Single mapping",
			[]wlan.PolicyListEntry{
				{
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "policy1"},
						},
					},
				},
			},
			[]wlan.WlanPolicy{
				{PolicyProfileName: "policy1"},
			},
			map[string]string{
				"profile1": "policy1",
			},
		},
		{
			"Multiple mappings",
			[]wlan.PolicyListEntry{
				{
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "policy1"},
							{WLANProfileName: "profile2", PolicyProfileName: "policy2"},
						},
					},
				},
			},
			[]wlan.WlanPolicy{
				{PolicyProfileName: "policy1"},
				{PolicyProfileName: "policy2"},
			},
			map[string]string{
				"profile1": "policy1",
				"profile2": "policy2",
			},
		},
		{
			"Missing policy reference",
			[]wlan.PolicyListEntry{
				{
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "nonexistent"},
						},
					},
				},
			},
			[]wlan.WlanPolicy{
				{PolicyProfileName: "policy1"},
			},
			map[string]string{},
		},
		{
			// One WLAN bound through two policy tags to two different policy profiles.
			// docs/collector.wlan.md states that the six policy series then report the
			// last binding the exporter can resolve, and nothing pinned that: every case
			// above binds each WLAN profile exactly once, so a first-wins rewrite would
			// pass them all. The controller returns these entries in datastore order,
			// which is not the order they were configured in, so which one wins is not a
			// property an operator can predict — the point of pinning it is that the
			// choice cannot change silently.
			"One WLAN bound through two tags reports the last binding",
			[]wlan.PolicyListEntry{
				{
					TagName: "tag1",
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "policy1"},
						},
					},
				},
				{
					TagName: "tag2",
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "policy2"},
						},
					},
				},
			},
			[]wlan.WlanPolicy{
				{PolicyProfileName: "policy1"},
				{PolicyProfileName: "policy2"},
			},
			map[string]string{
				"profile1": "policy2",
			},
		},
		{
			// The same shape with the second binding naming a policy profile the
			// controller did not return. The skip that handles it means an unresolvable
			// binding cannot displace a resolvable one, which is the other half of the
			// documented contract.
			"An unresolvable later binding leaves the earlier one in place",
			[]wlan.PolicyListEntry{
				{
					TagName: "tag1",
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "policy1"},
						},
					},
				},
				{
					TagName: "tag2",
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{WLANProfileName: "profile1", PolicyProfileName: "nonexistent"},
						},
					},
				},
			},
			[]wlan.WlanPolicy{
				{PolicyProfileName: "policy1"},
			},
			map[string]string{
				"profile1": "policy1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildWLANToPolicyMap(tt.policyListEntries, tt.wlanPolicies)

			if len(got) != len(tt.expectedMappings) {
				t.Errorf(
					"buildWLANToPolicyMap() returned map with %d entries, want %d",
					len(got),
					len(tt.expectedMappings),
				)
			}

			for wlanProfile, expectedPolicyName := range tt.expectedMappings {
				if policy, exists := got[wlanProfile]; !exists {
					t.Errorf("buildWLANToPolicyMap() missing WLAN profile %q", wlanProfile)
				} else if policy.PolicyProfileName != expectedPolicyName {
					t.Errorf(
						"buildWLANToPolicyMap()[%q].PolicyProfileName = %q, want %q",
						wlanProfile,
						policy.PolicyProfileName,
						expectedPolicyName,
					)
				}
			}
		})
	}
}

func TestDetermineWLANEnabledStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		wlanEntry wlan.WlanCfgEntry
		expected  int
	}{
		{
			"Enabled WLAN",
			wlan.WlanCfgEntry{
				APFVapIDData: &wlan.APFVapIDData{
					WlanStatus: true,
				},
			},
			1,
		},
		{
			"Disabled WLAN",
			wlan.WlanCfgEntry{
				APFVapIDData: &wlan.APFVapIDData{
					WlanStatus: false,
				},
			},
			0,
		},
		{
			"Nil APFVapIDData",
			wlan.WlanCfgEntry{
				APFVapIDData: nil,
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineWLANEnabledStatus(tt.wlanEntry); got != tt.expected {
				t.Errorf("determineWLANEnabledStatus() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDetermineWLANName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		wlanEntry wlan.WlanCfgEntry
		expected  string
	}{
		{
			"SSID present",
			wlan.WlanCfgEntry{
				ProfileName: "profile1",
				APFVapIDData: &wlan.APFVapIDData{
					SSID: "GuestNetwork",
				},
			},
			"GuestNetwork",
		},
		{
			"Empty SSID - use profile name",
			wlan.WlanCfgEntry{
				ProfileName: "profile1",
				APFVapIDData: &wlan.APFVapIDData{
					SSID: "",
				},
			},
			"profile1",
		},
		{
			"Nil APFVapIDData - use profile name",
			wlan.WlanCfgEntry{
				ProfileName:  "profile1",
				APFVapIDData: nil,
			},
			"profile1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineWLANName(tt.wlanEntry); got != tt.expected {
				t.Errorf("determineWLANName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetermineSessionTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   *wlan.WlanPolicy
		expected int
	}{
		{
			"Valid timeout",
			&wlan.WlanPolicy{
				WlanTimeout: &wlan.WlanTimeout{
					SessionTimeout: 3600,
				},
			},
			3600,
		},
		{
			"Zero timeout",
			&wlan.WlanPolicy{
				WlanTimeout: &wlan.WlanTimeout{
					SessionTimeout: 0,
				},
			},
			0,
		},
		{
			"Nil WlanTimeout",
			&wlan.WlanPolicy{
				WlanTimeout: nil,
			},
			0,
		},
		{
			"Nil policy",
			nil,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineSessionTimeout(tt.policy); got != tt.expected {
				t.Errorf("determineSessionTimeout() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDetermineCentralSwitchingValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   *wlan.WlanPolicy
		expected float64
	}{
		{
			"Central switching enabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralSwitching: true,
				},
			},
			1.0,
		},
		{
			"Central switching disabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralSwitching: false,
				},
			},
			0.0,
		},
		{
			"Nil WlanSwitchingPolicy",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: nil,
			},
			0.0,
		},
		{
			"Nil policy",
			nil,
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineCentralSwitchingValue(tt.policy); got != tt.expected {
				t.Errorf("determineCentralSwitchingValue() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestDetermineCentralAuthenticationValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   *wlan.WlanPolicy
		expected float64
	}{
		{
			"Central authentication enabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralAuthentication: true,
				},
			},
			1.0,
		},
		{
			"Central authentication disabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralAuthentication: false,
				},
			},
			0.0,
		},
		{
			"Nil WlanSwitchingPolicy",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: nil,
			},
			0.0,
		},
		{
			"Nil policy",
			nil,
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineCentralAuthenticationValue(tt.policy); got != tt.expected {
				t.Errorf("determineCentralAuthenticationValue() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestDetermineCentralDHCPValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   *wlan.WlanPolicy
		expected float64
	}{
		{
			"Central DHCP enabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralDHCP: true,
				},
			},
			1.0,
		},
		{
			"Central DHCP disabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralDHCP: false,
				},
			},
			0.0,
		},
		{
			"Nil WlanSwitchingPolicy",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: nil,
			},
			0.0,
		},
		{
			"Nil policy",
			nil,
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineCentralDHCPValue(tt.policy); got != tt.expected {
				t.Errorf("determineCentralDHCPValue() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestDetermineCentralAssocEnableValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		policy   *wlan.WlanPolicy
		expected float64
	}{
		{
			"Central association enabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralAssocEnable: true,
				},
			},
			1.0,
		},
		{
			"Central association disabled",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
					CentralAssocEnable: false,
				},
			},
			0.0,
		},
		{
			"Nil WlanSwitchingPolicy",
			&wlan.WlanPolicy{
				WlanSwitchingPolicy: nil,
			},
			0.0,
		},
		{
			"Nil policy",
			nil,
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineCentralAssocEnableValue(tt.policy); got != tt.expected {
				t.Errorf("determineCentralAssocEnableValue() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestWLANCollector_buildWLANInfoLabelValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		infoLabels []string
		id         string
		wlanName   string
		expected   []string
	}{
		{
			"Only required labels",
			[]string{"id"},
			"1",
			"GuestNetwork",
			[]string{"1"},
		},
		{
			"With name label",
			[]string{"id", "name"},
			"1",
			"GuestNetwork",
			[]string{"1", "GuestNetwork"},
		},
		{
			"Unknown label returns empty",
			[]string{"id", "unknown"},
			"1",
			"GuestNetwork",
			[]string{"1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &WLANCollector{
				infoLabelNames: tt.infoLabels,
			}
			got := collector.buildWLANInfoLabelValues(tt.id, tt.wlanName)

			if len(got) != len(tt.expected) {
				t.Errorf(
					"buildWLANInfoLabelValues() returned %d values, want %d",
					len(got),
					len(tt.expected),
				)
			}

			for i, expectedValue := range tt.expected {
				if got[i] != expectedValue {
					t.Errorf(
						"buildWLANInfoLabelValues()[%d] = %q, want %q",
						i,
						got[i],
						expectedValue,
					)
				}
			}
		})
	}
}

func TestWLANCollector_MetricNames(t *testing.T) {
	t.Parallel()

	collector := NewWLANCollector(nil, nil, WLANMetrics{
		General: true,
		Traffic: true,
		Config:  true,
		Info:    true,
	})

	tests := []struct {
		desc     *prometheus.Desc
		expected string
	}{
		{collector.enabledDesc, "wnc_wlan_enabled"},
		{collector.clientCountDesc, "wnc_wlan_clients"},
		{collector.onboardingDesc, "wnc_wlan_onboarding_clients"},
		{collector.authPskDesc, "wnc_wlan_auth_psk_enabled"},
		{collector.authDot1xDesc, "wnc_wlan_auth_dot1x_enabled"},
		{collector.wpa3EnabledDesc, "wnc_wlan_wpa3_enabled"},
		{collector.sessionTimeoutDesc, "wnc_wlan_session_timeout_seconds"},
		{collector.infoDesc, "wnc_wlan_info"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if tt.desc == nil {
				t.Errorf("Descriptor for %s is nil", tt.expected)
				return
			}

			descStr := tt.desc.String()
			if !strings.Contains(descStr, tt.expected) {
				t.Errorf("Descriptor %s does not contain expected metric name %s", descStr, tt.expected)
			}

			if !strings.Contains(descStr, "wnc_wlan_") {
				t.Errorf("Descriptor %s does not follow wnc_wlan_ naming convention", descStr)
			}
		})
	}
}

func TestWLANCollector_buildWLANStats(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		clientData []client.CommonOperData
		expected   map[int]wlanStats
	}{
		{
			"Empty data",
			[]client.CommonOperData{},
			map[int]wlanStats{},
		},
		{
			"Nil data",
			nil,
			map[int]wlanStats{},
		},
		{
			"Single client on WLAN 1",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
			},
			map[int]wlanStats{
				1: {clientCount: 1},
			},
		},
		{
			"Multiple clients on same WLAN",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 1, CoState: ClientStatusRun},
			},
			map[int]wlanStats{
				1: {clientCount: 2},
			},
		},
		{
			"Clients on different WLANs",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 2, CoState: ClientStatusRun},
			},
			map[int]wlanStats{
				1: {clientCount: 1},
				2: {clientCount: 1},
			},
		},
		{
			"Skip non-RUN clients",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 1, CoState: "IDLE"},
			},
			map[int]wlanStats{
				1: {clientCount: 1},
			},
		},
		{
			"Client without traffic stats",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
			},
			map[int]wlanStats{
				1: {clientCount: 1},
			},
		},
		{
			"Traffic stats unavailable",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
			},
			map[int]wlanStats{
				1: {clientCount: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &WLANCollector{}
			got := collector.buildWLANStats(tt.clientData)

			if len(got) != len(tt.expected) {
				t.Errorf(
					"buildWLANStats() returned map with %d entries, want %d",
					len(got),
					len(tt.expected),
				)
			}

			for wlanID, expectedStats := range tt.expected {
				if gotStats, exists := got[wlanID]; !exists {
					t.Errorf("buildWLANStats() missing WLAN ID %d", wlanID)
				} else if gotStats.clientCount != expectedStats.clientCount {
					t.Errorf(
						"buildWLANStats()[%d].clientCount = %d, want %d",
						wlanID,
						gotStats.clientCount,
						expectedStats.clientCount,
					)
				}
			}
		})
	}
}

func TestNewWLANCollector_InfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		infoLabels []string
		wantLabels []string
	}{
		{
			"No info labels",
			nil,
			[]string{"id"},
		},
		{
			"Custom info labels",
			[]string{"name"},
			[]string{"id", "name"},
		},
		{
			"Info labels with required label included",
			[]string{"id", "name"},
			[]string{"id", "name"},
		},
		{
			"Invalid labels filtered",
			[]string{"name", "invalid_label"},
			[]string{"id", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			metrics := WLANMetrics{
				Info:       true,
				InfoLabels: tt.infoLabels,
			}
			collector := NewWLANCollector(nil, nil, metrics)

			if collector.infoDesc == nil {
				t.Error("NewWLANCollector() info descriptor should not be nil when Info is enabled")
				return
			}

			if len(collector.infoLabelNames) != len(tt.wantLabels) {
				t.Errorf(
					"NewWLANCollector() infoLabelNames length = %d, want %d",
					len(collector.infoLabelNames),
					len(tt.wantLabels),
				)
				return
			}

			for _, expectedLabel := range tt.wantLabels {
				found := false
				for _, actualLabel := range collector.infoLabelNames {
					if actualLabel == expectedLabel {
						found = true
						break
					}
				}
				if !found {
					t.Errorf(
						"NewWLANCollector() missing expected label %q in infoLabelNames %v",
						expectedLabel,
						collector.infoLabelNames,
					)
				}
			}
		})
	}
}

func TestWLANCollector_Integration(t *testing.T) {
	t.Parallel()

	metrics := WLANMetrics{
		General:    true,
		Traffic:    true,
		Config:     true,
		Info:       true,
		InfoLabels: []string{"name"},
	}

	collector := NewWLANCollector(nil, nil, metrics)
	if collector == nil {
		t.Fatal("NewWLANCollector() returned nil")
	}

	var _ prometheus.Collector = collector

	ch := make(chan *prometheus.Desc, 50)
	go func() {
		defer close(ch)
		collector.Describe(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	if count == 0 {
		t.Error("Collector did not emit any descriptors")
	}

	expectedDescs := 22
	if count != expectedDescs {
		t.Errorf("Collector emitted %d descriptors, want %d", count, expectedDescs)
	}
}

func TestWLANCollector_Collect_EarlyReturn(t *testing.T) {
	t.Parallel()

	collector := &WLANCollector{
		metrics: WLANMetrics{},
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.Collect(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("Collect() emitted %d metrics when all flags disabled, want 0", count)
	}
}

func TestWLANStats_Type(t *testing.T) {
	t.Parallel()

	stats := wlanStats{clientCount: 10}

	if stats.clientCount != 10 {
		t.Errorf("wlanStats.clientCount = %d, want 10", stats.clientCount)
	}
}

func TestWLANCollector_collectInfoMetrics_LabelValues(t *testing.T) {
	t.Parallel()

	entry := wlan.WlanCfgEntry{
		WlanID: 1,
		APFVapIDData: &wlan.APFVapIDData{
			SSID: "GuestNetwork",
		},
	}

	metrics := WLANMetrics{
		Info:       true,
		InfoLabels: []string{"id", "name"},
	}

	collector := NewWLANCollector(nil, nil, metrics)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectInfoMetrics(ch, entry)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 1 {
		t.Errorf("collectInfoMetrics() emitted %d metrics, want 1", metricCount)
	}
}

func TestWLANCollector_collectGeneralMetrics(t *testing.T) {
	t.Parallel()

	entry := wlan.WlanCfgEntry{
		WlanID: 1,
		APFVapIDData: &wlan.APFVapIDData{
			SSID: "TestWLAN",
		},
	}

	collector := &WLANCollector{
		metrics:     WLANMetrics{General: true},
		enabledDesc: prometheus.NewDesc("test_enabled", "test", []string{"id"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectGeneralMetrics(ch, entry)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectGeneralMetrics() emitted 0 metrics, want > 0")
	}
}

func TestWLANCollector_collectTrafficMetrics(t *testing.T) {
	t.Parallel()

	entry := wlan.WlanCfgEntry{
		WlanID: 1,
		APFVapIDData: &wlan.APFVapIDData{
			SSID: "TestWLAN",
		},
	}

	statsMap := map[int]wlanStats{
		1: {clientCount: 10},
	}

	usageMap := map[int]uint64{
		1: 6884480,
	}

	onboardMap := map[int]ap.WlanClientStats{
		1: {WlanID: 1, ClientCurrStateL2Auth: 1},
	}

	collector := &WLANCollector{
		metrics:         WLANMetrics{Traffic: true},
		clientCountDesc: prometheus.NewDesc("test_client_count", "test", []string{"id"}, nil),
		dataUsageDesc:   prometheus.NewDesc("test_data_usage", "test", []string{"id"}, nil),
		onboardingDesc: prometheus.NewDesc(
			"test_onboarding", "test", []string{"id", "phase"}, nil,
		),
	}

	// The three series come from two data types, so each polarity has to be exercised
	// on its own: one map nil while another is populated is the case a shared early
	// return would get wrong. The byte counter and the phase counts share one fetch but
	// not one guard, because an unparsable byte leaf withholds only the counter.
	tests := []struct {
		name       string
		statsMap   map[int]wlanStats
		usageMap   map[int]uint64
		onboardMap map[int]ap.WlanClientStats
		want       int
		reason     string
	}{
		{
			name:       "Both data types available",
			statsMap:   statsMap,
			usageMap:   usageMap,
			onboardMap: onboardMap,
			want:       6,
			reason:     "the client count, the byte counter and one series per onboarding phase",
		},
		{
			name:       "Client data unavailable",
			statsMap:   nil,
			usageMap:   usageMap,
			onboardMap: onboardMap,
			want:       5,
			reason:     "a zero client count reads as an SSID with nobody on it, the rest still publishes",
		},
		{
			name:     "WLAN statistics unavailable",
			statsMap: statsMap,
			usageMap: nil,
			want:     1,
			reason:   "a zero byte counter reads as a reset, the client count still publishes",
		},
		{
			name:     "Neither data type available",
			statsMap: nil,
			usageMap: nil,
			want:     0,
			reason:   "nothing is published rather than fabricated zeros",
		},
		{
			name:       "WLAN missing from the statistics list",
			statsMap:   statsMap,
			usageMap:   map[int]uint64{99: 1},
			onboardMap: map[int]ap.WlanClientStats{99: {WlanID: 99}},
			want:       1,
			reason:     "both statistics series are absent for a WLAN the controller lists no record for",
		},
		{
			name:       "Byte leaf unparsable on a WLAN the controller does list",
			statsMap:   statsMap,
			usageMap:   nil,
			onboardMap: onboardMap,
			want:       5,
			reason:     "the phase counts survive a byte leaf the parse rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ch := make(chan prometheus.Metric, 10)
			go func() {
				defer close(ch)
				collector.collectTrafficMetrics(ch, entry, tt.statsMap, tt.usageMap, tt.onboardMap)
			}()

			metricCount := 0
			for range ch {
				metricCount++
			}

			if metricCount != tt.want {
				t.Errorf("collectTrafficMetrics() emitted %d metrics, want %d: %s",
					metricCount, tt.want, tt.reason)
			}
		})
	}
}

// TestWLANCollector_ConfigBooleansMatchLeaves binds each config boolean series to
// its own leaf. The ten series share one value domain, so a swapped descriptor, or a
// leaf swapped where the emit list reads it, keeps every count, every label and every
// value in range: only a fixture that raises one leaf at a time separates them, and
// only the registry route shows which descriptor carried the value.
//
// Raising one leaf at a time need not describe a WLAN anyone would deploy. The
// subject here is the wiring, not a deployable configuration.
// TestWLANCollector_ConfigStatesMatchLeaves pins each state series to the leaf it
// reads. Both leaves are value-typed strings, so swapping the two descriptors is a
// change no compiler and no count assertion catches — only the label value does.
// The empty case is asserted too, because a state label carrying an empty string
// reads as no label at all.
func TestWLANCollector_ConfigStatesMatchLeaves(t *testing.T) {
	t.Parallel()

	// stateOf returns the state label of the single series in each named family,
	// and reports whether that family carried a gauge at value 1 at all.
	stateOf := func(t *testing.T, mutate func(*wnc.WNCDataCache)) map[string]string {
		t.Helper()

		data := fullFixtureSnapshot()
		mutate(data)

		registry := prometheus.NewRegistry()
		registry.MustRegister(NewWLANCollector(
			wnc.NewWLANSource(fixtureSource{data: data}),
			wnc.NewClientSource(fixtureSource{data: data}),
			WLANMetrics{Config: true},
		))

		families, err := registry.Gather()
		if err != nil {
			t.Fatalf("Gather() error = %v, want nil", err)
		}

		states := make(map[string]string, len(families))
		for _, family := range families {
			metrics := family.GetMetric()
			if len(metrics) != 1 {
				continue
			}
			gauge := metrics[0].GetGauge()
			if gauge == nil {
				t.Errorf("%s is not a gauge", family.GetName())
				continue
			}
			if got := gauge.GetValue(); got != 1 {
				states[family.GetName()] = "value=" + strconv.FormatFloat(got, 'f', -1, 64)
				continue
			}
			for _, label := range metrics[0].GetLabel() {
				if label.GetName() == labelState {
					states[family.GetName()] = label.GetValue()
				}
			}
		}
		return states
	}

	t.Run("Each series carries its own leaf", func(t *testing.T) {
		t.Parallel()

		states := stateOf(t, func(*wnc.WNCDataCache) {})

		for name, want := range map[string]string{
			"wnc_wlan_pmf_state": fixturePMFOptions,
			"wnc_wlan_ft_state":  fixtureFTMode,
		} {
			got, ok := states[name]
			if !ok {
				t.Fatalf("%s carries no gauge at 1 with a state label, so its leaf is unpinned", name)
			}
			if got != want {
				t.Errorf("%s state = %q, want %q", name, got, want)
			}
		}
	})

	// Emptying one leaf while the other carries a value is what separates a per-leaf
	// skip from a loop that abandons the rest, and dropping the policy bindings is what
	// separates a leaf on the WLAN entry from one on the policy profile. Emptying both
	// at once, with a policy resolved, distinguishes neither.
	t.Run("One empty leaf leaves the other published", func(t *testing.T) {
		t.Parallel()

		states := stateOf(t, func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].PMFOptions = ""
			d.WLANPolicyListEntries = nil
		})

		if got, ok := states["wnc_wlan_pmf_state"]; ok {
			t.Errorf("wnc_wlan_pmf_state emitted state %q for an empty leaf, want no series", got)
		}
		if got := states["wnc_wlan_ft_state"]; got != fixtureFTMode {
			t.Errorf("wnc_wlan_ft_state state = %q, want %q — it reads the WLAN entry, "+
				"so neither the sibling leaf nor the policy binding gates it", got, fixtureFTMode)
		}
	})

	t.Run("An empty leaf publishes no series", func(t *testing.T) {
		t.Parallel()

		states := stateOf(t, func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].PMFOptions = ""
			d.WLANConfigEntries[0].FTMode = ""
		})

		for _, name := range []string{"wnc_wlan_pmf_state", "wnc_wlan_ft_state"} {
			if got, ok := states[name]; ok {
				t.Errorf("%s emitted state %q for an empty leaf, want no series", name, got)
			}
		}
	})
}

func TestWLANCollector_ConfigBooleansMatchLeaves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		raise func(*wnc.WNCDataCache)
	}{
		{"wnc_wlan_auth_psk_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].AuthKeyMgmtPsk = true
		}},
		{"wnc_wlan_auth_dot1x_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].AuthKeyMgmtDot1x = true
		}},
		{"wnc_wlan_auth_dot1x_sha256_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].AuthKeyMgmtDot1xSha256 = true
		}},
		{"wnc_wlan_wpa2_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].WPA2Enabled = true
		}},
		{"wnc_wlan_wpa3_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].WPA3Enabled = true
		}},
		{"wnc_wlan_load_balance_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].LoadBalance = true
		}},
		{"wnc_wlan_11k_neighbor_list_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].Wlan11kNeighList = true
		}},
		{"wnc_wlan_client_steering_enabled", func(d *wnc.WNCDataCache) {
			d.WLANConfigEntries[0].ClientSteering = true
		}},
		{"wnc_wlan_central_switching_enabled", func(d *wnc.WNCDataCache) {
			d.WLANPolicies[0].WlanSwitchingPolicy.CentralSwitching = true
		}},
		{"wnc_wlan_central_authentication_enabled", func(d *wnc.WNCDataCache) {
			d.WLANPolicies[0].WlanSwitchingPolicy.CentralAuthentication = true
		}},
		{"wnc_wlan_central_dhcp_enabled", func(d *wnc.WNCDataCache) {
			d.WLANPolicies[0].WlanSwitchingPolicy.CentralDHCP = true
		}},
		{"wnc_wlan_central_association_enabled", func(d *wnc.WNCDataCache) {
			d.WLANPolicies[0].WlanSwitchingPolicy.CentralAssocEnable = true
		}},
		{"wnc_wlan_policy_enabled", func(d *wnc.WNCDataCache) {
			d.WLANPolicies[0].Status = true
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			// The shared fixture raises three of these leaves, so the one-hot pattern
			// has to lower all three. No unread boolean is left on the entry to raise
			// as a decoy, so a descriptor folded onto one is caught by the pairwise
			// zeros this table asserts rather than by a decoy.
			data.WLANConfigEntries[0].AuthKeyMgmtPsk = false
			data.WLANConfigEntries[0].WPA2Enabled = false
			data.WLANPolicies[0].WlanSwitchingPolicy.CentralSwitching = false
			tt.raise(data)
			src := fixtureSource{data: data}

			registry := prometheus.NewRegistry()
			registry.MustRegister(NewWLANCollector(
				wnc.NewWLANSource(src), wnc.NewClientSource(src), WLANMetrics{Config: true},
			))

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}

			values := make(map[string]float64, len(families))
			for _, family := range families {
				if metrics := family.GetMetric(); len(metrics) == 1 {
					values[family.GetName()] = metrics[0].GetGauge().GetValue()
				}
			}

			for _, other := range tests {
				want := 0.0
				if other.name == tt.name {
					want = 1.0
				}
				got, ok := values[other.name]
				if !ok {
					t.Fatalf("%s carries no single series, so the raised leaf proves nothing", other.name)
				}
				if got != want {
					t.Errorf("%s = %v, want %v while only %s is raised", other.name, got, want, tt.name)
				}
			}
		})
	}
}

func TestWLANCollector_collectMetrics_NilSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "collectGeneralMetrics with minimal entry",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &WLANCollector{
					metrics:     WLANMetrics{General: true},
					enabledDesc: prometheus.NewDesc("test", "test", []string{"id"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectGeneralMetrics() panicked with minimal entry: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				entry := wlan.WlanCfgEntry{WlanID: 1}
				collector.collectGeneralMetrics(ch, entry)
			},
		},
		{
			name: "collectTrafficMetrics with empty maps",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &WLANCollector{
					metrics:         WLANMetrics{Traffic: true},
					clientCountDesc: prometheus.NewDesc("test", "test", []string{"id"}, nil),
					dataUsageDesc:   prometheus.NewDesc("test_usage", "test", []string{"id"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectTrafficMetrics() panicked with empty statsMap: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				entry := wlan.WlanCfgEntry{WlanID: 1}
				collector.collectTrafficMetrics(
					ch, entry, map[int]wlanStats{}, map[int]uint64{}, map[int]ap.WlanClientStats{},
				)
			},
		},
		{
			name: "collectConfigMetrics with empty policyMap",
			testFunc: func(t *testing.T) {
				t.Helper()
				collector := &WLANCollector{
					metrics:                   WLANMetrics{Config: true},
					authPskDesc:               prometheus.NewDesc("test", "test", []string{"id"}, nil),
					authDot1xDesc:             prometheus.NewDesc("test", "test", []string{"id"}, nil),
					authDot1xSha256Desc:       prometheus.NewDesc("test", "test", []string{"id"}, nil),
					wpa2EnabledDesc:           prometheus.NewDesc("test", "test", []string{"id"}, nil),
					wpa3EnabledDesc:           prometheus.NewDesc("test", "test", []string{"id"}, nil),
					sessionTimeoutDesc:        prometheus.NewDesc("test", "test", []string{"id"}, nil),
					loadBalanceDesc:           prometheus.NewDesc("test", "test", []string{"id"}, nil),
					wlan11kNeighDesc:          prometheus.NewDesc("test", "test", []string{"id"}, nil),
					clientSteeringDesc:        prometheus.NewDesc("test", "test", []string{"id"}, nil),
					centralSwitchingDesc:      prometheus.NewDesc("test", "test", []string{"id"}, nil),
					centralAuthenticationDesc: prometheus.NewDesc("test", "test", []string{"id"}, nil),
					centralDHCPDesc:           prometheus.NewDesc("test", "test", []string{"id"}, nil),
					centralAssocEnableDesc:    prometheus.NewDesc("test", "test", []string{"id"}, nil),
					policyEnabledDesc:         prometheus.NewDesc("test", "test", []string{"id"}, nil),
					pmfStateDesc:              prometheus.NewDesc("pmf", "pmf", []string{"id", "state"}, nil),
					ftStateDesc:               prometheus.NewDesc("ft", "ft", []string{"id", "state"}, nil),
				}
				ch := make(chan prometheus.Metric, 20)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectConfigMetrics() panicked with empty policyMap: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				entry := wlan.WlanCfgEntry{
					WlanID:      1,
					ProfileName: "unknown",
					PMFOptions:  fixturePMFOptions,
					FTMode:      fixtureFTMode,
				}
				collector.collectConfigMetrics(ch, entry, map[string]*wlan.WlanPolicy{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

// TestWLANCollector_EnabledReportsBothLowCauses pins both polarities of
// wnc_wlan_enabled and the fate of an absent container. The status lives in an
// optional container, so a guard that only checks the flag panics when the
// container is absent, one that only checks the container reports one for a
// disabled WLAN, and publishing zero for an absent container asserts that the
// operator disabled the WLAN.
func TestWLANCollector_EnabledReportsBothLowCauses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		vapIDData  *wlan.APFVapIDData
		wantSeries bool
		want       float64
	}{
		{"status flag set", &wlan.APFVapIDData{SSID: "TestWLAN", WlanStatus: true}, true, 1},
		{"status flag clear", &wlan.APFVapIDData{SSID: "TestWLAN", WlanStatus: false}, true, 0},
		{"container absent", nil, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.WLANConfigEntries[0].APFVapIDData = tt.vapIDData
			src := fixtureSource{data: data}

			registry := prometheus.NewRegistry()
			registry.MustRegister(NewWLANCollector(
				wnc.NewWLANSource(src), wnc.NewClientSource(src), WLANMetrics{General: true},
			))

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}

			var got float64
			found := false
			for _, family := range families {
				if family.GetName() != "wnc_wlan_enabled" {
					continue
				}
				got = family.GetMetric()[0].GetGauge().GetValue()
				found = true
			}
			if found != tt.wantSeries {
				t.Fatalf("wnc_wlan_enabled has a series = %v, want %v", found, tt.wantSeries)
			}
			if found && got != tt.want {
				t.Errorf("wnc_wlan_enabled = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWLANCollector_ConfigOmitsSeriesWhenContainerAbsent pins the contract that a
// policy-derived series is not published when the controller omits the container
// its leaf lives in. Zero is a legal value for all five, so a fabricated zero is
// indistinguishable from a measured one once it reaches Prometheus. The rows with
// zero leaves inside a present container pin the other half of the contract, where
// an omitted leaf and a configured zero decode alike and the series stays present.
// The entry-derived series are the control. Their leaves sit on the entry itself,
// so they are published either way.
func TestWLANCollector_ConfigOmitsSeriesWhenContainerAbsent(t *testing.T) {
	t.Parallel()

	timeoutDerived := []string{"wnc_wlan_session_timeout_seconds"}
	switchingDerived := []string{
		"wnc_wlan_central_switching_enabled",
		"wnc_wlan_central_authentication_enabled",
		"wnc_wlan_central_dhcp_enabled",
		"wnc_wlan_central_association_enabled",
	}
	entryDerived := []string{
		"wnc_wlan_auth_psk_enabled",
		"wnc_wlan_auth_dot1x_enabled",
		"wnc_wlan_auth_dot1x_sha256_enabled",
		"wnc_wlan_wpa3_enabled",
		"wnc_wlan_load_balance_enabled",
		"wnc_wlan_client_steering_enabled",
	}

	timeout := &wlan.WlanTimeout{SessionTimeout: 1800}
	switching := &wlan.WlanSwitchingPolicy{CentralSwitching: true}

	tests := []struct {
		name      string
		timeout   *wlan.WlanTimeout
		switching *wlan.WlanSwitchingPolicy
		absent    []string
	}{
		{"both containers present", timeout, switching, nil},
		{"both containers present with zero leaves", &wlan.WlanTimeout{}, &wlan.WlanSwitchingPolicy{}, nil},
		{"timeout container absent", nil, switching, timeoutDerived},
		{"switching container absent", timeout, nil, switchingDerived},
		{"both containers absent", nil, nil, slices.Concat(timeoutDerived, switchingDerived)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fullFixtureSnapshot()
			data.WLANPolicies[0].WlanTimeout = tt.timeout
			data.WLANPolicies[0].WlanSwitchingPolicy = tt.switching
			src := fixtureSource{data: data}

			registry := prometheus.NewRegistry()
			registry.MustRegister(NewWLANCollector(
				wnc.NewWLANSource(src), wnc.NewClientSource(src), WLANMetrics{Config: true},
			))

			families, err := registry.Gather()
			if err != nil {
				t.Fatalf("Gather() error = %v, want nil", err)
			}

			present := make(map[string]bool, len(families))
			for _, family := range families {
				present[family.GetName()] = len(family.GetMetric()) > 0
			}

			for _, name := range slices.Concat(entryDerived, timeoutDerived, switchingDerived) {
				want := !slices.Contains(tt.absent, name)
				if present[name] != want {
					t.Errorf("%s has a series = %v, want %v", name, present[name], want)
				}
			}
		})
	}
}

// TestBuildWLANDataUsageMap_SkipsAnUnreadableLeaf keeps a fabricated zero out of a
// counter. The leaf is a string on the wire, and the shared conversion reads both an
// omitted leaf and an unparsable one as zero, which on a counter Prometheus reads as a
// reset and then extrapolates from.
func TestBuildWLANDataUsageMap_SkipsAnUnreadableLeaf(t *testing.T) {
	t.Parallel()

	usage := buildWLANDataUsageMap([]ap.WlanClientStats{
		{WlanID: 1, DataUsage: "6884480"},
		{WlanID: 2, DataUsage: ""},
		{WlanID: 3, DataUsage: "not-a-number"},
		{WlanID: 4, DataUsage: "-1"},
		// The controller reports this counter as a string precisely because it exceeds
		// what a JSON number is safe for, so the width has to survive the parse.
		{WlanID: 5, DataUsage: "18446744073709551615"},
	})

	if got, ok := usage[1]; !ok || got != 6884480 {
		t.Errorf("usage[1] = %v (present %v), want 6884480", got, ok)
	}
	if got, ok := usage[5]; !ok || got != 18446744073709551615 {
		t.Errorf("usage[5] = %v (present %v), want the full 64-bit value", got, ok)
	}
	for _, id := range []int{2, 3, 4} {
		if got, ok := usage[id]; ok {
			t.Errorf("usage[%d] = %v, want the WLAN skipped rather than read as zero", id, got)
		}
	}
}

// TestWLANCollector_PolicyBindingIsOneSeriesPerResolvableBinding pins the series that
// makes an ambiguous binding observable. The six policy series name neither the tag nor
// the profile, so without this one an operator cannot tell which of several bound
// profiles they are reading.
func TestWLANCollector_PolicyBindingIsOneSeriesPerResolvableBinding(t *testing.T) {
	t.Parallel()

	data := fullFixtureSnapshot()

	// A second tag binding the same WLAN to a second profile: the state the series
	// exists to expose.
	data.WLANPolicies = append(data.WLANPolicies, wlan.WlanPolicy{PolicyProfileName: "second-policy"})
	data.WLANPolicyListEntries = append(data.WLANPolicyListEntries,
		wlan.PolicyListEntry{
			TagName: "second-tag",
			WLANPolicies: &wlan.WLANPolicies{WLANPolicy: []wlan.WLANPolicyMap{
				{WLANProfileName: fixtureProfile, PolicyProfileName: "second-policy"},
				// A tag naming a WLAN the controller does not define: real on a controller,
				// and it has no identifier to key a series by.
				{WLANProfileName: "undefined-wlan", PolicyProfileName: fixturePolicy},
				// A binding naming a policy profile the controller did not return, which the
				// six policy series skip as well.
				{WLANProfileName: fixtureProfile, PolicyProfileName: "unresolvable-policy"},
			}},
		},
		// A tag carrying no bindings at all.
		wlan.PolicyListEntry{TagName: "empty-tag"},
	)

	src := fixtureSource{data: data}
	registry := prometheus.NewRegistry()
	registry.MustRegister(NewWLANCollector(
		wnc.NewWLANSource(src), wnc.NewClientSource(src), WLANMetrics{Config: true},
	))

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	got := make(map[string]bool)
	for _, family := range families {
		if family.GetName() != "wnc_wlan_policy_binding" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, pair := range metric.GetLabel() {
				labels[pair.GetName()] = pair.GetValue()
			}
			got[labels[labelID]+"|"+labels[labelPolicyProfile]+"|"+labels[labelPolicyTag]] = true

			if metric.GetGauge().GetValue() != 1 {
				t.Errorf("wnc_wlan_policy_binding = %v, want 1", metric.GetGauge().GetValue())
			}
		}
	}

	want := map[string]bool{
		"1|" + fixturePolicy + "|test-tag": true,
		"1|second-policy|second-tag":       true,
	}
	if len(got) != len(want) {
		t.Errorf("wnc_wlan_policy_binding has %d series, want %d: %v", len(got), len(want), got)
	}
	for key := range want {
		if !got[key] {
			t.Errorf("wnc_wlan_policy_binding is missing the series %q", key)
		}
	}
}

// TestWLANCollector_OnboardingPhasesMatchLeaves binds each phase label value to its
// own leaf. The four counts share one descriptor and one value domain, so a pair
// swapped in onboardingPhases keeps the series count, the label names and the label
// values intact and changes only which number each phase reports.
func TestWLANCollector_OnboardingPhasesMatchLeaves(t *testing.T) {
	t.Parallel()

	families, err := fixtureRegistry(t, "").Gather()
	if err != nil {
		t.Fatalf("Gather() error = %v, want nil", err)
	}

	got := make(map[string]float64)
	for _, family := range families {
		if family.GetName() != "wnc_wlan_onboarding_clients" {
			continue
		}
		for _, metric := range family.GetMetric() {
			phase := ""
			for _, pair := range metric.GetLabel() {
				if pair.GetName() == labelPhase {
					phase = pair.GetValue()
				}
			}
			got[phase] = metric.GetGauge().GetValue()
		}
	}

	// The fixture record carries a distinct number per leaf, listed in the order the
	// container declares them.
	want := map[string]float64{
		"l2auth":          7103,
		"mobility":        7104,
		"iplearn":         7105,
		"webauth_pending": 7106,
	}

	if len(got) != len(want) {
		t.Errorf("wnc_wlan_onboarding_clients has %d series, want %d: %v", len(got), len(want), got)
	}
	for phase, value := range want {
		if got[phase] != value {
			t.Errorf("wnc_wlan_onboarding_clients{phase=%q} = %v, want %v", phase, got[phase], value)
		}
	}
}
