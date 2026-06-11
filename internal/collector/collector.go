// Package collector defines interfaces and common logic for metric collection.
//
// Key Components:
//   - Collector: Interface implemented by all subsystem collectors
//   - Registry: A manager for registered collectors
//
// Dependencies:
//   - context
//   - github.com/dlcuy22/exporter-agent/internal/config
//
// Error Types:
//   - None
package collector

import (
	"context"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
)

// Collector is the interface that all system metric collectors must implement.
type Collector interface {
	// Name returns the identifier of the collector.
	Name() string

	// Collect gathers telemetry data and populates the appropriate fields in the Envelope.
	Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error
}
