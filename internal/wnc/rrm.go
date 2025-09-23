// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains RRM (Radio Resource Management) related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// RRMSource provides access to RRM data from WNC via REST API.
type RRMSource interface {
	GetRRMMeasurement(ctx context.Context) (*rrm.RRMOperRRMMeasurement, error)
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

// GetRRMMeasurement returns RRM measurement data from WNC via SharedDataSource (cached).
func (s *rrmSource) GetRRMMeasurement(ctx context.Context) (*rrm.RRMOperRRMMeasurement, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.RRMMeasurements, nil
}
