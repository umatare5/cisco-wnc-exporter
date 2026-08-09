// Package wnc provides WNC data access and caching.
// This file holds the stale-while-revalidate primitive the data source is built on.
package wnc

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	// errNoSnapshot is returned until the first refresh produces data.
	errNoSnapshot = errors.New("no WNC data snapshot available yet")
	// errSnapshotWithheld is returned when the snapshot is too stale to serve.
	errSnapshotWithheld = errors.New("WNC data snapshot withheld")
	// errRefreshPanicked reports a recovered panic to the outcome handler.
	errRefreshPanicked = errors.New("WNC data refresh panicked")
)

// refresher serves the last snapshot immediately and refreshes it in the
// background, so a scrape never waits for the controller.
// The zero value is not usable: newRefresher captures a monotonic base.
type refresher struct {
	cur      atomic.Pointer[WNCDataCache]
	inflight atomic.Bool
	// nextAt is monotonic nanoseconds since base at which the next refresh may
	// start. The zero value means "now", so the first get always refreshes.
	nextAt atomic.Int64
	// base must come from time.Now() and must never be round-tripped through
	// Round, Truncate, UTC or marshaling, all of which strip the monotonic
	// reading and would make scheduling vulnerable to wall-clock steps.
	base time.Time
	ttl  time.Duration
	run  func(context.Context) (*WNCDataCache, error)
	// onDone reports every outcome, including a recovered panic.
	onDone func(err error, elapsed time.Duration)
}

func newRefresher(
	ttl time.Duration,
	run func(context.Context) (*WNCDataCache, error),
	onDone func(err error, elapsed time.Duration),
) *refresher {
	return &refresher{base: time.Now(), ttl: ttl, run: run, onDone: onDone}
}

// get returns the current snapshot, nil until the first refresh succeeds, and
// starts a background refresh when one is due. It never blocks.
func (r *refresher) get() *WNCDataCache {
	if r.due() && r.inflight.CompareAndSwap(false, true) {
		go r.refreshOnce(context.Background())
	}
	return r.cur.Load()
}

// due reports whether enough time has passed since the last refresh completed.
func (r *refresher) due() bool {
	return int64(time.Since(r.base)) >= r.nextAt.Load()
}

// stamp schedules the next refresh one TTL after this one completed, which keeps
// at least one TTL of idle between refreshes whatever the scrape rate.
func (r *refresher) stamp() {
	r.nextAt.Store(int64(time.Since(r.base)) + int64(r.ttl))
}

// refreshOnce runs one refresh. It is safe to call directly from a test.
func (r *refresher) refreshOnce(ctx context.Context) {
	// Registered first so it runs last: releasing the guard before stamping would
	// let the next caller start a refresh immediately.
	defer r.inflight.Store(false)

	// A caller that evaluated due() before the previous refresh stamped can win
	// the CAS after it released. Re-checking here, rather than in get, keeps the
	// deferred release above on every path.
	if !r.due() {
		return
	}

	start := time.Now()

	defer r.stamp()
	defer func() {
		if v := recover(); v != nil {
			slog.Error("WNC data refresh panicked", "panic", v, "stack", string(debug.Stack()))
			r.onDone(errRefreshPanicked, time.Since(start))
		}
	}()

	// Bracket the refresh so a long one is distinguishable from a wedged process.
	// This is the only progress signal at info level; per data type detail is at
	// debug level and in wnc_refresh_items.
	slog.Info("refreshing WNC data")

	data, err := r.run(ctx)
	elapsed := time.Since(start)

	if err != nil {
		slog.Error("failed to refresh WNC data", "error", err, "duration", elapsed)
	} else {
		// Publish before reporting success: onDone clears the failure counter, and
		// lifting the withhold before the new snapshot is visible would serve the
		// previous generation as if it were fresh.
		r.cur.Store(data)
		slog.Info("WNC data refreshed", "duration", elapsed)
	}

	r.onDone(err, elapsed)
}
