package wnc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"
	"testing"
	"time"
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
