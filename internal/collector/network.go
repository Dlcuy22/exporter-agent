// Package collector provides the data collection implementations.
package collector

import (
	"context"
	"time"

	"github.com/dlcuy22/exporter-agent/internal/config"
	"github.com/dlcuy22/exporter-agent/internal/model"
	"github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector gathers network interface and TCP connection metrics.
type NetworkCollector struct{}

func (c *NetworkCollector) Name() string { return "network" }

func (c *NetworkCollector) Collect(ctx context.Context, cfg *config.Config, env *model.Envelope) error {
	// Interface IO counters
	ioCounters, err := net.IOCountersWithContext(ctx, true) // true = per interface
	if err != nil {
		return err
	}

	// Interface addresses and properties
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return err
	}

	var result []model.Interface
	for _, iface := range interfaces {
		ips := []string{}
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}

		isUp := false
		for _, flag := range iface.Flags {
			if flag == "up" {
				isUp = true
				break
			}
		}

		i := model.Interface{
			Name:      iface.Name,
			IPs:       ips,
			MAC:       iface.HardwareAddr,
			MTU:       iface.MTU,
			SpeedMbps: 0, // gopsutil doesn't directly provide speed in basic net package
			IsUp:      isUp,
		}

		// Attach IO counters
		for _, io := range ioCounters {
			if io.Name == iface.Name {
				i.RxBytesTotal = int64(io.BytesRecv)
				i.TxBytesTotal = int64(io.BytesSent)
				i.RxPacketsTotal = int64(io.PacketsRecv)
				i.TxPacketsTotal = int64(io.PacketsSent)
				i.RxErrorsTotal = int64(io.Errin)
				i.TxErrorsTotal = int64(io.Errout)
				i.RxDroppedTotal = int64(io.Dropin)
				i.TxDroppedTotal = int64(io.Dropout)
				break
			}
		}
		result = append(result, i)
	}

	// Connections (TCP states)
	conns, err := net.ConnectionsWithContext(ctx, "tcp")
	if err == nil {
		tcpStates := map[string]int{
			"ESTABLISHED": 0,
			"TIME_WAIT":   0,
			"CLOSE_WAIT":  0,
			"LISTEN":      0,
			"SYN_SENT":    0,
			"SYN_RECV":    0,
			"FIN_WAIT1":   0,
			"FIN_WAIT2":   0,
			"CLOSED":      0,
		}

		for _, conn := range conns {
			tcpStates[conn.Status]++
		}

		env.Network.TCPStates = tcpStates
		env.Network.TotalConnections = len(conns)
	}

	// Note: DNS resolvers require parsing /etc/resolv.conf, skipped for minimal approach unless strictly required

	env.Network.Interfaces = result
	env.Network.CollectedAt = time.Now().Unix()
	return nil
}
