// Package collector provides collectors for cisco-wnc-exporter.
package collector

// Label constants used across collectors for Prometheus metrics.
const (
	labelBand = "band" // Radio frequency band (2.4GHz, 5GHz, 6GHz)
	labelMAC  = "mac"  // MAC address
	labelName = "name" // Human-readable name

	// AP-specific labels.
	labelChannel   = "channel"    // CAPWAP tunnel channel (control, data)
	labelEthMAC    = "eth_mac"    // AP Ethernet MAC address
	labelIP        = "ip"         // AP IP address
	labelModel     = "model"      // AP model number
	labelProfile   = "profile"    // RRM profile a radio is judged against
	labelRadio     = "radio"      // Radio slot identifier
	labelSerial    = "serial"     // AP serial number
	labelSWVersion = "sw_version" // AP software version

	// Client-specific labels.
	labelAP       = "ap"       // Access Point name
	labelIPv4     = "ipv4"     // Client IPv4 address
	labelIPv6     = "ipv6"     // Client IPv6 address
	labelUsername = "username" // Client authentication username
	labelWLAN     = "wlan"     // WLAN SSID name
	// labelWLANID is not spelled id: wnc_wlan_info already carries the WLAN identifier
	// under that name, and an id on a client series cannot say whose it is.
	labelWLANID     = "wlan_id"     // WLAN identifier the client is associated to
	labelDeviceType = "device_type" // Device class the controller assigned the client

	// WLAN-specific labels.
	labelID = "id" // WLAN identifier
	// labelPhase names one of four sibling leaves rather than one value of an enum, so
	// its values are this exporter's own and not the controller's spelling.
	labelPhase         = "phase"          // Onboarding phase a client is held in
	labelPolicyProfile = "policy_profile" // Policy profile a WLAN is bound to
	labelPolicyTag     = "policy_tag"     // Policy tag carrying the binding

	// Controller-specific labels.
	labelReason = "reason" // Reason a controller-wide counter is keyed by

	// Refresh health labels.
	labelData = "data" // WNC data type identifier
)
