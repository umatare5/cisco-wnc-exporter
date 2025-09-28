// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains WLAN-related functionality for accessing WNC via REST API.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
)

// WLANSource provides access to WLAN data from WNC via REST API.
type WLANSource interface {
	ListConfigEntries(ctx context.Context) ([]wlan.WlanCfgEntry, error)
	ListPolicies(ctx context.Context) ([]wlan.WlanPolicy, error)
	ListPolicyListEntries(ctx context.Context) ([]wlan.PolicyListEntry, error)
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
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.WLANConfigEntries, nil
}

// ListPolicies retrieves WLAN policies via SharedDataSource (cached).
func (s *wlanSource) ListPolicies(ctx context.Context) ([]wlan.WlanPolicy, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.WLANPolicies, nil
}

// ListPolicyListEntries retrieves policy list entries via SharedDataSource (cached).
func (s *wlanSource) ListPolicyListEntries(ctx context.Context) ([]wlan.PolicyListEntry, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.WLANPolicyListEntries, nil
}
