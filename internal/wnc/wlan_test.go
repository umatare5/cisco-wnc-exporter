package wnc

import (
	"context"
	"errors"
	"testing"

	"github.com/umatare5/cisco-ios-xe-wireless-go/service/wlan"
)

// mockWLANDataSource implements DataSource for testing WLAN operations.
type mockWLANDataSource struct {
	data *WNCDataCache
	err  error
}

func (m *mockWLANDataSource) GetCachedData(ctx context.Context) (*WNCDataCache, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// newMockWLANDataSource creates a mock DataSource with WLAN sample data based on live WNC.
func newMockWLANDataSource() *mockWLANDataSource {
	return &mockWLANDataSource{
		data: &WNCDataCache{
			WLANConfigEntries: []wlan.WlanCfgEntry{
				{
					ProfileName:      "labo-wlan",
					WlanID:           1,
					AuthKeyMgmtPsk:   true,
					AuthKeyMgmtDot1x: false,
					PSK:              "XNU]F]gFOUKOSYA]FKWIJd\\JG]LBKIfUWY_BFANFBC",
					PSKType:          "aes",
					FTMode:           "dot11r-disabled",
					APFVapIDData: &wlan.APFVapIDData{
						SSID:       "labo-wlan",
						WlanStatus: true,
					},
					MDNSSDMode: "mdns-sd-drop",
				},
				{
					ProfileName:            "labo-tls",
					WlanID:                 4,
					WPA2Enabled:            false,
					WPA3Enabled:            true,
					AuthKeyMgmtDot1x:       false,
					AuthKeyMgmtDot1xSha256: true,
					FTMode:                 "dot11r-disabled",
					PMFOptions:             "apf-vap-pmf-required",
					LoadBalance:            true,
					AuthenticationList:     "EAP_RADIUS",
					Wlan11kNeighList:       false,
					APFVapIDData: &wlan.APFVapIDData{
						SSID:       "labo-tls",
						WlanStatus: true,
					},
					APFVap80211vData: &wlan.APFVap80211vData{
						Dot11vDms: false,
					},
					MDNSSDMode:     "mdns-sd-drop",
					ClientSteering: true,
				},
			},
			WLANPolicies: []wlan.WlanPolicy{
				{
					PolicyProfileName: "labo-wlan-profile",
					Status:            true,
					InterfaceName:     "LAB-INTERNAL",
					WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
						CentralSwitching:      false,
						CentralAuthentication: false,
						CentralDHCP:           false,
						CentralAssocEnable:    false,
					},
					WlanTimeout: &wlan.WlanTimeout{
						SessionTimeout: 43200,
					},
					PerSsidQos: &wlan.PerSsidQos{
						IngressServiceName: "platinum-up",
						EgressServiceName:  "platinum",
					},
					DHCPParams: &wlan.DHCPParams{
						IsDHCPEnabled: true,
					},
					UmbrellaFlexParams: &wlan.UmbrellaFlexParams{
						DHCPDNSOptionEnable: false,
					},
				},
				{
					PolicyProfileName: "default-policy-profile",
					Description:       "Preconfigured default policy profile",
					Status:            true,
					WlanSwitchingPolicy: &wlan.WlanSwitchingPolicy{
						CentralSwitching: false,
						CentralDHCP:      false,
					},
				},
			},
			WLANPolicyListEntries: []wlan.PolicyListEntry{
				{
					TagName: "labo-wlan-flex",
					WLANPolicies: &wlan.WLANPolicies{
						WLANPolicy: []wlan.WLANPolicyMap{
							{
								WLANProfileName:   "labo-wlan",
								PolicyProfileName: "labo-wlan-profile",
							},
							{
								WLANProfileName:   "labo-psk",
								PolicyProfileName: "labo-wlan-profile",
							},
							{
								WLANProfileName:   "labo-tls",
								PolicyProfileName: "labo-wlan-profile",
							},
						},
					},
				},
				{
					TagName:     "default-policy-tag",
					Description: "Preconfigured default policy-tag",
				},
			},
		},
	}
}

func TestNewWLANSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		sharedDataSource DataSource
		expectNil        bool
	}{
		{
			name:             "With non-nil shared data source",
			sharedDataSource: newMockWLANDataSource(),
			expectNil:        false,
		},
		{
			name:             "With nil shared data source",
			sharedDataSource: nil,
			expectNil:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewWLANSource(tt.sharedDataSource)
			if source == nil && !tt.expectNil {
				t.Error("Expected non-nil WLANSource, got nil")
			}
			if source != nil && tt.expectNil {
				t.Error("Expected nil WLANSource, got non-nil")
			}
		})
	}
}

func TestWLANSource_ListConfigEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockWLANDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple WLAN entries",
			mock:    newMockWLANDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockWLANDataSource{
				data: &WNCDataCache{
					WLANConfigEntries: []wlan.WlanCfgEntry{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockWLANDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewWLANSource(tt.mock)
			entries, err := source.ListConfigEntries(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListConfigEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(entries) != tt.wantLen {
				t.Errorf("ListConfigEntries() got %d entries, want %d", len(entries), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstEntry := entries[0]
				if firstEntry.ProfileName == "" {
					t.Error("Expected ProfileName to be set")
				}
				if firstEntry.WlanID == 0 {
					t.Error("Expected WlanID to be set")
				}
				if firstEntry.APFVapIDData == nil {
					t.Error("Expected APFVapIDData to be set")
				}
			}
		})
	}
}

func TestWLANSource_ListPolicies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockWLANDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple policies",
			mock:    newMockWLANDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockWLANDataSource{
				data: &WNCDataCache{
					WLANPolicies: []wlan.WlanPolicy{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockWLANDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewWLANSource(tt.mock)
			policies, err := source.ListPolicies(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(policies) != tt.wantLen {
				t.Errorf("ListPolicies() got %d policies, want %d", len(policies), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstPolicy := policies[0]
				if firstPolicy.PolicyProfileName == "" {
					t.Error("Expected PolicyProfileName to be set")
				}
				if firstPolicy.WlanSwitchingPolicy == nil {
					t.Error("Expected WlanSwitchingPolicy to be set")
				}
			}
		})
	}
}

func TestWLANSource_ListPolicyListEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockWLANDataSource
		wantLen int
		wantErr bool
	}{
		{
			name:    "Success with multiple policy list entries",
			mock:    newMockWLANDataSource(),
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Success with empty data",
			mock: &mockWLANDataSource{
				data: &WNCDataCache{
					WLANPolicyListEntries: []wlan.PolicyListEntry{},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "Error from data source",
			mock: &mockWLANDataSource{
				err: errors.New("cache refresh failed"),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source := NewWLANSource(tt.mock)
			policyListEntries, err := source.ListPolicyListEntries(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPolicyListEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(policyListEntries) != tt.wantLen {
				t.Errorf("ListPolicyListEntries() got %d entries, want %d", len(policyListEntries), tt.wantLen)
			}

			// Verify structure for successful case with data
			if !tt.wantErr && tt.wantLen > 0 {
				firstEntry := policyListEntries[0]
				if firstEntry.TagName == "" {
					t.Error("Expected TagName to be set")
				}
				// WLANPolicies may be nil for entries like "default-policy-tag"
				if firstEntry.WLANPolicies != nil && len(firstEntry.WLANPolicies.WLANPolicy) > 0 {
					firstMapping := firstEntry.WLANPolicies.WLANPolicy[0]
					if firstMapping.WLANProfileName == "" {
						t.Error("Expected WLANProfileName to be set")
					}
					if firstMapping.PolicyProfileName == "" {
						t.Error("Expected PolicyProfileName to be set")
					}
				}
			}
		})
	}
}
