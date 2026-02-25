# External Integrations

**Analysis Date:** 2026-02-25

## APIs & External Services

**Prometheus Metrics:**
- Prometheus Server (pull-based)
  - SDK/Client: `github.com/prometheus/client_golang v1.20.4`
  - Method: HTTP GET to `/metrics` endpoint on port 3030
  - Metrics exposed: `pico_temperature` (celsius/fahrenheit), `pico_up` (server status)
  - Update interval: 2-second cache in exporter service

**Pico Temperature Server (HTTP API):**
- Service: Custom HTTP server running on Raspberry Pi Pico W
  - Endpoint: `http://192.168.57.213:80` (configurable via `PICO_SERVER_URL`)
  - Protocol: JSON over HTTP GET
  - Response format: `{"tempC": float64, "tempF": float64}`
  - Implementation: `picoserver/main.go`

## Data Storage

**Databases:**
- None detected

**File Storage:**
- None detected

**Caching:**
- In-memory cache in exporter service
  - Cache duration: 2 seconds (configurable in `exporter/picotempexport.go` line 74)
  - Stores: Last fetched temperature readings and server status

## Authentication & Identity

**Auth Provider:**
- None implemented
- No authentication between exporter and Pico server
- No authentication for Prometheus scraping

**Network:**
- Internal to home network (192.168.57.0/24 range)
- No TLS/HTTPS configured
- Direct HTTP communication

## Monitoring & Observability

**Error Tracking:**
- None (no external error tracking service)

**Logs:**
- Pico server: `slog` structured logging to serial port (`machine.Serial`)
  - Log level: INFO
  - Format: Text-based with structured fields
  - Example: `{"time":"...","level":"INFO","msg":"listening","addr":"http://..."}`
- Exporter: `log` package (stdlib) with println
  - Logs to stdout (captured by container)
  - Used for error reporting and request logging

**Metrics:**
- Prometheus format (`text/plain`)
- Exposed via `github.com/prometheus/client_golang/prometheus/promhttp` handler
- Endpoint: `http://exporter:3030/metrics`

## CI/CD & Deployment

**Hosting:**
- Kubernetes cluster (K3s)
  - Namespace: `gohome`
  - Service type: ClusterIP (internal only)
  - Replicas: 1 (Deployment in `exporter/mainfests/deployment.yaml`)

**Image Registry:**
- Docker registry: `registry.home-k3s.lab`
- Image: `registry.home-k3s.lab/gohome/picotempexport:v1`
- Image pull secret: `home-k3s-registry` (credentials required)

**Build:**
- Docker multi-stage build
  - Builder stage: Go 1.23.2 on golang:1.23.2 image
  - Runtime stage: Alpine Linux
  - Output binary: `picotempexport` (Linux, amd64)

**CI Pipeline:**
- None detected (no GitHub Actions, GitLab CI, Jenkins, etc.)

## Environment Configuration

**Required env vars:**
- `PICO_SERVER_URL` - Exporter service must know where Pico server is running
  - Default: None (required, will be empty string if not set)
  - Example: `http://192.168.57.213`
  - Set in Docker Compose: line 10 in `exporter/docker-compose.yml`
  - Set in K8s: line 26 in `exporter/mainfests/deployment.yaml`

**Optional env vars:**
- None detected

**Secrets location:**
- Image pull credentials: Kubernetes Secret `home-k3s-registry`
  - Used in: `exporter/mainfests/deployment.yaml` line 28
  - Type: docker-registry secret

**Timezone:**
- No timezone configuration detected
- System uses UTC or container default

## Webhooks & Callbacks

**Incoming:**
- None - Exporter pulls from Pico server (pull model, not push)

**Outgoing:**
- HTTP GET to Pico server (polling-based)
  - Frequency: Determined by Prometheus scrape interval (external)
  - Endpoint: `{PICO_SERVER_URL}/` (path is root)
  - Timeout: 10 seconds per request

## Network Communication

**Pico Server to Exporter:**
- HTTP/1.0 compatible
- TCP connections accepted on port 80
- Max concurrent connections: 3 (`picoserver/main.go` line 20)
- Connection timeout: 3 seconds
- TCP buffer sizes: 2030 bytes (RX and TX)
- Hostname resolution via DHCP: "picotemp"

**Exporter to Prometheus:**
- HTTP/1.1 via Prometheus scrape config
- Service discovery: `exporter/mainfests/servicemonitor.yaml`
- Prometheus scrape target: `picotempexport-v1:3030`

## Service Mesh

**Kubernetes Services:**
- Service name: `picotempexport-service`
  - Namespace: `gohome`
  - Port: 3030
  - Target port: `temp-export` (named port)
  - Type: ClusterIP (internal only)

**Prometheus Integration:**
- Service Monitor: `exporter/mainfests/servicemonitor.yaml`
- Job name: `picotempexport`
- Targets: `picotempexport-v1:3030`

---

*Integration audit: 2026-02-25*
