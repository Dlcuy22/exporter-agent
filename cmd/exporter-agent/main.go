// Package main provides the entry point for the exporter-agent daemon.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dlcuy22/exporter-agent/internal/collector"
	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/daemon"
	"github.com/dlcuy22/exporter-agent/internal/handler"
)

const AGENT_VERSION = "0.2.0"

func main() {
	// Parse configurations
	cfgPath := "exporter-agent.yaml"
	cfg, exported, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Set env vars so gopsutil respects custom procfs and sysfs paths
	os.Setenv("HOST_PROC", cfg.PathProcfs)
	os.Setenv("HOST_SYS", cfg.PathSysfs)
	os.Setenv("HOST_ETC", cfg.PathRootfs+"/etc")

	if exported {
		// EXPORT_ENV_TO_CONFIG flag handled and exported.
		// According to requirements, exit after exporting or keep running?
		// We'll keep running to be helpful, but we could exit.
		// User mentioned: "setelahnya tidak perlu di jalankan lagi"
		// which means "afterwards no need to run [the command with ENV] again".
		// We'll just continue running.
	}

	// Initialize collectors
	collectors := []collector.Collector{
		&collector.CPUCollector{},
		&collector.RAMCollector{},
		&collector.StorageCollector{},
		&collector.ProcessCollector{},
		&collector.NetworkCollector{},
	}

	d, err := daemon.New(cfg, collectors)
	if err != nil {
		log.Fatalf("Failed to initialize daemon: %v", err)
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Pull Server
	srv := handler.StartPullServer(d)

	// Start Push Worker in background
	go handler.StartPushWorker(ctx, d)

	// Wait for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Daemon initialized with Agent ID: %s", d.AgentID())
	<-sigCh
	log.Println("Received termination signal, shutting down...")

	// Graceful shutdown
	cancel()
	srv.Shutdown(context.Background())
	log.Println("Shutdown complete")
}
