// Package collector provides the data collection implementations.
package collector

import (
	"context"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessCollector gathers information about running processes.
type ProcessCollector struct{}

func (c *ProcessCollector) Name() string { return "process" }

func (c *ProcessCollector) Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error {
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return err
	}

	var processes []model.Process

	// To keep things lightweight, we might limit the number of processes,
	// but the requirement is to list them. We'll grab as much info as we can efficiently.
	for _, pid := range pids {
		p, err := process.NewProcessWithContext(ctx, pid)
		if err != nil {
			continue // Process might have exited
		}

		name, _ := p.NameWithContext(ctx)
		cmdline, _ := p.CmdlineWithContext(ctx)
		status, _ := p.StatusWithContext(ctx)
		user, _ := p.UsernameWithContext(ctx)

		cpuPercent, _ := p.CPUPercentWithContext(ctx)
		memPercent, _ := p.MemoryPercentWithContext(ctx)
		memInfo, _ := p.MemoryInfoWithContext(ctx)
		threads, _ := p.NumThreadsWithContext(ctx)
		nice, _ := p.NiceWithContext(ctx)
		fds, _ := p.NumFDsWithContext(ctx)
		io, _ := p.IOCountersWithContext(ctx)
		createTime, _ := p.CreateTimeWithContext(ctx)
		ppid, _ := p.PpidWithContext(ctx)

		var rss, vms int64
		if memInfo != nil {
			rss = int64(memInfo.RSS)
			vms = int64(memInfo.VMS)
		}

		var ioRead, ioWrite int64
		if io != nil {
			ioRead = int64(io.ReadBytes)
			ioWrite = int64(io.WriteBytes)
		}

		// Convert status slice to single string character if needed
		// gopsutil returns []string, we take the first element's first char ideally
		statusStr := ""
		if len(status) > 0 {
			statusStr = status[0]
		}

		processes = append(processes, model.Process{
			PID:            int(pid),
			PPID:           int(ppid),
			Name:           name,
			Cmdline:        cmdline,
			Status:         statusStr,
			User:           user,
			CPUPercent:     cpuPercent,
			MemoryRSSBytes: rss,
			MemoryVMSBytes: vms,
			MemoryPercent:  float64(memPercent),
			Threads:        int(threads),
			Nice:           int(nice),
			OpenFDs:        int(fds),
			IOReadBytes:    ioRead,
			IOWriteBytes:   ioWrite,
			CreatedAt:      createTime / 1000, // gopsutil returns ms, our schema is s
		})
	}

	env.Processes = processes
	return nil
}
