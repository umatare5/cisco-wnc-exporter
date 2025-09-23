// Package collector provides common utilities and constants for WNC collectors.
package collector

// Client state constants.
const (
	ClientStatusRun       = "client-status-run"
	ClientStateAssociated = 2
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

// DetermineBandFromRadioInfo determines radio band from slot ID and radio type.
func DetermineBandFromRadioInfo(radioSlotID int, radioType string) string {
	// Determine band from radio slot ID
	band := BandUnknown
	switch radioSlotID {
	case RadioSlot24GHz:
		band = Band24GHz
	case RadioSlot5GHz:
		band = Band5GHz
	case RadioSlot6GHz:
		band = Band6GHz
	}

	// Update band based on radio type if available
	switch radioType {
	case "dot11bg", "client-dot11ax-24ghz-prot", "client-dot11n-24-ghz-prot", "client-dot11bg-24-ghz-prot":
		band = Band24GHz
	case "dot11a", "client-dot11ax-5ghz-prot", "client-dot11ac-5-ghz-prot",
		"client-dot11n-5-ghz-prot", "client-dot11a-5-ghz-prot":
		band = Band5GHz
	case "client-dot11ax-6ghz-prot":
		band = Band6GHz
	}

	return band
}
