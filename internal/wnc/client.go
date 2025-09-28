// Package wnc provides thin interfaces and adapters for the Cisco WNC SDK.
// This file contains Client-related functionality.
package wnc

import (
	"context"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
)

// ClientSource provides access to client data from WNC via REST API.
type ClientSource interface {
	GetClientData(ctx context.Context) ([]client.CommonOperData, error)
	GetDeviceData(ctx context.Context) ([]client.DcInfo, error)
	GetDot11Data(ctx context.Context) ([]client.Dot11OperData, error)
	GetSISFDBData(ctx context.Context) ([]client.SisfDBMac, error)
	GetTrafficStats(ctx context.Context) ([]client.TrafficStats, error)
	GetMobilityHistory(ctx context.Context) ([]client.MmIfClientHistory, error)
}

// clientSource implements ClientSource using SharedDataSource for caching.
type clientSource struct {
	sharedDataSource DataSource
}

// NewClientSource creates a new ClientSource implementation that uses SharedDataSource for caching.
func NewClientSource(sharedDataSource DataSource) ClientSource {
	return &clientSource{
		sharedDataSource: sharedDataSource,
	}
}

// GetClientData returns client common operational data from WNC via SharedDataSource (cached).
func (s *clientSource) GetClientData(ctx context.Context) ([]client.CommonOperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.CommonOperData, nil
}

// GetDeviceData returns device classification info from WNC via SharedDataSource (cached).
func (s *clientSource) GetDeviceData(ctx context.Context) ([]client.DcInfo, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.DCInfo, nil
}

// GetDot11Data returns 802.11 operational data from WNC via SharedDataSource (cached).
func (s *clientSource) GetDot11Data(ctx context.Context) ([]client.Dot11OperData, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Dot11OperData, nil
}

// GetSISFDBData returns SISF database information from WNC via SharedDataSource (cached).
func (s *clientSource) GetSISFDBData(ctx context.Context) ([]client.SisfDBMac, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.SisfDBMac, nil
}

// GetTrafficStats returns traffic statistics from WNC via SharedDataSource (cached).
func (s *clientSource) GetTrafficStats(ctx context.Context) ([]client.TrafficStats, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.TrafficStats, nil
}

// GetMobilityHistory returns mobility manager interface client history from WNC via SharedDataSource (cached).
func (s *clientSource) GetMobilityHistory(ctx context.Context) ([]client.MmIfClientHistory, error) {
	data, err := s.sharedDataSource.GetCachedData(ctx)
	if err != nil {
		return nil, err
	}
	return data.MmIfClientHistory, nil
}
