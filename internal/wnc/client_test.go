package wnc

import (
	"context"
	"errors"
	"testing"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
)

// mockClientDataSource implements DataSource for testing Client operations.
type mockClientDataSource struct {
	data *WNCDataCache
	err  error
}

func (m *mockClientDataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// newMockClientDataSource creates a mock DataSource with client sample data based on live WNC.
func newMockClientDataSource() *mockClientDataSource {
	return &mockClientDataSource{
		data: &WNCDataCache{
			CommonOperData: []client.CommonOperData{
				{
					ClientMAC:   "aa:bb:cc:11:22:a9",
					ApName:      "TEST-AP01",
					MsApSlotID:  0,
					MsRadioType: "client-dot11n-24-ghz-prot",
					WlanID:      1,
					ClientType:  "dot11-client-normal",
					CoState:     "client-status-run",
				},
				{
					ClientMAC:   "aa:bb:cc:11:22:b0",
					ApName:      "TEST-AP02",
					MsApSlotID:  1,
					MsRadioType: "client-dot11ac-5-ghz-prot",
					WlanID:      2,
					ClientType:  "dot11-client-normal",
					CoState:     "client-status-run",
				},
			},
			DCInfo: []client.DcInfo{
				{
					ClientMAC:       "aa:bb:cc:11:22:a9",
					DeviceType:      "Un-Classified Device",
					ProtocolMap:     "protocol-map-oui",
					ConfidenceLevel: 0,
					DeviceName:      "Unknown Device",
				},
			},
			Dot11OperData: []client.Dot11OperData{
				{
					MsMACAddress:   "aa:bb:cc:11:22:a9",
					Dot11State:     "associated",
					MsBssid:        "aa:bb:cc:11:22:81",
					ApMACAddress:   "aa:bb:cc:11:22:10",
					CurrentChannel: 6,
					MsWlanID:       1,
					VapSsid:        "labo-wlan",
					PolicyProfile:  "labo-wlan-profile",
					MsApSlotID:     0,
					RadioType:      "dot11-radio-type-bg",
				},
				{
					MsMACAddress:   "aa:bb:cc:11:22:b0",
					Dot11State:     "associated",
					MsBssid:        "aa:bb:cc:11:22:91",
					ApMACAddress:   "aa:bb:cc:11:22:11",
					CurrentChannel: 36,
					MsWlanID:       2,
					VapSsid:        "guest",
					PolicyProfile:  "guest-profile",
					MsApSlotID:     1,
					RadioType:      "dot11-radio-type-a",
				},
			},
			SisfDBMac: []client.SisfDBMac{
				{
					MACAddr: "aa:bb:cc:11:22:a9",
					Ipv4Binding: struct {
						IPKey struct {
							ZoneID int    `json:"zone-id"`
							IPAddr string `json:"ip-addr"`
						} `json:"ip-key"`
					}{
						IPKey: struct {
							ZoneID int    `json:"zone-id"`
							IPAddr string `json:"ip-addr"`
						}{
							ZoneID: 0,
							IPAddr: "192.168.0.74",
						},
					},
				},
			},
			TrafficStats: []client.TrafficStats{
				{
					MsMACAddress:   "aa:bb:cc:11:22:a9",
					BytesRx:        "861981",
					BytesTx:        "29792",
					PolicyErrs:     "0",
					PktsRx:         "4330",
					PktsTx:         "280",
					DataRetries:    "0",
					MostRecentRSSI: -50,
					MostRecentSNR:  44,
					PowerSaveState: 1,
					CurrentRate:    "54.0",
					Speed:          54,
					SpatialStream:  0,
					ClientActive:   true,
				},
			},
			MmIfClientHistory: []client.MmIfClientHistory{
				{
					ClientMAC: "aa:bb:cc:11:22:a9",
				},
			},
		},
	}
}

func TestNewClientSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sharedDataSource DataSource
	}{
		{
			name:             "With non-nil shared data source",
			sharedDataSource: newMockClientDataSource(),
		},
		{
			name:             "With nil shared data source",
			sharedDataSource: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.sharedDataSource)
			if source == nil {
				t.Fatal("NewClientSource returned nil")
			}
		})
	}
}

func TestClientSource_GetClientData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple clients",
			mock:    newMockClientDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					CommonOperData: []client.CommonOperData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetClientData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClientData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetClientData() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify first item structure
				if data[0].ClientMAC == "" {
					t.Error("GetClientData() first item has empty ClientMAC")
				}
				if data[0].ApName == "" {
					t.Error("GetClientData() first item has empty ApName")
				}
			}
		})
	}
}

func TestClientSource_GetDeviceData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with device classification data",
			mock:    newMockClientDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					DCInfo: []client.DcInfo{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetDeviceData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetDeviceData() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify device classification structure
				if data[0].ClientMAC == "" {
					t.Error("GetDeviceData() first item has empty ClientMAC")
				}
				if data[0].DeviceType == "" {
					t.Error("GetDeviceData() first item has empty DeviceType")
				}
			}
		})
	}
}

func TestClientSource_GetDot11Data(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple dot11 clients",
			mock:    newMockClientDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					Dot11OperData: []client.Dot11OperData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetDot11Data(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDot11Data() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetDot11Data() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify dot11 structure
				if data[0].MsMACAddress == "" {
					t.Error("GetDot11Data() first item has empty MsMACAddress")
				}
				if data[0].Dot11State == "" {
					t.Error("GetDot11Data() first item has empty Dot11State")
				}
				if data[0].VapSsid == "" {
					t.Error("GetDot11Data() first item has empty VapSsid")
				}
			}
		})
	}
}

func TestClientSource_GetSISFDBData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with SISF data",
			mock:    newMockClientDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					SisfDBMac: []client.SisfDBMac{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetSISFDBData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSISFDBData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetSISFDBData() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify SISF structure
				if data[0].MACAddr == "" {
					t.Error("GetSISFDBData() first item has empty MACAddr")
				}
				if data[0].Ipv4Binding.IPKey.IPAddr == "" {
					t.Error("GetSISFDBData() first item has empty IPv4 address")
				}
			}
		})
	}
}

func TestClientSource_GetTrafficStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with traffic stats",
			mock:    newMockClientDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					TrafficStats: []client.TrafficStats{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetTrafficStats(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTrafficStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetTrafficStats() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify traffic stats structure
				if data[0].MsMACAddress == "" {
					t.Error("GetTrafficStats() first item has empty MsMACAddress")
				}
				if data[0].BytesRx == "" {
					t.Error("GetTrafficStats() first item has empty BytesRx")
				}
				if data[0].BytesTx == "" {
					t.Error("GetTrafficStats() first item has empty BytesTx")
				}
			}
		})
	}
}

func TestClientSource_GetMobilityHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockClientDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with mobility history",
			mock:    newMockClientDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockClientDataSource{
				data: &WNCDataCache{
					MmIfClientHistory: []client.MmIfClientHistory{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockClientDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewClientSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetMobilityHistory(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMobilityHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetMobilityHistory() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify mobility history structure
				if data[0].ClientMAC == "" {
					t.Error("GetMobilityHistory() first item has empty ClientMAC")
				}
			}
		})
	}
}
