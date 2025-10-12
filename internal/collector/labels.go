// Package collector provides collectors for cisco-wnc-exporter.
package collector

// Label constants used across collectors for Prometheus metrics.
const (
	labelBand = "band" // Radio frequency band (2.4GHz, 5GHz, 6GHz)
	labelMAC  = "mac"  // MAC address
	labelName = "name" // Human-readable name

	// AP-specific labels.
	labelEthMAC    = "eth_mac"    // AP Ethernet MAC address
	labelIP        = "ip"         // AP IP address
	labelModel     = "model"      // AP model number
	labelRadio     = "radio"      // Radio slot identifier
	labelSerial    = "serial"     // AP serial number
	labelSWVersion = "sw_version" // AP software version

	// Client-specific labels.
	labelAP       = "ap"       // Access Point name
	labelIPv4     = "ipv4"     // Client IPv4 address
	labelIPv6     = "ipv6"     // Client IPv6 address
	labelUsername = "username" // Client authentication username
	labelWLAN     = "wlan"     // WLAN SSID name

	// WLAN-specific labels.
	labelID = "id" // WLAN identifier
)
