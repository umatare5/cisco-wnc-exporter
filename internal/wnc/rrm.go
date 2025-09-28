// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains RRM (Radio Resource Management) related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// RRMSource provides access to RRM data from WNC via REST API.
type RRMSource interface {
	GetRRMMeasurements(ctx context.Context) ([]rrm.RRMMeasurement, error)
	GetRRMCoverage(ctx context.Context) ([]rrm.RRMCoverage, error)
	GetApDot11RadarData(ctx context.Context) ([]rrm.ApDot11RadarData, error)
}

// rrmSource implements RRMSource using SharedDataSource for caching.
type rrmSource struct {
	sharedDataSource DataSource
}

// NewRRMSource creates a new RRMSource implementation that uses SharedDataSource for caching.
func NewRRMSource(sharedDataSource DataSource) RRMSource {
	return &rrmSource{
		sharedDataSource: sharedDataSource,
	}
}

// GetRRMMeasurements returns RRM measurement data from WNC via SharedDataSource (cached).
func (s *rrmSource) GetRRMMeasurements(ctx context.Context) ([]rrm.RRMMeasurement, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RRMMeasurements, nil
}

// GetRRMCoverage returns RRM coverage data from WNC via SharedDataSource (cached).
func (s *rrmSource) GetRRMCoverage(ctx context.Context) ([]rrm.RRMCoverage, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RRMCoverage, nil
}

// GetApDot11RadarData returns AP radar detection data from WNC via SharedDataSource (cached).
func (s *rrmSource) GetApDot11RadarData(ctx context.Context) ([]rrm.ApDot11RadarData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.ApDot11RadarData, nil
}
