// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains WLAN-related functionality for accessing WNC via REST API.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
)

// WLANSource provides access to WLAN data from WNC via REST API.
type WLANSource interface {
	ListConfigEntries(ctx context.Context) ([]wlan.WlanCfgEntry, error)
	ListPolicies(ctx context.Context) ([]wlan.WlanPolicy, error)
	ListPolicyListEntries(ctx context.Context) ([]wlan.PolicyListEntry, error)
	ListClientStats(ctx context.Context) ([]ap.WlanClientStats, error)
}

// wlanSource implements WLANSource using SharedDataSource for caching.
type wlanSource struct {
	sharedDataSource DataSource
}

// NewWLANSource creates a new WLANSource implementation that uses SharedDataSource for caching.
func NewWLANSource(sharedDataSource DataSource) WLANSource {
	return &wlanSource{
		sharedDataSource: sharedDataSource,
	}
}

// ListConfigEntries retrieves WLAN configuration entries via SharedDataSource (cached).
func (s *wlanSource) ListConfigEntries(ctx context.Context) ([]wlan.WlanCfgEntry, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataWLANCfgEntries)
	if err != nil {
		return nil, err
	}
	return data.WLANConfigEntries, nil
}

// ListPolicies retrieves WLAN policies via SharedDataSource (cached).
func (s *wlanSource) ListPolicies(ctx context.Context) ([]wlan.WlanPolicy, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataWLANPolicies)
	if err != nil {
		return nil, err
	}
	return data.WLANPolicies, nil
}

// ListClientStats retrieves the per-WLAN client statistics via SharedDataSource
// (cached). The controller keeps them in the AP global operational subtree, but they
// are keyed by WLAN and read by the WLAN collector alone.
func (s *wlanSource) ListClientStats(ctx context.Context) ([]ap.WlanClientStats, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataWLANClientStats)
	if err != nil {
		return nil, err
	}
	return data.WLANClientStats, nil
}

// ListPolicyListEntries retrieves policy list entries via SharedDataSource (cached).
func (s *wlanSource) ListPolicyListEntries(ctx context.Context) ([]wlan.PolicyListEntry, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataWLANPolicyListEntries)
	if err != nil {
		return nil, err
	}
	return data.WLANPolicyListEntries, nil
}
