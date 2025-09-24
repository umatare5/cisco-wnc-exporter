// Package collector provides common utilities and constants for WNC collectors.
package collector

import "strings"

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

// WirelessProtocol represents 802.11 wireless protocols using iota for efficient encoding.
type WirelessProtocol int

const (
	ProtocolUnknown WirelessProtocol = iota // 0: Unknown or unrecognized protocol
	Protocol11A                             // 1: 802.11a (5GHz legacy)
	Protocol11B                             // 2: 802.11b (2.4GHz legacy)
	Protocol11G                             // 3: 802.11g (2.4GHz legacy)
	ProtocolN                               // 4: 802.11n (Wi-Fi 4)
	ProtocolAC                              // 5: 802.11ac (Wi-Fi 5)
	ProtocolAX                              // 6: 802.11ax (Wi-Fi 6/6E)
	ProtocolBE                              // 7: 802.11be (Wi-Fi 7)
)

// MapRadioTypeToSlot maps radio type string to radio slot ID.
func MapRadioTypeToSlot(radioType string) int {
	switch radioType {
	case "dot11bg", "client-dot11ax-24ghz-prot", "client-dot11n-24-ghz-prot", "client-dot11bg-24-ghz-prot":
		return RadioSlot24GHz
	case "dot11a", "client-dot11ax-5ghz-prot", "client-dot11ac-5-ghz-prot",
		"client-dot11n-5-ghz-prot", "client-dot11a-5-ghz-prot":
		return RadioSlot5GHz
	case "client-dot11ax-6ghz-prot":
		return RadioSlot6GHz
	default:
		return -1 // unknown radio type
	}
}

// MapRadioSlotToBand maps radio slot ID to band string.
// This function provides a direct mapping from radio slot ID to band designation.
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

// DetermineBandFromRadioInfo determines radio band from slot ID and radio type.
// This function uses the centralized mapping functions for consistency.
func DetermineBandFromRadioInfo(radioSlotID int, radioType string) string {
	// First try to determine band from radio slot ID
	band := MapRadioSlotToBand(radioSlotID)

	// If the slot-based band is unknown or if we have radio type info,
	// use the radio type for more accurate band determination
	if band == BandUnknown || radioType != "" {
		if typeBasedSlot := MapRadioTypeToSlot(radioType); typeBasedSlot != -1 {
			band = MapRadioSlotToBand(typeBasedSlot)
		}
	}

	return band
}

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

// MapWirelessProtocol maps WNC PHY type strings and radio information to WirelessProtocol enum values.
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
		// 802.11b/g mixed mode - determine by is-11g-client flag
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

// boolToFloat64 converts boolean to float64 for Prometheus metrics (0.0 or 1.0).
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
