# Codebase Concerns

**Analysis Date:** 2026-02-25

## Tech Debt

**Extensive TODO/FIXME markers indicating incomplete specifications:**
- Issue: 40+ TODO comments across the codebase, many without clear resolution paths
- Files: `cyw43439/bus.go` (lines 42, 122, 280, 325), `cyw43439/ioctl.go` (lines 122, 161, 270, 311, 419), `cyw43439/wifi.go` (line 136), `cyw43439/device.go` (lines 160, 301), `cyw43439/bluetooth.go` (line 477), `cyw43439/netif.go` (line 85)
- Impact: Developers cannot determine if incomplete code paths are intentional design decisions or unfinished work. Creates friction for contributors and reviewers.
- Fix approach: Audit each TODO; either implement, document rationale, or remove. Consider using issue tracker references in TODOs.

**Memory allocation workarounds in low-level bus code:**
- Issue: Stack-allocated fixed-size buffers with TODO comments indicating desire for heap allocation
- Files: `cyw43439/bus.go` lines 280 (`var buf [maxTxSize/4 + 1]uint32`) and 325 (commented out heapalloc attempt)
- Impact: Potential stack overflow on embedded systems with deep call stacks; fragile design that hasn't been revisited.
- Fix approach: Profile stack usage under load. Consider pooled buffer reuse pattern instead of dynamic heap allocation.

**Legacy code directory not removed:**
- Issue: `cyw43439/_legacy_cyrw/` directory contains 12 files (844 LOC in cy_ioctl.go alone) with duplicated functionality
- Files: All files in `cyw43439/_legacy_cyrw/` directory
- Impact: Code duplication increases maintenance burden and creates confusion about which implementation to use. Not all legacy code is referenced by active code paths.
- Fix approach: Audit which legacy functions are actually used by current codebase; remove unused legacy implementations and consolidate remaining code.

**Hardcoded timing delays scattered throughout initialization:**
- Issue: 30+ `time.Sleep()` calls with fixed durations (100ms, 150ms, 250ms, 500ms) without documented justification
- Files: `cyw43439/wifi.go` (lines 100, 101, 104, 106, 123, 128, 133), `cyw43439/device.go` (250ms after bus init), `cyw43439/bluetooth.go` (line 222)
- Impact: Slow startup path; some sleeps marked as "required" but commented as "not critical"; no adaptive timing mechanism if hardware responds faster.
- Fix approach: Document each sleep with hardware reference. Implement adaptive timing using readiness checks instead of fixed durations for non-critical sleeps.

**Deprecated API methods still present:**
- Issue: `MACAs6()` and `TryPoll()` methods marked deprecated but still functional
- Files: `cyw43439/deprecated.go`
- Impact: API surface confusion; unclear migration path for users.
- Fix approach: Set removal deadline; provide migration examples; communicate deprecation schedule.

## Known Bugs

**Potential buffer overrun in HTTP request handler:**
- Symptoms: HTTP handler reads incoming request into fixed-size buffer and discards it; no size checking
- Files: `picoserver/main.go` line 115: `conn.Read(discard[:])`
- Trigger: Sending HTTP request larger than `tcpbufsize` (2030 bytes)
- Workaround: Clients must send requests under 2030 bytes; no validation of actual bytes read
- Impact: Truncated request data could leave incomplete HTTP headers in buffer, causing next connection to receive stale data
- Fix approach: Check return value of Read(); handle partial reads in loop until EOF or error

**Ioctl polling timeout not fully investigated:**
- Symptoms: Comments indicate adding certain ioctl calls causes timeout errors
- Files: `cyw43439/_legacy_cyrw/cy_ioctl.go` lines 242, 341-342
- Trigger: Not documented; original attempt to fix caused regression
- Workaround: Disabled code path remains commented out
- Impact: Unknown feature (possibly timing-related) remains broken without root cause analysis
- Fix approach: Add instrumentation and trace logs; use logic analyzer to capture SPI timing patterns

**Endianness handling uncertain in multiple locations:**
- Symptoms: Comments asking "is this needed?" or "swap endianness on this read?"
- Files: `cyw43439/bluetooth.go` line 477, `cyw43439/bus.go` endianness handling in protocol
- Trigger: Likely edge cases on different word ordering scenarios
- Impact: May work on current hardware but fragile for port to other hardware or bitwidths
- Fix approach: Add comprehensive endianness tests; document byte order assumptions per register

## Security Considerations

**No input validation on temperature data:**
- Risk: HTTP response body parsed directly with no bounds checking or schema validation
- Files: `exporter/picotempexport.go` line 46: `json.NewDecoder(response.Body).Decode(tv)`
- Current mitigation: Prometheus itself validates metrics format; internal network assumed trusted
- Recommendations: Add explicit bounds checking (e.g., temperature within -50°C to +150°C); handle malformed JSON gracefully; add request size limits

**Panic on misconfiguration instead of graceful degradation:**
- Risk: Multiple panic() calls in initialization code paths
- Files: `picoserver/main.go` (lines 53, 71, 76), `bus_pico_pio.go` (lines 27, 31, 36, 56), `cyw43439/bus.go` line 207
- Current mitigation: Only during startup; hardware already initialized before user code
- Recommendations: Return errors instead of panicking in library code; only panic in examples/main applications; add recovery for alignment violations

**Unsafe pointer arithmetic without explicit validation:**
- Risk: `unsafeAs()` and `unsafeAsSlice()` functions perform alignment checks at runtime but panic on violation
- Files: `cyw43439/bus.go` lines 493-514
- Current mitigation: Alignment checked before cast; limited to internal use
- Recommendations: Consider compile-time assertion or bounds-checked safe wrapper; document alignment requirements clearly

**No authentication or authorization on HTTP endpoints:**
- Risk: Any network client can query temperature data and Prometheus metrics
- Files: `picoserver/main.go` HTTPHandler, `exporter/picotempexport.go` metrics endpoints
- Current mitigation: Assumes internal network; simple consumer device without secrets
- Recommendations: If expanding to internet-facing: add basic auth; rate limiting; HTTPS support

## Performance Bottlenecks

**Polling-based interrupt handling:**
- Problem: Comment indicates hardware interrupts not working; falls back to polling
- Files: `cyw43439/ioctl.go` line 311: `// TODO get real hw interrupts working and ditch polling`
- Cause: Hardware integration complexity or driver limitation
- Impact: CPU utilization higher than necessary; responsiveness tied to polling frequency
- Improvement path: Implement hardware interrupt handler; implement interrupt masking; measure power savings

**Slow startup sequence with sequential waits:**
- Problem: 13 sequential `time.Sleep()` calls during WiFi initialization (700ms+ total)
- Files: `cyw43439/wifi.go` lines 94-133
- Cause: Hardware requires settling time between operations; no parallelization possible
- Impact: On 3-second connection timeout, 700ms+ goes to initialization delays
- Improvement path: Implement state machine that overlaps operations where possible; measure actual settle times on target hardware

**TCP buffer draining on every HTTP request:**
- Problem: All incoming request data read into fixed buffer and discarded
- Files: `picoserver/main.go` line 115
- Cause: Comment says "to keep the TCP RX buffer clear"
- Impact: Unnecessary memory read operations; wasted bandwidth especially with large request bodies
- Improvement path: Only drain if buffer space low; implement proper HTTP request parsing instead of blind drain

**Fixed sleep durations create unnecessary latency:**
- Problem: 30+ hardcoded sleep calls prevent faster-than-expected hardware from responding quickly
- Files: Multiple locations in cyw43439/wifi.go, cyw43439/device.go
- Impact: Minimum 500ms+ device initialization time even if hardware ready faster
- Improvement path: Implement readiness polling with timeout fallback (sleep only until timeout)

## Fragile Areas

**CYW43439 SDPCM protocol implementation:**
- Files: `cyw43439/ioctl.go` (576 LOC), `cyw43439/wifi.go` (430 LOC), `cyw43439/whd/` (1322 LOC combined)
- Why fragile: Complex binary protocol with padding requirements, endianness handling, and SPI bus sequencing; multiple TODO comments in critical paths; sparse documentation of register meanings
- Safe modification: Add comprehensive unit tests with mock SPI responses before changing buffer layout or alignment code
- Test coverage: Basic protocol parsing tested in `whd/whd_test.go` but no integration tests for full control flow

**Concurrent access to Device struct:**
- Files: `cyw43439/device.go` (single `sync.Mutex`)
- Why fragile: Single mutex protecting entire device state (67 fields); no field-level locking; long hold times during ioctl operations
- Safe modification: Audit all lock paths to ensure no deadlocks; consider splitting state into independent sub-structures if lock contention observed
- Test coverage: No concurrency tests; no stress tests with multiple rapid API calls

**Bus-level SPI communication:**
- Files: `cyw43439/bus.go` (520 LOC)
- Why fragile: Low-level SPI handling with strict alignment requirements, buffer layout assumptions, and window-based addressing; multiple hardcoded constants; no CRC validation (by design for CLM, but no warning)
- Safe modification: Write comprehensive protocol state machine tests; ensure all address calculations verified with hardware; document all magic constants
- Test coverage: Only partial - `cmd_read`/`cmd_write` tested indirectly through integration tests

**Bluetooth and WiFi mode switching:**
- Files: `cyw43439/device.go`, `cyw43439/bluetooth.go`, `cyw43439/wifi.go`
- Why fragile: Code paths differ significantly between WiFi-only and BT+WiFi configurations; mode selection happens at init time; complex firmware loading sequences
- Safe modification: Add comprehensive matrix tests for all mode combinations; verify firmware loading on each target variant
- Test coverage: Modes tested in examples but no automated test suite covering all combinations

## Scaling Limits

**Maximum connections limited by fixed pool:**
- Current capacity: `picoserver/main.go` hardcoded `maxconns = 3` connections
- Limit: Stack-allocated connection buffers; TCP port count limited to 1; no connection pooling
- Scaling path: Would require redesign to streaming or async I/O; current microcontroller not suitable for more than handful of clients

**Memory constraints on device buffers:**
- Current capacity: Multiple 2048-byte buffers (mtuPrefix=2048) allocated in Device struct and as stack buffers
- Limit: 11KB total device buffers; fragmentation risk with large packets
- Scaling path: Implement buffer pooling; compress multiple small operations into single SPI transaction; consider larger MCU if needed

**Polling frequency bottleneck:**
- Current capacity: Polling loop must complete full cycle within connection timeout
- Limit: 3-second connection timeout with 700ms+ initialization creates tight budget for actual work
- Scaling path: Implement interrupt-driven model; optimize polling loop; parallelize independent operations

## Dependencies at Risk

**Go standard library constraints for embedded systems:**
- Risk: Code compiled for Raspberry Pi Pico with limited stdlib functions available
- Files: `picoserver/main.go`, `exporter/picotempexport.go` use `encoding/json`, `net`, `log/slog`, `time`
- Impact: No ability to add complex dependencies; must use existing minimal embedded Go environment
- Migration plan: If stdlib functions removed from embedded Go, will need minimal JSON parser replacement; consider pre-serializing responses

**Prometheus client library availability:**
- Risk: `github.com/prometheus/client_golang` must remain compatible with exporter's Go version
- Files: `exporter/picotempexport.go` (imports prometheus packages)
- Impact: Exporter cannot upgrade past versions breaking prometheus client compatibility
- Migration plan: Implement minimal Prometheus text format exporter if library becomes incompatible

**soypat/cyw43439 driver maturity:**
- Risk: This is custom open-source driver with 40+ TODOs; not production Broadcom support
- Files: All files in `cyw43439/` package
- Impact: No official support; complex protocol implementation may have bugs as new hardware variants emerge
- Migration plan: Monitor for upstream Pico SDK WiFi driver improvements; track Broadcom documentation updates

## Missing Critical Features

**No error recovery mechanism:**
- Problem: No automatic reconnection when WiFi drops; no retry logic for transient network failures
- Blocks: Unattended operation in unreliable networks; manual intervention required on disconnection
- Impact: Device becomes unmonitored if WiFi flakes out
- Fix: Implement connection watchdog with automatic reconnection attempts

**No metrics for device health:**
- Problem: Temperature is exported but no metrics for WiFi signal strength, connection uptime, packet loss, or device errors
- Blocks: Cannot diagnose why device stops reporting; cannot optimize placement
- Impact: Operator flying blind on device health
- Fix: Export WiFi RSSI, connection attempts, error counters

**No authentication on exporter HTTP server:**
- Problem: Any network client can access metrics and status pages
- Blocks: Cannot safely run on untrusted networks; metrics can be scraped by unauthorized consumers
- Impact: Information disclosure; potential for metric poisoning
- Fix: Add basic auth or mutual TLS

**No hot-reload of configuration:**
- Problem: WiFi credentials hardcoded at compile time or via environment variables at startup; cannot change without recompile/restart
- Blocks: Device must be physically accessed or reflashed to change network
- Impact: High operational friction; cannot quickly respond to WiFi network changes
- Fix: Implement configuration storage on device; add web UI for credential update

## Test Coverage Gaps

**No integration tests for full WiFi connection flow:**
- What's not tested: Complete `Device.Init()` → `JoinAP()` → data exchange sequence with simulated hardware
- Files: `cyw43439/device.go`, `cyw43439/wifi.go` main entry points
- Risk: Regressions in init sequence go undetected; different firmware versions untested
- Priority: High - this is the critical path

**No tests for concurrent device access:**
- What's not tested: Multiple goroutines calling Device methods simultaneously
- Files: `cyw43439/device.go` and all public API methods
- Risk: Race conditions, deadlocks, or data corruption under load
- Priority: High - only single mutex protecting all state

**No tests for error paths and timeout handling:**
- What's not tested: SPI timeouts, invalid responses, partial data reception
- Files: `cyw43439/bus.go`, `cyw43439/ioctl.go`
- Risk: Error handling code untested and potentially broken
- Priority: Medium - affects reliability but less critical path than happy path

**No tests for HTTP server request handling:**
- What's not tested: Oversized requests, malformed requests, rapid connection sequences
- Files: `picoserver/main.go` HTTPHandler
- Risk: Buffer overrun vulnerabilities, connection handling bugs
- Priority: Medium - affects robustness but current limited deployment

**No tests for metric export correctness:**
- What's not tested: Temperature encoding, JSON formatting, Prometheus label format
- Files: `exporter/picotempexport.go`
- Risk: Silently incorrect metric values exported
- Priority: Low - format simple but potential for subtle bugs

---

*Concerns audit: 2026-02-25*
