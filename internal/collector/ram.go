// Package collector provides the data collection implementations.
package collector

import (
	"context"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/shirou/gopsutil/v3/mem"
)

// RAMCollector gathers memory and swap metrics.
type RAMCollector struct{}

func (c *RAMCollector) Name() string { return "ram" }

func (c *RAMCollector) Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		env.RAM.Total = int64(v.Total)
		env.RAM.Used = int64(v.Used)
		env.RAM.Free = int64(v.Free)
		env.RAM.Shared = int64(v.Shared)
		env.RAM.BuffCache = int64(v.Buffers + v.Cached)
		env.RAM.Available = int64(v.Available)
		env.RAM.UsagePercent = v.UsedPercent
	}

	s, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		env.RAM.Swap = model.SwapInfo{
			Total:        int64(s.Total),
			Used:         int64(s.Used),
			Free:         int64(s.Free),
			UsagePercent: s.UsedPercent,
		}
	}

	env.RAM.CollectedAt = time.Now().Unix()
	return nil
}
