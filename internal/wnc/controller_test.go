package wnc

import (
	"context"
	"errors"
	"testing"
)

func TestControllerSource_GetBootTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *mockDataSource
		want    string
		wantErr bool
	}{
		{
			name: "Success with a boot instant",
			mock: &mockDataSource{
				data: &WNCDataCache{ControllerBootTime: "2026-01-13T00:00:00+00:00"},
			},
			want: "2026-01-13T00:00:00+00:00",
		},
		{
			// A controller that does not carry the leaf answers with no body, which the
			// refresh records as a successful read of nothing rather than as a failure.
			name:    "Empty when the controller carries no leaf",
			mock:    &mockDataSource{data: &WNCDataCache{}},
			want:    "",
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
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetBootTime() = %q, want %q", got, tt.want)
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
