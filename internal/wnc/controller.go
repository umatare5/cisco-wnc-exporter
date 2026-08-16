// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains controller-wide functionality.
package wnc

import (
	"context"
)

// ControllerSource provides access to controller-wide data from WNC via REST API.
type ControllerSource interface {
	GetBootTime(ctx context.Context) (string, error)
	GetClientDeleteReasons(ctx context.Context) (map[string]float64, error)
	GetClientRoamingStats(ctx context.Context) (map[string]float64, error)
}

// controllerSource implements ControllerSource using SharedDataSource for caching.
type controllerSource struct {
	sharedDataSource DataSource
}

// NewControllerSource creates a new ControllerSource implementation that uses SharedDataSource for caching.
func NewControllerSource(sharedDataSource DataSource) ControllerSource {
	return &controllerSource{
		sharedDataSource: sharedDataSource,
	}
}

// GetBootTime returns the controller boot time from WNC via SharedDataSource (cached).
// It is empty when the controller carries no such leaf, which the collector reads as
// absence rather than as an instant.
func (s *controllerSource) GetBootTime(ctx context.Context) (string, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataControllerBootTime)
	if err != nil {
		return "", err
	}
	return data.ControllerBootTime, nil
}

// GetClientRoamingStats returns the controller-wide roam counters from WNC via
// SharedDataSource (cached).
func (s *controllerSource) GetClientRoamingStats(ctx context.Context) (map[string]float64, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataClientRoamingStats)
	if err != nil {
		return nil, err
	}
	return data.ClientRoamingStats, nil
}

// GetClientDeleteReasons returns the per-reason client deletion counters from WNC via
// SharedDataSource (cached).
func (s *controllerSource) GetClientDeleteReasons(ctx context.Context) (map[string]float64, error) {
	data, err := snapshot(ctx, s.sharedDataSource, dataCoClientDelReason)
	if err != nil {
		return nil, err
	}
	return data.ClientDeleteReasons, nil
}
