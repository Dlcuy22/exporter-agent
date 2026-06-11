// Package config handles the parsing of YAML, ENV, and Flags to produce a unified configuration.
//
// Key Components:
//   - Load(): Reads YAML, ENV, and Flags in the correct priority (Flag > ENV > YAML)
//   - ExportConfig(): Writes the active configuration to exporter-agent.yaml if EXPORT_ENV_TO_CONFIG is set
//
// Dependencies:
//   - gopkg.in/yaml.v3: YAML parsing and generation
//   - flag: Command-line arguments parsing
//   - os: Environment variables reading
//
// Error Types:
//   - None
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds the complete configuration for the exporter-agent.
type Config struct {
	PullPort                int    `yaml:"pull_port"`
	PushURL                 string `yaml:"push_url"`
	PushInterval            string `yaml:"push_interval"` // e.g., "15s"
	AgentIDPath             string `yaml:"agent_id_path"`
	PathProcfs              string `yaml:"path_procfs"`
	PathSysfs               string `yaml:"path_sysfs"`
	PathRootfs              string `yaml:"path_rootfs"`
	MountPointsExcludeRegex string `yaml:"mount_points_exclude_regex"`
}

// defaultConfig returns the base configuration.
func defaultConfig() Config {
	return Config{
		PullPort:                8080,
		PushURL:                 "",
		PushInterval:            "15s",
		AgentIDPath:             "/var/lib/exporter-agent/state.json",
		PathProcfs:              "/proc",
		PathSysfs:               "/sys",
		PathRootfs:              "/",
		MountPointsExcludeRegex: "^/(sys|proc|dev|host|etc)($|/)",
	}
}

/*
Load reads configuration from YAML, ENV variables, and CLI flags.
Priority: CLI Flag > ENV > YAML > Default

	params:
		yamlPath: path to the configuration file
	returns:
		*Config: fully populated configuration
		bool: whether to exit immediately (true if config was exported)
		error: any error encountered during load
*/
func Load(yamlPath string) (*Config, bool, error) {
	cfg := defaultConfig()

	// 1. Read YAML (if exists)
	if data, err := os.ReadFile(yamlPath); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, false, fmt.Errorf("failed to parse yaml: %w", err)
		}
	}

	// 2. Read ENV variables
	if v := os.Getenv("PULL_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.PullPort = p
		}
	}
	if v := os.Getenv("PUSH_URL"); v != "" {
		cfg.PushURL = v
	}
	if v := os.Getenv("PUSH_INTERVAL"); v != "" {
		cfg.PushInterval = v
	}
	if v := os.Getenv("AGENT_ID_PATH"); v != "" {
		cfg.AgentIDPath = v
	}
	if v := os.Getenv("PATH_PROCFS"); v != "" {
		cfg.PathProcfs = v
	}
	if v := os.Getenv("PATH_SYSFS"); v != "" {
		cfg.PathSysfs = v
	}
	if v := os.Getenv("PATH_ROOTFS"); v != "" {
		cfg.PathRootfs = v
	}
	if v := os.Getenv("COLLECTOR_FILESYSTEM_MOUNT_POINTS_EXCLUDE"); v != "" {
		cfg.MountPointsExcludeRegex = v
	}

	// 3. Read Flags
	// We use flag.NewFlagSet to allow calling Load in tests easily without redefining flags
	fs := flag.NewFlagSet("exporter-agent", flag.ContinueOnError)

	pullPortFlag := fs.Int("pull.port", cfg.PullPort, "HTTP pull port")
	pushURLFlag := fs.String("push.url", cfg.PushURL, "Push endpoint URL")
	pushIntervalFlag := fs.String("push.interval", cfg.PushInterval, "Push interval (e.g., 15s)")
	agentIDPathFlag := fs.String("agent.id.path", cfg.AgentIDPath, "Path to store agent ID state")

	procfsFlag := fs.String("path.procfs", cfg.PathProcfs, "procfs mountpoint")
	sysfsFlag := fs.String("path.sysfs", cfg.PathSysfs, "sysfs mountpoint")
	rootfsFlag := fs.String("path.rootfs", cfg.PathRootfs, "rootfs mountpoint")
	mountExcludeFlag := fs.String("collector.filesystem.mount-points-exclude", cfg.MountPointsExcludeRegex, "Regexp of mount points to exclude for storage collector")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, false, err
	}

	// Override with flag values if they were explicitly provided
	// Since we set the default of flags to the ENV/YAML override above,
	// the flag value itself handles the priority correctly.
	cfg.PullPort = *pullPortFlag
	cfg.PushURL = *pushURLFlag
	cfg.PushInterval = *pushIntervalFlag
	cfg.AgentIDPath = *agentIDPathFlag
	cfg.PathProcfs = *procfsFlag
	cfg.PathSysfs = *sysfsFlag
	cfg.PathRootfs = *rootfsFlag
	cfg.MountPointsExcludeRegex = *mountExcludeFlag

	// Export check
	if os.Getenv("EXPORT_ENV_TO_CONFIG") == "true" || os.Getenv("EXPORT_ENV_TO_CONFIG") == "1" {
		if err := ExportConfig(yamlPath, &cfg); err != nil {
			return nil, false, fmt.Errorf("failed to export config: %w", err)
		}
		fmt.Printf("Configuration exported to %s. Exiting.\n", yamlPath)
		return &cfg, true, nil
	}

	return &cfg, false, nil
}

/*
ExportConfig writes the current configuration to the specified YAML file path.

	params:
		yamlPath: path to write the YAML file
		cfg: the configuration to serialize
	returns:
		error: if writing fails
*/
func ExportConfig(yamlPath string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(yamlPath, data, 0644)
}
