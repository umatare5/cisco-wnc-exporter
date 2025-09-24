// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains AP-related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
)

// APSource provides access to AP data from WNC via REST API.
type APSource interface {
	GetCAPWAPData(ctx context.Context) (*ap.ApOperCAPWAPData, error)
	GetAPOperData(ctx context.Context) (*ap.ApOperData, error)
	GetRadioData(ctx context.Context) (*ap.ApOperRadioOperData, error)
	GetRadioOperStats(ctx context.Context) (*ap.ApOperRadioOperStats, error)
	ListNameMACMaps(ctx context.Context) (*ap.ApOperApNameMACMap, error)
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
func (s *apSource) GetCAPWAPData(ctx context.Context) (*ap.ApOperCAPWAPData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if data.CAPWAPData == nil {
		return &ap.ApOperCAPWAPData{}, nil // Return empty struct instead of nil
	}
	return data.CAPWAPData, nil
}

// GetAPOperData returns AP operational data from WNC via SharedDataSource (cached).
func (s *apSource) GetAPOperData(ctx context.Context) (*ap.ApOperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if data.ApOperData == nil {
		return &ap.ApOperData{}, nil // Return empty struct instead of nil
	}

	// Validate and sanitize AP operational data
	validated := s.validateAPOperData(data.ApOperData)
	return validated, nil
}

// GetRadioData returns radio operational data from WNC via SharedDataSource (cached).
func (s *apSource) GetRadioData(ctx context.Context) (*ap.ApOperRadioOperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if data.RadioOperData == nil {
		return &ap.ApOperRadioOperData{}, nil // Return empty struct instead of nil
	}

	// Validate and sanitize radio operational data
	validated := s.validateRadioOperData(data.RadioOperData)
	return validated, nil
}

// GetRadioOperStats returns radio operational statistics from WNC via SharedDataSource (cached).
func (s *apSource) GetRadioOperStats(ctx context.Context) (*ap.ApOperRadioOperStats, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if data.RadioOperStats == nil {
		return &ap.ApOperRadioOperStats{}, nil // Return empty struct instead of nil
	}
	return data.RadioOperStats, nil
}

// ListNameMACMaps returns AP name to MAC mapping data from WNC via SharedDataSource (cached).
func (s *apSource) ListNameMACMaps(ctx context.Context) (*ap.ApOperApNameMACMap, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	if data.NameMACMaps == nil {
		return &ap.ApOperApNameMACMap{}, nil // Return empty struct instead of nil
	}

	// Validate and sanitize name-MAC mapping data
	validated := s.validateNameMACMaps(data.NameMACMaps)
	return validated, nil
}

// validateAPOperData validates and sanitizes AP operational data.
// NOTE: This function will be simplified further once SDK removes omitempty
// from Live: IOS-XE 17.12.5 fields (ApSysStats, PhyHtCfg, RadioBandInfo, etc.)
func (s *apSource) validateAPOperData(data *ap.ApOperData) *ap.ApOperData {
	if data == nil {
		return &ap.ApOperData{}
	}

	// TODO: Remove this validation once SDK removes omitempty for Live: IOS-XE 17.12.5 fields
	// Currently needed because ApSysStats has `json:"ap-sys-stats,omitempty"`
	validatedOperData := make([]ap.ApOperInternalData, len(data.OperData))
	for i, operData := range data.OperData {
		if operData.ApSysStats == nil {
			// Temporary workaround: provide default system stats
			operData.ApSysStats = &ap.ApSystemStats{}
		}
		validatedOperData[i] = operData
	}

	return &ap.ApOperData{
		OperData: validatedOperData,
	}
}

// validateRadioOperData validates and sanitizes radio operational data.
// NOTE: This function will be simplified further once SDK removes omitempty
// from Live: IOS-XE 17.12.5 fields (PhyHtCfg, RadioBandInfo, etc.)
func (s *apSource) validateRadioOperData(data *ap.ApOperRadioOperData) *ap.ApOperRadioOperData {
	if data == nil {
		return &ap.ApOperRadioOperData{}
	}

	// TODO: Remove this validation once SDK removes omitempty for Live: IOS-XE 17.12.5 fields
	// Currently needed because fields have omitempty tags
	validatedRadioData := make([]ap.RadioOperData, len(data.RadioOperData))
	for i, radioData := range data.RadioOperData {
		// Temporary workaround: ensure RadioBandInfo exists
		if len(radioData.RadioBandInfo) == 0 {
			radioData.RadioBandInfo = []ap.RadioBandInfo{{}} // Minimal default
		}

		// Temporary workaround: ensure PhyHtCfg exists
		if radioData.PhyHtCfg == nil {
			radioData.PhyHtCfg = &ap.PhyHtCfg{} // Minimal default
		}

		validatedRadioData[i] = radioData
	}

	return &ap.ApOperRadioOperData{
		RadioOperData: validatedRadioData,
	}
}

// validateNameMACMaps validates and sanitizes AP name-MAC mapping data.
// NOTE: This function will be simplified further once SDK removes omitempty
// from Live: IOS-XE 17.12.5 fields and ensures data completeness.
func (s *apSource) validateNameMACMaps(data *ap.ApOperApNameMACMap) *ap.ApOperApNameMACMap {
	if data == nil {
		return &ap.ApOperApNameMACMap{}
	}

	// TODO: Remove this validation once SDK ensures data completeness
	// Currently needed to filter out invalid/empty name-MAC mappings
	validatedMappings := make([]ap.ApNameMACMap, 0, len(data.ApNameMACMap))
	for _, mapping := range data.ApNameMACMap {
		// Only include mappings with both valid name and MAC
		if mapping.WtpName != "" && mapping.WtpMAC != "" {
			validatedMappings = append(validatedMappings, mapping)
		}
	}

	return &ap.ApOperApNameMACMap{
		ApNameMACMap: validatedMappings,
	}
}
