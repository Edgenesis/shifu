# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shifu is a Kubernetes-native IoT gateway (CNCF landscape project) that creates "digital twins" of physical IoT devices as Kubernetes pods. Each device gets a DeviceShifu pod that translates between the device's native protocol and a unified HTTP API.

## Build & Development Commands

```bash
# Build all targets
make build

# Run all tests (requires envtest setup — auto-downloaded)
make test

# Run a single package's tests
go test -v ./pkg/deviceshifu/deviceshifuhttp/...

# Format and vet
make fmt
make vet

# Install shifuctl CLI
make install

# Generate CRD manifests (from pkg/k8s/crd/)
cd pkg/k8s/crd && make manifests generate

# Build Docker images locally (single platform)
make buildx-build-image-deviceshifu-http-http
make buildx-build-image-shifu-controller
```

## Architecture

### Digital Twin Pattern

The core concept: for each physical IoT device, Shifu deploys a DeviceShifu pod in Kubernetes. This pod speaks the device's native protocol (MQTT, OPC UA, etc.) on one side and exposes a standard HTTP API on the other. Applications interact with devices through HTTP requests to the DeviceShifu pod, never directly.

### Key Components

**Kubernetes Controller** (`pkg/k8s/`):
- CRD types defined in `pkg/k8s/api/v1alpha1/` — `EdgeDevice` and `TelemetryService`
- Controllers in `pkg/k8s/controllers/` watch EdgeDevice CRs and manage DeviceShifu pod lifecycle
- Controller manager entry point: `pkg/k8s/crd/main.go`

**DeviceShifu Framework** (`pkg/deviceshifu/`):
- `deviceshifubase/` — Base implementation with `DeviceShifu` interface (`Start()`, `Stop()`), config loading from ConfigMaps, telemetry collection, and HTTP server
- Protocol implementations are in sibling packages: `deviceshifuhttp/`, `deviceshifumqtt/`, `deviceshifuopcua/`, `deviceshifusocket/`, `deviceshifutcp/`, `deviceshifulwm2m/`
- Each protocol impl has a `New()` factory function and a corresponding binary in `cmd/deviceshifu/cmd<protocol>/`

**Telemetry Service** (`pkg/telemetryservice/`):
- Collects device data and pushes to storage backends (MQTT, SQL, MinIO, TDengine)

**Gateway** (`pkg/gateway/`):
- Protocol gateways (e.g., LwM2M) that sit between devices and DeviceShifu pods

### Adding a New Protocol

1. Create `pkg/deviceshifu/deviceshifu<protocol>/` implementing the `DeviceShifu` interface
2. Create `cmd/deviceshifu/cmd<protocol>/main.go` as the binary entry point
3. Add a Dockerfile in `dockerfiles/`
4. `deviceshifutemplate/` serves as the reference implementation

### Configuration

Each DeviceShifu is configured via Kubernetes ConfigMaps with sections for:
- Device connection settings (address, protocol)
- Instruction mappings (device commands → HTTP endpoints)
- Telemetry collection settings

Example configurations live in `examples/deviceshifu/`.

## Testing

- Tests use `ginkgo/gomega` and `testify` frameworks
- Controller tests require envtest (kubebuilder test environment) — `make test` handles setup automatically
- Mock devices in `pkg/deviceshifu/mockdevice/` and `examples/deviceshifu/mockdevice/`
- Test helper utilities in `pkg/deviceshifu/unitest/`
- Telemetry service tests run separately without race detector (TDengine driver constraint)

## Linting

Uses golangci-lint v2 (`.golangci.yml`). Currently `errcheck` and `staticcheck` are disabled (tracked in #1361).

## Docker Images

- Registry: `edgehub/` on Docker Hub
- Version from `version.txt` (default: "nightly")
- Runtime base: `distroless/static-debian11`
- Multi-platform: linux/amd64, linux/arm64, linux/arm

## Repository Conventions

- Apache 2.0 license
- CRD group: `shifu.edgenesis.io`, version: `v1alpha1`
- After modifying CRD types in `pkg/k8s/api/v1alpha1/`, run `make manifests generate` from `pkg/k8s/crd/`
