// Package wnc provides WNC client creation and management functionality.
package wnc

import (
	wnc "github.com/umatare5/cisco-ios-xe-wireless-go"
	"github.com/umatare5/cisco-wnc-exporter/internal/config"
)

// createWNCClient creates a configured WNC client for REST API access.
//
// The SDK parses the controller address and refuses anything other than a host or
// host:port, so a value this package's own validation accepts can still fail here.
func createWNCClient(cfg config.WNC) (*wnc.Client, error) {
	options := []wnc.Option{
		wnc.WithTimeout(cfg.Timeout),
		wnc.WithInsecureSkipVerify(cfg.TLSSkipVerify),
	}

	return wnc.NewClient(cfg.Controller, cfg.AccessToken, options...)
}
