package wnc

import (
	"context"
	"errors"
	"testing"
	"time"
)

// noopDone satisfies the outcome handler for tests that do not inspect it.
func noopDone(error, time.Duration) {}

func TestRefresher_FirstGetIsDue(t *testing.T) {
	t.Parallel()

	r := newRefresher(time.Hour, func(context.Context) (*WNCDataCache, error) {
		return &WNCDataCache{}, nil
	}, noopDone)

	if !r.due() {
		t.Error("due() = false on a new refresher, want true: the zero nextAt has to mean now")
	}
	if r.cur.Load() != nil {
		t.Error("cur is set on a new refresher, want nil")
	}
}

// TestRefresher_Get_StartsBackgroundRefresh holds the refresh until the first get has
// returned: nothing orders the background store after that load, so an unheld refresh can
// publish first and fail the nil assertion for the wrong reason. The hold is bounded so a
// get that refreshes synchronously fails that assertion instead of hanging.
func TestRefresher_Get_StartsBackgroundRefresh(t *testing.T) {
	t.Parallel()

	want := &WNCDataCache{RefreshedAt: time.Now()}
	done := make(chan error, 1)
	release := make(chan struct{})

	r := newRefresher(time.Hour, func(context.Context) (*WNCDataCache, error) {
		select {
		case <-release:
		case <-time.After(5 * time.Second):
		}
		return want, nil
	}, func(err error, _ time.Duration) {
		done <- err
	})

	if snap := r.get(); snap != nil {
		t.Error("get() returned a snapshot before the first refresh completed, want nil")
	}
	close(release)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("refresh error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("get() did not start a background refresh")
	}

	if got := r.get(); got != want {
		t.Errorf("get() = %p, want the published snapshot %p", got, want)
	}
}

// TestRefresher_Get_SingleFlight covers the guard from the caller's side: a second
// scrape arriving while a refresh runs must not start another one.
func TestRefresher_Get_SingleFlight(t *testing.T) {
	t.Parallel()

	var calls int
	r := newRefresher(0, func(context.Context) (*WNCDataCache, error) {
		calls++
		return &WNCDataCache{}, nil
	}, noopDone)

	r.inflight.Store(true)

	if snap := r.get(); snap != nil {
		t.Error("get() returned a snapshot, want nil")
	}
	if calls != 0 {
		t.Errorf("run was called %d times while a refresh was in flight, want 0", calls)
	}
}

// TestRefresher_RefreshOnce_NotDueReleasesGuard pins why the CAS winner re-checks
// due inside refreshOnce: returning from get without starting the goroutine would
// leave inflight true for the life of the process.
func TestRefresher_RefreshOnce_NotDueReleasesGuard(t *testing.T) {
	t.Parallel()

	var calls int
	r := newRefresher(time.Hour, func(context.Context) (*WNCDataCache, error) {
		calls++
		return &WNCDataCache{}, nil
	}, noopDone)

	r.inflight.Store(true)
	r.stamp()

	r.refreshOnce(context.Background())

	if calls != 0 {
		t.Errorf("run was called %d times before the next refresh was due, want 0", calls)
	}
	if r.inflight.Load() {
		t.Error("inflight is still true, want false: no later refresh could ever start")
	}
}

func TestRefresher_RefreshOnce_PublishesOnSuccess(t *testing.T) {
	t.Parallel()

	want := &WNCDataCache{RefreshedAt: time.Now()}
	var gotErr error
	var gotElapsed time.Duration

	r := newRefresher(time.Hour, func(context.Context) (*WNCDataCache, error) {
		return want, nil
	}, func(err error, elapsed time.Duration) {
		gotErr = err
		gotElapsed = elapsed
	})

	r.refreshOnce(context.Background())

	if got := r.cur.Load(); got != want {
		t.Errorf("cur = %p, want the fetched snapshot %p", got, want)
	}
	if gotErr != nil {
		t.Errorf("onDone error = %v, want nil", gotErr)
	}
	if gotElapsed <= 0 {
		t.Errorf("onDone elapsed = %v, want a positive duration", gotElapsed)
	}
	if r.inflight.Load() {
		t.Error("inflight is still true after refreshOnce, want false")
	}
	if r.due() {
		t.Error("due() = true right after a refresh, want false")
	}
}

// TestRefresher_RefreshOnce_KeepsPreviousSnapshotOnError pins the invariant the
// grace period depends on: fetchAllData returns a nil snapshot when every data
// type fails, and storing it would drop the freshness series exactly when the
// staleness alert needs them.
func TestRefresher_RefreshOnce_KeepsPreviousSnapshotOnError(t *testing.T) {
	t.Parallel()

	published := &WNCDataCache{RefreshedAt: time.Now()}
	wantErr := errors.New("every data type failed")
	var gotErr error

	r := newRefresher(0, func(context.Context) (*WNCDataCache, error) {
		return nil, wantErr
	}, func(err error, _ time.Duration) {
		gotErr = err
	})
	r.cur.Store(published)

	r.refreshOnce(context.Background())

	if got := r.cur.Load(); got != published {
		t.Errorf("cur = %p, want the previous snapshot %p to keep being served", got, published)
	}
	if !errors.Is(gotErr, wantErr) {
		t.Errorf("onDone error = %v, want %v", gotErr, wantErr)
	}
}

// TestRefresher_RefreshOnce_ReportsPanic pins that a recovered panic still reaches
// the outcome handler. Bumping only the failure counter would freeze the refresh
// statistics while the data series drain away.
func TestRefresher_RefreshOnce_ReportsPanic(t *testing.T) {
	t.Parallel()

	var gotErr error
	var gotElapsed time.Duration

	r := newRefresher(time.Hour, func(context.Context) (*WNCDataCache, error) {
		panic("fetch blew up")
	}, func(err error, elapsed time.Duration) {
		gotErr = err
		gotElapsed = elapsed
	})

	r.refreshOnce(context.Background())

	if !errors.Is(gotErr, errRefreshPanicked) {
		t.Errorf("onDone error = %v, want errRefreshPanicked", gotErr)
	}
	if gotElapsed < 0 {
		t.Errorf("onDone elapsed = %v, want a non-negative duration", gotElapsed)
	}
	if r.cur.Load() != nil {
		t.Error("cur is set after a panicked refresh, want nil")
	}
	if r.inflight.Load() {
		t.Error("inflight is still true after a panicked refresh, want false")
	}
	if r.due() {
		t.Error("due() = true after a panicked refresh, want false: the panic path must stamp too")
	}
}

// TestRefresher_RefreshOnce_StampsAtCompletion uses a refresh that outlasts its own
// TTL. Stamping at the start would already be due on return, which is what turns a
// slow controller into back-to-back refreshes.
func TestRefresher_RefreshOnce_StampsAtCompletion(t *testing.T) {
	t.Parallel()

	const (
		ttl        = 250 * time.Millisecond
		refreshFor = 300 * time.Millisecond
	)

	r := newRefresher(ttl, func(context.Context) (*WNCDataCache, error) {
		time.Sleep(refreshFor)
		return &WNCDataCache{}, nil
	}, noopDone)

	r.refreshOnce(context.Background())

	if r.due() {
		t.Error("due() = true right after a refresh that outlasted the TTL, want false")
	}
}
