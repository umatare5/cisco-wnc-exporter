// Package wnc provides WNC data access and caching.
// This file holds the reads the SDK has no route or type for.
package wnc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strings"
)

// The RESTCONF paths of the containers the SDK has no route for. The SDK keeps its
// route table internal, so these are literals; the envelope check in soleValue is what
// catches a wrong one, and a wrong path answers 404 rather than an empty success.
const (
	routeControllerBootTime = "Cisco-IOS-XE-device-hardware-oper:device-hardware-data" +
		"/device-hardware/device-system-data/boot-time"
	routeCoClientDelReason = "Cisco-IOS-XE-wireless-client-global-oper:client-global-oper-data" +
		"/client-stats/co-client-del-reason"
	routeClientRoamingStats = "Cisco-IOS-XE-wireless-client-global-oper:client-global-oper-data" +
		"/client-dot11-stats/client-roaming-stats"
)

// restconfDataPath prefixes every path above, matching what the SDK builds for its
// own typed accessors.
const restconfDataPath = "/restconf/data/"

// rawGetter is the seam the SDK client already satisfies. Declaring the method here
// rather than naming the SDK type keeps these reads independent of a type the SDK
// exports from an internal package, and it is what a test substitutes.
type rawGetter interface {
	Do(ctx context.Context, method, path string) ([]byte, error)
}

// rawValue reads one RESTCONF container or leaf the SDK cannot reach and returns its
// value, reporting false when the controller carries nothing there.
//
// The read reuses the SDK client, so it inherits the credentials, the TLS settings,
// the request timeout, the connection pool and the *wnc.APIError typing of every
// typed accessor. What it gives up is the route constant and the SDK-owned struct,
// which is why the value is decoded without struct tags: a tag naming an envelope key
// the controller does not send decodes to nothing and publishes an empty family with
// no error, no log line and HTTP 200, which is the one failure mode this file must not
// reproduce.
func rawValue[V any](
	ctx context.Context, getter rawGetter, requestPath string,
) (value V, present bool, err error) {
	body, err := getter.Do(ctx, http.MethodGet, restconfDataPath+requestPath)
	if err != nil {
		return value, false, err
	}

	// An empty body is how this platform reports a container it does not carry: a list
	// elsewhere in its operational tree answers 204 with no body at all. The typed
	// accessors read an empty body as a successful fetch of nothing, so treating it
	// as absence here keeps both paths accounted alike. An HTTP error, 404 included,
	// has already been returned above: a path this exporter got wrong must not be
	// indistinguishable from a container an image does not have.
	if len(body) == 0 {
		return value, false, nil
	}

	var envelope map[string]V
	if err := json.Unmarshal(body, &envelope); err != nil {
		return value, false, fmt.Errorf("%s: %w", requestPath, err)
	}

	return soleValue(envelope, requestPath)
}

// soleValue returns the only value of a RESTCONF envelope, having checked that its key
// is the one the request asked for. A container or leaf read answers with exactly one
// key, module-qualified, whose local name is the last segment of the path.
//
// Counting the keys is not enough on its own. A well-formed envelope carrying some
// other key leaves the value at its zero, which publishes an empty family and reports
// success, so the local name is compared as well. That comparison is derived from the
// path rather than written out per read, which leaves nothing to keep in sync.
func soleValue[V any](envelope map[string]V, requestPath string) (value V, present bool, err error) {
	if len(envelope) != 1 {
		return value, false, fmt.Errorf("%s: response carries %d keys, want exactly 1",
			requestPath, len(envelope))
	}

	// A query parameter is not part of the node name. with-defaults is the one this
	// exporter has a use for, since the raw path cannot take the SDK's GetOption.
	trimmed, _, _ := strings.Cut(requestPath, "?")
	want := path.Base(trimmed)

	for key, sole := range envelope {
		module, local, qualified := strings.Cut(key, ":")
		if !qualified || module == "" || local != want {
			return value, false, fmt.Errorf("%s: response carries key %q, want a module-qualified %q",
				requestPath, key, want)
		}

		return sole, true, nil
	}

	return value, false, nil
}

// numericLeaves converts a container of leaves to their numeric values, skipping any
// leaf it cannot read rather than losing the container.
//
// This controller writes some 64-bit counters as JSON strings, so each leaf is decoded
// through json.Number, which accepts a number and a numeric string alike. Decoding the
// whole container into json.Number directly would abandon every leaf after the first
// one that is neither, which for this read would drop hundreds of counters because one
// leaf changed shape in a new release.
func numericLeaves(leaves map[string]json.RawMessage, name string) map[string]float64 {
	values := make(map[string]float64, len(leaves))

	for leaf, raw := range leaves {
		var number json.Number
		if err := json.Unmarshal(raw, &number); err != nil {
			slog.Debug("skipped a leaf that is not a number", "data", name, "leaf", leaf)
			continue
		}

		value, err := number.Float64()
		if err != nil {
			slog.Debug("skipped a leaf that does not parse as a number", "data", name, "leaf", leaf)
			continue
		}

		values[leaf] = value
	}

	return values
}
