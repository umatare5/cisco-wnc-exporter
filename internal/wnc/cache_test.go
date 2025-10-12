package wnc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

func TestNewDataSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.WNC
	}{
		{
			name: "With valid configuration",
			cfg: config.WNC{
				Controller:    "wnc1.example.internal",
				AccessToken:   "test-token",
				Timeout:       30 * time.Second,
				TLSSkipVerify: true,
				CacheTTL:      55 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewDataSource(tt.cfg)
			if source == nil {
				t.Fatal("NewDataSource returned nil")
			}
		})
	}
}

func TestDataSource_GetCachedData_MockSuccess(t *testing.T) {
	t.Parallel()

	mock := newMockDataSource()
	ctx := context.Background()

	data, err := mock.GetCachedData(ctx)
	if err != nil {
		t.Fatalf("GetCachedData() error = %v, want nil", err)
	}

	if data == nil {
		t.Fatal("GetCachedData() returned nil data")
	}

	if len(data.CAPWAPData) == 0 {
		t.Error("GetCachedData() CAPWAPData is empty")
	}

	if len(data.RadioOperData) == 0 {
		t.Error("GetCachedData() RadioOperData is empty")
	}

	if len(data.NameMACMaps) == 0 {
		t.Error("GetCachedData() NameMACMaps is empty")
	}
}

func TestDataSource_GetCachedData_MockError(t *testing.T) {
	t.Parallel()

	mock := &mockDataSource{
		err: errors.New("cache refresh failed"),
	}
	ctx := context.Background()

	data, err := mock.GetCachedData(ctx)

	if err == nil {
		t.Fatal("GetCachedData() error = nil, want error")
	}

	if data != nil {
		t.Error("GetCachedData() returned non-nil data on error")
	}
}

func TestWNCDataCache_Structure(t *testing.T) {
	t.Parallel()

	cache := &WNCDataCache{
		CAPWAPData: []ap.CAPWAPData{
			{
				WtpMAC: "aa:bb:cc:11:22:80",
				IPAddr: "192.168.255.11",
				Name:   "TEST-AP01",
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
				RadioType:   "radio-80211bg",
			},
		},
		RadioOperStats: []ap.RadioOperStats{
			{
				ApMAC:  "aa:bb:cc:11:22:80",
				SlotID: 0,
			},
		},
		RadioResetStats: []ap.RadioResetStats{
			{
				ApMAC:   "aa:bb:cc:11:22:80",
				RadioID: 0,
			},
		},
		NameMACMaps: []ap.ApNameMACMap{
			{
				WtpName: "TEST-AP01",
				EthMAC:  "aa:bb:cc:11:22:80",
			},
		},
		CommonOperData: []client.CommonOperData{
			{
				ClientMAC: "aa:bb:cc:11:22:a9",
				ApName:    "TEST-AP01",
			},
		},
		DCInfo: []client.DcInfo{
			{
				ClientMAC:  "aa:bb:cc:11:22:a9",
				DeviceType: "Un-Classified Device",
			},
		},
		Dot11OperData: []client.Dot11OperData{
			{
				MsMACAddress: "aa:bb:cc:11:22:a9",
				Dot11State:   "associated",
			},
		},
		SisfDBMac: []client.SisfDBMac{
			{
				MACAddr: "aa:bb:cc:11:22:a9",
			},
		},
		TrafficStats: []client.TrafficStats{
			{
				MsMACAddress: "aa:bb:cc:11:22:a9",
			},
		},
		MmIfClientHistory: []client.MmIfClientHistory{
			{
				ClientMAC: "aa:bb:cc:11:22:a9",
			},
		},
		RRMMeasurements: []rrm.RRMMeasurement{
			{
				WtpMAC: "aa:bb:cc:11:22:80",
			},
		},
		RRMCoverage: []rrm.RRMCoverage{
			{
				WtpMAC:      "aa:bb:cc:11:22:80",
				RadioSlotID: 0,
			},
		},
		ApDot11RadarData: []rrm.ApDot11RadarData{
			{
				WtpMAC: "aa:bb:cc:11:22:80",
			},
		},
		WLANConfigEntries: []wlan.WlanCfgEntry{
			{
				WlanID: 1,
			},
		},
		WLANPolicies: []wlan.WlanPolicy{
			{
				PolicyProfileName: "test-policy",
			},
		},
		WLANPolicyListEntries: []wlan.PolicyListEntry{
			{
				TagName: "test-tag",
			},
		},
	}

	if len(cache.CAPWAPData) != 1 {
		t.Error("CAPWAPData field not properly initialized")
	}
	if len(cache.ApOperData) != 1 {
		t.Error("ApOperData field not properly initialized")
	}
	if len(cache.RadioOperData) != 1 {
		t.Error("RadioOperData field not properly initialized")
	}
	if len(cache.RadioOperStats) != 1 {
		t.Error("RadioOperStats field not properly initialized")
	}
	if len(cache.RadioResetStats) != 1 {
		t.Error("RadioResetStats field not properly initialized")
	}
	if len(cache.NameMACMaps) != 1 {
		t.Error("NameMACMaps field not properly initialized")
	}
	if len(cache.CommonOperData) != 1 {
		t.Error("CommonOperData field not properly initialized")
	}
	if len(cache.DCInfo) != 1 {
		t.Error("DCInfo field not properly initialized")
	}
	if len(cache.Dot11OperData) != 1 {
		t.Error("Dot11OperData field not properly initialized")
	}
	if len(cache.SisfDBMac) != 1 {
		t.Error("SisfDBMac field not properly initialized")
	}
	if len(cache.TrafficStats) != 1 {
		t.Error("TrafficStats field not properly initialized")
	}
	if len(cache.MmIfClientHistory) != 1 {
		t.Error("MmIfClientHistory field not properly initialized")
	}
	if len(cache.RRMMeasurements) != 1 {
		t.Error("RRMMeasurements field not properly initialized")
	}
	if len(cache.RRMCoverage) != 1 {
		t.Error("RRMCoverage field not properly initialized")
	}
	if len(cache.ApDot11RadarData) != 1 {
		t.Error("ApDot11RadarData field not properly initialized")
	}
	if len(cache.WLANConfigEntries) != 1 {
		t.Error("WLANConfigEntries field not properly initialized")
	}
	if len(cache.WLANPolicies) != 1 {
		t.Error("WLANPolicies field not properly initialized")
	}
	if len(cache.WLANPolicyListEntries) != 1 {
		t.Error("WLANPolicyListEntries field not properly initialized")
	}
}

func TestWNCDataCache_EmptyInitialization(t *testing.T) {
	t.Parallel()

	cache := &WNCDataCache{}

	if cache.CAPWAPData != nil {
		t.Error("CAPWAPData should be nil when not initialized")
	}
	if cache.ApOperData != nil {
		t.Error("ApOperData should be nil when not initialized")
	}
	if cache.RadioOperData != nil {
		t.Error("RadioOperData should be nil when not initialized")
	}
	if cache.CommonOperData != nil {
		t.Error("CommonOperData should be nil when not initialized")
	}
	if cache.WLANConfigEntries != nil {
		t.Error("WLANConfigEntries should be nil when not initialized")
	}
}

func TestDataFetcher_Structure(t *testing.T) {
	t.Parallel()

	fetcher := dataFetcher{
		name:     "test-fetcher",
		required: true,
		fetch: func(ctx context.Context, cache *WNCDataCache) error {
			return nil
		},
	}

	if fetcher.name != "test-fetcher" {
		t.Errorf("dataFetcher name = %s, want test-fetcher", fetcher.name)
	}

	if !fetcher.required {
		t.Error("dataFetcher required = false, want true")
	}

	if fetcher.fetch == nil {
		t.Error("dataFetcher fetch function is nil")
	}

	cache := &WNCDataCache{}
	ctx := context.Background()
	err := fetcher.fetch(ctx, cache)
	if err != nil {
		t.Errorf("dataFetcher fetch() error = %v, want nil", err)
	}
}
