# Exporter Agent

This project is part of the Esempe.id Minecraft hosting service. It is a node monitoring agent designed to collect system metrics such as CPU, RAM, storage, network, and process information. The agent can serve metrics via a pull endpoint or push them to a centralized server.

## Stacks Used

- Go: The primary programming language used for the agent.
- gopsutil: A library for retrieving system and hardware information.
- yaml.v3: Used for parsing configuration files.

## Getting Started

You can run the agent using Docker or install it directly using Go.

### Using Docker

To run the agent using Docker, execute the following command:

```bash
docker run --rm --network host \
  -v "/proc:/host/proc:ro" \
  -v "/sys:/host/sys:ro" \
  -v "/:/rootfs:ro" \
  -e PATH_PROCFS=/host/proc \
  -e PATH_SYSFS=/host/sys \
  -e PATH_ROOTFS=/rootfs \
  ghcr.io/dlcuy22/exporter-agent:main
```

### Using Go Install

Alternatively, you can install the agent directly using `go install`:

```bash
go install github.com/dlcuy22/exporter-agent@latest
```

## Configuration

The configuration is loaded by reading from YAML, ENV, and CLI Flags in the following priority (Flag > ENV > YAML). Additionally, you can write the active configuration back to the `exporter-agent.yaml` file if the `EXPORT_ENV_TO_CONFIG` environment variable is set.

The agent uses a configuration file named `exporter-agent.yaml`. Here are the configuration options available along with their environment variable equivalents:

- pull_port (ENV: PULL_PORT): The port used to expose the metrics for pulling.
- push_url (ENV: PUSH_URL): The endpoint URL where the agent will push metrics. Leave this empty to disable pushing.
- push_interval (ENV: PUSH_INTERVAL): The time interval between push requests.
- agent_id_path (ENV: AGENT_ID_PATH): The file path where the agent stores its unique identifier.
- path_procfs (ENV: PATH_PROCFS): The path to the proc filesystem.
- path_sysfs (ENV: PATH_SYSFS): The path to the sys filesystem.
- path_rootfs (ENV: PATH_ROOTFS): The path to the root filesystem.
- mount_points_exclude_regex (ENV: COLLECTOR_FILESYSTEM_MOUNT_POINTS_EXCLUDE): A regular expression used to exclude specific mount points from storage monitoring.

### Example Configuration

```yaml
pull_port: 9090
push_url: ""
push_interval: 15s
agent_id_path: /var/lib/exporter-agent/state.json
path_procfs: /proc
path_sysfs: /sys
path_rootfs: /
mount_points_exclude_regex: ^/(sys|proc|dev|host|etc)($|/)
```

Made with ❤️ by Esempe hosting team
