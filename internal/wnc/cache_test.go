package wnc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestDataSource_FetchAllData_Success(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(mockServerConfig{
		apCAPWAPSuccess:    true,
		apOperSuccess:      true,
		radioOperSuccess:   true,
		nameMACSuccess:     true,
		rrmMeasSuccess:     true,
		wlanCfgSuccess:     true,
		wlanPolicySuccess:  true,
		wlanPolicyListSucc: true,
		clientCommonSucc:   true,
		clientDCSuccess:    true,
		clientDot11Succ:    true,
		clientSISFSuccess:  true,
		clientTrafficSucc:  true,
		clientMobilitySucc: true,
		radioStatsSuccess:  true,
		radioResetSuccess:  true,
		rrmCoverageSuccess: true,
		apRadarSuccess:     true,
	})
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	data, err := ds.GetCachedData(ctx)
	if err != nil {
		t.Fatalf("GetCachedData() error = %v, want nil", err)
	}

	if len(data.CAPWAPData) != 1 {
		t.Errorf("CAPWAPData length = %d, want 1", len(data.CAPWAPData))
	}
	if len(data.ApOperData) != 1 {
		t.Errorf("ApOperData length = %d, want 1", len(data.ApOperData))
	}
	if len(data.RadioOperData) != 1 {
		t.Errorf("RadioOperData length = %d, want 1", len(data.RadioOperData))
	}
}

func TestDataSource_FetchAllData_RequiredFetcherFailure(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(mockServerConfig{
		apCAPWAPSuccess: false,
	})
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	_, err := ds.GetCachedData(ctx)
	if err == nil {
		t.Fatal("GetCachedData() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "failed to fetch AP CAPWAP") {
		t.Errorf("error message = %v, want to contain 'failed to fetch AP CAPWAP'", err)
	}
}

func TestDataSource_FetchAllData_OptionalFetcherFailure(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(mockServerConfig{
		apCAPWAPSuccess:    true,
		apOperSuccess:      true,
		radioOperSuccess:   true,
		nameMACSuccess:     true,
		rrmMeasSuccess:     true,
		wlanCfgSuccess:     true,
		wlanPolicySuccess:  true,
		wlanPolicyListSucc: true,
		clientCommonSucc:   true,
		clientDCSuccess:    true,
		clientDot11Succ:    true,
		clientSISFSuccess:  true,
		clientTrafficSucc:  true,
		clientMobilitySucc: true,
		radioStatsSuccess:  false,
		radioResetSuccess:  true,
		rrmCoverageSuccess: true,
		apRadarSuccess:     true,
	})
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	data, err := ds.GetCachedData(ctx)
	if err != nil {
		t.Fatalf("GetCachedData() error = %v, want nil (optional failure should not error)", err)
	}

	if len(data.RadioOperStats) != 0 {
		t.Errorf(
			"RadioOperStats length = %d, want 0 (optional fetcher failed)",
			len(data.RadioOperStats),
		)
	}

	if len(data.CAPWAPData) != 1 {
		t.Errorf("CAPWAPData length = %d, want 1", len(data.CAPWAPData))
	}
}

func TestDataSource_FetchAllData_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	_, err := ds.GetCachedData(ctx)
	if err == nil {
		t.Fatal("GetCachedData() error = nil, want error")
	}
}

func TestDataSource_FetchAllData_Timeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       100 * time.Millisecond,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	_, err := ds.GetCachedData(ctx)
	if err == nil {
		t.Fatal("GetCachedData() error = nil, want timeout error")
	}
}

func TestDataSource_FetchAllData_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "invalid-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	_, err := ds.GetCachedData(ctx)
	if err == nil {
		t.Fatal("GetCachedData() error = nil, want unauthorized error")
	}
}

func TestDataSource_FetchAllData_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yang-data+json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	cfg := config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}

	ds := NewDataSource(cfg)
	ctx := context.Background()

	_, err := ds.GetCachedData(ctx)
	if err == nil {
		t.Fatal("GetCachedData() error = nil, want JSON parse error")
	}
}

// mockServerConfig controls which endpoints return success responses.
type mockServerConfig struct {
	apCAPWAPSuccess    bool
	apOperSuccess      bool
	radioOperSuccess   bool
	nameMACSuccess     bool
	rrmMeasSuccess     bool
	wlanCfgSuccess     bool
	wlanPolicySuccess  bool
	wlanPolicyListSucc bool
	clientCommonSucc   bool
	clientDCSuccess    bool
	clientDot11Succ    bool
	clientSISFSuccess  bool
	clientTrafficSucc  bool
	clientMobilitySucc bool
	radioStatsSuccess  bool
	radioResetSuccess  bool
	rrmCoverageSuccess bool
	apRadarSuccess     bool
}

func newMockWNCServer(cfg mockServerConfig) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yang-data+json")

		switch {
		case strings.Contains(r.URL.Path, "capwap-data"):
			if cfg.apCAPWAPSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:capwap-data":` +
						`[{"wtp-mac":"aa:bb:cc:11:22:80","ip-addr":"192.168.255.11","name":"TEST-AP01"}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "oper-data") && !strings.Contains(r.URL.Path, "radio-oper-data"):
			if cfg.apOperSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:oper-data":` +
						`[{"wtp-mac":"aa:bb:cc:11:22:80","radio-id":0}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "radio-oper-data") && !strings.Contains(r.URL.Path, "stats"):
			if cfg.radioOperSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:radio-oper-data":` +
						`[{"wtp-mac":"aa:bb:cc:11:22:80","radio-slot-id":0}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "ap-name-mac-map"):
			if cfg.nameMACSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:ap-name-mac-map":` +
						`[{"wtp-name":"TEST-AP01","eth-mac":"aa:bb:cc:11:22:80"}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "rrm-measurement"):
			if cfg.rrmMeasSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-rrm-oper:rrm-measurement":[{"wtp-mac":"aa:bb:cc:11:22:80"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "wlan-cfg-data/wlan-cfg-entries"):
			if cfg.wlanCfgSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-wlan-cfg:wlan-cfg-entries":{"wlan-cfg-entry":[{"wlan-id":1}]}}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "wlan-policies"):
			if cfg.wlanPolicySuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-wlan-cfg:wlan-policies":` +
						`{"wlan-policy":[{"policy-profile-name":"test-policy"}]}}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "policy-list-entries"):
			if cfg.wlanPolicyListSucc {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-wlan-cfg:policy-list-entries":` +
						`{"policy-list-entry":[{"tag-name":"test-tag"}]}}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "common-oper-data"):
			if cfg.clientCommonSucc {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-client-oper:common-oper-data":[{"client-mac":"aa:bb:cc:11:22:a9"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "dc-info"):
			if cfg.clientDCSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-client-oper:dc-info":[{"client-mac":"aa:bb:cc:11:22:a9"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "dot11-oper-data"):
			if cfg.clientDot11Succ {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-client-oper:dot11-oper-data":` +
						`[{"ms-mac-address":"aa:bb:cc:11:22:a9"}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "sisf-db-mac"):
			if cfg.clientSISFSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-client-oper:sisf-db-mac":[{"mac-addr":"aa:bb:cc:11:22:a9"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "traffic-stats"):
			if cfg.clientTrafficSucc {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-client-oper:traffic-stats":` +
						`[{"ms-mac-address":"aa:bb:cc:11:22:a9"}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "mm-if-client-history"):
			if cfg.clientMobilitySucc {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-client-oper:mm-if-client-history":` +
						`[{"client-mac":"aa:bb:cc:11:22:a9"}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "radio-oper-data/radio-oper-stats"):
			if cfg.radioStatsSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:radio-oper-stats":` +
						`[{"ap-mac":"aa:bb:cc:11:22:80","slot-id":0}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "radio-reset-stats"):
			if cfg.radioResetSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(
					`{"Cisco-IOS-XE-wireless-access-point-oper:radio-reset-stats":` +
						`[{"ap-mac":"aa:bb:cc:11:22:80","radio-id":0}]}`,
				))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "rrm-coverage"):
			if cfg.rrmCoverageSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-rrm-oper:rrm-coverage":[{"wtp-mac":"aa:bb:cc:11:22:80"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		case strings.Contains(r.URL.Path, "ap-dot11-radar-data"):
			if cfg.apRadarSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write(
					[]byte(
						`{"Cisco-IOS-XE-wireless-rrm-oper:ap-dot11-radar-data":[{"wtp-mac":"aa:bb:cc:11:22:80"}]}`,
					),
				)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func extractHostFromURL(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
}
