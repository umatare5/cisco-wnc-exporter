// Package main provides the main entry point for the cisco-wnc-exporter.
package main

import (
	"github.com/umatare5/cisco-wnc-exporter/internal/cli"
)

// main is the entry point of the application.
func main() {
	cli.NewApp()
}
