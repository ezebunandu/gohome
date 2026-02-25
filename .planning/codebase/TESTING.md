# Testing Patterns

**Analysis Date:** 2026-02-25

## Test Framework

**Runner:**
- Go standard `testing` package (no external test framework)
- Run tests: `go test ./...`
- Watch mode: Use third-party tool (not built-in) or IDE
- Coverage: `go test -cover ./...`

**Assertion Library:**
- Go standard library only (no assertion libraries like testify)
- Manual comparisons and conditional testing

## Test File Organization

**Location:**
- Co-located with source code in same package
- Naming: `*_test.go` suffix

**Naming:**
- Test functions: `TestXxx(t *testing.T)` format (PascalCase after `Test`)
- Examples: `TestInterpretBytes`, `TestParseAsyncEvent`

**Current Test Files:**
- `cyw43439/cmd/cywanalyze/main_test.go` - Tests for command-line analyzer
- `cyw43439/whd/whd_test.go` - Tests for WiFi Hardware Driver package

**Test Coverage:**
- Only 2 test files identified in project
- Limited test coverage across large codebase (11,138 total lines of Go code)
- Most coverage in `cyw43439` package (core device logic)

## Test Structure

**Suite Organization:**
```go
func TestInterpretBytes(t *testing.T) {
	bus := BusCtl{
		Order:           binary.LittleEndian,
		WordInterpreter: binary.BigEndian,
	}
	data := []byte{0x01, 0x02, 0x03, 0x04}
	bus.interpretBytes(data)
	if !bytes.Equal(data, []byte{0x04, 0x03, 0x02, 0x01}) {
		t.Error("expected big endian", data)
	}
	// ... additional test cases in same function
}
```

**Patterns:**
- **Single test function with multiple cases:** Setup shared data, run multiple assertions in same test
- **No table-driven tests:** Not used in project (see `main_test.go` structure)
- **Direct comparison:** Use `if` statements with manual checks
- **Setup:** Inline struct initialization within test function
- **Teardown:** Not explicitly used (Go `t.Cleanup()` not observed)
- **Test isolation:** Each test can be run independently

## Assertion Patterns

**Error Assertion:**
```go
if err != nil {
    t.Fatal(err)
}
```
- `t.Fatal()`: Stops test immediately on error condition
- `t.Error()`: Continues running test, marks as failed

**Value Assertion:**
```go
if ev.Flags != 515 {
    t.Error("bad flags")
}
if !bytes.Equal(data, []byte{0x04, 0x03, 0x02, 0x01}) {
    t.Error("expected big endian", data)
}
```

**Pattern Characteristics:**
- Simple conditional checks with descriptive error messages
- Messages are lowercase and brief: `"bad flags"`, `"bad event type"`, `"bad status"`, `"bad reason"`
- Compound messages: `t.Error("expected big endian", data)` - message followed by actual value

## Mocking

**Framework:** Not detected - no mocking libraries imported

**Approach:**
- Unit tests work directly with concrete types
- No dependency injection for testability observed
- Hardware communication mocked implicitly through test data structures

**Interface-Based Testing:**
- `cmdBus` interface used for SPI abstraction (allows mock implementation)
- Passed to `New(pwr, cs outputPin, spi cmdBus)` constructor in `bus.go`

**What to Mock:**
- External services would use the interface pattern
- Hardware peripherals (SPI, UART) would implement cmdBus interface

**What NOT to Mock:**
- Internal state and data structures - test with real structs
- Binary encoding/decoding - test with actual byte values
- Device status operations - test with concrete Status type

## Test Data & Fixtures

**Test Data Patterns:**
```go
var buf [48]byte
for i := range buf {
    buf[i] = byte(i)
}
ev, err := ParseAsyncEvent(binary.LittleEndian, buf[:])
```

**Location:**
- Test data created inline in test functions
- No separate fixtures or factory files
- Constants used: `binary.LittleEndian`, `binary.BigEndian`

**Data Builders:**
- Constructor functions: `BusCtl{Order: binary.LittleEndian, WordInterpreter: binary.BigEndian}`
- Struct literals with explicit field initialization

## Coverage

**Requirements:** No coverage requirement detected (no CI/CD enforcing minimum)

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Current State:**
- Limited tests (2 test functions) in large codebase
- `cyw43439` package has more test coverage than others
- `exporter/`, `picoserver/`, `prometheus/` have no tests

## Test Types

**Unit Tests:**
- **Scope:** Individual functions and small components
- **Approach:** Direct function calls with test data
- **Examples:**
  - `TestInterpretBytes` - Tests byte interpretation in binary encoding
  - `TestParseAsyncEvent` - Tests async event parsing from binary buffer
- **Environment:** Standard Go testing with embedded test data

**Integration Tests:**
- Not explicitly identified
- Full Device initialization would require real SPI hardware
- No mock SPI interface in test files observed

**End-to-End Tests:**
- Not implemented
- Would require actual hardware (CYW43439 WiFi chip, Raspberry Pi Pico)

**Acceptance Tests:**
- Examples provided in `cyw43439/examples/` directory serve as semi-acceptance tests
- Real usage patterns shown in `picoserver/` and `exporter/` applications

## Async/Concurrency Testing

**Not used in current test suite**

**Patterns in Code:**
```go
// In regular code, goroutines used:
go d.irqPoll()

// Mutex protection observed:
type Device struct {
    mu sync.Mutex
    // ...
}
```

**Example Async Operations:**
- Device initialization involves timing-critical operations (time.Sleep)
- No tests for concurrent access patterns observed
- Real concurrency tested in deployed applications (picoserver)

## Error Testing

**Pattern:**
```go
ev, err := ParseAsyncEvent(binary.LittleEndian, buf[:])
if err != nil {
    t.Fatal(err)
}
```

**Error Test Cases:**
- Expected failures: Tests verify success path, not error cases
- No negative tests observed (tests that verify errors are properly returned)

**What's Not Tested:**
- Error conditions for WiFi operations
- Timeout scenarios in device communication
- Invalid configuration parameters to Device.Init()
- Network connectivity failures in exporter

## Test Execution

**Run all tests:**
```bash
go test ./...
```

**Run specific package tests:**
```bash
go test ./cyw43439/cmd/cywanalyze
go test ./cyw43439/whd
```

**Run with verbose output:**
```bash
go test -v ./...
```

**Run single test:**
```bash
go test -run TestInterpretBytes ./cyw43439/cmd/cywanalyze
```

## Common Testing Patterns

**1. Binary Data Testing:**
```go
data := []byte{0x01, 0x02, 0x03, 0x04}
bus.interpretBytes(data)
if !bytes.Equal(data, []byte{0x04, 0x03, 0x02, 0x01}) {
    t.Error("expected big endian", data)
}
```

**2. Struct Field Verification:**
```go
if ev.Flags != 515 {
    t.Error("bad flags")
}
if ev.EventType != 67438087 {
    t.Error("bad event type")
}
```

**3. Multi-Case Testing (Same Function):**
- Multiple independent test cases with different setup values in single test function
- Used in `TestInterpretBytes` with 4 different bus configurations

**4. Test Initialization:**
```go
func TestXxx(t *testing.T) {
    // Inline setup
    bus := BusCtl{...}
    // Execute
    bus.interpretBytes(data)
    // Assert
    if !bytes.Equal(...) {
        t.Error("message")
    }
}
```

## Test Coverage Gaps

**Untested Packages:**
- `exporter/picotempexport.go` - Prometheus exporter (HTTP client, metrics)
- `picoserver/main.go` - HTTP server application
- `cyw43439/ioctl.go` - IOCTL communication (576 lines)
- `cyw43439/wifi.go` - WiFi operations (430 lines)
- `cyw43439/bluetooth.go` - Bluetooth operations (584 lines)
- `cyw43439/device.go` - Device initialization (370 lines)
- `cyw43439/_legacy_cyrw/` - All legacy code untested

**Critical Gaps:**
- Device.Init() initialization sequence
- WiFi connection join/authentication flows
- Error conditions in IOCTL operations
- Concurrent access to Device
- Event mask operations (enabled in tests but incomplete)

**Risk:** Changes to core device communication code lack test validation, increasing regression risk

## Testing Philosophy

**Current Approach:**
- Minimal testing, focused on low-level binary protocol handling
- Testing added incrementally as needed for critical functionality
- Hardware limitations prevent comprehensive testing (requires actual WiFi chip and Pico board)

**Recommended Additions:**
- Mock interfaces for SPI and other hardware abstractions (already exists with `cmdBus`)
- Table-driven tests for protocol parsing functions
- Negative tests for error conditions
- Integration tests using mock hardware interface

---

*Testing analysis: 2026-02-25*
