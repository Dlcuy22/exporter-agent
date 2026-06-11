// Package collector provides the data collection implementations.
package collector

import (
	"context"
	"regexp"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/shirou/gopsutil/v3/disk"
)

// StorageCollector gathers filesystem and disk I/O metrics.
type StorageCollector struct {
	excludeRegex *regexp.Regexp
}

func (c *StorageCollector) Name() string { return "storage" }

func (c *StorageCollector) Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error {
	// Compile regex if not already done
	if c.excludeRegex == nil && cfg.MountPointsExcludeRegex != "" {
		re, err := regexp.Compile(cfg.MountPointsExcludeRegex)
		if err == nil {
			c.excludeRegex = re
		}
	}

	partitions, err := disk.PartitionsWithContext(ctx, true) // true = all partitions (including virtual)
	if err != nil {
		return err
	}

	ioCounters, _ := disk.IOCountersWithContext(ctx)

	var result []model.Partition
	for _, p := range partitions {
		// Filter by regex
		if c.excludeRegex != nil && c.excludeRegex.MatchString(p.Mountpoint) {
			continue
		}

		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}

		part := model.Partition{
			Device:       p.Device,
			Mountpoint:   p.Mountpoint,
			FSType:       p.Fstype,
			TotalBytes:   int64(usage.Total),
			UsedBytes:    int64(usage.Used),
			FreeBytes:    int64(usage.Free),
			UsagePercent: usage.UsedPercent,
			InodesTotal:  int64(usage.InodesTotal),
			InodesUsed:   int64(usage.InodesUsed),
			InodesFree:   int64(usage.InodesFree),
		}

		// Attach I/O counters if device matches (naive matching)
		if io, ok := ioCounters[p.Device]; ok {
			part.ReadBytesTotal = int64(io.ReadBytes)
			part.WriteBytesTotal = int64(io.WriteBytes)
			part.ReadOpsTotal = int64(io.ReadCount)
			part.WriteOpsTotal = int64(io.WriteCount)
		} else {
			// gopsutil IO counters keys are usually basenames (e.g., sda1)
			// p.Device is usually /dev/sda1
			// Let's try to extract the basename
			// Since we want minimal logic, we'll try stripping "/dev/"
			devName := p.Device
			if len(devName) > 5 && devName[:5] == "/dev/" {
				devName = devName[5:]
			}
			if io, ok := ioCounters[devName]; ok {
				part.ReadBytesTotal = int64(io.ReadBytes)
				part.WriteBytesTotal = int64(io.WriteBytes)
				part.ReadOpsTotal = int64(io.ReadCount)
				part.WriteOpsTotal = int64(io.WriteCount)
			}
		}

		result = append(result, part)
	}

	env.Storage.Partitions = result
	env.Storage.CollectedAt = time.Now().Unix()
	return nil
}
