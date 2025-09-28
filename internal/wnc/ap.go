// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains AP-related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
)

// APSource provides access to AP data from WNC via REST API.
type APSource interface {
	GetCAPWAPData(ctx context.Context) ([]ap.CAPWAPData, error)
	GetAPOperData(ctx context.Context) ([]ap.OperData, error)
	GetRadioData(ctx context.Context) ([]ap.RadioOperData, error)
	GetRadioOperStats(ctx context.Context) ([]ap.RadioOperStats, error)
	GetRadioResetStats(ctx context.Context) ([]ap.RadioResetStats, error)
	ListNameMACMaps(ctx context.Context) ([]ap.ApNameMACMap, error)
}

// apSource implements APSource using SharedDataSource for caching.
type apSource struct {
	sharedDataSource DataSource
}

// NewAPSource creates a new APSource implementation that uses SharedDataSource for caching.
func NewAPSource(sharedDataSource DataSource) APSource {
	return &apSource{
		sharedDataSource: sharedDataSource,
	}
}

// GetCAPWAPData returns CAPWAP data from WNC via SharedDataSource (cached).
func (s *apSource) GetCAPWAPData(ctx context.Context) ([]ap.CAPWAPData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CAPWAPData, nil
}

// GetAPOperData retrieves AP operational data via SharedDataSource (cached).
func (s *apSource) GetAPOperData(ctx context.Context) ([]ap.OperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.ApOperData, nil
}

// GetRadioData returns radio operational data from WNC via SharedDataSource (cached).
func (s *apSource) GetRadioData(ctx context.Context) ([]ap.RadioOperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RadioOperData, nil
}

// GetRadioOperStats returns radio operational statistics from WNC via SharedDataSource (cached).
func (s *apSource) GetRadioOperStats(ctx context.Context) ([]ap.RadioOperStats, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RadioOperStats, nil
}

// GetRadioResetStats returns radio reset statistics from WNC via SharedDataSource (cached).
func (s *apSource) GetRadioResetStats(ctx context.Context) ([]ap.RadioResetStats, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RadioResetStats, nil
}

// ListNameMACMaps returns AP name to MAC mapping data from WNC via SharedDataSource (cached).
func (s *apSource) ListNameMACMaps(ctx context.Context) ([]ap.ApNameMACMap, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.NameMACMaps, nil
}
