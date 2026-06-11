// Package collector provides the data collection implementations.
package collector

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
)

// CPUCollector gathers CPU metrics.
type CPUCollector struct {
	mu         sync.Mutex
	maxPercent float64
	avg1m      float64
	avg5m      float64
	avg15m     float64
	lastUpdate time.Time
}

func (c *CPUCollector) Name() string { return "cpu" }

func (c *CPUCollector) Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Hardware info
	info, err := cpu.InfoWithContext(ctx)
	if err == nil && len(info) > 0 {
		env.CPU.Model = info[0].ModelName
	}

	// Cores count
	phys, _ := cpu.CountsWithContext(ctx, false)
	logical, _ := cpu.CountsWithContext(ctx, true)
	env.CPU.PhysicalCores = phys
	env.CPU.LogicalCores = logical

	// Current percent (overall)
	// Passing interval 0 means it returns immediately (since last call)
	percents, err := cpu.PercentWithContext(ctx, 0, false)
	var current float64
	if err == nil && len(percents) > 0 {
		current = percents[0]
		env.CPU.CurrentPercent = current
	}

	// Track Max
	if current > c.maxPercent {
		c.maxPercent = current
	}
	env.CPU.MaxPercent = c.maxPercent

	// Calculate EWMA for 1m, 5m, 15m if we have a previous update
	if !c.lastUpdate.IsZero() && current > 0 {
		dt := now.Sub(c.lastUpdate).Seconds()
		if dt > 0 {
			alpha1m := 1.0 - math.Exp(-dt/60.0)
			alpha5m := 1.0 - math.Exp(-dt/300.0)
			alpha15m := 1.0 - math.Exp(-dt/900.0)

			c.avg1m = alpha1m*current + (1.0-alpha1m)*c.avg1m
			c.avg5m = alpha5m*current + (1.0-alpha5m)*c.avg5m
			c.avg15m = alpha15m*current + (1.0-alpha15m)*c.avg15m
		}
	} else if c.lastUpdate.IsZero() {
		// First initialization
		c.avg1m = current
		c.avg5m = current
		c.avg15m = current
	}
	c.lastUpdate = now

	env.CPU.Avg = model.CPUAvg{
		OneMin:     c.avg1m,
		FiveMin:    c.avg5m,
		FifteenMin: c.avg15m,
	}

	// Current percent (per core)
	corePercents, err := cpu.PercentWithContext(ctx, 0, true)
	if err == nil {
		env.CPU.PerCore = make([]model.CPUPerCore, len(corePercents))
		for i, p := range corePercents {
			env.CPU.PerCore[i] = model.CPUPerCore{
				Core:    i,
				Percent: p,
			}
		}
	}

	// Load Average
	avg, err := load.AvgWithContext(ctx)
	if err == nil {
		env.CPU.LoadAvg = model.CPULoadAvg{
			OneMin:     avg.Load1,
			FiveMin:    avg.Load5,
			FifteenMin: avg.Load15,
		}
	}

	env.CPU.CollectedAt = now.Unix()
	return nil
}
