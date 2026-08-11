// Package collector provides utilities for WNC collectors.
package collector

import (
	"slices"
	"strconv"
	"strings"
)

// percentScale converts a ratio to a percentage.
const percentScale = 100.0

// Client state constants.
const (
	ClientStatusRun = "client-status-run"
)

// AP state constants.
const (
	APRadioStateUp      = "radio-up"
	APAdminStateEnabled = "enabled"
)

type WirelessProtocol int

const (
	ProtocolUnknown WirelessProtocol = iota
	Protocol11A
	Protocol11B
	Protocol11G
	ProtocolN
	ProtocolAC
	ProtocolAX
	ProtocolBE
)

// MapWirelessProtocol maps WNC PHY type strings to WirelessProtocol enum values.
//
// phyType is ms-phy-radio-type and radioType is ms-radio-type; they are different
// typedefs. is-11g-client is not an input: the controller sets it on clients whose
// PHY type is not 802.11g, so using it as a fallback reported that generation for
// clients on bands where the PHY does not exist.
func MapWirelessProtocol(phyType, radioType string) WirelessProtocol {
	switch {
	case strings.Contains(phyType, "dot11be"):
		return ProtocolBE
	case strings.Contains(phyType, "dot11ax"):
		return ProtocolAX
	case strings.Contains(phyType, "dot11ac"):
		return ProtocolAC
	case strings.Contains(phyType, "dot11n"):
		return ProtocolN
	// ms-phy-radio-type spells the legacy generations without a band suffix, so these
	// must be matched before the substrings above would catch a longer name.
	case phyType == "client-dot11g":
		return Protocol11G
	case phyType == "client-dot11b":
		return Protocol11B
	case phyType == "client-dot11a":
		return Protocol11A
	case radioType == "dot11-radio-type-a":
		// dot11-oper-data/radio-type names the band, not the generation. It is only
		// consulted when the PHY type is absent, and 802.11a is the only generation
		// the band alone implies.
		return Protocol11A
	default:
		return ProtocolUnknown
	}
}

func stringToUint64(value string) uint64 {
	if value == "" {
		return 0
	}
	if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
		return parsed
	}
	return 0
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// buildInfoLabels constructs info labels slice based on configuration.
func buildInfoLabels(requiredLabels, configuredLabels, availableLabels []string) []string {
	labels := make([]string, 0, len(requiredLabels)+len(configuredLabels))
	labels = append(labels, requiredLabels...)

	for _, label := range availableLabels {
		if slices.Contains(configuredLabels, label) && !slices.Contains(labels, label) {
			labels = append(labels, label)
		}
	}

	return labels
}

// IsEnabled returns true if any of the provided boolean values is true.
func IsEnabled(flags ...bool) bool {
	for _, enabled := range flags {
		if enabled {
			return true
		}
	}
	return false
}
