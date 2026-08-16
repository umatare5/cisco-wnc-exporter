// Package wnc provides WNC data access and caching.
// This file holds the index of RESTCONF data types a refresh walks.
package wnc

import (
	"context"
	"encoding/json"

	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// dataTypeNames lists every data type a refresh attempts, in fetch order.
// A refresh truncated by its deadline drops the tail, so ap_capwap_data comes
// first: the AP inventory is what every other AP series is labeled from.
var dataTypeNames = []string{
	dataAPCAPWAPData,
	dataAPOperData,
	dataAPRadioOperData,
	dataAPNameMACMap,
	dataAPJoinStats,
	dataRRMMeasurement,
	dataWLANCfgEntries,
	dataWLANPolicies,
	dataWLANPolicyListEntries,
	dataWLANClientStats,
	dataControllerBootTime,
	dataCoClientDelReason,
	dataClientRoamingStats,
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

// boolToInt reports one item for a leaf the controller carries and none for one it
// omits, so wnc_refresh_items reads as the count of what the read published.
func boolToInt(present bool) int {
	if present {
		return 1
	}

	return 0
}

// anyOf reports whether any flag is set. internal/collector has the same predicate,
// but importing it here would be a cycle.
func anyOf(flags ...bool) bool {
	for _, flag := range flags {
		if flag {
			return true
		}
	}

	return false
}

// requiredDataTypes returns the data types the enabled modules read, in the order
// of dataTypeNames. A data type no enabled module reads costs one serial request
// per refresh and buys nothing.
func requiredDataTypes(modules config.Collectors) []string {
	names := make([]string, 0, len(dataTypeNames))

	for _, name := range dataTypeNames {
		if isDataTypeRequired(name, modules) {
			names = append(names, name)
		}
	}

	return names
}

// isDataTypeRequired mirrors the source calls each Collect method makes under its
// module guards, so it has to be updated alongside them. Three data types are read
// before any guard, and one is read by three modules across two collectors, so the
// relation is a union over the enabled modules rather than a per-module list.
//
// An unlisted data type is fetched. Paying for a request is recoverable, while
// withholding one relies on the caller marking it absent.
func isDataTypeRequired(name string, modules config.Collectors) bool {
	anyAP := anyOf(modules.AP.General, modules.AP.Radio,
		modules.AP.Traffic, modules.AP.Errors, modules.AP.Info)
	anyClient := anyOf(modules.Client.General, modules.Client.Radio,
		modules.Client.Traffic, modules.Client.Errors, modules.Client.Info)
	anyWLAN := anyOf(modules.WLAN.General, modules.WLAN.Traffic,
		modules.WLAN.Config, modules.WLAN.Info)

	switch name {
	case dataAPCAPWAPData, dataAPRadioOperData:
		return anyAP
	case dataAPOperData:
		return modules.AP.General
	case dataAPNameMACMap, dataRRMMeasurement:
		return modules.AP.Radio
	case dataAPJoinStats:
		// The join module is keyed by the statistics list itself, which keeps a record
		// for an AP the inventory has dropped, so it reads no other AP data type.
		return modules.AP.Join
	case dataAPRadioOperStats:
		return anyOf(modules.AP.Traffic, modules.AP.Errors)
	case dataAPRadioResetStats, dataRRMCoverage, dataRRMAPDot11RadarData:
		return modules.AP.Errors
	case dataControllerBootTime, dataCoClientDelReason, dataClientRoamingStats:
		return modules.Controller.General
	case dataWLANCfgEntries:
		return anyWLAN
	case dataWLANPolicies, dataWLANPolicyListEntries:
		return modules.WLAN.Config
	case dataWLANClientStats:
		return modules.WLAN.Traffic
	case dataClientCommonOperData:
		// The per-radio and per-WLAN client counts read it through their own
		// collectors, so a client module is not the only reason to fetch it.
		return anyOf(anyClient, modules.AP.Radio, modules.WLAN.Traffic)
	case dataClientDCInfo, dataClientSISFDBMac:
		return modules.Client.Info
	case dataClientDot11OperData:
		return anyOf(modules.Client.General, modules.Client.Radio, modules.Client.Info)
	case dataClientTrafficStats:
		return anyOf(modules.Client.General, modules.Client.Radio,
			modules.Client.Traffic, modules.Client.Errors)
	case dataClientMMIFHistory:
		return modules.Client.General
	default:
		return true
	}
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
		{dataAPJoinStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListAPJoinStats(ctx)
			if err != nil {
				return 0, err
			}
			c.JoinStats = data.ApJoinStats
			return len(c.JoinStats), nil
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
		{dataWLANClientStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			data, err := s.client.AP().ListWLANClientStats(ctx)
			if err != nil {
				return 0, err
			}
			c.WLANClientStats = data.WlanClientStats
			return len(c.WLANClientStats), nil
		}},
		{dataControllerBootTime, func(ctx context.Context, c *WNCDataCache) (int, error) {
			bootTime, present, err := rawValue[string](ctx, s.client.Core(), routeControllerBootTime)
			if err != nil {
				return 0, err
			}
			c.ControllerBootTime = bootTime
			return boolToInt(present), nil
		}},
		{dataCoClientDelReason, func(ctx context.Context, c *WNCDataCache) (int, error) {
			leaves, present, err := rawValue[map[string]json.RawMessage](
				ctx, s.client.Core(), routeCoClientDelReason,
			)
			if err != nil {
				return 0, err
			}
			if !present {
				return 0, nil
			}
			c.ClientDeleteReasons = numericLeaves(leaves, dataCoClientDelReason)
			return len(c.ClientDeleteReasons), nil
		}},
		{dataClientRoamingStats, func(ctx context.Context, c *WNCDataCache) (int, error) {
			leaves, present, err := rawValue[map[string]json.RawMessage](
				ctx, s.client.Core(), routeClientRoamingStats,
			)
			if err != nil {
				return 0, err
			}
			if !present {
				return 0, nil
			}
			c.ClientRoamingStats = numericLeaves(leaves, dataClientRoamingStats)
			return len(c.ClientRoamingStats), nil
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
