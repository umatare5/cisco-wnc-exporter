package wnc

import (
	"context"
	"errors"
	"testing"
	"time"
)

// testBootInstant is the boot instant the source is asked to carry through unchanged.
var testBootInstant = time.Date(2026, 1, 13, 0, 0, 0, 0, time.UTC)

func TestControllerSource_GetBootTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		want    *time.Time
		wantErr bool
	}{
		{
			name: "Success with a boot instant",
			mock: &mockDataSource{
				data: &WNCDataCache{ControllerBootTime: ptr(testBootInstant)},
			},
			want: ptr(testBootInstant),
		},
		{
			// A controller that does not carry the leaf answers with no body, which the
			// refresh records as a successful read of nothing rather than as a failure.
			name:    "Nil when the controller carries no leaf",
			mock:    &mockDataSource{data: &WNCDataCache{}},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Error from data source",
			mock:    &mockDataSource{err: errors.New("cache refresh failed")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewControllerSource(tt.mock)

			got, err := source.GetBootTime(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBootTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			switch {
			case got == nil && tt.want == nil:
			case got == nil || tt.want == nil:
				t.Errorf("GetBootTime() = %v, want %v", got, tt.want)
			case !got.Equal(*tt.want):
				t.Errorf("GetBootTime() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestControllerSource_GetClientRoamingStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name: "Success with roam counters",
			mock: &mockDataSource{
				data: &WNCDataCache{ClientRoamingStats: map[string]float64{"ap-auth-roams": 30829}},
			},
			wantLen: 1,
		},
		{
			name:    "Nil when the controller carries no container",
			mock:    &mockDataSource{data: &WNCDataCache{}},
			wantLen: 0,
		},
		{
			name:    "Error from data source",
			mock:    &mockDataSource{err: errors.New("cache refresh failed")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewControllerSource(tt.mock)

			got, err := source.GetClientRoamingStats(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClientRoamingStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("GetClientRoamingStats() returned %d leaves, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestControllerSource_GetClientDeleteReasons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		wantLen int
		wantErr bool
	}{
		{
			name: "Success with reasons",
			mock: &mockDataSource{
				data: &WNCDataCache{ClientDeleteReasons: map[string]float64{"ap-delete": 1, "bssid-down": 0}},
			},
			wantLen: 2,
		},
		{
			name:    "Nil when the controller carries no container",
			mock:    &mockDataSource{data: &WNCDataCache{}},
			wantLen: 0,
		},
		{
			name:    "Error from data source",
			mock:    &mockDataSource{err: errors.New("cache refresh failed")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := NewControllerSource(tt.mock)

			got, err := source.GetClientDeleteReasons(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClientDeleteReasons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("GetClientDeleteReasons() returned %d reasons, want %d", len(got), tt.wantLen)
			}
		})
	}
}
