// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains RRM (Radio Resource Management) related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
)

// Default values for RRM validation.
const (
	defaultNoiseFloor = -90 // Default noise floor in dBm
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
	if data.RRMMeasurements == nil {
		return &rrm.RRMOperRRMMeasurement{}, nil // Return empty struct instead of nil
	}

	// Validate and sanitize RRM measurement data
	validated := s.validateRRMMeasurements(data.RRMMeasurements)
	return validated, nil
}

// validateRRMMeasurements validates and sanitizes RRM measurement data to ensure collectors
// can use it directly without additional type checks.
func (s *rrmSource) validateRRMMeasurements(data *rrm.RRMOperRRMMeasurement) *rrm.RRMOperRRMMeasurement {
	if data == nil {
		return &rrm.RRMOperRRMMeasurement{}
	}

	// Sanitize each measurement entry
	validatedMeasurements := make([]rrm.RRMMeasurement, len(data.RRMMeasurement))
	for i, measurement := range data.RRMMeasurement {
		validatedMeasurements[i] = s.validateSingleMeasurement(measurement)
	}

	return &rrm.RRMOperRRMMeasurement{
		RRMMeasurement: validatedMeasurements,
	}
}

// validateSingleMeasurement validates and sanitizes a single RRM measurement.
// NOTE: This function will be simplified further once SDK removes omitempty
// from Live: IOS-XE 17.12.5 fields (Load, Noise, etc.)
func (s *rrmSource) validateSingleMeasurement(measurement rrm.RRMMeasurement) rrm.RRMMeasurement {
	// TODO: Remove this validation once SDK removes omitempty for Live: IOS-XE 17.12.5 fields
	// Currently needed because Load and Noise have `json:"load,omitempty"` and `json:"noise,omitempty"`

	// Temporary workaround: ensure Load data exists
	if measurement.Load == nil {
		measurement.Load = &rrm.Load{} // Minimal default, Go zero values are sufficient
	}

	// Temporary workaround: ensure Noise data structure exists
	if measurement.Noise == nil || len(measurement.Noise.Noise.NoiseData) == 0 {
		measurement.Noise = &rrm.Noise{
			Noise: rrm.NoiseData{
				NoiseData: []rrm.NoiseDataItem{
					{Noise: defaultNoiseFloor}, // Minimal functional default
				},
			},
		}
	}

	return measurement
}
