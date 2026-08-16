package wnc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
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

	ds := newTestDataSource(t, "wnc1.example.internal", 55*time.Second)
	suppressBackgroundRefresh(ds)

	stats := ds.Stats()
	if len(stats.Errors) != len(dataTypeNames) {
		t.Fatalf("Stats().Errors covers %d data types, want %d seeded at zero",
			len(stats.Errors), len(dataTypeNames))
	}
	for _, name := range dataTypeNames {
		if got, ok := stats.Errors[name]; !ok || got != 0 {
			t.Errorf("Stats().Errors[%s] = %d (present=%t), want 0", name, got, ok)
		}
	}

	if stats.Attempted {
		t.Error("Stats().Attempted = true before any refresh, want false")
	}
	if stats.Up {
		t.Error("Stats().Up = true before any refresh, want false")
	}
	if !stats.RefreshedAt.IsZero() {
		t.Error("Stats().RefreshedAt is set before any refresh, want zero")
	}
	if len(stats.Items) != 0 {
		t.Errorf("Stats().Items = %v before any refresh, want empty", stats.Items)
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

func TestSnapshot_SkipsFailedDataType(t *testing.T) {
	t.Parallel()

	src := &mockDataSource{data: &WNCDataCache{
		CAPWAPData:  []ap.CAPWAPData{{WtpMAC: mockAPMAC}},
		FetchErrors: map[string]error{dataAPCAPWAPData: errors.New("fetch failed")},
	}}
	ctx := context.Background()

	if _, err := snapshot(ctx, src, dataAPCAPWAPData); err == nil {
		t.Errorf("snapshot(%s) error = nil, want the recorded fetch error", dataAPCAPWAPData)
	}

	data, err := snapshot(ctx, src, dataAPOperData)
	if err != nil {
		t.Fatalf("snapshot(%s) error = %v, want nil for a data type that did not fail",
			dataAPOperData, err)
	}
	if len(data.CAPWAPData) != 1 {
		t.Errorf("snapshot() CAPWAPData length = %d, want 1", len(data.CAPWAPData))
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
	if cache.FetchErrors != nil {
		t.Error("FetchErrors should be nil when not initialized")
	}
}

func TestDataFetcher_Structure(t *testing.T) {
	t.Parallel()

	fetcher := dataFetcher{
		name: "test-fetcher",
		fetch: func(ctx context.Context, cache *WNCDataCache) (int, error) {
			return 1, nil
		},
	}

	if fetcher.name != "test-fetcher" {
		t.Errorf("dataFetcher name = %s, want test-fetcher", fetcher.name)
	}

	if fetcher.fetch == nil {
		t.Error("dataFetcher fetch function is nil")
	}

	cache := &WNCDataCache{}
	ctx := context.Background()
	count, err := fetcher.fetch(ctx, cache)
	if err != nil {
		t.Errorf("dataFetcher fetch() error = %v, want nil", err)
	}
	if count != 1 {
		t.Errorf("dataFetcher fetch() count = %d, want 1", count)
	}
}

// TestFetchers_MatchDataTypeNames guards the fetch order the deadline truncates
// from: dataTypeNames is what the error and item series are seeded from, so a
// fetcher missing from it would never be reported.
func TestFetchers_MatchDataTypeNames(t *testing.T) {
	t.Parallel()

	ds := newTestDataSource(t, "wnc1.example.internal", 55*time.Second)
	suppressBackgroundRefresh(ds)

	fetchers := ds.fetchers()
	if len(fetchers) != len(dataTypeNames) {
		t.Fatalf("fetchers() returned %d fetchers, want %d", len(fetchers), len(dataTypeNames))
	}
	for i, f := range fetchers {
		if f.name != dataTypeNames[i] {
			t.Errorf("fetchers()[%d].name = %s, want %s", i, f.name, dataTypeNames[i])
		}
	}
	if dataTypeNames[0] != dataAPCAPWAPData {
		t.Errorf("dataTypeNames[0] = %s, want %s: the AP inventory labels every other AP series",
			dataTypeNames[0], dataAPCAPWAPData)
	}
}

func TestMockEndpoints_CoverEveryDataType(t *testing.T) {
	t.Parallel()

	if len(mockEndpoints) != len(dataTypeNames) {
		t.Fatalf("mockEndpoints covers %d data types, want %d", len(mockEndpoints), len(dataTypeNames))
	}

	covered := make(map[string]string, len(mockEndpoints))
	for segment, ep := range mockEndpoints {
		if other, dup := covered[ep.dataType]; dup {
			t.Errorf("segments %q and %q both serve %s", other, segment, ep.dataType)
		}
		covered[ep.dataType] = segment
	}
	for _, name := range dataTypeNames {
		if _, ok := covered[name]; !ok {
			t.Errorf("mockEndpoints has no endpoint for %s", name)
		}
	}
}

func TestDataSource_FetchAllData_Success(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing())
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 55*time.Second)

	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil", err)
	}
	if len(data.FetchErrors) != 0 {
		t.Errorf("FetchErrors = %v, want empty", data.FetchErrors)
	}
	if data.RefreshedAt.IsZero() {
		t.Error("RefreshedAt is zero, want the refresh start time")
	}

	suppressBackgroundRefresh(ds)
	stats := ds.Stats()
	if !stats.Up {
		t.Error("Stats().Up = false, want true")
	}
	if !stats.Attempted {
		t.Error("Stats().Attempted = false, want true after a completed refresh")
	}
	for _, name := range dataTypeNames {
		if stats.Items[name] != 1 {
			t.Errorf("Stats().Items[%s] = %d, want 1", name, stats.Items[name])
		}
		if stats.Errors[name] != 0 {
			t.Errorf("Stats().Errors[%s] = %d, want 0", name, stats.Errors[name])
		}
	}
}

func TestDataSource_FetchAllData_PartialFailure(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing(dataAPCAPWAPData))
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 55*time.Second)

	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil: a partial failure must not discard the snapshot", err)
	}
	if data.FetchErrors[dataAPCAPWAPData] == nil {
		t.Errorf("FetchErrors[%s] = nil, want the fetch error so collectors skip its series", dataAPCAPWAPData)
	}
	if err := data.FetchErrors[dataAPRadioOperData]; err != nil {
		t.Errorf("FetchErrors[%s] = %v, want nil", dataAPRadioOperData, err)
	}
	if len(data.CAPWAPData) != 0 {
		t.Errorf("CAPWAPData length = %d, want 0 because the fetch failed", len(data.CAPWAPData))
	}
	if len(data.RadioOperData) != 1 {
		t.Errorf("RadioOperData length = %d, want 1", len(data.RadioOperData))
	}

	suppressBackgroundRefresh(ds)
	stats := ds.Stats()
	if !stats.Up {
		t.Error("Stats().Up = false, want true while at least one data type succeeds")
	}
	if stats.Errors[dataAPCAPWAPData] != 1 {
		t.Errorf("Stats().Errors[%s] = %d, want 1", dataAPCAPWAPData, stats.Errors[dataAPCAPWAPData])
	}
	if _, ok := stats.Items[dataAPCAPWAPData]; ok {
		t.Errorf("Stats().Items records %s, want it absent so wnc_refresh_items reports no zero",
			dataAPCAPWAPData)
	}
	if stats.Items[dataAPRadioOperData] != 1 {
		t.Errorf("Stats().Items[%s] = %d, want 1", dataAPRadioOperData, stats.Items[dataAPRadioOperData])
	}
}

func TestDataSource_FetchAllData_AllFail(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing(dataTypeNames...))
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 55*time.Second)

	data, err := ds.fetchAllData(context.Background())
	if err == nil {
		t.Fatal("fetchAllData() error = nil, want error when every data type fails")
	}
	if data != nil {
		t.Error("fetchAllData() returned a snapshot, want nil so the previous one keeps being served")
	}
	if errors.Unwrap(err) == nil {
		t.Errorf("fetchAllData() error = %q, want the last fetch error wrapped so the cause is diagnosable",
			err)
	}
	if want := fmt.Sprintf("all %d", len(dataTypeNames)); !strings.Contains(err.Error(), want) {
		t.Errorf("fetchAllData() error = %q, want it to report %q", err, want)
	}

	suppressBackgroundRefresh(ds)
	stats := ds.Stats()
	if stats.Up {
		t.Error("Stats().Up = true, want false when every data type failed")
	}
	if len(stats.Items) != 0 {
		t.Errorf("Stats().Items = %v, want empty when every data type failed", stats.Items)
	}
	for _, name := range dataTypeNames {
		if stats.Errors[name] != 1 {
			t.Errorf("Stats().Errors[%s] = %d, want 1", name, stats.Errors[name])
		}
	}
}

// TestDataSource_FetchAllData_TruncatedByDeadline answers the first data type
// immediately so the snapshot survives, then stalls the rest. Failing the first
// one instead would make every data type fail and discard the snapshot, leaving
// the truncation bookkeeping unverifiable.
func TestDataSource_FetchAllData_TruncatedByDeadline(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yang-data+json")

		ep, ok := mockEndpoints[path.Base(r.URL.Path)]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if ep.dataType != dataAPCAPWAPData {
			<-r.Context().Done()
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ep.body))
	}))
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 100*time.Millisecond)

	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil while the first data type succeeds", err)
	}
	if err := data.FetchErrors[dataAPCAPWAPData]; err != nil {
		t.Errorf("FetchErrors[%s] = %v, want nil", dataAPCAPWAPData, err)
	}
	if len(data.CAPWAPData) != 1 {
		t.Errorf("CAPWAPData length = %d, want 1", len(data.CAPWAPData))
	}

	last := dataTypeNames[len(dataTypeNames)-1]
	if !errors.Is(data.FetchErrors[last], context.DeadlineExceeded) {
		t.Errorf("FetchErrors[%s] = %v, want a deadline error: a data type the deadline never reached "+
			"must not read as a successful empty fetch", last, data.FetchErrors[last])
	}

	suppressBackgroundRefresh(ds)
	stats := ds.Stats()
	if stats.Errors[last] != 1 {
		t.Errorf("Stats().Errors[%s] = %d, want 1 so truncation is visible in wnc_refresh_errors_total",
			last, stats.Errors[last])
	}
	if !stats.Up {
		t.Error("Stats().Up = false, want true while the first data type succeeded")
	}
}

func TestDataSource_FetchAllData_TotalFailureVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "Unauthorized",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized"}`))
			},
		},
		{
			name: "Internal server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal server error"}`))
			},
		},
		{
			name: "Invalid JSON",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/yang-data+json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid json}`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(tt.handler)
			defer server.Close()

			ds := newTestDataSource(t, server.URL, 55*time.Second)

			if _, err := ds.fetchAllData(context.Background()); err == nil {
				t.Fatal("fetchAllData() error = nil, want error when every data type fails")
			}
		})
	}
}

// TestDataSource_GetCachedData_ColdStart pins the property the whole design
// exists for: the first scrape returns immediately instead of waiting for the
// controller.
func TestDataSource_GetCachedData_ColdStart(t *testing.T) {
	t.Parallel()

	ds := newTestDataSource(t, "wnc1.example.internal", 55*time.Second)
	suppressBackgroundRefresh(ds)

	data, err := ds.GetCachedData(context.Background())
	if !errors.Is(err, errNoSnapshot) {
		t.Fatalf("GetCachedData() error = %v, want errNoSnapshot", err)
	}
	if data != nil {
		t.Error("GetCachedData() returned data before the first refresh completed")
	}
}

func TestDataSource_GetCachedData_ServesAfterRefresh(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing())
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 55*time.Second)
	ds.refresher.refreshOnce(context.Background())

	data, err := ds.GetCachedData(context.Background())
	if err != nil {
		t.Fatalf("GetCachedData() error = %v, want nil after a completed refresh", err)
	}
	if len(data.CAPWAPData) != 1 {
		t.Errorf("CAPWAPData length = %d, want 1", len(data.CAPWAPData))
	}

	stats := ds.Stats()
	if stats.RefreshedAt != data.RefreshedAt {
		t.Errorf("Stats().RefreshedAt = %v, want the snapshot's %v", stats.RefreshedAt, data.RefreshedAt)
	}
}

func TestDataSource_GetCachedData_WithheldAfterConsecutiveFailures(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing())
	defer server.Close()

	ds := newTestDataSource(t, server.URL, 55*time.Second)
	ds.refresher.refreshOnce(context.Background())
	suppressBackgroundRefresh(ds)

	for range maxConsecutiveRefreshFailures - 1 {
		ds.onRefreshDone(errors.New("refresh failed"), time.Second)
	}
	if _, err := ds.GetCachedData(context.Background()); err != nil {
		t.Fatalf("GetCachedData() error = %v, want nil below the failure threshold", err)
	}

	ds.onRefreshDone(errors.New("refresh failed"), time.Second)
	if _, err := ds.GetCachedData(context.Background()); !errors.Is(err, errSnapshotWithheld) {
		t.Fatalf("GetCachedData() error = %v, want errSnapshotWithheld", err)
	}

	// Freshness has to stay observable exactly while data is being withheld,
	// otherwise the staleness alert has nothing to fire on.
	if ds.Stats().RefreshedAt.IsZero() {
		t.Error("Stats().RefreshedAt is zero while the snapshot is withheld, want it observable")
	}

	ds.onRefreshDone(nil, time.Second)
	if _, err := ds.GetCachedData(context.Background()); err != nil {
		t.Errorf("GetCachedData() error = %v, want nil after a successful refresh clears the count", err)
	}
}

func TestDataSource_OnRefreshDone_PanicRecordsEveryDataType(t *testing.T) {
	t.Parallel()

	ds := newTestDataSource(t, "wnc1.example.internal", 55*time.Second)
	suppressBackgroundRefresh(ds)

	ds.onRefreshDone(fmt.Errorf("run: %w", errRefreshPanicked), 2*time.Second)

	stats := ds.Stats()
	if stats.Up {
		t.Error("Stats().Up = true after a panicked refresh, want false")
	}
	if !stats.Attempted {
		t.Error("Stats().Attempted = false after a panicked refresh, want true")
	}
	if stats.Duration != 2*time.Second {
		t.Errorf("Stats().Duration = %v, want 2s", stats.Duration)
	}
	for _, name := range dataTypeNames {
		if stats.Errors[name] != 1 {
			t.Errorf("Stats().Errors[%s] = %d, want 1", name, stats.Errors[name])
		}
	}
}

func TestConfig_RefreshDeadlineExceedsPerRequestTimeout(t *testing.T) {
	t.Parallel()

	// The per-request timeout defaults to the cache TTL, so a one-times-TTL
	// deadline would give every sequential request the budget of one.
	if refreshDeadlineFactor < 2 {
		t.Errorf("refreshDeadlineFactor = %d, want at least 2", refreshDeadlineFactor)
	}
	if got := refreshDeadlineFactor * config.DefaultWNCCacheTTL; got <= config.DefaultWNCTimeout {
		t.Errorf("default refresh deadline = %v, want more than the per-request timeout %v",
			got, config.DefaultWNCTimeout)
	}
}

// RESTCONF module names the mock replies are keyed by.
const (
	mockAPGlobalOperModule   = "Cisco-IOS-XE-wireless-ap-global-oper"
	mockClientGlobalModule   = "Cisco-IOS-XE-wireless-client-global-oper"
	mockDeviceHardwareModule = "Cisco-IOS-XE-device-hardware-oper"
	mockAPOperModule         = "Cisco-IOS-XE-wireless-access-point-oper"
	mockClientOperModule     = "Cisco-IOS-XE-wireless-client-oper"
	mockRRMOperModule        = "Cisco-IOS-XE-wireless-rrm-oper"
	mockRRMGlobalOperModule  = "Cisco-IOS-XE-wireless-rrm-global-oper"
	mockWLANCfgModule        = "Cisco-IOS-XE-wireless-wlan-cfg"
)

const (
	mockAPMAC     = "aa:bb:cc:11:22:80"
	mockClientMAC = "aa:bb:cc:11:22:a9"
)

// mockEndpoint is the reply the mock server serves for one data type.
type mockEndpoint struct {
	dataType string
	body     string
}

// mockEndpoints maps the final RESTCONF path segment to the data type it serves.
// Those segments are unique across every data type, so the handler matches
// them exactly: substring matching routed every client and RRM request into the AP
// branch, which left most of the per-data-type failure switches unexercised.
var mockEndpoints = map[string]mockEndpoint{
	"capwap-data": {dataAPCAPWAPData, mockList(mockAPOperModule, "capwap-data",
		`{"wtp-mac":"`+mockAPMAC+`","ip-addr":"192.168.255.11","name":"TEST-AP01"}`)},
	"oper-data": {dataAPOperData, mockList(mockAPOperModule, "oper-data",
		`{"wtp-mac":"`+mockAPMAC+`","radio-id":0}`)},
	"radio-oper-data": {dataAPRadioOperData, mockList(mockAPOperModule, "radio-oper-data",
		`{"wtp-mac":"`+mockAPMAC+`","radio-slot-id":0}`)},
	"ap-name-mac-map": {dataAPNameMACMap, mockList(mockAPOperModule, "ap-name-mac-map",
		`{"wtp-name":"TEST-AP01","eth-mac":"`+mockAPMAC+`"}`)},
	"radio-oper-stats": {dataAPRadioOperStats, mockList(mockAPOperModule, "radio-oper-stats",
		`{"ap-mac":"`+mockAPMAC+`","slot-id":0}`)},
	"radio-reset-stats": {dataAPRadioResetStats, mockList(mockAPOperModule, "radio-reset-stats",
		`{"ap-mac":"`+mockAPMAC+`","radio-id":0}`)},
	"ap-join-stats": {dataAPJoinStats, mockList(mockAPGlobalOperModule, "ap-join-stats",
		`{"wtp-mac":"`+mockAPMAC+`","ap-join-info":{"ap-name":"TEST-AP01","is-joined":true}}`)},
	"common-oper-data": {dataClientCommonOperData, mockList(mockClientOperModule, "common-oper-data",
		`{"client-mac":"`+mockClientMAC+`"}`)},
	"dc-info": {dataClientDCInfo, mockList(mockClientOperModule, "dc-info",
		`{"client-mac":"`+mockClientMAC+`"}`)},
	"dot11-oper-data": {dataClientDot11OperData, mockList(mockClientOperModule, "dot11-oper-data",
		`{"ms-mac-address":"`+mockClientMAC+`"}`)},
	"sisf-db-mac": {dataClientSISFDBMac, mockList(mockClientOperModule, "sisf-db-mac",
		`{"mac-addr":"`+mockClientMAC+`"}`)},
	"traffic-stats": {dataClientTrafficStats, mockList(mockClientOperModule, "traffic-stats",
		`{"ms-mac-address":"`+mockClientMAC+`"}`)},
	"mm-if-client-history": {dataClientMMIFHistory, mockList(mockClientOperModule, "mm-if-client-history",
		`{"client-mac":"`+mockClientMAC+`"}`)},
	"rrm-measurement": {dataRRMMeasurement, mockList(mockRRMOperModule, "rrm-measurement",
		`{"wtp-mac":"`+mockAPMAC+`"}`)},
	"rrm-coverage": {dataRRMCoverage, mockList(mockRRMGlobalOperModule, "rrm-coverage",
		`{"wtp-mac":"`+mockAPMAC+`","radio-slot-id":0}`)},
	"ap-dot11-radar-data": {dataRRMAPDot11RadarData, mockList(mockRRMOperModule, "ap-dot11-radar-data",
		`{"wtp-mac":"`+mockAPMAC+`"}`)},
	// The two raw reads answer with the node itself as the only key rather than with a
	// list, which is what mockContainer wraps and mockList cannot.
	"boot-time": {dataControllerBootTime, mockContainer(mockDeviceHardwareModule, "boot-time",
		`"2026-01-01T00:00:00+00:00"`)},
	"co-client-del-reason": {dataCoClientDelReason, mockContainer(mockClientGlobalModule, "co-client-del-reason",
		// One leaf, because every mock here answers with exactly one item.
		`{"ap-delete":24665}`)},
	"wlan-cfg-entries": {dataWLANCfgEntries, mockNestedList(mockWLANCfgModule, "wlan-cfg-entries",
		"wlan-cfg-entry", `{"wlan-id":1}`)},
	"wlan-policies": {dataWLANPolicies, mockNestedList(mockWLANCfgModule, "wlan-policies",
		"wlan-policy", `{"policy-profile-name":"test-policy"}`)},
	"policy-list-entries": {dataWLANPolicyListEntries, mockNestedList(mockWLANCfgModule, "policy-list-entries",
		"policy-list-entry", `{"tag-name":"test-tag"}`)},
}

// mockList wraps one entry in a module-qualified YANG list.
func mockList(module, container, entry string) string {
	return `{"` + module + `:` + container + `":[` + entry + `]}`
}

// mockContainer wraps the value of a leaf or container read. Such a read answers with
// the node itself as its only key, where a list read answers with the list.
func mockContainer(module, node, value string) string {
	return `{"` + module + `:` + node + `":` + value + `}`
}

// mockNestedList wraps one entry in a container that nests the list, which is how
// the WLAN configuration subtree is shaped.
func mockNestedList(module, container, list, entry string) string {
	return `{"` + module + `:` + container + `":{"` + list + `":[` + entry + `]}}`
}

// mockServerConfig selects which data types the mock server answers with HTTP 500.
type mockServerConfig struct {
	fail map[string]bool
}

// failing returns a config where exactly the named data types fail.
func failing(dataTypes ...string) mockServerConfig {
	fail := make(map[string]bool, len(dataTypes))
	for _, name := range dataTypes {
		fail[name] = true
	}
	return mockServerConfig{fail: fail}
}

func newMockWNCServer(cfg mockServerConfig) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yang-data+json")

		ep, ok := mockEndpoints[path.Base(r.URL.Path)]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if cfg.fail[ep.dataType] {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ep.body))
	}))
}

// newTestDataSource returns the concrete data source so tests can drive
// fetchAllData and the refresher directly instead of racing the background one.
func newTestDataSource(t *testing.T, controllerURL string, ttl time.Duration) *dataSource {
	t.Helper()

	ds, ok := NewDataSource(config.WNC{
		Controller:    extractHostFromURL(controllerURL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      ttl,
	}, allModules()).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}
	return ds
}

// allModules enables every collector module, which is what makes a refresh walk
// every data type. The assertions over dataTypeNames describe that deployment
// rather than the default one, where no module is enabled.
func allModules() config.Collectors {
	return config.Collectors{
		AP: config.APCollectorModules{
			General: true, Radio: true, Traffic: true, Errors: true, Join: true, Info: true,
		},
		Client: config.ClientCollectorModules{
			General: true, Radio: true, Traffic: true, Errors: true, Info: true,
		},
		WLAN: config.WLANCollectorModules{
			General: true, Traffic: true, Config: true, Info: true,
		},
		Controller: config.ControllerCollectorModules{General: true},
	}
}

// suppressBackgroundRefresh keeps Stats and GetCachedData from starting a refresh
// that would race assertions on statistics a test produced synchronously.
func suppressBackgroundRefresh(ds *dataSource) {
	ds.refresher.inflight.Store(true)
}

func extractHostFromURL(url string) string {
	return strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
}
