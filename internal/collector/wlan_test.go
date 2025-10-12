package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
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
			3, // client_count, rx_bytes, tx_bytes
		},
		{
			"Config module only",
			WLANMetrics{Config: true},
			13, // auth_psk, auth_dot1x, auth_dot1x_sha256, wpa2, wpa3, session_timeout, load_balance, 11k, steering, central_switching, central_auth, central_dhcp, central_assoc
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
			18, // 1+3+13+1
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
		{collector.clientCountDesc, "wnc_wlan_clients_total"},
		{collector.bytesRxDesc, "wnc_wlan_rx_bytes_total"},
		{collector.bytesTxDesc, "wnc_wlan_tx_bytes_total"},
		{collector.authPskDesc, "wnc_wlan_auth_psk_enabled"},
		{collector.authDot1xDesc, "wnc_wlan_auth_dot1x_enabled"},
		{collector.wpa2EnabledDesc, "wnc_wlan_wpa2_enabled"},
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
		name         string
		clientData   []client.CommonOperData
		trafficStats []client.TrafficStats
		expected     map[int]wlanStats
	}{
		{
			"Empty data",
			[]client.CommonOperData{},
			[]client.TrafficStats{},
			map[int]wlanStats{},
		},
		{
			"Nil data",
			nil,
			nil,
			map[int]wlanStats{},
		},
		{
			"Single client on WLAN 1",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
			},
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
			},
			map[int]wlanStats{
				1: {clientCount: 1, bytesRx: 1000, bytesTx: 2000},
			},
		},
		{
			"Multiple clients on same WLAN",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 1, CoState: ClientStatusRun},
			},
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
				{MsMACAddress: "11:22:33:44:55:66", BytesRx: "1500", BytesTx: "2500"},
			},
			map[int]wlanStats{
				1: {clientCount: 2, bytesRx: 2500, bytesTx: 4500},
			},
		},
		{
			"Clients on different WLANs",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 2, CoState: ClientStatusRun},
			},
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
				{MsMACAddress: "11:22:33:44:55:66", BytesRx: "1500", BytesTx: "2500"},
			},
			map[int]wlanStats{
				1: {clientCount: 1, bytesRx: 1000, bytesTx: 2000},
				2: {clientCount: 1, bytesRx: 1500, bytesTx: 2500},
			},
		},
		{
			"Skip non-RUN clients",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
				{ClientMAC: "11:22:33:44:55:66", WlanID: 1, CoState: "IDLE"},
			},
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
				{MsMACAddress: "11:22:33:44:55:66", BytesRx: "1500", BytesTx: "2500"},
			},
			map[int]wlanStats{
				1: {clientCount: 1, bytesRx: 1000, bytesTx: 2000},
			},
		},
		{
			"Client without traffic stats",
			[]client.CommonOperData{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", WlanID: 1, CoState: ClientStatusRun},
			},
			[]client.TrafficStats{},
			map[int]wlanStats{
				1: {clientCount: 1, bytesRx: 0, bytesTx: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &WLANCollector{}
			got := collector.buildWLANStats(tt.clientData, tt.trafficStats)

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
				} else {
					if gotStats.clientCount != expectedStats.clientCount {
						t.Errorf(
							"buildWLANStats()[%d].clientCount = %d, want %d",
							wlanID,
							gotStats.clientCount,
							expectedStats.clientCount,
						)
					}
					if gotStats.bytesRx != expectedStats.bytesRx {
						t.Errorf(
							"buildWLANStats()[%d].bytesRx = %d, want %d",
							wlanID,
							gotStats.bytesRx,
							expectedStats.bytesRx,
						)
					}
					if gotStats.bytesTx != expectedStats.bytesTx {
						t.Errorf(
							"buildWLANStats()[%d].bytesTx = %d, want %d",
							wlanID,
							gotStats.bytesTx,
							expectedStats.bytesTx,
						)
					}
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

	expectedDescs := 18
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

	stats := wlanStats{
		clientCount: 10,
		bytesRx:     1000,
		bytesTx:     2000,
		packetsRx:   100,
		packetsTx:   200,
	}

	if stats.clientCount != 10 {
		t.Errorf("wlanStats.clientCount = %d, want 10", stats.clientCount)
	}
	if stats.bytesRx != 1000 {
		t.Errorf("wlanStats.bytesRx = %d, want 1000", stats.bytesRx)
	}
	if stats.bytesTx != 2000 {
		t.Errorf("wlanStats.bytesTx = %d, want 2000", stats.bytesTx)
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

// TestWLANCollector_collectGeneralMetrics tests basic metric emission
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

// TestWLANCollector_collectTrafficMetrics tests basic metric emission
func TestWLANCollector_collectTrafficMetrics(t *testing.T) {
	t.Parallel()

	entry := wlan.WlanCfgEntry{
		WlanID: 1,
		APFVapIDData: &wlan.APFVapIDData{
			SSID: "TestWLAN",
		},
	}

	statsMap := map[int]wlanStats{
		1: {
			clientCount: 10,
			bytesRx:     1000000,
			bytesTx:     2000000,
			packetsRx:   10000,
			packetsTx:   20000,
		},
	}

	collector := &WLANCollector{
		metrics:         WLANMetrics{Traffic: true},
		clientCountDesc: prometheus.NewDesc("test_client_count", "test", []string{"id"}, nil),
		bytesRxDesc:     prometheus.NewDesc("test_bytes_rx", "test", []string{"id"}, nil),
		bytesTxDesc:     prometheus.NewDesc("test_bytes_tx", "test", []string{"id"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectTrafficMetrics(ch, entry, statsMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectTrafficMetrics() emitted 0 metrics, want > 0")
	}
}

// TestWLANCollector_collectConfigMetrics tests basic metric emission
func TestWLANCollector_collectConfigMetrics(t *testing.T) {
	t.Parallel()

	entry := wlan.WlanCfgEntry{
		WlanID:                 1,
		ProfileName:            "test-profile",
		AuthKeyMgmtPsk:         true,
		AuthKeyMgmtDot1x:       false,
		AuthKeyMgmtDot1xSha256: false,
		WPA2Enabled:            true,
		WPA3Enabled:            false,
		LoadBalance:            true,
		Wlan11kNeighList:       true,
		ClientSteering:         true,
		APFVapIDData: &wlan.APFVapIDData{
			SSID: "TestWLAN",
		},
	}

	policyMap := map[string]*wlan.WlanPolicy{
		"test-profile": {},
	}

	collector := &WLANCollector{
		metrics:                   WLANMetrics{Config: true},
		authPskDesc:               prometheus.NewDesc("test_auth_psk", "test", []string{"id"}, nil),
		authDot1xDesc:             prometheus.NewDesc("test_auth_dot1x", "test", []string{"id"}, nil),
		authDot1xSha256Desc:       prometheus.NewDesc("test_auth_dot1x_sha256", "test", []string{"id"}, nil),
		wpa2EnabledDesc:           prometheus.NewDesc("test_wpa2_enabled", "test", []string{"id"}, nil),
		wpa3EnabledDesc:           prometheus.NewDesc("test_wpa3_enabled", "test", []string{"id"}, nil),
		sessionTimeoutDesc:        prometheus.NewDesc("test_session_timeout", "test", []string{"id"}, nil),
		loadBalanceDesc:           prometheus.NewDesc("test_load_balance", "test", []string{"id"}, nil),
		wlan11kNeighDesc:          prometheus.NewDesc("test_11k_neigh", "test", []string{"id"}, nil),
		clientSteeringDesc:        prometheus.NewDesc("test_client_steering", "test", []string{"id"}, nil),
		centralSwitchingDesc:      prometheus.NewDesc("test_central_switching", "test", []string{"id"}, nil),
		centralAuthenticationDesc: prometheus.NewDesc("test_central_auth", "test", []string{"id"}, nil),
		centralDHCPDesc:           prometheus.NewDesc("test_central_dhcp", "test", []string{"id"}, nil),
		centralAssocEnableDesc:    prometheus.NewDesc("test_central_assoc", "test", []string{"id"}, nil),
	}

	ch := make(chan prometheus.Metric, 20)
	go func() {
		defer close(ch)
		collector.collectConfigMetrics(ch, entry, policyMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectConfigMetrics() emitted 0 metrics, want > 0")
	}
}

// TestWLANCollector_collectMetrics_NilSafety tests nil safety
func TestWLANCollector_collectMetrics_NilSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "collectGeneralMetrics with minimal entry",
			testFunc: func(t *testing.T) {
				t.Parallel()
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
			name: "collectTrafficMetrics with empty statsMap",
			testFunc: func(t *testing.T) {
				t.Parallel()
				collector := &WLANCollector{
					metrics:         WLANMetrics{Traffic: true},
					clientCountDesc: prometheus.NewDesc("test", "test", []string{"id"}, nil),
					bytesRxDesc:     prometheus.NewDesc("test", "test", []string{"id"}, nil),
					bytesTxDesc:     prometheus.NewDesc("test", "test", []string{"id"}, nil),
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
				collector.collectTrafficMetrics(ch, entry, map[int]wlanStats{})
			},
		},
		{
			name: "collectConfigMetrics with empty policyMap",
			testFunc: func(t *testing.T) {
				t.Parallel()
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
				entry := wlan.WlanCfgEntry{WlanID: 1, ProfileName: "unknown"}
				collector.collectConfigMetrics(ch, entry, map[string]*wlan.WlanPolicy{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
