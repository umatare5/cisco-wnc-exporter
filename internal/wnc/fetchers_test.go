package wnc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// withDefaultsParam is the query the exporter sends to ask for the value in force
// on every leaf. The tests below read it off the wire rather than trusting the
// call site, because a missing parameter is invisible in the decoded result.
const withDefaultsParam = "with-defaults=report-all"

// queryRecorder answers like the mock WNC server and records the query string of
// every request, keyed by the last path segment.
type queryRecorder struct {
	mu      sync.Mutex
	queries map[string][]string
	// reject names the data types that answer 400 for a request carrying a query.
	reject map[string]bool
	// status is the code those data types answer with. 400 unless a test says otherwise.
	status int
}

func newQueryRecorder(reject ...string) *queryRecorder {
	rec := &queryRecorder{
		queries: make(map[string][]string),
		reject:  make(map[string]bool, len(reject)),
		status:  http.StatusBadRequest,
	}
	for _, name := range reject {
		rec.reject[name] = true
	}
	return rec
}

func (r *queryRecorder) server() *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		segment := path.Base(req.URL.Path)

		r.mu.Lock()
		r.queries[segment] = append(r.queries[segment], req.URL.RawQuery)
		reject := r.reject[segment] && req.URL.RawQuery != ""
		status := r.status
		r.mu.Unlock()

		w.Header().Set("Content-Type", "application/yang-data+json")

		ep, ok := mockEndpoints[segment]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if reject {
			w.WriteHeader(status)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ep.body))
	}))
}

func (r *queryRecorder) get(segment string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.queries[segment]...)
}

// TestFetchers_AskForValuesInForceOnConfigRoutesOnly pins which routes carry the
// parameter. A leaf left at its default is omitted from a plain read and decodes
// to zero, so the two WLAN config routes must ask for the value in force — and
// the operational routes must not, because the parameter buys nothing there and
// every one of them is read on every refresh.
func TestFetchers_AskForValuesInForceOnConfigRoutesOnly(t *testing.T) {
	t.Parallel()

	rec := newQueryRecorder()
	server := rec.server()
	defer server.Close()

	ds := newTestDataSource(t, server.URL, time.Minute)
	if _, err := ds.fetchAllData(context.Background()); err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil", err)
	}

	for _, segment := range []string{"wlan-cfg-entries", "wlan-policies"} {
		queries := rec.get(segment)
		if len(queries) != 1 {
			t.Fatalf("%s was requested %d times, want 1", segment, len(queries))
		}
		if queries[0] != withDefaultsParam {
			t.Errorf("%s query = %q, want %q", segment, queries[0], withDefaultsParam)
		}
	}

	// policy-list-entries is a config route too, but the exporter reads only the
	// two profile names off it, and neither is defaultable.
	for segment, queries := range rec.queries {
		if segment == "wlan-cfg-entries" || segment == "wlan-policies" {
			continue
		}
		for _, query := range queries {
			if query != "" {
				t.Errorf("%s query = %q, want no query", segment, query)
			}
		}
	}
}

// TestFetchers_FallBackToPlainReadOnRejection covers the controller that rejects
// the parameter. Without the fallback the whole WLAN family disappears, including
// the series that do not read a defaultable leaf.
func TestFetchers_FallBackToPlainReadOnRejection(t *testing.T) {
	t.Parallel()

	rec := newQueryRecorder("wlan-cfg-entries", "wlan-policies")
	server := rec.server()
	defer server.Close()

	ds := newTestDataSource(t, server.URL, time.Minute)
	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil", err)
	}

	for _, id := range []string{dataWLANCfgEntries, dataWLANPolicies} {
		if fetchErr := data.FetchErrors[id]; fetchErr != nil {
			t.Errorf("FetchErrors[%s] = %v, want nil after the plain re-read", id, fetchErr)
		}
	}
	if len(data.WLANConfigEntries) != 1 {
		t.Errorf("WLANConfigEntries has %d entries, want 1 from the plain re-read",
			len(data.WLANConfigEntries))
	}
	if len(data.WLANPolicies) != 1 {
		t.Errorf("WLANPolicies has %d entries, want 1 from the plain re-read", len(data.WLANPolicies))
	}

	for _, segment := range []string{"wlan-cfg-entries", "wlan-policies"} {
		queries := rec.get(segment)
		if len(queries) != 2 {
			t.Fatalf("%s was requested %d times, want 2: one rejected and one plain",
				segment, len(queries))
		}
		if queries[0] != withDefaultsParam || queries[1] != "" {
			t.Errorf("%s queries = %q, want [%q, \"\"]", segment, queries, withDefaultsParam)
		}
	}

	if got := ds.Stats().DefaultsFallbacks; got != 2 {
		t.Errorf("DefaultsFallbacks = %d, want 2", got)
	}
}

// TestFetchers_NoFallBackOnOtherStatus keeps the retry narrow. Any status other
// than 400 is a fetch failure, and retrying it would double the request rate
// against a controller that is already answering badly.
func TestFetchers_NoFallBackOnOtherStatus(t *testing.T) {
	t.Parallel()

	rec := newQueryRecorder("wlan-cfg-entries")
	rec.status = http.StatusInternalServerError
	server := rec.server()
	defer server.Close()

	ds := newTestDataSource(t, server.URL, time.Minute)
	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil: one failed data type is not a failed refresh", err)
	}

	if data.FetchErrors[dataWLANCfgEntries] == nil {
		t.Error("FetchErrors[wlan_cfg_entries] is nil, want the fetch error recorded")
	}
	if queries := rec.get("wlan-cfg-entries"); len(queries) != 1 {
		t.Errorf("wlan-cfg-entries was requested %d times, want 1: no retry on a non-400", len(queries))
	}
	if got := ds.Stats().DefaultsFallbacks; got != 0 {
		t.Errorf("DefaultsFallbacks = %d, want 0", got)
	}
}

// TestRequiredDataTypes_MatchesWhatTheModulesRead pins the table the fetch loop is
// gated on. A data type missing from it is never fetched, and its collector then
// takes an empty slice for a successful empty fetch — the map builders return a
// non-nil empty map, so the nil guards in the collectors do not catch it. Every
// module's route set is therefore asserted rather than sampled.
func TestRequiredDataTypes_MatchesWhatTheModulesRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		modules config.Collectors
		want    []string
	}{
		{
			"AP general reads the inventory, the radios and the AP oper data",
			config.Collectors{AP: config.APCollectorModules{General: true}},
			[]string{dataAPCAPWAPData, dataAPOperData, dataAPRadioOperData},
		},
		{
			"AP radio reads client data for the per-radio client count",
			config.Collectors{AP: config.APCollectorModules{Radio: true}},
			[]string{
				dataAPCAPWAPData, dataAPRadioOperData, dataAPNameMACMap,
				dataRRMMeasurement, dataClientCommonOperData,
			},
		},
		{
			"AP info reads only the two the AP collector fetches unconditionally",
			config.Collectors{AP: config.APCollectorModules{Info: true}},
			[]string{dataAPCAPWAPData, dataAPRadioOperData},
		},
		{
			"AP errors reads the reset stats and both RRM error routes",
			config.Collectors{AP: config.APCollectorModules{Errors: true}},
			[]string{
				dataAPCAPWAPData, dataAPRadioOperData, dataAPRadioOperStats,
				dataAPRadioResetStats, dataRRMCoverage, dataRRMAPDot11RadarData,
			},
		},
		{
			"WLAN config reads all three configuration routes",
			config.Collectors{WLAN: config.WLANCollectorModules{Config: true}},
			[]string{dataWLANCfgEntries, dataWLANPolicies, dataWLANPolicyListEntries},
		},
		{
			"WLAN traffic reads client data instead of the policy routes",
			config.Collectors{WLAN: config.WLANCollectorModules{Traffic: true}},
			[]string{dataWLANCfgEntries, dataClientCommonOperData},
		},
		{
			"WLAN general reads one data type of the whole set",
			config.Collectors{WLAN: config.WLANCollectorModules{General: true}},
			[]string{dataWLANCfgEntries},
		},
		{
			"client traffic reads two data types",
			config.Collectors{Client: config.ClientCollectorModules{Traffic: true}},
			[]string{dataClientCommonOperData, dataClientTrafficStats},
		},
		{
			"client info reads the two identity routes and the dot11 data",
			config.Collectors{Client: config.ClientCollectorModules{Info: true}},
			[]string{
				dataClientCommonOperData, dataClientDCInfo,
				dataClientDot11OperData, dataClientSISFDBMac,
			},
		},
		{
			"client general also reads the mobility history",
			config.Collectors{Client: config.ClientCollectorModules{General: true}},
			[]string{
				dataClientCommonOperData, dataClientDot11OperData,
				dataClientTrafficStats, dataClientMMIFHistory,
			},
		},
		{"every module reads every data type, in fetch order", allModules(), dataTypeNames},
		{"no module reads nothing", config.Collectors{}, []string{}},
	}

	// A data type with no case is fetched rather than skipped: paying for a request
	// is recoverable, while withholding one relies on the mark reaching its reader.
	t.Run("an unlisted data type is fetched", func(t *testing.T) {
		t.Parallel()

		if !isDataTypeRequired("data_type_with_no_case", config.Collectors{}) {
			t.Error("isDataTypeRequired() = false for an unlisted data type, want true")
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := requiredDataTypes(tt.modules); !slices.Equal(got, tt.want) {
				t.Errorf("requiredDataTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFetchers_SkipWhatNoEnabledModuleReads proves the gate reaches the wire and
// that a skipped data type is marked rather than left looking successful.
// Asserting the name set alone would pass on a loop that still fetched all
// the whole set, which is the regression this exists to catch.
func TestFetchers_SkipWhatNoEnabledModuleReads(t *testing.T) {
	t.Parallel()

	rec := newQueryRecorder()
	server := rec.server()
	defer server.Close()

	ds, ok := NewDataSource(config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      time.Minute,
	}, config.Collectors{
		Client: config.ClientCollectorModules{Traffic: true},
	}).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}

	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil", err)
	}

	required := []string{dataClientCommonOperData, dataClientTrafficStats}
	for segment, endpoint := range mockEndpoints {
		want := 0
		if slices.Contains(required, endpoint.dataType) {
			want = 1
		}
		if got := len(rec.get(segment)); got != want {
			t.Errorf("%s was requested %d times, want %d", segment, got, want)
		}
	}

	// A skipped data type must be marked, or a collector reading it takes the
	// empty slice in the snapshot for data the controller actually returned.
	for _, name := range dataTypeNames {
		fetchErr := data.FetchErrors[name]
		if slices.Contains(required, name) {
			if fetchErr != nil {
				t.Errorf("FetchErrors[%s] = %v, want nil for a required data type", name, fetchErr)
			}
			continue
		}
		if !errors.Is(fetchErr, errDataTypeNotRequested) {
			t.Errorf("FetchErrors[%s] = %v, want errDataTypeNotRequested", name, fetchErr)
		}
	}

	// The mark must not reach the error counter or wnc_up: it reports a request
	// nobody wanted, not a controller that failed one.
	stats := ds.Stats()
	if len(stats.Errors) != len(required) {
		t.Errorf("Stats().Errors covers %d data types, want %d seeded from the enabled modules",
			len(stats.Errors), len(required))
	}
	for _, name := range required {
		if count, present := stats.Errors[name]; !present || count != 0 {
			t.Errorf("Stats().Errors[%s] = %d (present %v), want 0", name, count, present)
		}
	}
	if !stats.Up {
		t.Error("Stats().Up = false, want true when every required data type succeeded")
	}
}

// TestFetchers_DownWhenEveryRequiredDataTypeFails is the regression test for the
// defect the gating exists to fix. Before it, wnc_up was judged against all
// every data type, so a deployment running one module could fail every data
// type it needs, publish nothing, and still report up — and because fetchAllData
// returned no error the consecutive-failure counter reset, so the
// withhold-after-three-failures path never armed. The sixteen data types nothing
// reads answer normally here, which is what made the old denominator read healthy.
func TestFetchers_DownWhenEveryRequiredDataTypeFails(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(failing(dataClientCommonOperData, dataClientTrafficStats))
	defer server.Close()

	ds, ok := NewDataSource(config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      time.Minute,
	}, config.Collectors{
		Client: config.ClientCollectorModules{Traffic: true},
	}).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}

	if _, err := ds.fetchAllData(context.Background()); err == nil {
		t.Error("fetchAllData() error = nil, want an error when every required data type failed")
	}

	if stats := ds.Stats(); stats.Up {
		t.Error("Stats().Up = true, want false when every required data type failed")
	}
}

// TestFetchers_SkipBeforeTheDeadlineCheck pins the order of the two branches. A
// data type nobody asked for must be skipped before the deadline is consulted:
// recording it as truncated would seed a `data` label for a request that was
// never going to be sent, and would make the refresh report a failure it did not
// have — which is the wnc_up defect this gating fixes, reintroduced.
func TestFetchers_SkipBeforeTheDeadlineCheck(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(mockServerConfig{})
	defer server.Close()

	ds, ok := NewDataSource(config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      time.Minute,
	}, config.Collectors{
		WLAN: config.WLANCollectorModules{General: true},
	}).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ds.fetchAllData(ctx); err == nil {
		t.Error("fetchAllData() error = nil, want an error when the deadline stopped every required fetch")
	}

	// Only the required data type may carry a label. The others were skipped, and a
	// skipped data type is not a truncated one.
	stats := ds.Stats()
	if len(stats.Errors) != 1 {
		t.Errorf("Stats().Errors covers %d data types, want 1", len(stats.Errors))
	}
	if count, present := stats.Errors[dataWLANCfgEntries]; !present || count != 1 {
		t.Errorf("Stats().Errors[%s] = %d (present %v), want 1",
			dataWLANCfgEntries, count, present)
	}
}

// TestDataSource_PanicCountsOnlyTheRequiredDataTypes complements the panic test that
// uses every module: with one module enabled, a panic must not invent a `data` label
// for the data types no module reads, because recordRefresh increments the counter
// for every name it is handed.
func TestDataSource_PanicCountsOnlyTheRequiredDataTypes(t *testing.T) {
	t.Parallel()

	ds, ok := NewDataSource(config.WNC{
		Controller:    "wnc1.example.internal",
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      55 * time.Second,
	}, config.Collectors{
		WLAN: config.WLANCollectorModules{General: true},
	}).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}
	suppressBackgroundRefresh(ds)

	ds.onRefreshDone(fmt.Errorf("run: %w", errRefreshPanicked), 2*time.Second)

	stats := ds.Stats()
	if len(stats.Errors) != 1 {
		t.Errorf("Stats().Errors covers %d data types after a panic, want 1", len(stats.Errors))
	}
	if count := stats.Errors[dataWLANCfgEntries]; count != 1 {
		t.Errorf("Stats().Errors[%s] = %d, want 1", dataWLANCfgEntries, count)
	}
	if stats.Up {
		t.Error("Stats().Up = true after a panicked refresh, want false")
	}
}

// stubSource serves one snapshot so snapshot() can be driven without the refresher.
type stubSource struct {
	data *WNCDataCache
}

func (s stubSource) GetCachedData(context.Context) (*WNCDataCache, error) {
	return s.data, nil
}

// TestSnapshot_RejectsADataTypeNoModuleDeclared covers the diagnostic path. A correct
// table never reaches it, because a collector only asks for a data type it reads — so
// reaching it means the table has drifted from the guards it mirrors, and the sentinel
// is what turns that into an omitted series rather than an empty slice read as data.
func TestSnapshot_RejectsADataTypeNoModuleDeclared(t *testing.T) {
	t.Parallel()

	server := newMockWNCServer(mockServerConfig{})
	defer server.Close()

	ds, ok := NewDataSource(config.WNC{
		Controller:    extractHostFromURL(server.URL),
		AccessToken:   "test-token",
		Timeout:       5 * time.Second,
		TLSSkipVerify: true,
		CacheTTL:      time.Minute,
	}, config.Collectors{
		WLAN: config.WLANCollectorModules{General: true},
	}).(*dataSource)
	if !ok {
		t.Fatal("NewDataSource did not return *dataSource")
	}

	data, err := ds.fetchAllData(context.Background())
	if err != nil {
		t.Fatalf("fetchAllData() error = %v, want nil", err)
	}

	src := stubSource{data: data}

	if _, err := snapshot(context.Background(), src, dataAPCAPWAPData); !errors.Is(err, errDataTypeNotRequested) {
		t.Errorf("snapshot() error = %v, want errDataTypeNotRequested", err)
	}

	if _, err := snapshot(context.Background(), src, dataWLANCfgEntries); err != nil {
		t.Errorf("snapshot() error = %v, want nil for a data type a module declared", err)
	}
}

// TestRequiredDataTypes_EveryModuleFlagReadsSomething closes the one break mode the
// per-module cases above cannot. Those name the flags that exist today, so a flag
// added later and left out of the unions in isDataTypeRequired escapes them — and
// its collector would then publish nothing at all, because every route it reads is
// skipped. Walking the struct covers a flag nobody remembered to list.
func TestRequiredDataTypes_EveryModuleFlagReadsSomething(t *testing.T) {
	t.Parallel()

	var empty config.Collectors
	groups := reflect.ValueOf(&empty).Elem()

	for i := range groups.NumField() {
		group := groups.Field(i)
		if group.Kind() != reflect.Struct {
			continue
		}

		for j := range group.NumField() {
			if group.Field(j).Kind() != reflect.Bool {
				continue
			}

			name := groups.Type().Field(i).Name + "." + group.Type().Field(j).Name
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				var one config.Collectors
				reflect.ValueOf(&one).Elem().Field(i).Field(j).SetBool(true)

				if got := requiredDataTypes(one); len(got) == 0 {
					t.Errorf("requiredDataTypes() is empty with only %s enabled, so every "+
						"route that module reads is skipped and it publishes nothing", name)
				}
			})
		}
	}
}
