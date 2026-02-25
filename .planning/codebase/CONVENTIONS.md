# Coding Conventions

**Analysis Date:** 2026-02-25

## Language & Runtime

**Language:** Go
**Convention Style:** Follows Go standard idioms and practices as defined in [Effective Go](https://golang.org/doc/effective_go)

## Naming Patterns

**Files:**
- Lowercase with underscores: `main.go`, `bus.go`, `ioctl.go`
- Test files: `_test.go` suffix (e.g., `main_test.go`, `whd_test.go`)
- Package-specific files by function (e.g., `wifi.go`, `device.go`, `ioctl.go`)

**Packages:**
- Lowercase, single word where possible: `cyw43439`, `whd`, `common`
- Directory name matches package name (e.g., `cyw43439/whd/` contains `package whd`)

**Functions:**
- PascalCase for exported functions (public): `DefaultWifiConfig()`, `Init()`, `GPIOSet()`, `ParseAsyncEvent()`
- camelCase for unexported functions (private): `clmLoad()`, `initBus()`, `cmd_read()`, `csEnable()`
- Receiver methods use shorthand receiver names: `d *Device`, `s Status`, `e *eventMask`, `pm powerManagementMode`

**Variables:**
- camelCase: `buf`, `offset`, `remaining`, `chunk`, `flag`, `header`
- Private package variables: `logger`, `_sendIoctlBuf`, `_iovarBuf`, `_rxBuf` (underscore prefix for unexported module-level buffers)
- Constants: All caps with underscores for multi-word: `WordLengthPos`, `HiSpeedModePos`, `TEST_PATTERN`, `RWTestPattern`
- Constants may also use camelCase in some contexts: `connTimeout`, `maxconns`, `tcpbufsize`, `hostname`, `listenPort`

**Types:**
- PascalCase: `Device`, `Config`, `Status`, `Interrupts`, `Function`, `eventMask`, `opMode`, `linkState`, `tempValues`, `metrics`
- Type aliases with clear purpose: `outputPin` (func type), `ioctlType`, `powerManagementMode`
- Single-letter receiver types used sparingly with clear meaning in context

**Interfaces:**
- Named with -er suffix or descriptive name: `cmdBus` (implicit interface)

**Errors:**
- Exported package-level errors start with `Err`: `ErrDataNotAvailable`
- Unexported errors start with `err`: `errJoinAuth`, `errJoinSetSSID`, `errTxPacketTooLarge`, `errLinkDown`
- Error messages are lowercase, no punctuation unless wrapping: `errors.New("spi test failed:" + hex32(got))`

## Code Style

**Formatting:**
- Standard Go formatting (enforced implicitly by `gofmt`)
- 4-space indentation (Go standard)
- Opening braces on same line: `func() {`, `if x {`, `for {`
- Line length: Natural, no strict limit enforced (some lines exceed 100 chars)

**Linting:**
- No `.golangci.yml` detected in project root
- Convention: Code follows idiomatic Go patterns
- Imports organized by standard library, then third-party, then local packages

**Naming with Underscores:**
- Private unexported names can use underscores in some cases: `_sendIoctlBuf`, `_iovarBuf`, `_rxBuf`, `_traceenabled`
- Private method names with underscores: `core_disable()`, `bt_mode_enabled()`, `bt_init()`, `get_iovar()`, `set_iovar()`, `get_iovar_n()`, `set_iovar_n()`, `set_ioctl()`, `doIoctlSet()`, `update_credit()`, `read32_swapped()`, `write32_swapped()`
- This convention is used for lower-level device communication functions

## Import Organization

**Order:**
1. Standard library imports (`"encoding/binary"`, `"errors"`, `"runtime"`, `"sync"`, `"time"`, `"log/slog"`, etc.)
2. Third-party imports (`"github.com/..."`)
3. Local package imports (`"github.com/soypat/cyw43439/whd"`)

**Path Aliases:**
- Not commonly used
- Relative imports not used (Go best practice of using full import paths)

**Blank Imports:**
- Used for side effects: `_ "embed"` in `exporter/picotempexport.go`

## Error Handling

**Pattern: Explicit Error Check:**
```go
if err != nil {
    return err  // Propagate directly
}
```

**Pattern: Log and Return:**
```go
if err != nil {
    logger.Error("failed to change LED state: ", slog.String("err", err.Error()))
}
```

**Pattern: Error Wrapping:**
```go
err := fmt.Errorf("%w: invalid status code: %s", errInvalidResponse, response.Status)
```

**Pattern: Panic for Fatal Errors:**
```go
if err != nil {
    panic("setup DHCP:" + err.Error())
}
```

**Multiple Errors:**
- Function `errjoin()` used to combine multiple errors (defined in `def.go`)
- Custom error types like `joinError` with Error() method implementation

## Logging

**Framework:** `log/slog` (standard Go structured logging, available in Go 1.21+)

**Patterns:**
- Logger initialized in `init()` functions
- Output to `machine.Serial` or standard output
- Log level configuration: `slog.HandlerOptions{Level: slog.LevelInfo}`
- Structured logging with key-value pairs: `logger.Error("message", slog.String("key", value))`
- Info level: `logger.Info()` with context fields
- Error level: `logger.Error()` for error conditions
- Direct string concatenation for error messages sometimes used: `"failed: " + err.Error()`

**Usage Examples:**
- `logger.Error("failed to change LED state: ", slog.String("err", err.Error()))`
- `logger.Info("listening", slog.String("addr", "http://"+listenAddr.String()))`
- `d.debug("initControl", slog.Int("clm_len", len(clm)))` (Device has debug method wrapper)

## Comments

**When to Comment:**
- Explain the "why" not the "what"
- Complex algorithms and non-obvious logic
- External references to specifications or other implementations

**Doc Comments (Package-Level):**
- Package comments are minimal
- Type and function doc comments in standard format

**Inline Comments:**
- Double slash with space: `// comment`
- Block comments for complex sections
- Comments above code explaining intent: `// Set Antenna to chip antenna.`

**Examples from Codebase:**
```go
// opMode determines the enabled modes of operation as a bitfield.
// To select multiple modes use OR operation:
//
//	mode := ModeWifi | ModeBluetooth
type opMode uint32

// Status supports status notification to the host after a read/write
// transaction over gSPI...
type Status uint32
```

**Reference Comments:**
- Include links to external sources: `// reference: https://github.com/embassy-rs/embassy/...`
- Attribution comments: `// This file borrows heavily from control.rs from the reference:`

## Function Design

**Size:** Functions generally 20-80 lines, with larger functions in hardware communication layers

**Parameters:**
- Receiver types: short variables (`d *Device`, `s Status`)
- Mix of simple and struct types
- Variadic not common, use slices instead
- No function options pattern observed

**Return Values:**
- Single return: `(value)` or `(err)`
- Multiple return: `(value, err)` - error always last
- Named return values used rarely: `func (d *Device) Init(cfg Config) (err error)`

**Function Groups:**
- Related functions grouped together by feature (WiFi, Bluetooth, bus, ioctl)
- Methods on types grouped by receiver type

## Module Design

**Exports:**
- Explicit capitalization for public API
- Internal utility functions unexported (lowercase)
- Device struct with many unexported fields (`mu sync.Mutex`, `spi spibus`, `pwr outputPin`)

**Package Organization:**
- `cyw43439/` - Main package with device, WiFi, Bluetooth functionality
- `cyw43439/whd/` - WiFi Hardware Driver abstractions
- `cyw43439/internal/netlink/` - Internal netlink utilities
- `cyw43439/cmd/` - Command-line tools
- `cyw43439/examples/` - Example usage
- `cyw43439/_legacy_cyrw/` - Legacy code (not actively maintained)
- `exporter/` - Prometheus metrics exporter
- `picoserver/` - HTTP server example
- `prometheus/` - Prometheus configuration

**Barrel Files:**
- Not used extensively
- Each package has discrete responsibility

## Memory & Performance Patterns

**Buffer Alignment:**
- Buffers declared as `[N]uint32` to ensure alignment: `_sendIoctlBuf [2048 / 4]uint32`
- Comment explaining why: `// uint32 buffers to ensure alignment of buffers.`
- Helpers for uint32/uint8 conversion: `u32AsU8()`

**Type Constraints:**
- Uses `golang.org/x/exp/constraints` for generic functions
- Example: `func alignup[T constraints.Unsigned](val, align T) T`

## Testing Patterns

**See TESTING.md for full testing conventions**

---

*Convention analysis: 2026-02-25*
