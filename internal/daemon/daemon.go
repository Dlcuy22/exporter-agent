// Package daemon orchestrates the lifecycle, collectors, and push/pull handlers.
//
// Key Components:
//   - Daemon: The main application struct holding state and configuration
//   - Start(): Initializes the server, background workers, and OS signal listening
//
// Dependencies:
//   - github.com/google/uuid
//   - encoding/json
//   - os
//
// Error Types:
//   - None
package daemon

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/collector"
	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/host"
)

const AGENT_VERSION = "0.1.2"

/*
State represents the on-disk state.
*/
type State struct {
	AgentID string `json:"agent_id"`
}

/*
Daemon is the central orchestrator.
*/
type Daemon struct {
	cfg        *config.Config
	agentID    string
	collectors []collector.Collector

	// Cache for pull requests
	mu          sync.RWMutex
	lastMetrics *model.Envelope
	lastCollect time.Time
	sequence    uint64
}

/*
New creates a new Daemon instance, loading or generating the Agent ID.

	params:
		cfg: The runtime configuration
		collectors: List of system metric collectors to run
	returns:
		*Daemon: initialized daemon
		error: if state initialization fails
*/
func New(cfg *config.Config, collectors []collector.Collector) (*Daemon, error) {
	d := &Daemon{
		cfg:        cfg,
		collectors: collectors,
	}

	id, err := d.loadOrInitState()
	if err != nil {
		return nil, err
	}
	d.agentID = id

	return d, nil
}

func (d *Daemon) loadOrInitState() (string, error) {
	path := d.cfg.AgentIDPath

	// Try reading
	data, err := os.ReadFile(path)
	if err == nil {
		var state State
		if err := json.Unmarshal(data, &state); err == nil && state.AgentID != "" {
			return state.AgentID, nil
		}
	}

	// Create dir if not exists. Fallback to current directory if permission denied.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Warning: failed to create state dir %s: %v. Falling back to ./state.json", dir, err)
		path = "./state.json"
		d.cfg.AgentIDPath = path
	}

	// Generate new UUID
	id := uuid.New().String()
	state := State{AgentID: id}
	out, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		log.Printf("Warning: failed to save state to %s: %v", path, err)
		// Return ID anyway so it can run ephemerally
		return id, nil
	}

	log.Printf("Initialized new Agent ID: %s at %s", id, path)
	return id, nil
}

/*
CollectAll runs all registered collectors sequentially and constructs the payload Envelope.

	params:
		ctx: Context for timeout/cancellation
	returns:
		*model.Envelope: Populated telemetry metrics
		error: if a critical failure occurs
*/
func (d *Daemon) CollectAll(ctx context.Context) (*model.Envelope, error) {
	d.mu.Lock()
	d.sequence++
	currentSeq := d.sequence
	d.mu.Unlock()

	now := time.Now().Unix()

	hostInfo, err := host.InfoWithContext(ctx)
	hostname := "unknown"
	osName := "linux"
	arch := ""
	kernel := ""

	if err == nil {
		hostname = hostInfo.Hostname
		osName = hostInfo.OS
		arch = hostInfo.KernelArch
		kernel = hostInfo.KernelVersion
	} else {
		if h, err := os.Hostname(); err == nil {
			hostname = h
		}
	}

	// Default meta info
	env := &model.Envelope{
		Meta: model.Meta{
			AgentID:      d.agentID,
			Hostname:     hostname,
			CollectedAt:  now,
			OS:           osName,
			Arch:         arch,
			Kernel:       kernel,
			AgentVersion: AGENT_VERSION,
		},
	}

	// Set collected_at for sub-structs
	env.CPU.CollectedAt = now
	env.RAM.CollectedAt = now
	env.Storage.CollectedAt = now
	env.Network.CollectedAt = now

	for _, c := range d.collectors {
		if err := c.Collect(ctx, d.cfg, env); err != nil {
			log.Printf("Error collecting from %s: %v", c.Name(), err)
		}
	}

	d.mu.Lock()
	d.lastMetrics = env
	d.lastCollect = time.Now()
	d.mu.Unlock()

	log.Printf("Metrics collected. Sequence: %d, Time: %d", currentSeq, now)
	return env, nil
}

/*
GetLatestMetrics returns the most recently collected metrics, or forces a collection if empty or stale.
*/
func (d *Daemon) GetLatestMetrics(ctx context.Context) (*model.Envelope, error) {
	d.mu.RLock()
	metrics := d.lastMetrics
	lastCollect := d.lastCollect
	d.mu.RUnlock()

	// Refresh cache if older than 1 second to avoid duplicate collections in same burst
	if metrics == nil || time.Since(lastCollect) > time.Second {
		return d.CollectAll(ctx)
	}
	return metrics, nil
}

/*
AgentID returns the unique agent ID.
*/
func (d *Daemon) AgentID() string {
	return d.agentID
}

/*
Config returns the current daemon configuration.
*/
func (d *Daemon) Config() *config.Config {
	return d.cfg
}
