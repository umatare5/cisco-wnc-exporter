// Package wnc provides WNC data access and caching.
// This file holds the index of RESTCONF data types a refresh walks.
package wnc

import "context"

// dataTypeNames lists every data type a refresh attempts, in fetch order.
// A refresh truncated by its deadline drops the tail, so ap_capwap_data comes
// first: the AP inventory is what every other AP series is labeled from.
var dataTypeNames = []string{
	dataAPCAPWAPData,
	dataAPOperData,
	dataAPRadioOperData,
	dataAPNameMACMap,
	dataRRMMeasurement,
	dataWLANCfgEntries,
	dataWLANPolicies,
	dataWLANPolicyListEntries,
	dataClientCommonOperData,
	dataClientDCInfo,
	dataClientDot11OperData,
	dataClientSISFDBMac,
	dataClientTrafficStats,
	dataClientMMIFHistory,
	dataAPRadioOperStats,
	dataAPRadioResetStats,
	dataRRMCoverage,
	dataRRMAPDot11RadarData,
}

// fetchers returns the data fetchers in the order of dataTypeNames.
func (s *dataSource) fetchers() []dataFetcher {
	return []dataFetcher{
		{dataAPCAPWAPData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListCAPWAPData(ctx)
			if err != nil {
				return 0, err
			}
			c.CAPWAPData = data.CAPWAPData
			return len(c.CAPWAPData), nil
		}},
		{dataAPOperData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListApOperData(ctx)
			if err != nil {
				return 0, err
			}
			c.ApOperData = data.OperData
			return len(c.ApOperData), nil
		}},
		{dataAPRadioOperData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListRadioData(ctx)
			if err != nil {
				return 0, err
			}
			c.RadioOperData = data.RadioOperData
			return len(c.RadioOperData), nil
		}},
		{dataAPNameMACMap, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListNameMACMaps(ctx)
			if err != nil {
				return 0, err
			}
			c.NameMACMaps = data.ApNameMACMap
			return len(c.NameMACMaps), nil
		}},
		{dataRRMMeasurement, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.RRM().ListRRMMeasurement(ctx)
			if err != nil {
				return 0, err
			}
			c.RRMMeasurements = data.RRMMeasurement
			return len(c.RRMMeasurements), nil
		}},
		{dataWLANCfgEntries, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := readEffective(ctx, &s.defaultsFallbacks, s.client.WLAN().ListWlanCfgEntries)
			if err != nil {
				return 0, err
			}
			if data != nil && data.WlanCfgEntries != nil {
				c.WLANConfigEntries = data.WlanCfgEntries.WlanCfgEntry
			}
			return len(c.WLANConfigEntries), nil
		}},
		{dataWLANPolicies, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := readEffective(ctx, &s.defaultsFallbacks, s.client.WLAN().ListWlanPolicies)
			if err != nil {
				return 0, err
			}
			if data != nil && data.WlanPolicies != nil {
				c.WLANPolicies = data.WlanPolicies.WlanPolicy
			}
			return len(c.WLANPolicies), nil
		}},
		{dataWLANPolicyListEntries, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.WLAN().ListCfgPolicyListEntries(ctx)
			if err != nil {
				return 0, err
			}
			if data != nil && data.PolicyListEntries != nil {
				c.WLANPolicyListEntries = data.PolicyListEntries.PolicyListEntry
			}
			return len(c.WLANPolicyListEntries), nil
		}},
		{dataClientCommonOperData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListCommonInfo(ctx)
			if err != nil {
				return 0, err
			}
			c.CommonOperData = data.CommonOperData
			return len(c.CommonOperData), nil
		}},
		{dataClientDCInfo, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListDCInfo(ctx)
			if err != nil {
				return 0, err
			}
			c.DCInfo = data.DcInfo
			return len(c.DCInfo), nil
		}},
		{dataClientDot11OperData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListDot11Info(ctx)
			if err != nil {
				return 0, err
			}
			c.Dot11OperData = data.Dot11OperData
			return len(c.Dot11OperData), nil
		}},
		{dataClientSISFDBMac, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListSISFDB(ctx)
			if err != nil {
				return 0, err
			}
			c.SisfDBMac = data.SisfDBMac
			return len(c.SisfDBMac), nil
		}},
		{dataClientTrafficStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListTrafficStats(ctx)
			if err != nil {
				return 0, err
			}
			c.TrafficStats = data.TrafficStats
			return len(c.TrafficStats), nil
		}},
		{dataClientMMIFHistory, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.Client().ListMMIFClientHistory(ctx)
			if err != nil {
				return 0, err
			}
			c.MmIfClientHistory = data.MmIfClientHistory
			return len(c.MmIfClientHistory), nil
		}},
		{dataAPRadioOperStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListRadioOperStats(ctx)
			if err != nil {
				return 0, err
			}
			c.RadioOperStats = data.RadioOperStats
			return len(c.RadioOperStats), nil
		}},
		{dataAPRadioResetStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListRadioResetStats(ctx)
			if err != nil {
				return 0, err
			}
			c.RadioResetStats = data.RadioResetStats
			return len(c.RadioResetStats), nil
		}},
		{dataRRMCoverage, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.RRM().ListRRMCoverage(ctx)
			if err != nil {
				return 0, err
			}
			c.RRMCoverage = data.RRMCoverage
			return len(c.RRMCoverage), nil
		}},
		{dataRRMAPDot11RadarData, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.RRM().ListApDot11RadarData(ctx)
			if err != nil {
				return 0, err
			}
			c.ApDot11RadarData = data.ApDot11RadarData
			return len(c.ApDot11RadarData), nil
		}},
	}
}
