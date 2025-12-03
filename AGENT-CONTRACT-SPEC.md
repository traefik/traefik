# Baqup Agent Contract Specification

**Version:** 1.0  
**Status:** Draft  
**Date:** 2025-12-01

---

## Overview

This specification defines the contract that all baqup agents must implement. Agents are stateless, short-lived containers that perform backup and restore operations on behalf of the baqup orchestrator.

### Design Principles

1. **Stateless execution** - All configuration via environment variables
2. **Self-describing** - Schema and capabilities discoverable without execution
3. **Fail fast** - Validate early, exit with clear error codes
4. **Trust but verify** - Checksums mandatory, manifests are contracts
5. **Controller owns state** - Agents produce artifacts, controller manages lifecycle

---

## 1. Lifecycle

### State Machine

```
INITIALIZING → RUNNING → COMPLETING → TERMINATED
     │            │           │
     └──► FAILED ◄┴───────────┘
```

| State | Description | Timeout |
|-------|-------------|---------|
| `INITIALIZING` | Config validation, bus connection, first heartbeat | 30s |
| `RUNNING` | Active work, heartbeats every N seconds | Job timeout |
| `COMPLETING` | Writing final artifacts, checksums, status | 60s |
| `TERMINATED` | Exit 0, status reported | - |
| `FAILED` | Exit non-zero, error reported if possible | - |

### Signal Handling

| Signal | Behaviour |
|--------|-----------|
| `SIGTERM` | Begin graceful shutdown. If in RUNNING, abort and clean up. If in COMPLETING, finish atomic move then exit. |
| `SIGKILL` | Immediate termination. Controller reconstructs state from last heartbeat. |
| `SIGUSR1` | Emit immediate heartbeat (used by controller for liveness check). |

### Heartbeat

Agents emit heartbeats at configurable intervals (default 10s) to indicate liveness.

```json
{
  "type": "heartbeat",
  "job_id": "abc123",
  "agent_id": "agent-fs-7f8a9b",
  "timestamp": "2024-01-15T03:00:30Z",
  "state": "running",
  "intent": {
    "operation": "checksum_large_file",
    "expected_duration_seconds": 120,
    "detail": "/staging/abc123/data.tar.zst"
  }
}
```

**Intent signalling:** For long operations (large file checksums, compression), agents declare expected duration. Controller extends timeout accordingly.

**Zombie detection:**
- 3 missed heartbeats (30s silence) = presumed dead
- Controller sends SIGUSR1 to request immediate heartbeat
- No response within 5s = declare dead, kill container
- Hard ceiling: 10 minutes without heartbeat regardless of intent

---

## 2. Configuration Interface

### Environment Variable Categories

| Prefix | Purpose | Example |
|--------|---------|---------|
| `BAQUP_JOB_*` | Job context | `BAQUP_JOB_ID`, `BAQUP_JOB_TIMEOUT` |
| `BAQUP_AGENT_*` | Agent behaviour | `BAQUP_AGENT_HEARTBEAT_INTERVAL`, `BAQUP_AGENT_RETRY_MAX` |
| `BAQUP_STAGING_*` | Staging area location | `BAQUP_STAGING_DIR` |
| `BAQUP_ARTIFACT_*` | Backup file format | `BAQUP_ARTIFACT_COMPRESS`, `BAQUP_ARTIFACT_CHECKSUM_ALGORITHM` |
| `BAQUP_BUS_*` | Communication | `BAQUP_BUS_URL`, `BAQUP_BUS_KEY_PREFIX` |
| `BAQUP_SOURCE_*` | Data origin | `BAQUP_SOURCE_PATH`, `BAQUP_SOURCE_HOST` |
| `BAQUP_TARGET_*` | Data destination | `BAQUP_TARGET_PATH`, `BAQUP_TARGET_HOST` |
| `BAQUP_SECRET_*` | Sensitive values | `BAQUP_SECRET_DB_PASSWORD` |

### Secret Handling

Variables prefixed with `BAQUP_SECRET_*` are sensitive and MUST:
- Never be logged
- Never appear in status reports or error messages
- Be cleared from memory after use where possible

Controller resolves secret references before spawning agent. Agent receives plaintext but is unaware of original reference.

### Multi-Value Fields

Arrays are parsed in priority order:
1. Starts with `[` → parse as JSON
2. Contains `|` but no `,` → split on `|`
3. Default → split on `,`

```bash
# Standard (comma)
BAQUP_SOURCE_EXCLUDE=*.log,*.tmp,node_modules

# Values containing commas (pipe)
BAQUP_SOURCE_EXCLUDE=file1|file,with,commas|file2

# Complex (JSON)
BAQUP_SOURCE_EXCLUDE=["file|with|pipes", "file,with,commas"]
```

### Schema-Driven Validation

Each agent ships with `agent-schema.json` defining its configuration contract. Schema declares:
- Required vs optional variables
- Types (string, int, bool, enum, array)
- Defaults
- Constraints (min/max, patterns, allowed values)
- Conditional requirements

**Injection at build time:**
```bash
docker build \
  --label "io.baqup.schema=$(cat agent-schema.json | jq -c .)" \
  -t ghcr.io/baqup/agent-fs:latest .
```

**Validation runs:**
1. Agent startup (INITIALIZING phase)
2. Dry-run mode (`--validate` flag)
3. Controller pre-flight (optional, via OCI registry API)

### Type Coercion

Environment variables are always strings. Agent coerces based on schema-declared type:
- `"true"` / `"false"` → boolean
- `"123"` → integer
- `"3.14"` → float
- Invalid values → fail with clear error

Written spellings (e.g., `"three"`) are not coerced.

### Defaults and Precedence

- Defaults defined in schema
- Agent applies defaults if variable unset
- Controller resolves precedence (labels → config → CLI)
- Agent sees only final env vars—no precedence logic in agent

---

## 3. Communication Protocol

### Message Types

| Type | Direction | Purpose |
|------|-----------|---------|
| Heartbeat | Agent → Bus | Liveness with state and intent |
| Progress | Agent → Bus | Percentage, bytes, files processed |
| Status | Agent → Bus | Job completion with artifacts and metrics |
| Event | Agent → Bus | Notable occurrences (warnings, errors) |

### Message Structures

**Progress:**
```json
{
  "type": "progress",
  "job_id": "abc123",
  "timestamp": "2024-01-15T03:00:45Z",
  "percent_complete": 45,
  "bytes_processed": 2147483648,
  "files_processed": 1250,
  "current_file": "/data/uploads/image.png",
  "estimated_remaining_seconds": 180
}
```

**Status:**
```json
{
  "type": "status",
  "job_id": "abc123",
  "agent": "agent-fs",
  "role": "snapshot",
  "status": "success",
  "started_at": "2024-01-15T03:00:00Z",
  "completed_at": "2024-01-15T03:05:47Z",
  "files": [
    {
      "path": "/staging/abc123/data.tar.zst",
      "size_bytes": 1048576,
      "checksum_sha256": "a7f3b9c..."
    }
  ],
  "metrics": {
    "duration_seconds": 347,
    "bytes_processed": 5368709120,
    "files_processed": 12450,
    "compression_ratio": 0.42
  },
  "warnings_summary": {
    "PERMISSION_DENIED": 23,
    "SYMLINK_SKIPPED": 5
  },
  "error": null
}
```

**Event:**
```json
{
  "type": "event",
  "job_id": "abc123",
  "timestamp": "2024-01-15T03:02:15Z",
  "severity": "warn",
  "code": "PERMISSION_DENIED",
  "message": "Permission denied on 847 files",
  "first_occurrence": "2024-01-15T03:02:15Z",
  "last_occurrence": "2024-01-15T03:02:47Z",
  "count": 847,
  "sample_paths": ["/data/secrets/key1", "/data/secrets/key2"]
}
```

### Redis Topology

| Purpose | Structure | Key Pattern |
|---------|-----------|-------------|
| Heartbeats | Hash with TTL | `baqup:heartbeat:{agent_id}` |
| Progress | Hash (overwritten) | `baqup:progress:{job_id}` |
| Status | Hash (persisted) | `baqup:status:{job_id}` |
| Events | Stream | `baqup:events:{job_id}` |
| Live notifications | Pub/Sub | `baqup:notify` |

### Message Ordering

Timestamps are truth. Each message is self-contained. Controller takes latest by timestamp, discards stale.

Events use Redis Streams for ordering guarantee within single producer.

### Bus Unavailability

If Redis is unreachable:
1. Agent continues work
2. Messages buffered to `BAQUP_BUS_FALLBACK_DIR`
3. Agent retries connection with exponential backoff
4. On reconnect, agent replays buffer
5. If job completes before reconnect, agent stays alive for grace period
6. If grace period expires, agent writes `.baqup-status.json` to staging and exits with code 75

Controller checks for orphaned status files during cleanup.

---

## 4. Output Contract

### Directory Structure

```
/staging/{job_id}/
├── .staging/                    # Work in progress (invisible to controller)
│   ├── data.tar.zst
│   └── manifest.json
├── data.tar.zst                 # Final artifact
└── manifest.json                # Completion signal
```

### Atomic Completion Pattern

1. Write artifacts to `.staging/`
2. Compute checksums
3. Write manifest to `.staging/`
4. Atomic move: `.staging/*` → parent directory

**Filesystem requirement:** Staging directory MUST be a single filesystem mount. Agent validates at startup:

```python
if os.stat(staging_dir).st_dev != os.stat(work_dir).st_dev:
    fail_fast("Staging subdirectories must be on same filesystem")
```

### Manifest Schema

```json
{
  "version": "1.0",
  "job_id": "abc123",
  "agent": "agent-fs",
  "agent_version": "1.2.0",
  "role": "snapshot",
  "created_at": "2024-01-15T03:05:47Z",
  "artifacts": [
    {
      "filename": "data.tar.zst",
      "size_bytes": 1048576,
      "checksum": {
        "algorithm": "sha256",
        "value": "a7f3b9c..."
      },
      "compression": "zstd",
      "encrypted": false
    }
  ],
  "source_metadata": {
    "type": "filesystem",
    "path": "/data",
    "files_included": 12450,
    "files_excluded": 23,
    "total_bytes_source": 5368709120
  }
}
```

### Validation Rules

- No manifest = incomplete, discard
- Manifest present but checksums fail = corrupt, discard
- Multi-artifact jobs: all artifacts must be present and valid
- Controller verifies every checksum before trusting

### Cleanup Responsibility

Controller owns lifecycle. Agent never deletes.

| State | Cleanup Trigger |
|-------|-----------------|
| Success | After transfer verification |
| Failure | After `BAQUP_ARTIFACT_CLEANUP_DELAY` expires |
| Manual | Operator intervention |

---

## 5. Error Taxonomy

### Exit Codes

| Code | Meaning | Retriable |
|------|---------|-----------|
| 0 | Success | - |
| 1 | General failure | Maybe |
| 64 | Usage/config error | No |
| 65 | Data error | No |
| 69 | Resource unavailable | Yes |
| 70 | Internal error | Maybe |
| 73 | Can't create output | Yes |
| 74 | I/O error | Yes |
| 75 | Completed but unreported | Yes |
| 76 | Partial failure | Configurable |

### Error Categories

| Category | Meaning | Default Retry |
|----------|---------|---------------|
| `config` | Bad configuration | No |
| `transient` | Temporary, may self-resolve | Yes |
| `permanent` | Requires intervention | No |
| `partial` | Some components succeeded | Configurable |
| `internal` | Agent bug | No |

### Error Structure

```json
{
  "error": {
    "code": "SOURCE_UNREACHABLE",
    "category": "transient",
    "retriable": true,
    "message": "Connection failed after 3 attempts",
    "attempts": [
      {"attempt": 1, "timestamp": "...", "error": "ETIMEDOUT", "duration_ms": 30000},
      {"attempt": 2, "timestamp": "...", "error": "ETIMEDOUT", "duration_ms": 30000},
      {"attempt": 3, "timestamp": "...", "error": "ECONNREFUSED", "duration_ms": 50}
    ],
    "root_cause": {
      "code": "SOURCE_UNREACHABLE",
      "message": "Connection to source database timed out"
    },
    "cascading_effects": [
      {"code": "SIZE_ESTIMATION_SKIPPED", "caused_by": "SOURCE_UNREACHABLE"}
    ],
    "suggestions": [
      "Verify network connectivity",
      "Check source database status"
    ]
  }
}
```

### Partial Failure

Per-component status with aggregate summary:

```json
{
  "status": "partial",
  "summary": {"total": 5, "succeeded": 4, "failed": 1},
  "components": [
    {"name": "users_db", "status": "success", "artifact": "users_db.sql.zst"},
    {"name": "legacy_db", "status": "failed", "error": {"code": "PERMISSION_DENIED"}}
  ]
}
```

Successful artifacts are still valid. `BAQUP_AGENT_PARTIAL_FAILURE_MODE` determines exit code.

---

## 6. Observability Contract

### Logging

**Format:** JSON, structured, always.

```json
{
  "timestamp": "2024-01-15T03:02:15.123456Z",
  "level": "info",
  "job_id": "abc123",
  "agent_id": "agent-fs-7f8a9b",
  "component": "archiver",
  "message": "Starting compression",
  "context": {
    "file": "/data/uploads/large.bin",
    "size_bytes": 1073741824
  }
}
```

**Required fields:** `timestamp`, `level`, `job_id`, `message`

**Log levels:** trace, debug, info, warn, error, fatal

**Volume control:** Log transitions not iterations. Per-file logging at debug/trace only.

### Sensitive Data in Logs

**Never log:**
- `BAQUP_SECRET_*` values
- Credentials, tokens, keys
- Full file contents

**Sanitise paths:** Configurable via `BAQUP_AGENT_LOG_SANITIZE`:
- `strict`: Redact all user-identifiable paths
- `moderate`: Redact known sensitive patterns (default)
- `none`: Development only

### Metrics

Agent does NOT expose Prometheus endpoint directly.

Metrics included in status message:
```json
{
  "metrics": {
    "duration_seconds": 347,
    "bytes_processed": 5368709120,
    "files_processed": 12450,
    "compression_ratio": 0.42
  }
}
```

Orchestrator aggregates and exposes to Prometheus. This avoids:
- Cardinality explosion from per-job labels
- Missed scrapes for short-lived containers

### Tracing (Optional)

OpenTelemetry integration when:
- `BAQUP_AGENT_TRACING_ENABLED=true`
- `BAQUP_AGENT_TRACING_ENDPOINT` configured
- Trace context via `BAQUP_JOB_TRACE_PARENT`

Spans: `agent.init`, `agent.connect_source`, `agent.snapshot`, `agent.compress`, `agent.write_artifact`, `agent.report_status`

---

## 7. Security Requirements

### Container Execution

- Run as non-root (UID 1000+)
- No privileged mode
- Read-only root filesystem where possible
- Drop unnecessary capabilities

### Path Handling

**Traversal prevention:**
```python
real_path = os.path.realpath(path)
if not real_path.startswith(allowed_root + os.sep):
    raise SecurityViolation("Path escapes boundary")
```

**Symlink policy** (`BAQUP_SOURCE_FOLLOW_SYMLINKS`):

| Policy | Behaviour |
|--------|-----------|
| `follow` | Follow, validate resolved path within boundary |
| `skip` | Log warning, exclude |
| `fail` | Abort on any symlink |
| `shallow` | Follow internal, skip external (default) |

### Archive Extraction (Zip Slip Prevention)

Validate every member before extraction:
- Reject `../` sequences
- Reject absolute paths
- Resolve and validate destination within boundary

### Secret Handling

**Wrapper type pattern:**
```python
class Secret:
    def __str__(self):
        return "[REDACTED]"
    def reveal(self):
        return self._value
```

**Sanitisation layers:**
1. Secret wrapper type (prevents accidental logging)
2. Connection string builders exclude secrets
3. Error sanitiser as safety net

### Resource Limits

**Streaming:** Never load full file list into memory

**Configurable caps:**
- `BAQUP_SOURCE_MAX_FILES`
- `BAQUP_SOURCE_MAX_DEPTH`
- `BAQUP_AGENT_MEMORY_LIMIT`

**Anomaly detection:** Abort on pathological patterns (e.g., 99% empty files)

**Container limits:** Memory cap as final backstop

---

## 8. Self-Description

### Image Labels

```dockerfile
LABEL io.baqup.agent="true"
LABEL io.baqup.type="fs"
LABEL io.baqup.version="1.2.0"
LABEL io.baqup.roles="snapshot,restore"
LABEL io.baqup.schema='{"..."}'
```

Schema injected at build time:
```bash
docker build --label "io.baqup.schema=$(cat agent-schema.json | jq -c .)" ...
```

### Runtime Commands

**Info mode:**
```bash
docker run ghcr.io/baqup/agent-fs:latest --info
```
```json
{
  "agent": "agent-fs",
  "version": "1.2.0",
  "build": {"commit": "a7f3b9c", "date": "2024-01-10T14:30:00Z"},
  "roles": ["snapshot", "restore"],
  "capabilities": {
    "compression": ["zstd", "gzip", "lz4"],
    "checksums": ["sha256", "sha512", "blake3"],
    "parallel": true
  },
  "requirements": {
    "min_memory": "64M",
    "recommended_memory": "256M"
  }
}
```

**Validation mode:**
```bash
docker run -e BAQUP_SOURCE_PATH=/data ghcr.io/baqup/agent-fs:latest --validate
```
Exit 0 if valid, exit 64 with details if invalid.

### Schema Retrieval (Controller)

Via OCI registry API without pulling image:
1. Get auth token: `GET https://ghcr.io/token?service=ghcr.io&scope=repository:baqup/agent-fs:pull`
2. Fetch manifest: `GET https://ghcr.io/v2/baqup/agent-fs/manifests/latest`
3. Fetch config blob: `GET https://ghcr.io/v2/baqup/agent-fs/blobs/{digest}`
4. Extract `io.baqup.schema` from labels

Controller caches per image digest.

---

## Appendix A: Environment Variable Reference

### BAQUP_JOB_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_JOB_ID` | string | Yes | - | Unique job identifier (UUID) |
| `BAQUP_JOB_TIMEOUT` | int | No | 3600 | Maximum job duration in seconds |
| `BAQUP_JOB_RETRY_MAX` | int | No | 3 | Maximum retry attempts |
| `BAQUP_JOB_RETRY_BACKOFF` | enum | No | exponential | linear, exponential, fibonacci, constant |
| `BAQUP_JOB_TRACE_PARENT` | string | No | - | OpenTelemetry trace context |

### BAQUP_AGENT_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_AGENT_HEARTBEAT_INTERVAL` | int | No | 10 | Seconds between heartbeats |
| `BAQUP_AGENT_LOG_LEVEL` | enum | No | info | trace, debug, info, warn, error, fatal |
| `BAQUP_AGENT_LOG_FORMAT` | enum | No | json | json, text, logfmt |
| `BAQUP_AGENT_LOG_SANITIZE` | enum | No | moderate | strict, moderate, none |
| `BAQUP_AGENT_PARTIAL_FAILURE_MODE` | enum | No | fail | fail, warn, skip |
| `BAQUP_AGENT_MEMORY_LIMIT` | string | No | auto | Memory cap (auto, 512M, 2G) |
| `BAQUP_AGENT_PARALLELISM` | int | No | 0 | Worker count (0 = auto) |

### BAQUP_STAGING_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_STAGING_DIR` | path | Yes | - | Root staging directory |

### BAQUP_ARTIFACT_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_ARTIFACT_COMPRESS` | bool | No | true | Enable compression |
| `BAQUP_ARTIFACT_COMPRESS_ALGORITHM` | enum | No | zstd | zstd, gzip, lz4, none |
| `BAQUP_ARTIFACT_COMPRESS_LEVEL` | int | No | 3 | Compression level |
| `BAQUP_ARTIFACT_CHECKSUM_ALGORITHM` | enum | No | sha256 | sha256, sha512, blake3 |
| `BAQUP_ARTIFACT_ENCRYPT` | bool | No | false | Enable encryption |
| `BAQUP_ARTIFACT_ENCRYPT_ALGORITHM` | enum | No | aes-256-gcm | aes-256-gcm, chacha20-poly1305 |
| `BAQUP_ARTIFACT_ENCRYPT_KEY_ID` | string | No | - | Encryption key identifier |
| `BAQUP_ARTIFACT_SPLIT_SIZE` | string | No | 0 | Split large files (0 = disabled) |
| `BAQUP_ARTIFACT_RETAIN_ON_FAILURE` | bool | No | true | Keep artifacts on failure |
| `BAQUP_ARTIFACT_CLEANUP_DELAY` | int | No | 3600 | Seconds before cleanup |

### BAQUP_BUS_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_BUS_URL` | uri | Yes | - | Redis connection URL |
| `BAQUP_BUS_KEY_PREFIX` | string | No | baqup: | Redis key prefix |
| `BAQUP_BUS_FALLBACK_DIR` | path | No | /tmp/baqup-buffer | Buffer directory |
| `BAQUP_BUS_RECONNECT_MAX_ATTEMPTS` | int | No | 10 | Reconnection attempts |
| `BAQUP_BUS_RECONNECT_BACKOFF` | enum | No | exponential | linear, exponential, constant |

### BAQUP_SOURCE_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_SOURCE_PATH` | path | Conditional | - | Source path (filesystem agent) |
| `BAQUP_SOURCE_HOST` | hostname | Conditional | - | Source host (database agents) |
| `BAQUP_SOURCE_PORT` | int | No | varies | Source port |
| `BAQUP_SOURCE_EXCLUDE` | array | No | - | Exclusion patterns |
| `BAQUP_SOURCE_FOLLOW_SYMLINKS` | enum | No | shallow | follow, skip, fail, shallow |
| `BAQUP_SOURCE_MAX_FILES` | int | No | 0 | File count limit (0 = unlimited) |
| `BAQUP_SOURCE_MAX_DEPTH` | int | No | 0 | Directory depth limit |

### BAQUP_TARGET_*

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `BAQUP_TARGET_PATH` | path | Conditional | - | Target path for restore |
| `BAQUP_TARGET_HOST` | hostname | Conditional | - | Target host for restore |
| `BAQUP_TARGET_PORT` | int | No | varies | Target port |

### BAQUP_SECRET_*

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `BAQUP_SECRET_*` | string | Varies | Sensitive values (never logged) |

---

## Appendix B: Event Codes

| Code | Severity | Description |
|------|----------|-------------|
| `PERMISSION_DENIED` | warn | Access denied to file/resource |
| `SYMLINK_SKIPPED` | info | Symlink excluded per policy |
| `SYMLINK_EXTERNAL` | warn | Symlink points outside boundary |
| `FILE_CHANGED` | warn | File modified during backup |
| `FILE_VANISHED` | warn | File deleted during backup |
| `COMPRESSION_FALLBACK` | info | Fell back to alternative algorithm |
| `CHECKSUM_RETRY` | warn | Checksum mismatch, retrying |
| `CONNECTION_RETRY` | info | Retrying connection |
| `RESOURCE_LIMIT_APPROACHING` | warn | Near configured limit |
| `RESOURCE_LIMIT_EXCEEDED` | error | Limit exceeded, aborting |
| `ANOMALY_DETECTED` | warn | Unusual pattern detected |

---

## Appendix C: Manifest Schema (JSON Schema)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["version", "job_id", "agent", "role", "created_at", "artifacts"],
  "properties": {
    "version": {"type": "string", "const": "1.0"},
    "job_id": {"type": "string", "format": "uuid"},
    "agent": {"type": "string"},
    "agent_version": {"type": "string"},
    "role": {"type": "string", "enum": ["snapshot", "restore", "migrate", "sync"]},
    "created_at": {"type": "string", "format": "date-time"},
    "artifacts": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "required": ["filename", "size_bytes", "checksum"],
        "properties": {
          "filename": {"type": "string"},
          "size_bytes": {"type": "integer", "minimum": 0},
          "checksum": {
            "type": "object",
            "required": ["algorithm", "value"],
            "properties": {
              "algorithm": {"type": "string"},
              "value": {"type": "string"}
            }
          },
          "compression": {"type": "string"},
          "encrypted": {"type": "boolean"}
        }
      }
    },
    "source_metadata": {"type": "object"}
  }
}
```

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-12-01 | Initial specification |
