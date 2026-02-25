# Codebase Structure

**Analysis Date:** 2026-02-25

## Directory Layout

```
tempmonitor/
├── picoserver/             # Embedded HTTP server for Pico W
│   ├── main.go             # Entry point, temperature HTTP endpoint
│   ├── go.mod              # Go module for Pico W build
│   ├── go.sum
│   └── main.uf2            # Compiled Pico W firmware image
├── exporter/               # Prometheus exporter service
│   ├── picotempexport.go   # Exporter logic, metrics, HTTP handlers
│   ├── rootPage.html       # Simple HTML root page
│   ├── go.mod              # Go module for exporter
│   ├── go.sum
│   ├── Dockerfile          # Container image definition
│   ├── docker-compose.yml  # Local compose setup
│   └── mainfests/          # Kubernetes manifests
│       ├── deployment.yaml # Exporter deployment spec
│       ├── service.yaml    # Kubernetes service
│       ├── servicemonitor.yaml  # Prometheus ServiceMonitor
│       └── kustomization.yaml   # Kustomize base
├── cyw43439/               # CYW43439 WiFi/Bluetooth driver library
│   ├── device.go           # Main Device type, public API
│   ├── wifi.go             # WiFi join/leave/scan logic
│   ├── bluetooth.go        # Bluetooth initialization
│   ├── bus.go              # SPI bus abstraction
│   ├── bus_native.go       # Native SPI implementation
│   ├── bus_pico_pio.go     # PIO-based SPI bit-bang (TinyGo)
│   ├── netif.go            # Network interface (Ethernet sender/receiver)
│   ├── ioctl.go            # ioctl command handling
│   ├── def.go              # Constants and definitions
│   ├── debug.go            # Debug logging helpers
│   ├── firmware_embed.go   # Embedded firmware binary resources
│   ├── deprecated.go       # Legacy functions (unused)
│   ├── interrupts_string.go  # Generated enum string methods
│   ├── go.mod              # Module definition
│   ├── go.sum
│   ├── whd/                # Cypress WiFi Host Driver protocol
│   │   ├── whd.go          # Country code utilities
│   │   ├── protocol.go     # SDPCM, CDC, BDC headers and parsing
│   │   ├── asyncevent.go   # Event packet definitions
│   │   ├── asyncevent_type_string.go  # Generated enum methods
│   │   ├── sdpcm_command_string.go    # Generated enum methods
│   │   └── whd_test.go     # Protocol parsing tests
│   ├── internal/           # Internal-only code
│   │   └── netlink/        # Netlink protocol implementation
│   │       └── netlink.go
│   ├── cmd/                # Command-line tools
│   │   ├── cywparse/       # WHD packet parser
│   │   │   └── main.go
│   │   └── cywanalyze/     # Signal analysis tool
│   │       ├── main.go
│   │       ├── main_test.go
│   │       └── ref/        # Reference data
│   ├── examples/           # Example applications
│   │   ├── common/         # Shared setup utilities (SetupWithDHCP)
│   │   ├── blinky/
│   │   ├── http-client/
│   │   ├── http-server/
│   │   ├── mqtt/
│   │   ├── tcp*/
│   │   └── ...
│   ├── firmware/           # Firmware binary files (not shown, embedded)
│   ├── _legacy_cyrw/       # Deprecated implementation (unused)
│   └── [generated files]   # Enum string methods, etc.
├── prometheus/             # Prometheus configuration reference (minimal)
├── README.md               # Project overview
└── .planning/              # GSD planning documents
    └── codebase/
        ├── ARCHITECTURE.md
        ├── STRUCTURE.md (this file)
        └── [other analysis docs]
```

## Directory Purposes

**picoserver/**
- Purpose: Embedded HTTP server running on Raspberry Pi Pico W
- Contains: Single-file Go application for microcontroller
- Key files: `main.go` (entire application), `main.uf2` (compiled binary)
- Build target: TinyGo compiled to Pico W ARM binary format

**exporter/**
- Purpose: Prometheus metrics exporter service
- Contains: Go HTTP server, Prometheus client integration, Docker/Kubernetes deployment configs
- Key files: `picotempexport.go` (metrics registration and scraping logic), `Dockerfile` (container build)
- Build target: Standard Go binary, containerized with Docker

**cyw43439/**
- Purpose: Reusable WiFi/Bluetooth chipset driver library
- Contains: Hardware abstraction, protocol implementation, reference examples
- Key files: `device.go` (public API), `wifi.go` (WiFi state machine), `bus.go` (SPI abstraction)
- Status: Published as external module (`github.com/soypat/cyw43439`); local version at `../cyw43439` for picoserver

**cyw43439/whd/**
- Purpose: Broadcom Cypress WiFi Host Driver protocol definitions
- Contains: Protocol structures, parsing utilities, type-safe representations
- Key files: `protocol.go` (all header types), `whd.go` (utility functions)
- Role: Foundation for device communication; consumed by `cyw43439` device layer

**cyw43439/internal/**
- Purpose: Private implementation details not exported
- Contains: Netlink protocol code for internal use
- Access: Imported only within cyw43439 package

**cyw43439/cmd/**
- Purpose: Standalone command-line utilities
- Contains: Packet parser, signal analyzer
- Role: Development/debugging tools; not part of main application binary

**cyw43439/examples/**
- Purpose: Reference implementations showing how to use cyw43439 library
- Contains: HTTP client/server, MQTT, TCP listener/server examples
- Key file: `common/setup.go` (reusable initialization; imported by picoserver)
- Status: Educational; `common` subpackage provides foundational setup

**exporter/mainfests/**
- Purpose: Kubernetes deployment specifications
- Contains: YAML manifests for container orchestration
- Key files: `deployment.yaml` (exporter pod spec), `servicemonitor.yaml` (Prometheus discovery)
- Usage: Applied via `kubectl apply -k` (Kustomize)

## Key File Locations

**Entry Points:**
- `picoserver/main.go`: Pico W application entry; `main()` initializes device and runs server loop
- `exporter/picotempexport.go`: Exporter entry; `main()` starts HTTP server on port 3030
- `cyw43439/examples/common/`: Initialization helpers for device setup (not an entry point; imported by picoserver)

**Configuration:**
- `picoserver/main.go`: Hardcoded constants at top (hostname="picotemp", listenPort=80, maxconns=3, tcpbufsize=2030, connTimeout=3s)
- `exporter/picotempexport.go`: Environment variable `PICO_SERVER_URL` (required at runtime)
- `exporter/mainfests/deployment.yaml`: Container env vars for exporter configuration
- `cyw43439/device.go`: Config struct for firmware selection, mode (WiFi/BT/both), logger

**Core Logic:**
- `picoserver/main.go:HTTPHandler()`: Temperature JSON serialization and HTTP response
- `picoserver/main.go:handleConnection()`: Connection loop and LED signaling
- `exporter/picotempexport.go:getMetrics()`: Cache logic and Pico W polling
- `cyw43439/wifi.go:Join()`: WiFi network association state machine
- `cyw43439/device.go:New()`: Device initialization and SPI setup
- `cyw43439/bus.go:cmd_read()`, `cmd_write()`: SPI command handling

**Testing:**
- `cyw43439/whd/whd_test.go`: Protocol parsing unit tests
- `cyw43439/cmd/cywanalyze/main_test.go`: Signal analysis tool tests
- No tests in picoserver or exporter (embedded and service respectively)

## Naming Conventions

**Files:**
- **Go source**: `*.go` (e.g., `device.go`, `wifi.go`, `main.go`)
- **Generated**: `*_string.go` (e.g., `interrupts_string.go`) - auto-generated enum string methods
- **Manifests**: `*.yaml` (e.g., `deployment.yaml`, `service.yaml`)
- **Binary artifacts**: `*.uf2` (Pico W firmware), `Dockerfile`, `docker-compose.yml`
- **Test files**: `*_test.go` (e.g., `whd_test.go`)

**Directories:**
- **Lowercase with underscores**: Most directories (`picoserver`, `cyw43439`, `whd`, `cmd`)
- **Private prefixed with underscore**: `_legacy_cyrw/` (deprecated), `internal/` (Go convention for unexported)
- **examples/**: Reference implementations grouped by use case

**Functions:**
- **Exported (Capitalized)**: Public API intended for external use (e.g., `Device.Join()`, `Device.Scan()`, `New()`)
- **Unexported (lowercase)**: Internal functions (e.g., `clmLoad()`, `initControl()`, `handleConnection()`)
- **Receivers on types**: Methods use receiver shorthand (e.g., `(d *Device)`, `(m *metrics)`, `(tv *tempValues)`)

**Types:**
- **Capitalized struct names**: `Device`, `Config`, `SDPCMHeader`, `CDCHeader`, `BDCHeader`, `AsyncEvent`
- **Lowercase interface names**: `spibus`, `cmdBus`, `logstate` (internal/unexported)
- **Enum types**: Defined as base type (e.g., `type opMode uint32`, `type linkState uint8`)

**Variables & Constants:**
- **Capitalized**: Exported constants (e.g., `MTU`, `CONTROL_HEADER`)
- **Lowercase**: Unexported constants and variables (e.g., `connTimeout`, `maxconns`, `tcpbufsize`)
- **Snake_case in JSON tags**: `json:"tempC"`, `json:"tempF"` (Go conventions)

## Where to Add New Code

**New Feature (e.g., humidity sensor addition):**
- Primary code: `picoserver/main.go` - Add sensor read function, JSON struct field, HTTP endpoint
- Tests: None currently; would add `picoserver/*_test.go`
- Example: Copy `getTemperature()` pattern for new sensor

**New Exporter Metric:**
- Implementation: `exporter/picotempexport.go` - Register new `promauto.NewGaugeFunc()` in `newMux()`
- Logic: Add new field to `tempValues`, implement getter methods, call `getTempValues()` to populate
- Example: Prometheus gauge setup starts at line ~108

**New Driver Feature (WiFi scan, Bluetooth pairing):**
- Core: `cyw43439/device.go` - Add public method with receiver `(d *Device)`
- Protocol: `cyw43439/wifi.go` or `bluetooth.go` - Add state machine logic or command handling
- Protocol definitions: `cyw43439/whd/protocol.go` - Add new header or event types if needed
- Example: `Device.Scan()` spans device.go and wifi.go

**New Kubernetes Resource:**
- Manifests: `exporter/mainfests/` - Add YAML file (e.g., `ingress.yaml`, `hpa.yaml`)
- Integration: Update `kustomization.yaml` to include new resource in `resources:` list
- Example: Copy format from `service.yaml`

**Utilities or Helpers:**
- Shared driver utilities: `cyw43439/def.go` or new file in cyw43439
- Exporter utilities: New file in `exporter/` root (e.g., `exporter/cache.go`)
- Common examples: `cyw43439/examples/common/` - Reusable setup functions

**Command-line Tools:**
- Location: `cyw43439/cmd/[toolname]/main.go`
- Example: `cyw43439/cmd/cywparse/` for packet parsing

## Special Directories

**cyw43439/firmware/**
- Purpose: Stores firmware binary files for embedded device
- Generated: Yes (by Broadcom/Cypress; sourced externally)
- Committed: Not shown in listing but referenced in `firmware_embed.go`
- Accessed: Embedded via Go's `//go:embed` into binaries

**exporter/mainfests/**
- Purpose: Kubernetes manifests for deployment
- Generated: No (hand-written YAML)
- Committed: Yes (in git)
- Applied by: `kubectl apply -k exporter/mainfests/` or GitOps operator

**cyw43439/_legacy_cyrw/**
- Purpose: Legacy implementation (deprecated, unused)
- Generated: No
- Committed: Yes (for historical reference)
- Status: Do not use; kept for reference only

**cyw43439/internal/**
- Purpose: Go convention for unexported packages
- Generated: No
- Committed: Yes
- Access: Only imported within `cyw43439` package; users cannot import `internal/`

**cyw43439/examples/**
- Purpose: Reference implementations
- Generated: No
- Committed: Yes (in main repo)
- Status: Not built by default; used as examples and in tests

---

*Structure analysis: 2026-02-25*
