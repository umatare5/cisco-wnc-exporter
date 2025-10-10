package wnc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// mockRRMDataSource implements DataSource for testing RRM operations.
type mockRRMDataSource struct {
	data *WNCDataCache
	err  error
}

func (m *mockRRMDataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// newMockRRMDataSource creates a mock DataSource with RRM sample data based on live WNC.
func newMockRRMDataSource() *mockRRMDataSource {
	return &mockRRMDataSource{
		data: &WNCDataCache{
			RRMMeasurements: []rrm.RRMMeasurement{
				{
					WtpMAC:      "aa:bb:cc:11:22:80",
					RadioSlotID: 0,
					Foreign: &rrm.Foreign{
						Foreign: rrm.ForeignData{
							ForeignData: []rrm.ForeignDataItem{
								{
									Chan:                1,
									Power:               -128,
									Rogue20Count:        1,
									Rogue40PrimaryCount: 0,
									Rogue80PrimaryCount: 0,
									ChanUtil:            0,
								},
								{
									Chan:                6,
									Power:               -73,
									Rogue20Count:        6,
									Rogue40PrimaryCount: 0,
									Rogue80PrimaryCount: 0,
									ChanUtil:            6,
								},
							},
						},
					},
					Noise: &rrm.Noise{
						Noise: rrm.NoiseData{
							NoiseData: []rrm.NoiseDataItem{
								{Chan: 1, Noise: -94},
								{Chan: 6, Noise: -94},
							},
						},
					},
				},
				{
					WtpMAC:      "aa:bb:cc:11:22:80",
					RadioSlotID: 1,
					Foreign: &rrm.Foreign{
						Foreign: rrm.ForeignData{
							ForeignData: []rrm.ForeignDataItem{
								{
									Chan:                36,
									Power:               -85,
									Rogue20Count:        2,
									Rogue40PrimaryCount: 1,
									Rogue80PrimaryCount: 0,
									ChanUtil:            4,
								},
							},
						},
					},
					Noise: &rrm.Noise{
						Noise: rrm.NoiseData{
							NoiseData: []rrm.NoiseDataItem{
								{Chan: 36, Noise: -92},
							},
						},
					},
				},
			},
			RRMCoverage: []rrm.RRMCoverage{
				{
					WtpMAC:            "aa:bb:cc:11:22:80",
					RadioSlotID:       0,
					FailedClientCount: 2,
					SNRInfo: []rrm.SNRInfo{
						{SNR: 20, NumClients: 5},
						{SNR: 25, NumClients: 3},
					},
					RSSIInfo: []rrm.RSSIInfo{
						{RSSI: -50, NumClients: 4},
						{RSSI: -60, NumClients: 4},
					},
				},
				{
					WtpMAC:            "aa:bb:cc:11:22:80",
					RadioSlotID:       1,
					FailedClientCount: 0,
					SNRInfo: []rrm.SNRInfo{
						{SNR: 30, NumClients: 2},
					},
					RSSIInfo: []rrm.RSSIInfo{
						{RSSI: -45, NumClients: 2},
					},
				},
			},
			ApDot11RadarData: []rrm.ApDot11RadarData{
				{
					WtpMAC:           "aa:bb:cc:11:22:80",
					RadioSlotID:      0,
					LastRadarOnRadio: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					WtpMAC:           "aa:bb:cc:11:22:80",
					RadioSlotID:      1,
					LastRadarOnRadio: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					WtpMAC:           "aa:bb:cc:11:22:20",
					RadioSlotID:      0,
					LastRadarOnRadio: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
}

func TestNewRRMSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sharedDataSource DataSource
		expectNil        bool
	}{
		{
			name:             "With non-nil shared data source",
			sharedDataSource: newMockRRMDataSource(),
			expectNil:        false,
		},
		{
			name:             "With nil shared data source",
			sharedDataSource: nil,
			expectNil:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewRRMSource(tt.sharedDataSource)
			if source == nil && !tt.expectNil {
				t.Error("Expected non-nil RRMSource, got nil")
			}
			if source != nil && tt.expectNil {
				t.Error("Expected nil RRMSource, got non-nil")
			}
		})
	}
}

func TestRRMSource_GetRRMMeasurements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockRRMDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple measurements",
			mock:    newMockRRMDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockRRMDataSource{
				data: &WNCDataCache{
					RRMMeasurements: []rrm.RRMMeasurement{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockRRMDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewRRMSource(tt.mock)
			measurements, err := source.GetRRMMeasurements(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRRMMeasurements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(measurements) != tt.wantLen {
				t.Errorf("GetRRMMeasurements() got %d measurements, want %d", len(measurements), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstMeasurement := measurements[0]
				if firstMeasurement.WtpMAC == "" {
					t.Error("Expected WtpMAC to be set")
				}
				if firstMeasurement.Foreign == nil {
					t.Error("Expected Foreign data to be set")
				}
				if firstMeasurement.Noise == nil {
					t.Error("Expected Noise data to be set")
				}
			}
		})
	}
}

func TestRRMSource_GetRRMCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockRRMDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple coverage entries",
			mock:    newMockRRMDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockRRMDataSource{
				data: &WNCDataCache{
					RRMCoverage: []rrm.RRMCoverage{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockRRMDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewRRMSource(tt.mock)
			coverage, err := source.GetRRMCoverage(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRRMCoverage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(coverage) != tt.wantLen {
				t.Errorf("GetRRMCoverage() got %d coverage entries, want %d", len(coverage), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstCoverage := coverage[0]
				if firstCoverage.WtpMAC == "" {
					t.Error("Expected WtpMAC to be set")
				}
				if len(firstCoverage.SNRInfo) == 0 {
					t.Error("Expected SNRInfo to be set")
				}
				if len(firstCoverage.RSSIInfo) == 0 {
					t.Error("Expected RSSIInfo to be set")
				}
			}
		})
	}
}

func TestRRMSource_GetApDot11RadarData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockRRMDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple radar entries",
			mock:    newMockRRMDataSource(),
			wantLen: 3,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockRRMDataSource{
				data: &WNCDataCache{
					ApDot11RadarData: []rrm.ApDot11RadarData{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockRRMDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewRRMSource(tt.mock)
			radarData, err := source.GetApDot11RadarData(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetApDot11RadarData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(radarData) != tt.wantLen {
				t.Errorf("GetApDot11RadarData() got %d radar entries, want %d", len(radarData), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstRadar := radarData[0]
				if firstRadar.WtpMAC == "" {
					t.Error("Expected WtpMAC to be set")
				}
				if firstRadar.LastRadarOnRadio.IsZero() {
					t.Error("Expected LastRadarOnRadio to be set")
				}
			}
		})
	}
}
