// Package wnc provides WNC data access and caching.
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

// WNCDataCache contains operational data from WNC.
type WNCDataCache struct {
	CAPWAPData      []ap.CAPWAPData
	ApOperData      []ap.OperData
	RadioOperData   []ap.RadioOperData
	RadioOperStats  []ap.RadioOperStats
	RadioResetStats []ap.RadioResetStats
	NameMACMaps     []ap.ApNameMACMap

	CommonOperData    []client.CommonOperData
	DCInfo            []client.DcInfo
	Dot11OperData     []client.Dot11OperData
	SisfDBMac         []client.SisfDBMac
	TrafficStats      []client.TrafficStats
	MmIfClientHistory []client.MmIfClientHistory

	// RRM data
	RRMMeasurements  []rrm.RRMMeasurement
	RRMCoverage      []rrm.RRMCoverage
	ApDot11RadarData []rrm.ApDot11RadarData

	// WLAN data
	WLANConfigEntries     []wlan.WlanCfgEntry
	WLANPolicies          []wlan.WlanPolicy
	WLANPolicyListEntries []wlan.PolicyListEntry
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

func (s *dataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	return s.cache.Get(func() (*WNCDataCache, error) {
		data, err := s.fetchAllData(ctx)
		if err != nil {
			return nil, fmt.Errorf("cache refresh failed: %w", err)
		}
		return data, nil
	})
}

func (s *dataSource) fetchAllData(ctx context.Context) (*WNCDataCache, error) {
	data := &WNCDataCache{}

	// Define all data fetchers with their criticality
	fetchers := []dataFetcher{
		{"AP CAPWAP", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListCAPWAPData(ctx)
			if err != nil {
				return err
			}
			c.CAPWAPData = data.CAPWAPData
			slog.Info(
				"AP CAPWAP data retrieved from WNC successfully",
				"count",
				len(data.CAPWAPData),
			)
			return nil
		}},
		{"AP operational", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListApOperData(ctx)
			if err != nil {
				return err
			}
			c.ApOperData = data.OperData
			slog.Info(
				"AP operational data retrieved from WNC successfully",
				"count",
				len(data.OperData),
			)
			return nil
		}},
		{"radio operational", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListRadioData(ctx)
			if err != nil {
				return err
			}
			c.RadioOperData = data.RadioOperData
			slog.Info(
				"radio operational data retrieved from WNC successfully",
				"count",
				len(data.RadioOperData),
			)
			return nil
		}},
		{"AP name to MAC mapping", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListNameMACMaps(ctx)
			if err != nil {
				return err
			}
			c.NameMACMaps = data.ApNameMACMap
			slog.Info(
				"AP name to MAC mapping data retrieved from WNC successfully",
				"count",
				len(data.ApNameMACMap),
			)
			return nil
		}},
		{"RRM measurements", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.RRM().ListRRMMeasurement(ctx)
			if err != nil {
				return err
			}
			c.RRMMeasurements = data.RRMMeasurement
			slog.Info(
				"RRM measurement data retrieved from WNC successfully",
				"count",
				len(data.RRMMeasurement),
			)
			return nil
		}},
		{"WLAN configuration entries", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListWlanCfgEntries(ctx)
			if err != nil {
				return err
			}
			if data != nil && data.WlanCfgEntries != nil {
				c.WLANConfigEntries = data.WlanCfgEntries.WlanCfgEntry
			} else {
				c.WLANConfigEntries = []wlan.WlanCfgEntry{}
			}
			slog.Info(
				"WLAN configuration entries retrieved from WNC successfully",
				"count",
				len(c.WLANConfigEntries),
			)
			return nil
		}},
		{"WLAN policies", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListWlanPolicies(ctx)
			if err != nil {
				return err
			}
			if data != nil && data.WlanPolicies != nil {
				c.WLANPolicies = data.WlanPolicies.WlanPolicy
			} else {
				c.WLANPolicies = []wlan.WlanPolicy{}
			}
			slog.Info("WLAN policies retrieved from WNC successfully", "count", len(c.WLANPolicies))
			return nil
		}},
		{"WLAN policy list entries", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.WLAN().ListCfgPolicyListEntries(ctx)
			if err != nil {
				return err
			}
			if data != nil && data.PolicyListEntries != nil {
				c.WLANPolicyListEntries = data.PolicyListEntries.PolicyListEntry
			} else {
				c.WLANPolicyListEntries = []wlan.PolicyListEntry{}
			}
			slog.Info("WLAN policy list entries retrieved from WNC successfully", "count", len(c.WLANPolicyListEntries))
			return nil
		}},
		{"client common", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListCommonInfo(ctx)
			if err != nil {
				return err
			}
			c.CommonOperData = data.CommonOperData
			slog.Info(
				"client common data retrieved from WNC successfully",
				"count",
				len(data.CommonOperData),
			)
			return nil
		}},
		{"client device classification", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListDCInfo(ctx)
			if err != nil {
				return err
			}
			c.DCInfo = data.DcInfo
			slog.Info(
				"client device classification data retrieved from WNC successfully",
				"count",
				len(data.DcInfo),
			)
			return nil
		}},
		{"client 802.11", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListDot11Info(ctx)
			if err != nil {
				return err
			}
			c.Dot11OperData = data.Dot11OperData
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
			c.SisfDBMac = data.SisfDBMac
			slog.Info(
				"SISF database data retrieved from WNC successfully",
				"count",
				len(data.SisfDBMac),
			)
			return nil
		}},
		{"client traffic stats", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListTrafficStats(ctx)
			if err != nil {
				return err
			}
			c.TrafficStats = data.TrafficStats
			slog.Info(
				"client traffic stats data retrieved from WNC successfully",
				"count",
				len(data.TrafficStats),
			)
			return nil
		}},
		{"client mobility history", true, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.Client().ListMMIFClientHistory(ctx)
			if err != nil {
				return err
			}
			c.MmIfClientHistory = data.MmIfClientHistory
			slog.Info(
				"client mobility history data retrieved from WNC successfully",
				"count",
				len(data.MmIfClientHistory),
			)
			return nil
		}},
		{"radio statistics", false, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListRadioOperStats(ctx)
			if err != nil {
				return err
			}
			c.RadioOperStats = data.RadioOperStats
			slog.Info(
				"radio statistics data retrieved from WNC successfully",
				"count",
				len(data.RadioOperStats),
			)
			return nil
		}},
		{"radio reset statistics", false, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.AP().ListRadioResetStats(ctx)
			if err != nil {
				return err
			}
			c.RadioResetStats = data.RadioResetStats
			slog.Info(
				"radio reset statistics retrieved from WNC successfully",
				"count",
				len(data.RadioResetStats),
			)
			return nil
		}},
		{"RRM coverage", false, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.RRM().ListRRMCoverage(ctx)
			if err != nil {
				return err
			}
			c.RRMCoverage = data.RRMCoverage
			slog.Info(
				"RRM coverage data retrieved from WNC successfully",
				"count",
				len(data.RRMCoverage),
			)
			return nil
		}},
		{"AP radar data", false, func(ctx context.Context, c *WNCDataCache) error {
			data, err := s.client.RRM().ListApDot11RadarData(ctx)
			if err != nil {
				return err
			}
			c.ApDot11RadarData = data.ApDot11RadarData
			slog.Info(
				"AP radar data retrieved from WNC successfully",
				"count",
				len(data.ApDot11RadarData),
			)
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
