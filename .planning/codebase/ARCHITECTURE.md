# Architecture

**Analysis Date:** 2026-02-25

## Pattern Overview

**Overall:** Distributed sensor-to-metrics pipeline with three distinct layers: embedded device, scraper/exporter, and metrics aggregation.

**Key Characteristics:**
- Microcontroller runs a minimal HTTP server exposing JSON temperature readings
- Dedicated Prometheus exporter scrapes the device on polling intervals
- Codebase includes a reusable WiFi/Bluetooth chipset driver (`cyw43439`) as a foundational library
- Loose coupling via HTTP JSON payload exchange between Pico W and exporter
- Hardware abstractions for device-level operations (SPI, GPIO, TCP stack)

## Layers

**Embedded Device Layer (Pico W):**
- Purpose: Collect temperature readings and serve via HTTP
- Location: `picoserver/`
- Contains: HTTP server, temperature sensor polling, LED control, TCP connection management
- Depends on: `cyw43439` (WiFi driver), `github.com/soypat/seqs` (TCP/HTTP stack)
- Used by: Prometheus exporter via HTTP requests

**WiFi Driver Layer (CYW43439):**
- Purpose: Interface with Broadcom CYW43439 WiFi/Bluetooth chipset at protocol and hardware level
- Location: `cyw43439/`
- Contains: Device initialization, firmware loading, WiFi state machine, SDPCM protocol handling, Bluetooth stack, SPI/PIO bus abstraction
- Depends on: WHD protocol definitions, machine-level SPI/GPIO access
- Used by: Picoserver (for network connectivity)

**WHD Protocol Layer:**
- Purpose: Define Cypress WiFi Host Driver protocol structures and utilities
- Location: `cyw43439/whd/`
- Contains: SDPCM headers, CDC headers, async events, country code tables, protocol parsing
- Depends on: Standard library, `github.com/soypat/seqs` for Ethernet types
- Used by: cyw43439 device implementation

**Prometheus Exporter Layer:**
- Purpose: Scrape temperature from Pico W and expose metrics for Prometheus
- Location: `exporter/`
- Contains: HTTP client for polling Pico W, Prometheus metric registration, caching layer, mux routing
- Depends on: `github.com/prometheus/client_golang`
- Used by: Prometheus server (via `/metrics` endpoint)

**Kubernetes Manifests Layer:**
- Purpose: Define deployment configuration for exporter
- Location: `exporter/mainfests/`
- Contains: Deployment specs, service definitions, ServiceMonitor for Prometheus discovery
- Depends on: None (declarative configuration)
- Used by: Kubernetes cluster operator

## Data Flow

**Temperature Sensor Read → HTTP Response:**

1. Pico W calls `machine.ReadTemperature()` on HTTP request
2. `getTemperature()` converts raw sensor value (millidegrees) to Celsius and Fahrenheit
3. `temp` struct is JSON marshaled
4. `HTTPHandler()` writes HTTP response with JSON body
5. TCP connection closes

**Exporter Metrics Collection:**

1. Prometheus periodically makes GET request to exporter `/metrics` endpoint (via Kubernetes ServiceMonitor or direct scrape config)
2. `newMux()` handler invokes registered gauge functions
3. `getMetrics()` checks cache expiry (2-second window)
4. If expired, `getTempValues()` makes HTTP GET to Pico W (configurable via `PICO_SERVER_URL` env var)
5. JSON response decoded into `tempValues` struct
6. Metrics registered via `promauto.NewGaugeFunc()` return current cached values
7. If Pico W unreachable, `pico_up` metric set to 0, temperatures to 0
8. Metrics rendered in Prometheus text format

**Device Initialization Chain:**

1. `main()` calls `setUpDevice()`
2. `SetupWithDHCP()` from `cyw43439/examples/common` initializes CYW43439 via SPI bus
3. Device acquires MAC address, loads WiFi firmware, performs DHCP negotiation
4. `PortStack` and `Device` returned; LED turned on
5. `newListener()` creates TCP listener on port 80
6. `blinkLED()` goroutine started to handle LED blink requests via channel
7. `handleConnection()` goroutine started to accept and process HTTP requests

**State Management:**

- **Device State**: CYW43439 maintains link state machine (`linkState` enum in `device.go`): Down → WaitForSSID → Up/Failed/AuthFailed → WaitForReconnect
- **Exporter Cache**: `metrics` struct holds locked RWMutex-protected cached temperature and status with expiry time
- **Connection Pool**: Pico W listener maintains `maxconns=3` concurrent connections with 2030-byte RX/TX buffers
- **Firmware State**: Device maintains SDPCM sequence numbers, backplane window pointers, ioctls in flight

## Key Abstractions

**cyw43439.Device:**
- Purpose: Single unified interface to CYW43439 hardware
- Examples: `cyw43439/device.go`, `cyw43439/bus.go`
- Pattern: Mutex-protected state machine with method receivers; internal protocol handling abstracted from caller
- Public API: `Join()`, `Scan()`, `Leave()`, `SendEth()`, `RecvEthHandle()`, `GPIOSet()`, `ReadTemperature()` indirectly via `machine` package

**WHD Protocol Structures:**
- Purpose: Type-safe representation of Broadcom protocol frames
- Examples: `SDPCMHeader`, `CDCHeader`, `BDCHeader`, `AsyncEvent` in `cyw43439/whd/protocol.go`
- Pattern: Structs with binary.ByteOrder-aware encoding/decoding; methods for type introspection

**metrics (Exporter):**
- Purpose: Thread-safe cached metrics with lazy evaluation
- Examples: `exporter/picotempexport.go:metrics` type
- Pattern: RWMutex guards state; gauge functions invoked by Prometheus registry on scrape
- Behavior: Caches temperature readings with 2-second TTL to avoid hammering Pico W

**SPI Bus Abstraction:**
- Purpose: Hardware-agnostic interface for command read/write
- Examples: `spibus` struct in `cyw43439/bus.go`
- Pattern: Interface `cmdBus` allows multiple implementations (native, PIO)
- Implementations: `bus_native.go` (direct SPI), `bus_pico_pio.go` (bit-banged via PIO)

## Entry Points

**Pico W Server:**
- Location: `picoserver/main.go:main()`
- Triggers: Microcontroller boot
- Responsibilities: Device initialization, network setup, connection acceptance loop, temperature reading on demand

**Prometheus Exporter:**
- Location: `exporter/picotempexport.go:main()`
- Triggers: Container startup (via Docker/Kubernetes)
- Responsibilities: HTTP server startup on port 3030, Prometheus metric registration, polling loop via gauge functions

**WHD Command Interface:**
- Location: `cyw43439/` package exports (no single entry point; library consumed by picoserver examples)
- Triggers: Application calls to `Device` methods
- Responsibilities: Protocol translation, hardware register access, firmware command sequencing

## Error Handling

**Strategy:** Explicit error returns from functions; panics only on initialization failures (picoserver) or irrecoverable hardware issues.

**Patterns:**

- **Initialization Panics**: Unrecoverable conditions panic immediately:
  ```go
  // picoserver/main.go
  if err != nil {
      panic("setup DHCP:" + err.Error())
  }
  ```

- **Connection-Level Logging**: Individual request errors logged but don't crash server:
  ```go
  // picoserver/main.go:handleConnection()
  if err != nil {
      logger.Error("listener accept:", slog.String("err", err.Error()))
      time.Sleep(time.Second)
      continue
  }
  ```

- **Graceful Degradation (Exporter)**: Pico W unreachable → metrics report 0 with `pico_up=0`:
  ```go
  // exporter/picotempexport.go
  if err := m.results.getTempValues(client, url); err != nil {
      m.up = 0
      m.results.TempC = 0
      m.results.TempF = 0
  }
  ```

- **Hardware Protocol Errors**: WHD layer returns detailed error types for protocol violations:
  ```go
  // cyw43439/whd/protocol.go
  var errSDPCMHeaderSizeComplementMismatch = errors.New("sdpcm hdr size complement mismatch")
  ```

- **Timeout Management**: Connection deadline set per-request to avoid indefinite hangs:
  ```go
  // picoserver/main.go
  err = conn.SetDeadline(time.Now().Add(connTimeout))
  ```

## Cross-Cutting Concerns

**Logging:**
- **Pico W**: `log/slog` with `TextHandler` to `machine.Serial` (UART); `Info` level default
- **Exporter**: Standard `log` package; no structured logging
- **CYW43439**: Optional `slog.Logger` passed in `Config`; mostly silent unless debug enabled

**Validation:**

- **HTTP Protocol**: Manual request draining (`conn.Read(discard[:])`) to clear RX buffer before writing response; prevents Tx buffer overload
- **Metrics Format**: JSON unmarshaling validates temperature JSON structure; invalid data logged as error
- **Temperature Bounds**: No explicit bounds checking; raw sensor values passed through

**Authentication:**

- **HTTP**: None; runs on local network or behind firewall
- **WiFi**: SSID/passphrase provided via `SetupWithDHCP()` config (from common examples)
- **Prometheus**: No auth; assumes scraper runs in same cluster/network

**Resource Constraints:**

- **Memory**: Fixed TCP buffers (`tcpbufsize=2030`), fixed internal device buffers (2048 bytes for SDPCM, iovar, ioctl, RX)
- **Concurrency**: Pico W limited to `maxconns=3` connections; goroutine-per-request model in exporter (no pooling)
- **SPI Access**: Mutex protects device state; no interrupt-driven concurrency on embedded side

**Network Resilience:**

- **Pico W Connection Timeout**: 3-second deadline per request; connection closed after response
- **Exporter Retry**: Client timeout 10 seconds; no retry logic; next scrape in ~15 seconds (Prometheus default)
- **Cache Layer**: 2-second TTL in exporter reduces polling frequency to Pico W by 3-5x

---

*Architecture analysis: 2026-02-25*
