package collector

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
)

func TestNewClientCollector(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		metrics ClientMetrics
		wantNil bool
	}{
		{
			"All modules disabled",
			ClientMetrics{},
			false,
		},
		{
			"General module enabled",
			ClientMetrics{General: true},
			false,
		},
		{
			"Radio module enabled",
			ClientMetrics{Radio: true},
			false,
		},
		{
			"Traffic module enabled",
			ClientMetrics{Traffic: true},
			false,
		},
		{
			"Errors module enabled",
			ClientMetrics{Errors: true},
			false,
		},
		{
			"Info module enabled",
			ClientMetrics{Info: true},
			false,
		},
		{
			"All modules enabled",
			ClientMetrics{
				General: true,
				Radio:   true,
				Traffic: true,
				Errors:  true,
				Info:    true,
			},
			false,
		},
		{
			"Info module with labels",
			ClientMetrics{
				Info:       true,
				InfoLabels: []string{"ap", "band", "wlan", "name", "username", "ipv4", "ipv6"},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewClientCollector(nil, tt.metrics)

			if tt.wantNil {
				if collector != nil {
					t.Errorf("NewClientCollector() = %v, want nil", collector)
				}
			} else {
				if collector == nil {
					t.Errorf("NewClientCollector() = nil, want non-nil")
					return
				}
				if collector.metrics.General != tt.metrics.General {
					t.Errorf(
						"NewClientCollector() General = %v, want %v",
						collector.metrics.General,
						tt.metrics.General,
					)
				}
				if collector.metrics.Info != tt.metrics.Info {
					t.Errorf("NewClientCollector() Info = %v, want %v", collector.metrics.Info, tt.metrics.Info)
				}
			}
		})
	}
}

func TestClientCollector_isAnyMetricFlagEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metrics  ClientMetrics
		expected bool
	}{
		{
			"All disabled",
			ClientMetrics{},
			false,
		},
		{
			"General enabled",
			ClientMetrics{General: true},
			true,
		},
		{
			"Radio enabled",
			ClientMetrics{Radio: true},
			true,
		},
		{
			"Traffic enabled",
			ClientMetrics{Traffic: true},
			true,
		},
		{
			"Errors enabled",
			ClientMetrics{Errors: true},
			true,
		},
		{
			"Info enabled",
			ClientMetrics{Info: true},
			true,
		},
		{
			"Multiple enabled",
			ClientMetrics{General: true, Radio: true},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := &ClientCollector{metrics: tt.metrics}
			if got := collector.isAnyMetricFlagEnabled(); got != tt.expected {
				t.Errorf("isAnyMetricFlagEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClientCollector_Describe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		metrics     ClientMetrics
		expectDescs int
	}{
		{
			"No modules enabled",
			ClientMetrics{},
			0,
		},
		{
			"General module only",
			ClientMetrics{General: true},
			4, // state, uptime, state_transition, power_save_state
		},
		{
			"Radio module only",
			ClientMetrics{Radio: true},
			6, // protocol, mcs_index, spatial_streams, speed, rssi, snr
		},
		{
			"Traffic module only",
			ClientMetrics{Traffic: true},
			4, // rx_bytes, tx_bytes, rx_packets, tx_packets
		},
		{
			"Errors module only",
			ClientMetrics{Errors: true},
			12, // retry_ratio, policy_errors, duplicate_received, decryption_failed, mic_mismatch, mic_missing, excessive_retries, rx_group_counter, tx_drops, data_retries, rts_retries, tx_retries
		},
		{
			"Info module only",
			ClientMetrics{Info: true},
			1, // info metric
		},
		{
			"All modules enabled",
			ClientMetrics{
				General: true,
				Radio:   true,
				Traffic: true,
				Errors:  true,
				Info:    true,
			},
			27, // 4+6+4+12+1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			collector := NewClientCollector(nil, tt.metrics)

			ch := make(chan *prometheus.Desc, 100)
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

func TestBuildDeviceMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []client.DcInfo
		expected map[string]client.DcInfo
	}{
		{
			"Empty slice",
			[]client.DcInfo{},
			map[string]client.DcInfo{},
		},
		{
			"Single device",
			[]client.DcInfo{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
			},
			map[string]client.DcInfo{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
			},
		},
		{
			"Multiple devices",
			[]client.DcInfo{
				{ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
				{ClientMAC: "11:22:33:44:55:66", DeviceName: "MacBook"},
			},
			map[string]client.DcInfo{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
				"11:22:33:44:55:66": {ClientMAC: "11:22:33:44:55:66", DeviceName: "MacBook"},
			},
		},
		{
			"Empty MAC address",
			[]client.DcInfo{
				{ClientMAC: "", DeviceName: "Unknown"},
			},
			map[string]client.DcInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildDeviceMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildDeviceMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac, expectedData := range tt.expected {
				if gotData, exists := got[mac]; !exists {
					t.Errorf("buildDeviceMap() missing MAC %q", mac)
				} else if gotData.ClientMAC != expectedData.ClientMAC || gotData.DeviceName != expectedData.DeviceName {
					t.Errorf("buildDeviceMap()[%q] = %+v, want %+v", mac, gotData, expectedData)
				}
			}
		})
	}
}

func TestBuildDot11Map(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []client.Dot11OperData
		expected map[string]client.Dot11OperData
	}{
		{
			"Empty slice",
			[]client.Dot11OperData{},
			map[string]client.Dot11OperData{},
		},
		{
			"Single entry",
			[]client.Dot11OperData{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", VapSsid: "GuestNetwork"},
			},
			map[string]client.Dot11OperData{
				"aa:bb:cc:dd:ee:ff": {MsMACAddress: "aa:bb:cc:dd:ee:ff", VapSsid: "GuestNetwork"},
			},
		},
		{
			"Multiple entries",
			[]client.Dot11OperData{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", VapSsid: "GuestNetwork"},
				{MsMACAddress: "11:22:33:44:55:66", VapSsid: "CorpNetwork"},
			},
			map[string]client.Dot11OperData{
				"aa:bb:cc:dd:ee:ff": {MsMACAddress: "aa:bb:cc:dd:ee:ff", VapSsid: "GuestNetwork"},
				"11:22:33:44:55:66": {MsMACAddress: "11:22:33:44:55:66", VapSsid: "CorpNetwork"},
			},
		},
		{
			"Empty MAC address",
			[]client.Dot11OperData{
				{MsMACAddress: "", VapSsid: "GuestNetwork"},
			},
			map[string]client.Dot11OperData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildDot11Map(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildDot11Map() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac := range tt.expected {
				if _, exists := got[mac]; !exists {
					t.Errorf("buildDot11Map() missing MAC %q", mac)
				}
			}
		})
	}
}

func TestBuildSISFMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []client.SisfDBMac
		expected map[string]client.SisfDBMac
	}{
		{
			"Empty slice",
			[]client.SisfDBMac{},
			map[string]client.SisfDBMac{},
		},
		{
			"Single entry",
			[]client.SisfDBMac{
				{MACAddr: "aa:bb:cc:dd:ee:ff"},
			},
			map[string]client.SisfDBMac{
				"aa:bb:cc:dd:ee:ff": {MACAddr: "aa:bb:cc:dd:ee:ff"},
			},
		},
		{
			"Multiple entries",
			[]client.SisfDBMac{
				{MACAddr: "aa:bb:cc:dd:ee:ff"},
				{MACAddr: "11:22:33:44:55:66"},
			},
			map[string]client.SisfDBMac{
				"aa:bb:cc:dd:ee:ff": {MACAddr: "aa:bb:cc:dd:ee:ff"},
				"11:22:33:44:55:66": {MACAddr: "11:22:33:44:55:66"},
			},
		},
		{
			"Empty MAC address",
			[]client.SisfDBMac{
				{MACAddr: ""},
			},
			map[string]client.SisfDBMac{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildSISFMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildSISFMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac := range tt.expected {
				if _, exists := got[mac]; !exists {
					t.Errorf("buildSISFMap() missing MAC %q", mac)
				}
			}
		})
	}
}

func TestBuildTrafficMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []client.TrafficStats
		expected map[string]client.TrafficStats
	}{
		{
			"Empty slice",
			[]client.TrafficStats{},
			map[string]client.TrafficStats{},
		},
		{
			"Single entry",
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
			},
			map[string]client.TrafficStats{
				"aa:bb:cc:dd:ee:ff": {MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
			},
		},
		{
			"Multiple entries",
			[]client.TrafficStats{
				{MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
				{MsMACAddress: "11:22:33:44:55:66", BytesRx: "1500", BytesTx: "2500"},
			},
			map[string]client.TrafficStats{
				"aa:bb:cc:dd:ee:ff": {MsMACAddress: "aa:bb:cc:dd:ee:ff", BytesRx: "1000", BytesTx: "2000"},
				"11:22:33:44:55:66": {MsMACAddress: "11:22:33:44:55:66", BytesRx: "1500", BytesTx: "2500"},
			},
		},
		{
			"Empty MAC address",
			[]client.TrafficStats{
				{MsMACAddress: "", BytesRx: "1000"},
			},
			map[string]client.TrafficStats{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildTrafficMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildTrafficMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac := range tt.expected {
				if _, exists := got[mac]; !exists {
					t.Errorf("buildTrafficMap() missing MAC %q", mac)
				}
			}
		})
	}
}

func TestBuildMobilityMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []client.MmIfClientHistory
		expected map[string]client.MmIfClientHistory
	}{
		{
			"Empty slice",
			[]client.MmIfClientHistory{},
			map[string]client.MmIfClientHistory{},
		},
		{
			"Single entry",
			[]client.MmIfClientHistory{
				{ClientMAC: "aa:bb:cc:dd:ee:ff"},
			},
			map[string]client.MmIfClientHistory{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff"},
			},
		},
		{
			"Multiple entries",
			[]client.MmIfClientHistory{
				{ClientMAC: "aa:bb:cc:dd:ee:ff"},
				{ClientMAC: "11:22:33:44:55:66"},
			},
			map[string]client.MmIfClientHistory{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff"},
				"11:22:33:44:55:66": {ClientMAC: "11:22:33:44:55:66"},
			},
		},
		{
			"Empty MAC address",
			[]client.MmIfClientHistory{
				{ClientMAC: ""},
			},
			map[string]client.MmIfClientHistory{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildMobilityMap(tt.input)

			if len(got) != len(tt.expected) {
				t.Errorf("buildMobilityMap() returned map with %d entries, want %d", len(got), len(tt.expected))
			}

			for mac := range tt.expected {
				if _, exists := got[mac]; !exists {
					t.Errorf("buildMobilityMap() missing MAC %q", mac)
				}
			}
		})
	}
}

func TestDetermineDeviceNameFromDeviceMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		deviceMap map[string]client.DcInfo
		mac       string
		expected  string
	}{
		{
			"Found in map",
			map[string]client.DcInfo{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
			},
			"aa:bb:cc:dd:ee:ff",
			"iPhone",
		},
		{
			"Not found in map",
			map[string]client.DcInfo{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff", DeviceName: "iPhone"},
			},
			"11:22:33:44:55:66",
			"",
		},
		{
			"Empty map",
			map[string]client.DcInfo{},
			"aa:bb:cc:dd:ee:ff",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineDeviceNameFromDeviceMap(tt.deviceMap, tt.mac); got != tt.expected {
				t.Errorf("determineDeviceNameFromDeviceMap() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetermineIPv4FromSISF(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		sisf     client.SisfDBMac
		expected string
	}{
		{
			"Valid IPv4",
			func() client.SisfDBMac {
				var sisf client.SisfDBMac
				sisf.Ipv4Binding.IPKey.IPAddr = "192.168.1.100"
				return sisf
			}(),
			"192.168.1.100",
		},
		{
			"Empty IPv4",
			client.SisfDBMac{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineIPv4FromSISF(tt.sisf); got != tt.expected {
				t.Errorf("determineIPv4FromSISF() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetermineIPv6FromSISF(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		sisf     client.SisfDBMac
		expected string
	}{
		{
			"Valid IPv6",
			func() client.SisfDBMac {
				var sisf client.SisfDBMac
				sisf.Ipv6Binding = make([]struct {
					Ipv6BindingIPKey struct {
						ZoneID int64  `json:"zone-id"`
						IPAddr string `json:"ip-addr"`
					} `json:"ip-key"`
				}, 1)
				sisf.Ipv6Binding[0].Ipv6BindingIPKey.IPAddr = "2001:db8::1"
				return sisf
			}(),
			"2001:db8::1",
		},
		{
			"Link-local IPv6 skipped",
			func() client.SisfDBMac {
				var sisf client.SisfDBMac
				sisf.Ipv6Binding = make([]struct {
					Ipv6BindingIPKey struct {
						ZoneID int64  `json:"zone-id"`
						IPAddr string `json:"ip-addr"`
					} `json:"ip-key"`
				}, 2)
				sisf.Ipv6Binding[0].Ipv6BindingIPKey.IPAddr = "fe80::1"
				sisf.Ipv6Binding[1].Ipv6BindingIPKey.IPAddr = "2001:db8::1"
				return sisf
			}(),
			"2001:db8::1",
		},
		{
			"Only link-local IPv6",
			func() client.SisfDBMac {
				var sisf client.SisfDBMac
				sisf.Ipv6Binding = make([]struct {
					Ipv6BindingIPKey struct {
						ZoneID int64  `json:"zone-id"`
						IPAddr string `json:"ip-addr"`
					} `json:"ip-key"`
				}, 1)
				sisf.Ipv6Binding[0].Ipv6BindingIPKey.IPAddr = "fe80::1"
				return sisf
			}(),
			"",
		},
		{
			"Empty IPv6 binding",
			client.SisfDBMac{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineIPv6FromSISF(tt.sisf); got != tt.expected {
				t.Errorf("determineIPv6FromSISF() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDetermineLastRunLatency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		mobilityMap map[string]client.MmIfClientHistory
		clientMAC   string
		expected    float64
	}{
		{
			"Found in map with latency 1500ms",
			func() map[string]client.MmIfClientHistory {
				var hist client.MmIfClientHistory
				hist.ClientMAC = "aa:bb:cc:dd:ee:ff"
				hist.MobilityHistory.Entry = make([]struct {
					InstanceID    int       `json:"instance-id"`
					MsApSlotID    int       `json:"ms-ap-slot-id"`
					MsAssocTime   time.Time `json:"ms-assoc-time"`
					Role          string    `json:"role"`
					Bssid         string    `json:"bssid"`
					ApName        string    `json:"ap-name"`
					RunLatency    int       `json:"run-latency"`
					Dot11RoamType string    `json:"dot11-roam-type"`
				}, 1)
				hist.MobilityHistory.Entry[0].RunLatency = 1500
				return map[string]client.MmIfClientHistory{
					"aa:bb:cc:dd:ee:ff": hist,
				}
			}(),
			"aa:bb:cc:dd:ee:ff",
			1.5,
		},
		{
			"Not found in map",
			map[string]client.MmIfClientHistory{
				"aa:bb:cc:dd:ee:ff": {ClientMAC: "aa:bb:cc:dd:ee:ff"},
			},
			"11:22:33:44:55:66",
			0.0,
		},
		{
			"Empty map",
			map[string]client.MmIfClientHistory{},
			"aa:bb:cc:dd:ee:ff",
			0.0,
		},
		{
			"Zero latency",
			func() map[string]client.MmIfClientHistory {
				var hist client.MmIfClientHistory
				hist.ClientMAC = "aa:bb:cc:dd:ee:ff"
				hist.MobilityHistory.Entry = make([]struct {
					InstanceID    int       `json:"instance-id"`
					MsApSlotID    int       `json:"ms-ap-slot-id"`
					MsAssocTime   time.Time `json:"ms-assoc-time"`
					Role          string    `json:"role"`
					Bssid         string    `json:"bssid"`
					ApName        string    `json:"ap-name"`
					RunLatency    int       `json:"run-latency"`
					Dot11RoamType string    `json:"dot11-roam-type"`
				}, 1)
				hist.MobilityHistory.Entry[0].RunLatency = 0
				return map[string]client.MmIfClientHistory{
					"aa:bb:cc:dd:ee:ff": hist,
				}
			}(),
			"aa:bb:cc:dd:ee:ff",
			0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := determineLastRunLatency(tt.mobilityMap, tt.clientMAC); got != tt.expected {
				t.Errorf("determineLastRunLatency() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestParseMCSIndex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		currentRate string
		expected    int
	}{
		{
			"MCS 0",
			"m0 ",
			0,
		},
		{
			"MCS 7",
			"m7 ",
			7,
		},
		{
			"MCS 11",
			"m11 ",
			11,
		},
		{
			"Legacy rate (not MCS)",
			"54 Mbps",
			-1,
		},
		{
			"Empty string",
			"",
			-1,
		},
		{
			"Invalid MCS format",
			"mX ",
			-1,
		},
		{
			"MCS with additional text",
			"m9 2ss short",
			9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseMCSIndex(tt.currentRate); got != tt.expected {
				t.Errorf("parseMCSIndex(%q) = %d, want %d", tt.currentRate, got, tt.expected)
			}
		})
	}
}

func TestClientCollector_MetricNames(t *testing.T) {
	t.Parallel()

	collector := NewClientCollector(nil, ClientMetrics{
		General: true,
		Radio:   true,
		Traffic: true,
		Errors:  true,
		Info:    true,
	})

	tests := []struct {
		desc     *prometheus.Desc
		expected string
	}{
		{collector.stateDesc, "wnc_client_state"},
		{collector.associationUptimeDesc, "wnc_client_uptime_seconds"},
		{collector.stateTransitionSecondsDesc, "wnc_client_state_transition_seconds"},
		{collector.powerSaveStateDesc, "wnc_client_power_save_state"},
		{collector.protocolDesc, "wnc_client_protocol"},
		{collector.mcsIndexDesc, "wnc_client_mcs_index"},
		{collector.spatialStreamsDesc, "wnc_client_spatial_streams"},
		{collector.speedDesc, "wnc_client_speed_mbps"},
		{collector.rssiDesc, "wnc_client_rssi_dbm"},
		{collector.snrDesc, "wnc_client_snr_decibels"},
		{collector.bytesRxDesc, "wnc_client_rx_bytes_total"},
		{collector.bytesTxDesc, "wnc_client_tx_bytes_total"},
		{collector.packetsRxDesc, "wnc_client_rx_packets_total"},
		{collector.packetsTxDesc, "wnc_client_tx_packets_total"},
		{collector.retryRatioDesc, "wnc_client_retry_ratio_percent"},
		{collector.policyErrorsDesc, "wnc_client_policy_errors_total"},
		{collector.infoDesc, "wnc_client_info"},
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

			if !strings.Contains(descStr, "wnc_client_") {
				t.Errorf("Descriptor %s does not follow wnc_client_ naming convention", descStr)
			}
		})
	}
}

func TestNewClientCollector_InfoLabels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		infoLabels []string
		wantLabels []string
	}{
		{
			"No info labels",
			nil,
			[]string{"mac"},
		},
		{
			"Custom info labels",
			[]string{"ap", "band", "wlan", "name", "username"},
			[]string{"mac", "ap", "band", "wlan", "name", "username"},
		},
		{
			"Info labels with required labels included",
			[]string{"mac", "ap", "wlan"},
			[]string{"mac", "ap", "wlan"},
		},
		{
			"Invalid labels filtered",
			[]string{"ap", "invalid_label", "wlan"},
			[]string{"mac", "ap", "wlan"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			metrics := ClientMetrics{
				Info:       true,
				InfoLabels: tt.infoLabels,
			}
			collector := NewClientCollector(nil, metrics)

			if collector.infoDesc == nil {
				t.Error("NewClientCollector() info descriptor should not be nil when Info is enabled")
				return
			}

			if len(collector.infoLabelNames) != len(tt.wantLabels) {
				t.Errorf(
					"NewClientCollector() infoLabelNames length = %d, want %d",
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
						"NewClientCollector() missing expected label %q in infoLabelNames %v",
						expectedLabel,
						collector.infoLabelNames,
					)
				}
			}
		})
	}
}

func TestClientCollector_Integration(t *testing.T) {
	t.Parallel()

	metrics := ClientMetrics{
		General:    true,
		Radio:      true,
		Traffic:    true,
		Errors:     true,
		Info:       true,
		InfoLabels: []string{"ap", "band", "wlan", "name", "username", "ipv4", "ipv6"},
	}

	collector := NewClientCollector(nil, metrics)
	if collector == nil {
		t.Fatal("NewClientCollector() returned nil")
	}

	var _ prometheus.Collector = collector

	ch := make(chan *prometheus.Desc, 100)
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

	expectedDescs := 27
	if count != expectedDescs {
		t.Errorf("Collector emitted %d descriptors, want %d", count, expectedDescs)
	}
}

func TestClientCollector_Collect_EarlyReturn(t *testing.T) {
	t.Parallel()

	collector := &ClientCollector{
		metrics: ClientMetrics{},
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

func TestClientCollector_collectInfoMetrics_LabelValues(t *testing.T) {
	t.Parallel()

	data := client.CommonOperData{
		ClientMAC:   "aa:bb:cc:dd:ee:ff",
		ApName:      "AP-01",
		Username:    "user@example.com",
		MsRadioType: "dot11-5ghz-radio",
	}

	dot11Map := map[string]client.Dot11OperData{
		"aa:bb:cc:dd:ee:ff": {
			MsMACAddress: "aa:bb:cc:dd:ee:ff",
			VapSsid:      "GuestNetwork",
			MsAssocTime:  time.Now().Add(-1 * time.Hour),
		},
	}

	deviceMap := map[string]client.DcInfo{
		"aa:bb:cc:dd:ee:ff": {
			ClientMAC:  "aa:bb:cc:dd:ee:ff",
			DeviceName: "iPhone",
		},
	}

	sisfMap := func() map[string]client.SisfDBMac {
		var sisf client.SisfDBMac
		sisf.MACAddr = "aa:bb:cc:dd:ee:ff"
		sisf.Ipv4Binding.IPKey.IPAddr = "192.168.1.100"
		sisf.Ipv6Binding = make([]struct {
			Ipv6BindingIPKey struct {
				ZoneID int64  `json:"zone-id"`
				IPAddr string `json:"ip-addr"`
			} `json:"ip-key"`
		}, 1)
		sisf.Ipv6Binding[0].Ipv6BindingIPKey.IPAddr = "2001:db8::1"
		return map[string]client.SisfDBMac{
			"aa:bb:cc:dd:ee:ff": sisf,
		}
	}()

	metrics := ClientMetrics{
		Info:       true,
		InfoLabels: []string{"ap", "band", "wlan", "name", "username", "ipv4", "ipv6"},
	}

	collector := NewClientCollector(nil, metrics)

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectInfoMetrics(ch, data, dot11Map, deviceMap, sisfMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount != 1 {
		t.Errorf("collectInfoMetrics() emitted %d metrics, want 1", metricCount)
	}
}

// TestClientCollector_collectGeneralMetrics tests basic metric emission
func TestClientCollector_collectGeneralMetrics(t *testing.T) {
	t.Parallel()

	data := client.CommonOperData{
		ClientMAC:   "aa:bb:cc:dd:ee:ff",
		ApName:      "AP-01",
		CoState:     "associated",
		MsRadioType: "dot11-5ghz-radio",
	}

	trafficMap := map[string]client.TrafficStats{
		"aa:bb:cc:dd:ee:ff": {
			PowerSaveState: 0,
		},
	}

	dot11Map := map[string]client.Dot11OperData{
		"aa:bb:cc:dd:ee:ff": {
			MsMACAddress: "aa:bb:cc:dd:ee:ff",
			MsAssocTime:  time.Now().Add(-1 * time.Hour),
		},
	}

	mobilityMap := map[string]client.MmIfClientHistory{}

	collector := &ClientCollector{
		metrics:                    ClientMetrics{General: true},
		stateDesc:                  prometheus.NewDesc("test_state", "test", []string{"mac"}, nil),
		associationUptimeDesc:      prometheus.NewDesc("test_uptime", "test", []string{"mac"}, nil),
		stateTransitionSecondsDesc: prometheus.NewDesc("test_transition", "test", []string{"mac"}, nil),
		powerSaveStateDesc:         prometheus.NewDesc("test_power_save", "test", []string{"mac"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectGeneralMetrics(ch, data, trafficMap, dot11Map, mobilityMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectGeneralMetrics() emitted 0 metrics, want > 0")
	}
}

// TestClientCollector_collectRadioMetrics tests basic metric emission
func TestClientCollector_collectRadioMetrics(t *testing.T) {
	t.Parallel()

	data := client.CommonOperData{
		ClientMAC:   "aa:bb:cc:dd:ee:ff",
		MsRadioType: "dot11-5ghz-radio",
	}

	trafficMap := map[string]client.TrafficStats{
		"aa:bb:cc:dd:ee:ff": {
			Speed:          866,
			SpatialStream:  2,
			CurrentRate:    "mcs9",
			MostRecentRSSI: -45,
			MostRecentSNR:  35,
		},
	}

	dot11Map := map[string]client.Dot11OperData{
		"aa:bb:cc:dd:ee:ff": {
			MsMACAddress:  "aa:bb:cc:dd:ee:ff",
			EwlcMsPhyType: "ax",
			RadioType:     "dot11-5ghz-radio",
			Is11GClient:   false,
		},
	}

	collector := &ClientCollector{
		metrics:            ClientMetrics{Radio: true},
		protocolDesc:       prometheus.NewDesc("test_protocol", "test", []string{"mac"}, nil),
		speedDesc:          prometheus.NewDesc("test_speed", "test", []string{"mac"}, nil),
		spatialStreamsDesc: prometheus.NewDesc("test_spatial_streams", "test", []string{"mac"}, nil),
		mcsIndexDesc:       prometheus.NewDesc("test_mcs_index", "test", []string{"mac"}, nil),
		rssiDesc:           prometheus.NewDesc("test_rssi", "test", []string{"mac"}, nil),
		snrDesc:            prometheus.NewDesc("test_snr", "test", []string{"mac"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectRadioMetrics(ch, data, trafficMap, dot11Map)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectRadioMetrics() emitted 0 metrics, want > 0")
	}
}

// TestClientCollector_collectTrafficMetrics tests basic metric emission
func TestClientCollector_collectTrafficMetrics(t *testing.T) {
	t.Parallel()

	data := client.CommonOperData{
		ClientMAC: "aa:bb:cc:dd:ee:ff",
	}

	trafficMap := map[string]client.TrafficStats{
		"aa:bb:cc:dd:ee:ff": {
			BytesRx: "1000000",
			BytesTx: "2000000",
			PktsRx:  "10000",
			PktsTx:  "20000",
		},
	}

	collector := &ClientCollector{
		metrics:       ClientMetrics{Traffic: true},
		bytesRxDesc:   prometheus.NewDesc("test_bytes_rx", "test", []string{"mac"}, nil),
		bytesTxDesc:   prometheus.NewDesc("test_bytes_tx", "test", []string{"mac"}, nil),
		packetsRxDesc: prometheus.NewDesc("test_packets_rx", "test", []string{"mac"}, nil),
		packetsTxDesc: prometheus.NewDesc("test_packets_tx", "test", []string{"mac"}, nil),
	}

	ch := make(chan prometheus.Metric, 10)
	go func() {
		defer close(ch)
		collector.collectTrafficMetrics(ch, data, trafficMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectTrafficMetrics() emitted 0 metrics, want > 0")
	}
}

// TestClientCollector_collectErrorMetrics tests basic metric emission
func TestClientCollector_collectErrorMetrics(t *testing.T) {
	t.Parallel()

	data := client.CommonOperData{
		ClientMAC: "aa:bb:cc:dd:ee:ff",
	}

	trafficMap := map[string]client.TrafficStats{
		"aa:bb:cc:dd:ee:ff": {
			PolicyErrs:         "10",
			DuplicateRcv:       "5",
			DecryptFailed:      "2",
			MicMismatch:        "1",
			MicMissing:         "0",
			TxExcessiveRetries: "3",
			RxGroupCounter:     "0",
			TxTotalDrops:       "4",
			DataRetries:        "100",
			RtsRetries:         "20",
			TxRetries:          "50",
			PktsTx:             "10000",
		},
	}

	collector := &ClientCollector{
		metrics:               ClientMetrics{Errors: true},
		retryRatioDesc:        prometheus.NewDesc("test_retry_ratio", "test", []string{"mac"}, nil),
		policyErrorsDesc:      prometheus.NewDesc("test_policy_errors", "test", []string{"mac"}, nil),
		duplicateReceivedDesc: prometheus.NewDesc("test_duplicate_received", "test", []string{"mac"}, nil),
		decryptionFailedDesc:  prometheus.NewDesc("test_decryption_failed", "test", []string{"mac"}, nil),
		micMismatchDesc:       prometheus.NewDesc("test_mic_mismatch", "test", []string{"mac"}, nil),
		micMissingDesc:        prometheus.NewDesc("test_mic_missing", "test", []string{"mac"}, nil),
		excessiveRetriesDesc:  prometheus.NewDesc("test_excessive_retries", "test", []string{"mac"}, nil),
		rxGroupCounterDesc:    prometheus.NewDesc("test_rx_group_counter", "test", []string{"mac"}, nil),
		txTotalDropsDesc:      prometheus.NewDesc("test_tx_total_drops", "test", []string{"mac"}, nil),
		dataRetriesDesc:       prometheus.NewDesc("test_data_retries", "test", []string{"mac"}, nil),
		rtsRetriesDesc:        prometheus.NewDesc("test_rts_retries", "test", []string{"mac"}, nil),
		txRetriesDesc:         prometheus.NewDesc("test_tx_retries", "test", []string{"mac"}, nil),
	}

	ch := make(chan prometheus.Metric, 20)
	go func() {
		defer close(ch)
		collector.collectErrorMetrics(ch, data, trafficMap)
	}()

	metricCount := 0
	for range ch {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("collectErrorMetrics() emitted 0 metrics, want > 0")
	}
}

// TestClientCollector_collectMetrics_NilSafety tests nil safety
func TestClientCollector_collectMetrics_NilSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "collectGeneralMetrics with empty maps",
			testFunc: func(t *testing.T) {
				t.Parallel()
				collector := &ClientCollector{
					metrics:                    ClientMetrics{General: true},
					stateDesc:                  prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					associationUptimeDesc:      prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					stateTransitionSecondsDesc: prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					powerSaveStateDesc:         prometheus.NewDesc("test", "test", []string{"mac"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectGeneralMetrics() panicked with empty maps: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				data := client.CommonOperData{ClientMAC: "aa:bb:cc:dd:ee:ff"}
				collector.collectGeneralMetrics(ch, data, map[string]client.TrafficStats{}, map[string]client.Dot11OperData{}, map[string]client.MmIfClientHistory{})
			},
		},
		{
			name: "collectRadioMetrics with empty maps",
			testFunc: func(t *testing.T) {
				t.Parallel()
				collector := &ClientCollector{
					metrics:            ClientMetrics{Radio: true},
					protocolDesc:       prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					speedDesc:          prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					spatialStreamsDesc: prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					mcsIndexDesc:       prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					rssiDesc:           prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					snrDesc:            prometheus.NewDesc("test", "test", []string{"mac"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectRadioMetrics() panicked with empty maps: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				data := client.CommonOperData{ClientMAC: "aa:bb:cc:dd:ee:ff"}
				collector.collectRadioMetrics(ch, data, map[string]client.TrafficStats{}, map[string]client.Dot11OperData{})
			},
		},
		{
			name: "collectTrafficMetrics with empty maps",
			testFunc: func(t *testing.T) {
				t.Parallel()
				collector := &ClientCollector{
					metrics:       ClientMetrics{Traffic: true},
					bytesRxDesc:   prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					bytesTxDesc:   prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					packetsRxDesc: prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					packetsTxDesc: prometheus.NewDesc("test", "test", []string{"mac"}, nil),
				}
				ch := make(chan prometheus.Metric, 10)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectTrafficMetrics() panicked with empty maps: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				data := client.CommonOperData{ClientMAC: "aa:bb:cc:dd:ee:ff"}
				collector.collectTrafficMetrics(ch, data, map[string]client.TrafficStats{})
			},
		},
		{
			name: "collectErrorMetrics with empty maps",
			testFunc: func(t *testing.T) {
				t.Parallel()
				collector := &ClientCollector{
					metrics:               ClientMetrics{Errors: true},
					retryRatioDesc:        prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					policyErrorsDesc:      prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					duplicateReceivedDesc: prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					decryptionFailedDesc:  prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					micMismatchDesc:       prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					micMissingDesc:        prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					excessiveRetriesDesc:  prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					rxGroupCounterDesc:    prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					txTotalDropsDesc:      prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					dataRetriesDesc:       prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					rtsRetriesDesc:        prometheus.NewDesc("test", "test", []string{"mac"}, nil),
					txRetriesDesc:         prometheus.NewDesc("test", "test", []string{"mac"}, nil),
				}
				ch := make(chan prometheus.Metric, 20)
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("collectErrorMetrics() panicked with empty maps: %v", r)
					}
					close(ch)
					for range ch {
					}
				}()
				data := client.CommonOperData{ClientMAC: "aa:bb:cc:dd:ee:ff"}
				collector.collectErrorMetrics(ch, data, map[string]client.TrafficStats{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
