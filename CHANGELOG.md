# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2026-06-12
### Added
- Token-based authentication using `X-Token` header for the pull endpoint (`/metrics`).
- Configuration options `pull_auth` and `pull_token` for authentication.
- Automatic UUID token generation on first run if `PULL_AUTH=true` and no token is provided.
- Added GitHub Actions workflow to automatically build and push Docker images to `ghcr.io`.
- Updated `README.md` with explicit details about the Push and Pull modes, and configuration.
- Added `.gitignore` to omit binaries, state files, and logs.

### Fixed
- Addressed `cmd/exporter-agent` directory being accidentally ignored in Git.

## [0.1.0] - Initial Release
- Basic implementation of `CPU`, `RAM`, `Storage`, `Process`, and `Network` collectors.
- Support for `Pull` and `Push` mechanisms.
- Configuration parsing from YAML, ENV variables, and CLI Flags.
