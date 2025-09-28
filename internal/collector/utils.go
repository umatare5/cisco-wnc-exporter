// Package collector provides utilities for WNC collectors.
package collector

import (
	"slices"
	"strconv"
	"strings"
)

// Client state constants.
const (
	ClientStatusRun       = "client-status-run"
	ClientStateAssociated = 2
)

// AP state constants.
const (
	APRadioStateUp      = "radio-up"
	APAdminStateEnabled = "enabled"
)

// Band constants for radio bands.
const (
	BandUnknown = "unknown"
	Band24GHz   = "2.4"
	Band5GHz    = "5"
	Band6GHz    = "6"
)

// Radio slot constants for band mapping.
const (
	RadioSlot24GHz = 0
	RadioSlot5GHz  = 1
	RadioSlot6GHz  = 2
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

// MapClientState maps client operational state to numeric value.
func MapClientState(state string) int {
	switch state {
	case ClientStatusRun:
		return ClientStateAssociated // associated
	case "client-status-authenticated":
		return 1 // authenticated
	default:
		return 0 // disconnected
	}
}

// MapWirelessProtocol maps WNC PHY type strings to WirelessProtocol enum values.
func MapWirelessProtocol(phyType, radioType string, is11GClient bool) WirelessProtocol {
	switch {
	case strings.Contains(phyType, "dot11n"):
		return ProtocolN
	case strings.Contains(phyType, "dot11ac"):
		return ProtocolAC
	case strings.Contains(phyType, "dot11ax"):
		return ProtocolAX
	case strings.Contains(phyType, "dot11be"), strings.Contains(phyType, "eht"):
		return ProtocolBE
	case strings.Contains(phyType, "dot11bg"):
		if is11GClient {
			return Protocol11G
		}
		return Protocol11B
	case strings.Contains(phyType, "dot11a") || radioType == "dot11-radio-type-a":
		return Protocol11A
	case strings.Contains(phyType, "dot11g") || is11GClient:
		return Protocol11G
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

// DetermineBandFromRadioInfo determines radio band from slot ID and radio type.
func DetermineBandFromRadioInfo(radioSlotID int, radioType string) string {
	band := MapRadioSlotToBand(radioSlotID)

	if band == BandUnknown || radioType != "" {
		if typeBasedSlot := MapRadioTypeToSlot(radioType); typeBasedSlot != -1 {
			band = MapRadioSlotToBand(typeBasedSlot)
		}
	}

	return band
}

// MapRadioTypeToSlot maps radio type string to radio slot ID.
func MapRadioTypeToSlot(radioType string) int {
	switch radioType {
	case "dot11bg",
		"client-dot11ax-24ghz-prot", "client-dot11n-24-ghz-prot", "client-dot11bg-24-ghz-prot":
		return RadioSlot24GHz
	case "dot11a",
		"client-dot11ax-5ghz-prot", "client-dot11ac-5-ghz-prot", "client-dot11n-5-ghz-prot", "client-dot11a-5-ghz-prot":
		return RadioSlot5GHz
	case "client-dot11ax-6ghz-prot":
		return RadioSlot6GHz
	default:
		return 0 // unknown state
	}
}

// MapRadioSlotToBand maps radio slot ID to band string.
func MapRadioSlotToBand(radioSlotID int) string {
	switch radioSlotID {
	case RadioSlot24GHz:
		return Band24GHz
	case RadioSlot5GHz:
		return Band5GHz
	case RadioSlot6GHz:
		return Band6GHz
	default:
		return BandUnknown
	}
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
