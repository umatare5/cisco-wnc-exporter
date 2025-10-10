package wnc

import (
	"context"
	"errors"
	"testing"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
)

// mockDataSource implements DataSource for testing.
type mockDataSource struct {
	data *WNCDataCache
	err  error
}

func (m *mockDataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// newMockDataSource creates a mock DataSource with sample data based on live WNC.
func newMockDataSource() *mockDataSource {
	return &mockDataSource{
		data: &WNCDataCache{
			CAPWAPData: []ap.CAPWAPData{
				{
					WtpMAC: "aa:bb:cc:11:22:80",
					IPAddr: "192.168.255.11",
					Name:   "TEST-AP01",
				},
				{
					WtpMAC: "aa:bb:cc:11:22:90",
					IPAddr: "192.168.255.12",
					Name:   "TEST-AP02",
				},
			},
			ApOperData: []ap.OperData{
				{
					WtpMAC:  "aa:bb:cc:11:22:80",
					RadioID: 0,
				},
			},
			RadioOperData: []ap.RadioOperData{
				{
					WtpMAC:      "aa:bb:cc:11:22:80",
					RadioSlotID: 0,
					SlotID:      0,
					RadioType:   "radio-80211bg",
					AdminState:  "enabled",
					OperState:   "radio-up",
					PhyHtCfg: &ap.PhyHtCfg{
						CfgData: ap.PhyHtCfgData{
							CurrFreq: 6,
						},
					},
				},
				{
					WtpMAC:      "aa:bb:cc:11:22:80",
					RadioSlotID: 1,
					SlotID:      1,
					RadioType:   "radio-80211a",
					AdminState:  "enabled",
					OperState:   "radio-up",
					PhyHtCfg: &ap.PhyHtCfg{
						CfgData: ap.PhyHtCfgData{
							CurrFreq: 36,
						},
					},
				},
			},
			RadioOperStats: []ap.RadioOperStats{
				{
					ApMAC:           "aa:bb:cc:11:22:80",
					SlotID:          0,
					RxDataPktCount:  1000,
					TxFragmentCount: 2000,
					FcsErrorCount:   10,
					FailedCount:     5,
					NoiseFloor:      -90,
				},
			},
			RadioResetStats: []ap.RadioResetStats{
				{
					ApMAC:       "aa:bb:cc:11:22:80",
					RadioID:     0,
					Cause:       "user-initiated",
					DetailCause: "Manual reset",
					Count:       2,
				},
			},
			NameMACMaps: []ap.ApNameMACMap{
				{
					WtpName: "TEST-AP01",
					EthMAC:  "aa:bb:cc:11:22:80",
				},
				{
					WtpName: "TEST-AP02",
					EthMAC:  "aa:bb:cc:11:22:90",
				},
			},
		},
	}
}

func TestNewAPSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sharedDataSource DataSource
	}{
		{
			name:             "With non-nil shared data source",
			sharedDataSource: newMockDataSource(),
		},
		{
			name:             "With nil shared data source",
			sharedDataSource: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.sharedDataSource)
			if source == nil {
				t.Fatal("NewAPSource returned nil")
			}
		})
	}
}

func TestAPSource_GetCAPWAPData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple APs",
			mock:    newMockDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					CAPWAPData: []ap.CAPWAPData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetCAPWAPData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCAPWAPData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetCAPWAPData() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify first item structure
				if data[0].WtpMAC == "" {
					t.Error("GetCAPWAPData() first item has empty WtpMAC")
				}
				if data[0].Name == "" {
					t.Error("GetCAPWAPData() first item has empty Name")
				}
			}
		})
	}
}

func TestAPSource_GetAPOperData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with AP operational data",
			mock:    newMockDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					ApOperData: []ap.OperData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetAPOperData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPOperData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetAPOperData() returned %d items, want %d", len(data), tt.wantLen)
			}
		})
	}
}

func TestAPSource_GetRadioData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple radios",
			mock:    newMockDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					RadioOperData: []ap.RadioOperData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetRadioData(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRadioData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetRadioData() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify radio structure
				if data[0].WtpMAC == "" {
					t.Error("GetRadioData() first item has empty WtpMAC")
				}
				if data[0].RadioType == "" {
					t.Error("GetRadioData() first item has empty RadioType")
				}
			}
		})
	}
}

func TestAPSource_GetRadioOperStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with radio stats",
			mock:    newMockDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					RadioOperStats: []ap.RadioOperStats{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetRadioOperStats(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRadioOperStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetRadioOperStats() returned %d items, want %d", len(data), tt.wantLen)
			}
		})
	}
}

func TestAPSource_GetRadioResetStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with reset stats",
			mock:    newMockDataSource(),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					RadioResetStats: []ap.RadioResetStats{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.GetRadioResetStats(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRadioResetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("GetRadioResetStats() returned %d items, want %d", len(data), tt.wantLen)
			}
		})
	}
}

func TestAPSource_ListNameMACMaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with name-MAC mappings",
			mock:    newMockDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockDataSource{
				data: &WNCDataCache{
					NameMACMaps: []ap.ApNameMACMap{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewAPSource(tt.mock)
			ctx := context.Background()

			data, err := source.ListNameMACMaps(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListNameMACMaps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(data) != tt.wantLen {
				t.Errorf("ListNameMACMaps() returned %d items, want %d", len(data), tt.wantLen)
			}

			if !tt.wantErr && tt.wantLen > 0 {
				// Verify mapping structure
				if data[0].WtpName == "" {
					t.Error("ListNameMACMaps() first item has empty WtpName")
				}
				if data[0].EthMAC == "" {
					t.Error("ListNameMACMaps() first item has empty EthMAC")
				}
			}
		})
	}
}
