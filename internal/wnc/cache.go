// Package wnc provides caching functionality for WNC data.
package wnc

import (
	"context"
	"fmt"
	"log/slog"

	wnc "github.com/umatare5/cisco-ios-xe-wireless-go"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/ap"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/client"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/rrm"
	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
	"github.com/umatare5/cisco-wnc-exporter/internal/cache"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// WNCDataCache contains all operational data from WNC that can be cached and shared between collectors.
type WNCDataCache struct {
	// AP data
	CAPWAPData     *ap.ApOperCAPWAPData
	ApOperData     *ap.ApOperData
	RadioOperData  *ap.ApOperRadioOperData
	RadioOperStats *ap.ApOperRadioOperStats
	NameMACMaps    *ap.ApOperApNameMACMap

	// Client data
	CommonOperData *client.ClientOperCommonOperData
	DCInfo         *client.ClientOperDcInfo
	Dot11OperData  *client.ClientOperDot11OperData
	SisfDBMac      *client.ClientOperSisfDBMac
	TrafficStats   *client.ClientOperTrafficStats

	// RRM data
	RRMMeasurements *rrm.RRMOperRRMMeasurement

	// WLAN data
	WLANConfigEntries *wlan.WlanCfgWlanCfgEntries
	WLANOperInfo      *wlan.WlanGlobalOperWlanInfo
	WLANPolicies      *wlan.WlanCfgWlanPolicies
}

// DataSource provides cached access to WNC operational data.
type DataSource interface {
	GetCachedData(ctx context.Context) (*WNCDataCache, error)
}

// dataSource implements DataSource with caching to minimize WNC requests.
type dataSource struct {
	client *wnc.Client
	cache  *cache.Cache[*WNCDataCache]
}

// dataFetcher defines a WNC API call that can fail.
type dataFetcher struct {
	name     string
	required bool
	fetch    func(ctx context.Context, cache *WNCDataCache) error
}

// NewDataSource creates a new shared data source.
func NewDataSource(cfg config.WNC) DataSource {
	return &dataSource{
		client: createWNCClient(cfg),
		cache:  cache.New[*WNCDataCache](cfg.CacheTTL, "WNC API data cache"),
	}
}

// GetCachedData retrieves all operational data from WNC using specific List methods.
// Data is cached with configurable TTL optimized for Prometheus scrape intervals.
func (s *dataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	return s.cache.Get(func() (*WNCDataCache, error) {
		data, err := s.fetchAllData(ctx)
		if err != nil {
			return nil, fmt.Errorf("cache refresh failed: %w", err)
		}
		return data, nil
	})
}

// fetchAllData performs all WNC API calls and returns complete cache data.
func (s *dataSource) fetchAllData(ctx context.Context) (*WNCDataCache, error) {
	data := &WNCDataCache{}

	// Define all data fetchers with their criticality
	fetchers := []dataFetcher{
		{"AP CAPWAP", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListCAPWAPData(ctx)
			if err != nil {
				return err
			}
			c.CAPWAPData = data
			slog.Info("AP CAPWAP data retrieved from WNC successfully", "count", len(data.CAPWAPData))
			return nil
		}},
		{"AP operational", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListApOperData(ctx)
			if err != nil {
				return err
			}
			c.ApOperData = data
			slog.Info("AP operational data retrieved from WNC successfully", "count", len(data.OperData))
			return nil
		}},
		{"radio operational", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListRadioData(ctx)
			if err != nil {
				return err
			}
			c.RadioOperData = data
			slog.Info("radio operational data retrieved from WNC successfully", "count", len(data.RadioOperData))
			return nil
		}},
		{"AP name to MAC mapping", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListNameMACMaps(ctx)
			if err != nil {
				return err
			}
			c.NameMACMaps = data
			slog.Info("AP name to MAC mapping data retrieved from WNC successfully", "count", len(data.ApNameMACMap))
			return nil
		}},
		{"RRM measurements", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.RRM().ListRRMMeasurement(ctx)
			if err != nil {
				return err
			}
			c.RRMMeasurements = data
			slog.Info("RRM measurement data retrieved from WNC successfully", "count", len(data.RRMMeasurement))
			return nil
		}},
		{"WLAN configuration entries", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListWlanCfgEntries(ctx)
			if err != nil {
				return err
			}
			c.WLANConfigEntries = data
			var profileCount int
			if data != nil && data.WlanCfgEntries != nil {
				profileCount = len(data.WlanCfgEntries.WlanCfgEntry)
			}
			slog.Info("WLAN configuration entries retrieved from WNC successfully", "count", profileCount)
			return nil
		}},
		{"WLAN operational info", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListWlanInfo(ctx)
			if err != nil {
				return err
			}
			c.WLANOperInfo = data
			var count int
			if data != nil && data.CiscoIOSXEWirelessWlanGlobalOperData.WlanInfo != nil {
				count = len(data.CiscoIOSXEWirelessWlanGlobalOperData.WlanInfo)
			}
			slog.Info("WLAN operational info retrieved from WNC successfully", "count", count)
			return nil
		}},
		{"WLAN policies", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListWlanPolicies(ctx)
			if err != nil {
				return err
			}
			c.WLANPolicies = data
			var count int
			if data != nil && data.CiscoIOSXEWirelessWlanCfgData.WlanPolicies != nil {
				count = len(data.CiscoIOSXEWirelessWlanCfgData.WlanPolicies.WlanPolicy)
			}
			slog.Info("WLAN policies retrieved from WNC successfully", "count", count)
			return nil
		}},
		{"client common", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListCommonInfo(ctx)
			if err != nil {
				return err
			}
			c.CommonOperData = data
			slog.Info("client common data retrieved from WNC successfully", "count", len(data.CommonOperData))
			return nil
		}},
		{"client device classification", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListDCInfo(ctx)
			if err != nil {
				return err
			}
			c.DCInfo = data
			slog.Info("client device classification data retrieved from WNC successfully", "count", len(data.DcInfo))
			return nil
		}},
		{"client 802.11", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListDot11Info(ctx)
			if err != nil {
				return err
			}
			c.Dot11OperData = data
			slog.Info("client 802.11 data retrieved from WNC successfully",
				"count", len(data.Dot11OperData),
			)
			return nil
		}},
		{"SISF database", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListSISFDB(ctx)
			if err != nil {
				return err
			}
			c.SisfDBMac = data
			slog.Info("SISF database data retrieved from WNC successfully", "count", len(data.SisfDBMac))
			return nil
		}},
		{"client traffic stats", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListTrafficStats(ctx)
			if err != nil {
				return err
			}
			c.TrafficStats = data
			slog.Info("client traffic stats data retrieved from WNC successfully", "count", len(data.TrafficStats))
			return nil
		}},

		// Optional data - failures are logged but don't abort cache refresh
		{"radio statistics", false, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListRadioOperStats(ctx)
			if err != nil {
				return err
			}
			c.RadioOperStats = data
			slog.Info("radio statistics data retrieved from WNC successfully", "count", len(data.RadioOperStats))
			return nil
		}},
	}

	// Execute all fetchers
	for _, f := range fetchers {
		if err := f.fetch(ctx, data); err != nil {
			if f.required {
				return nil, fmt.Errorf("failed to fetch %s: %w", f.name, err)
			}
			slog.Warn("optional data fetch failed", "data_type", f.name, "error", err)
		}
	}

	return data, nil
}
