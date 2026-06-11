// Package model provides data structures matching the JSON schema for the agent payload.
//
// Key Components:
//   - Envelope: Main struct encapsulating all system metrics
//
// Dependencies:
//   - None
//
// Error Types:
//   - None
package model

/*
Meta holds metadata about the agent and host system.
*/
type Meta struct {
	AgentID      string `json:"agent_id"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	Kernel       string `json:"kernel"`
	AgentVersion string `json:"agent_version"`
	CollectedAt  int64  `json:"collected_at"`
}

/*
CPU holds CPU usage and hardware information.
*/
type CPU struct {
	Model          string       `json:"model"`
	PhysicalCores  int          `json:"physical_cores"`
	LogicalCores   int          `json:"logical_cores"`
	CurrentPercent float64      `json:"current_percent"`
	MaxPercent     float64      `json:"max_percent"`
	Avg            CPUAvg       `json:"avg"`
	LoadAvg        CPULoadAvg   `json:"load_avg"`
	PerCore        []CPUPerCore `json:"per_core"`
	CollectedAt    int64        `json:"collected_at"`
}

type CPUAvg struct {
	OneMin     float64 `json:"1m"`
	FiveMin    float64 `json:"5m"`
	FifteenMin float64 `json:"15m"`
}

type CPULoadAvg struct {
	OneMin     float64 `json:"1m"`
	FiveMin    float64 `json:"5m"`
	FifteenMin float64 `json:"15m"`
}

type CPUPerCore struct {
	Core    int     `json:"core"`
	Percent float64 `json:"percent"`
}

/*
RAM holds memory and swap usage information.
*/
type RAM struct {
	Total        int64    `json:"total"`
	Used         int64    `json:"used"`
	Free         int64    `json:"free"`
	Shared       int64    `json:"shared"`
	BuffCache    int64    `json:"buff_cache"`
	Available    int64    `json:"available"`
	UsagePercent float64  `json:"usage_percent"`
	Swap         SwapInfo `json:"swap"`
	CollectedAt  int64    `json:"collected_at"`
}

type SwapInfo struct {
	Total        int64   `json:"total"`
	Used         int64   `json:"used"`
	Free         int64   `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

/*
Storage holds information about mounted partitions and disk I/O.
*/
type Storage struct {
	Partitions  []Partition `json:"partitions"`
	CollectedAt int64       `json:"collected_at"`
}

type Partition struct {
	Device          string  `json:"device"`
	Mountpoint      string  `json:"mountpoint"`
	FSType          string  `json:"fstype"`
	TotalBytes      int64   `json:"total_bytes"`
	UsedBytes       int64   `json:"used_bytes"`
	FreeBytes       int64   `json:"free_bytes"`
	UsagePercent    float64 `json:"usage_percent"`
	InodesTotal     int64   `json:"inodes_total"`
	InodesUsed      int64   `json:"inodes_used"`
	InodesFree      int64   `json:"inodes_free"`
	ReadBytesTotal  int64   `json:"read_bytes_total"`
	WriteBytesTotal int64   `json:"write_bytes_total"`
	ReadOpsTotal    int64   `json:"read_ops_total"`
	WriteOpsTotal   int64   `json:"write_ops_total"`
}

/*
Process holds system process list details.
*/
type Process struct {
	PID            int     `json:"pid"`
	PPID           int     `json:"ppid"`
	Name           string  `json:"name"`
	Cmdline        string  `json:"cmdline"`
	Status         string  `json:"status"`
	User           string  `json:"user"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryRSSBytes int64   `json:"memory_rss_bytes"`
	MemoryVMSBytes int64   `json:"memory_vms_bytes"`
	MemoryPercent  float64 `json:"memory_percent"`
	Threads        int     `json:"threads"`
	Nice           int     `json:"nice"`
	OpenFDs        int     `json:"open_fds"`
	IOReadBytes    int64   `json:"io_read_bytes"`
	IOWriteBytes   int64   `json:"io_write_bytes"`
	CreatedAt      int64   `json:"created_at"`
}

/*
Network holds network interface statistics and TCP states.
*/
type Network struct {
	Interfaces       []Interface    `json:"interfaces"`
	TCPStates        map[string]int `json:"tcp_states"`
	TotalConnections int            `json:"total_connections"`
	DNSResolvers     []string       `json:"dns_resolvers"`
	CollectedAt      int64          `json:"collected_at"`
}

type Interface struct {
	Name           string   `json:"name"`
	IPs            []string `json:"ips"`
	MAC            string   `json:"mac"`
	MTU            int      `json:"mtu"`
	SpeedMbps      int      `json:"speed_mbps"`
	IsUp           bool     `json:"is_up"`
	RxBytesTotal   int64    `json:"rx_bytes_total"`
	TxBytesTotal   int64    `json:"tx_bytes_total"`
	RxPacketsTotal int64    `json:"rx_packets_total"`
	TxPacketsTotal int64    `json:"tx_packets_total"`
	RxErrorsTotal  int64    `json:"rx_errors_total"`
	TxErrorsTotal  int64    `json:"tx_errors_total"`
	RxDroppedTotal int64    `json:"rx_dropped_total"`
	TxDroppedTotal int64    `json:"tx_dropped_total"`
}

/*
Envelope is the complete payload wrapping all telemetry data.
*/
type Envelope struct {
	Meta      Meta      `json:"meta"`
	CPU       CPU       `json:"cpu"`
	RAM       RAM       `json:"ram"`
	Storage   Storage   `json:"storage"`
	Processes []Process `json:"processes"`
	Network   Network   `json:"network"`
}
