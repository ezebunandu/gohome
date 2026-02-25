# Technology Stack

**Analysis Date:** 2026-02-25

## Languages

**Primary:**
- Go 1.23.2 - Exporter service (`exporter/`) and PicoServer application (`picoserver/`)
- Go 1.20 - CYW43439 WiFi driver library (`cyw43439/`)

**Secondary:**
- YAML - Kubernetes manifests and Docker Compose configuration
- HTML - Embedded root page in exporter service (`exporter/rootPage.html`)

## Runtime

**Environment:**
- Go runtime (compiled to native binaries)
- TinyGo support for Raspberry Pi Pico W microcontroller

**Package Manager:**
- Go modules (go.mod/go.sum)
- Lockfiles present for all three modules

## Frameworks

**Core:**
- Standard library `net/http` - HTTP server implementation
- `log/slog` - Structured logging (picoserver)
- `encoding/json` - JSON serialization/deserialization

**Microcontroller/Embedded:**
- github.com/soypat/cyw43439 v0.0.0-20240321235513-d28d7f302509 - CYW43439 WiFi driver for Pico W
- github.com/soypat/seqs v0.0.0-20240527012110-1201bab640ef - Protocol stacks and HTTP handling
- github.com/tinygo-org/pio v0.0.0-20231216154340-cd888eb58899 - PIO support for Pico

**Testing:**
- Standard Go testing (no external framework detected)

**Build/Dev:**
- Docker (multi-stage builds)
- Go build toolchain (CGO_ENABLED=0 for static binaries)

## Key Dependencies

**Critical:**
- github.com/prometheus/client_golang v1.20.4 - Prometheus metrics client and HTTP handler
- github.com/soypat/cyw43439 v0.0.0-20240321235513-d28d7f302509 - WiFi driver for Raspberry Pi Pico W (local replace in picoserver/go.mod)
- github.com/soypat/seqs v0.0.0-20240527012110-1201bab640ef - Network protocol stacks (TCP, HTTP)
- github.com/soypat/natiu-mqtt v0.5.1 - MQTT protocol implementation (cyw43439 only, for examples)

**Infrastructure/Metrics:**
- github.com/prometheus/client_model v0.6.1 - Prometheus metrics protocol buffers
- github.com/prometheus/common v0.55.0 - Prometheus common utilities
- github.com/prometheus/procfs v0.15.1 - Linux /proc filesystem utilities
- google.golang.org/protobuf v1.34.2 - Protocol buffer runtime

**Compression:**
- github.com/klauspost/compress v1.17.9 - Data compression library

**Utilities:**
- golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 - Experimental Go features
- golang.org/x/sys v0.22.0 - System calls for Linux

## Configuration

**Environment:**
- `PICO_SERVER_URL` - Required at runtime for exporter service to know where Pico server is running
- Set via environment variables in Docker Compose and Kubernetes Deployment manifests
- Example value: `http://192.168.57.213` (hardcoded in deployment configs)

**Build:**
- Dockerfile: `exporter/Dockerfile` - Multi-stage build with Go 1.23.2 builder and Alpine runtime
- CGO disabled for cross-platform compilation
- Strip binaries with `-ldflags="-s -w"` for smaller images

## Platform Requirements

**Development:**
- Go 1.20 or 1.23.2
- Standard Unix toolchain (make, etc.)

**Runtime - Exporter Service:**
- Linux container (Alpine-based Docker image)
- Exposes port 3030
- Requires network access to Pico server (TCP on port 80)

**Runtime - PicoServer:**
- Raspberry Pi Pico W microcontroller
- WiFi connectivity via CYW43439 chip
- Listens on HTTP port 80
- Hostname configured as "picotemp"

**Runtime - CYW43439 Driver:**
- Targets Raspberry Pi Pico (ARM Cortex-M0+)
- Runs on bare metal or with minimal OS

**Production:**
- Kubernetes cluster (K3s) running in namespace `gohome`
- Docker registry at `registry.home-k3s.lab`
- Image pull secret: `home-k3s-registry`
- Prometheus server scrapes metrics on port 3030 at `/metrics` endpoint

---

*Stack analysis: 2026-02-25*
