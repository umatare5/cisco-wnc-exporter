// Package wnc provides WNC data access and caching.
package wnc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	wnc "github.com/umatare5/cisco-ios-xe-wireless-go"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// Data type identifiers. The values mirror the RESTCONF container names with
// hyphens replaced by underscores, prefixed by their subject when the container
// name does not already start with it. They are the `data` label values, the
// RefreshStats map keys and the FetchErrors keys, so they must not drift.
const (
	// gosec G101 matches the "pw" inside CAPWAP, not a credential.
	dataAPCAPWAPData          = "ap_capwap_data" //nolint:gosec
	dataAPOperData            = "ap_oper_data"
	dataAPRadioOperData       = "ap_radio_oper_data"
	dataAPNameMACMap          = "ap_name_mac_map"
	dataAPRadioOperStats      = "ap_radio_oper_stats"
	dataAPRadioResetStats     = "ap_radio_reset_stats"
	dataClientCommonOperData  = "client_common_oper_data"
	dataClientDCInfo          = "client_dc_info"
	dataClientDot11OperData   = "client_dot11_oper_data"
	dataClientSISFDBMac       = "client_sisf_db_mac"
	dataClientTrafficStats    = "client_traffic_stats"
	dataClientMMIFHistory     = "client_mm_if_client_history"
	dataRRMMeasurement        = "rrm_measurement"
	dataRRMCoverage           = "rrm_coverage"
	dataRRMAPDot11RadarData   = "rrm_ap_dot11_radar_data"
	dataWLANCfgEntries        = "wlan_cfg_entries"
	dataWLANPolicies          = "wlan_policies"
	dataWLANPolicyListEntries = "wlan_policy_list_entries"
)

// refreshDeadlineFactor bounds a whole refresh at this multiple of the cache TTL.
// One times the TTL is not enough: the per-request timeout defaults to the same
// value as the TTL, so a single slow request would consume the entire budget.
const refreshDeadlineFactor = 2

// maxConsecutiveRefreshFailures is how many refreshes may fail in a row before
// data series are withheld. Serving a frozen snapshot indefinitely defeats
// Prometheus staleness marking and makes rate() report a false zero.
const maxConsecutiveRefreshFailures = 3

// WNCDataCache contains operational data from WNC.
// All fields are read-only once the snapshot is published.
type WNCDataCache struct {
	CAPWAPData      []ap.CAPWAPData
	ApOperData      []ap.OperData
	RadioOperData   []ap.RadioOperData
	RadioOperStats  []ap.RadioOperStats
	RadioResetStats []ap.RadioResetStats
	NameMACMaps     []ap.ApNameMACMap

	CommonOperData    []client.CommonOperData
	DCInfo            []client.DcInfo
	Dot11OperData     []client.Dot11OperData
	SisfDBMac         []client.SisfDBMac
	TrafficStats      []client.TrafficStats
	MmIfClientHistory []client.MmIfClientHistory

	// RRM data
	RRMMeasurements  []rrm.RRMMeasurement
	RRMCoverage      []rrm.RRMCoverage
	ApDot11RadarData []rrm.ApDot11RadarData

	// WLAN data
	WLANConfigEntries     []wlan.WlanCfgEntry
	WLANPolicies          []wlan.WlanPolicy
	WLANPolicyListEntries []wlan.PolicyListEntry

	// FetchErrors records the failure per data type so callers skip derived
	// metrics instead of publishing a fabricated zero.
	FetchErrors map[string]error
	// RefreshedAt is when the refresh that produced this snapshot started, which
	// bounds the age of every datum it carries.
	RefreshedAt time.Time
}

// DataSource provides cached access to WNC operational data.
type DataSource interface {
	GetCachedData(ctx context.Context) (*WNCDataCache, error)
}

// RefreshStats reports the outcome of WNC data refreshes.
type RefreshStats struct {
	// Up reports whether the last completed refresh reached the controller.
	// It is not a claim about data completeness or about the current scrape.
	Up bool
	// Attempted is false until the first refresh completes.
	Attempted bool
	// RefreshedAt is the start time of the refresh that produced the snapshot.
	RefreshedAt time.Time
	// Duration is how long the last refresh attempt took.
	Duration time.Duration
	// Errors counts failures per data type since process start.
	Errors map[string]int
	// Items counts what each data type returned, recorded on success only.
	Items map[string]int
	// DefaultsFallbacks counts WLAN configuration fetches that asked for the
	// values in force and had to settle for a plain read, since process start.
	DefaultsFallbacks int64
}

// StatsProvider is implemented by data sources that report refresh statistics.
type StatsProvider interface {
	Stats() RefreshStats
}

// dataSource implements DataSource with caching to minimize WNC requests.
type dataSource struct {
	client    *wnc.Client
	refresher *refresher
	cacheTTL  time.Duration

	// names lists the data types the enabled modules read, in fetch order. It is
	// the seeded `data` label set and the denominator every refresh outcome is
	// judged against, so a refresh that failed everything those modules need
	// reports wnc_up 0 however many other data types exist.
	names []string

	// failures counts consecutive failed refreshes. It must stay atomic rather
	// than mu-protected: Stats holds mu and calls serving, and sync.Mutex is not
	// reentrant.
	failures atomic.Int64

	// defaultsFallbacks counts WLAN configuration fetches that fell back to a
	// plain read. It is written by the refresh goroutine and read by a scrape.
	defaultsFallbacks atomic.Int64

	mu        sync.Mutex
	errors    map[string]int
	items     map[string]int
	duration  time.Duration
	up        bool
	attempted bool
}

// dataFetcher defines a WNC API call that can fail. It reports the item count on success.
type dataFetcher struct {
	name  string
	fetch func(ctx context.Context, c *WNCDataCache) (int, error)
}

// NewDataSource creates a new shared data source. It reads only the data types the
// enabled modules need, so enabling one module does not poll the controller for the
// data the others would have read.
func NewDataSource(cfg config.WNC, modules config.Collectors) DataSource {
	names := requiredDataTypes(modules)
	s := &dataSource{
		client:   createWNCClient(cfg),
		cacheTTL: cfg.CacheTTL,
		names:    names,
		errors:   make(map[string]int, len(names)),
	}

	// Seed every data type so the error series exist on the first scrape, which
	// is the scrape most likely to be reporting a failure. Seeding the required
	// set is what lets an operator tell a failed fetch from one never requested.
	for _, name := range names {
		s.errors[name] = 0
	}

	s.refresher = newRefresher(cfg.CacheTTL, s.fetchAllData, s.onRefreshDone)
	return s
}

// snapshot returns the cached data unless the given data type failed to fetch.
// Reading FetchErrors from a nil map yields nil, so snapshots built in tests
// need no guard.
func snapshot(ctx context.Context, src DataSource, id string) (*WNCDataCache, error) {
	data, err := src.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if err := data.FetchErrors[id]; err != nil {
		return nil, err
	}
	return data, nil
}

// GetCachedData returns the current snapshot. It never blocks on the controller:
// the refresh runs in the background and ctx is not used for it.
func (s *dataSource) GetCachedData(_ context.Context) (*WNCDataCache, error) {
	data, withheld := s.serving()
	if data == nil {
		return nil, errNoSnapshot
	}
	if withheld {
		return nil, fmt.Errorf("%w after %d consecutive failed refreshes",
			errSnapshotWithheld, maxConsecutiveRefreshFailures)
	}
	return data, nil
}

// serving returns the current snapshot and whether it must be withheld from data
// collectors. Stats deliberately ignores withheld: freshness has to stay
// observable exactly while data is being withheld.
func (s *dataSource) serving() (snap *WNCDataCache, withheld bool) {
	snap = s.refresher.get()
	return snap, s.failures.Load() >= maxConsecutiveRefreshFailures
}

// Stats returns a snapshot of the refresh statistics.
func (s *dataSource) Stats() RefreshStats {
	snap, _ := s.serving()

	s.mu.Lock()
	defer s.mu.Unlock()

	st := RefreshStats{
		Up:        s.up,
		Attempted: s.attempted,
		Duration:  s.duration,
		Errors:    maps.Clone(s.errors),
		Items:     maps.Clone(s.items),
	}
	st.DefaultsFallbacks = s.defaultsFallbacks.Load()
	if snap != nil {
		st.RefreshedAt = snap.RefreshedAt
	}
	return st
}

// onRefreshDone tracks the consecutive failure count and, for a recovered panic,
// publishes the statistics that fetchAllData never got to record.
func (s *dataSource) onRefreshDone(err error, elapsed time.Duration) {
	if err == nil {
		s.failures.Store(0)
		return
	}

	s.failures.Add(1)
	if errors.Is(err, errRefreshPanicked) {
		s.recordRefresh(nil, s.names, elapsed)
	}
}

// recordRefresh publishes the outcome of a refresh attempt for the collector.
func (s *dataSource) recordRefresh(items map[string]int, failures []string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, name := range failures {
		s.errors[name]++
	}

	s.items = items
	s.duration = duration
	s.attempted = true
	// A source with no enabled module needs nothing, so nothing failing is up.
	s.up = len(failures) == 0 || len(failures) < len(s.names)
}

func (s *dataSource) fetchAllData(ctx context.Context) (*WNCDataCache, error) {
	ctx, cancel := context.WithTimeout(ctx, refreshDeadlineFactor*s.cacheTTL)
	defer cancel()

	start := time.Now()
	data := &WNCDataCache{
		FetchErrors: make(map[string]error),
		RefreshedAt: start,
	}

	items := make(map[string]int, len(s.names))
	failures := make([]string, 0, len(s.names))
	var lastErr error

	for _, f := range s.fetchers() {
		// A data type no enabled module reads is marked rather than fetched, and it
		// is neither an item nor a failure: recording it as failed would raise the
		// error counter for a request nobody wanted. The mark is what makes a
		// collector reading it anyway omit its series, because an unmarked skip
		// returns the snapshot and the collector takes the empty slice for data.
		if !slices.Contains(s.names, f.name) {
			data.FetchErrors[f.name] = fmt.Errorf("%s: %w", f.name, errDataTypeNotRequested)
			continue
		}

		// A data type the deadline never reached must be recorded as failed.
		// Leaving it out would make FetchErrors report it as a successful empty
		// fetch, and the collectors would publish fabricated zeros for it.
		if ctxErr := ctx.Err(); ctxErr != nil {
			err := fmt.Errorf("%s: %w", f.name, ctxErr)
			failures = append(failures, f.name)
			data.FetchErrors[f.name] = err
			lastErr = err
			continue
		}

		fetchStart := time.Now()
		count, err := f.fetch(ctx, data)
		if err != nil {
			failures = append(failures, f.name)
			data.FetchErrors[f.name] = err
			lastErr = err
		} else {
			items[f.name] = count
		}
		slog.Debug("data fetch completed", "data", f.name,
			"count", count, "duration", time.Since(fetchStart))
	}

	s.recordRefresh(items, failures, time.Since(start))

	if len(failures) > 0 {
		slog.Info("WNC data refreshed with failures",
			"failed_data", failures, "total", len(s.names),
			"duration", time.Since(start))
	}

	// The first condition is not redundant: a source with no enabled module has
	// nothing to fail, and wrapping a nil error would print a formatting verb.
	if len(failures) > 0 && len(failures) == len(s.names) {
		return nil, fmt.Errorf("all %d WNC data fetches failed: %w", len(s.names), lastErr)
	}

	return data, nil
}

// readEffective asks the controller to report the value in force on every leaf,
// so a leaf left at its default is reported instead of omitted. A controller
// that rejects the parameter answers 400, and the plain re-read then keeps the
// WLAN series alive at the cost of reading an omitted leaf as its zero value.
// A controller that accepts the parameter and ignores it answers 200, so the
// fallback counter cannot report that case.
func readEffective[T any](
	ctx context.Context,
	fallbacks *atomic.Int64,
	list func(context.Context, ...wnc.GetOption) (*T, error),
) (*T, error) {
	data, err := list(ctx, wnc.WithDefaults(wnc.ReportAll))
	if err == nil {
		return data, nil
	}

	var apiErr *wnc.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		return nil, err
	}

	fallbacks.Add(1)
	slog.Warn("controller rejected the request for values in force, re-reading without it",
		"status", apiErr.StatusCode)

	return list(ctx)
}
