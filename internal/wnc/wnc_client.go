// Package wnc provides WNC client creation and management functionality.
package wnc

import (
	"fmt"

	wnc "github.com/umatare5/cisco-ios-xe-wireless-go"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// createWNCClient creates a configured WNC client for REST API access.
func createWNCClient(cfg config.WNC) *wnc.Client {
	options := []wnc.Option{
		wnc.WithTimeout(cfg.Timeout),
		wnc.WithInsecureSkipVerify(cfg.TLSSkipVerify),
	}

	// Create WNC client
	wncClient, err := wnc.NewClient(cfg.Controller, cfg.AccessToken, options...)
	if err != nil {
		panic(fmt.Sprintf("Failed to create WNC client: %v", err))
	}

	return wncClient
}
