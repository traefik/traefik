# baqup Design Document

> Label-driven Docker backup controller with Traefik-style autodiscovery

**Version**: 1.0.0-draft
**Last Updated**: 2024-01-15

---

## Table of Contents

1. [Overview](#overview)
2. [Design Philosophy](#design-philosophy)
3. [Architecture](#architecture)
4. [Label Schema](#label-schema)
5. [Core Concepts](#core-concepts)
6. [Agents](#agents)
7. [Communication & Storage](#communication--storage)
8. [Configuration Resolution](#configuration-resolution)
9. [Pipeline & Workflows](#pipeline--workflows)
10. [Validation](#validation)
11. [Restore & Models](#restore--models)
12. [Security & Secrets](#security--secrets)
13. [Implementation Considerations](#implementation-considerations)
14. [Examples](#examples)
15. [Roadmap](#roadmap)
16. [Appendices](#appendices)

---

## Overview

baqup is a Docker backup orchestration system inspired by Traefik's label-based autodiscovery pattern. Containers declare their backup requirements via Docker labels, and the baqup controller orchestrates stateless agents to perform snapshots, transfers, validation, and restoration.

### Core Goals

- **Zero-config simplicity**: Sensible defaults, minimal labels required
- **Full control when needed**: Every aspect configurable via labels or config
- **Separation of concerns**: Controller orchestrates, agents execute
- **Multi-destination**: Single snapshot to multiple storage backends
- **Validation-first**: Verify backups are restorable, not just present
- **Cloud-native**: Works on single Docker, Swarm, or Kubernetes

### Name Origin

"baqup" is a phonetic spelling of "backup", following Traefik's pattern (phonetic for "traffic"). The name was chosen because:

1. It's short and memorable
2. It follows an established pattern in the ecosystem
3. It's unique in the software landscape (baqup was taken, hence the organisation baqupio)

### Why baqup Exists

Existing backup solutions fall into two categories:

1. **Database-native tools** (docker-db-backup, etc.) - Understand database semantics but lack comprehensive UI and multi-destination support
2. **File-based tools with hooks** (Backrest, Borgmatic, etc.) - Good UI but require manual database dump orchestration

baqup bridges this gap by:
- Understanding both database and filesystem semantics via typed agents
- Providing Traefik-style autodiscovery for zero-touch configuration
- Supporting multiple storage destinations with per-destination retention
- Validating backups at every pipeline stage
- Planning for a web UI from the start

---

## Design Philosophy

### Traefik-Inspired Autodiscovery

Just as Traefik discovers HTTP routes from container labels, baqup discovers backup targets:

```yaml
# Traefik pattern
labels:
  - "traefik.http.routers.myapp.rule=Host(`example.com`)"

# baqup pattern
labels:
  - "baqup.snapshot.main-db.type=postgres"
```

**Why this matters**: The compose file becomes self-documenting. Looking at a service definition immediately shows what gets backed up, when, and where - without consulting external configuration.

### Declarative Over Imperative

Containers declare *what* should be backed up, not *how*. The controller determines execution strategy.

**Example**: A container declares `baqup.snapshot.db.type=postgres`. The controller:
1. Determines which agent to use (agent-postgres)
2. Discovers agent capabilities via registry inspection
3. Spawns the agent with appropriate network/volume configuration
4. Manages the pipeline through validation and transfer

The container owner never needs to know about pg_dump, tar streams, or rclone internals.

### Stateless Agents

Agents are ephemeral workers spawned per-job. They:
- Receive configuration via environment variables
- Execute a single operation
- Report results via Redis
- Exit and are removed

**Design rationale**: This pattern was chosen over a sidecar or persistent agent model because:

| Approach | Pros | Cons |
|----------|------|------|
| Sidecar per container | Always available | Resource overhead, version coupling |
| Persistent agent pool | Connection reuse | State management complexity, scaling issues |
| **Ephemeral per-job** | Isolated, right-sized, easy updates | Startup overhead per job |

The ephemeral model wins because:
1. **Isolation**: A failed postgres dump doesn't affect filesystem backups
2. **Right-sized**: Each agent contains only the tools it needs
3. **Updates**: Pull new image, next job uses it - no restart coordination
4. **Parallelism**: Spawn as many agents as needed, naturally scales

### First-Class Entities

Rather than embedding configuration, baqup treats these as independently configurable entities:

| Entity | Role | Why First-Class? |
|--------|------|------------------|
| **Schedule** | When to run | Same schedule reused across snapshots |
| **Retention** | How long to keep | Different retention per storage location |
| **Storage** | Where to send | Multiple destinations, independent config |
| **Validation** | How to verify | Different validation depth per context |
| **Model** | Restore environment | Reusable templates for testing |

**Example of the benefit**: Without first-class retention, you couldn't express "keep 30 on local NAS, 7 on Google Drive, 4 on S3" for the same snapshot.

### The `.default` Pattern

Throughout the schema, we use a `.default` pattern for overrideable settings:

```yaml
- "baqup.snapshot.main-db.retention.default=critical"    # Applied to all storage
- "baqup.snapshot.main-db.retention.offsite=minimal"     # Override for this storage
```

**Why this pattern?**:
1. Explicit about what's a default vs an override
2. Extensible - new storage locations inherit the default
3. Self-documenting - seeing `.default` signals "this can be overridden"
4. Future-proof - allows adding more override dimensions later

---

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         baqup controller                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────────────┐ │
│  │ Discovery │ │ Scheduler │ │Orchestrator│ │   Agent Registry   │ │
│  │           │ │           │ │           │ │                     │ │
│  │ - Docker  │ │ - Cron    │ │ - Jobs    │ │ - GHCR discovery    │ │
│  │   API     │ │ - Triggers│ │ - Pipeline│ │ - Capability cache  │ │
│  │ - Labels  │ │           │ │           │ │                     │ │
│  └───────────┘ └───────────┘ └───────────┘ └─────────────────────┘ │
│                                                                     │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────────────┐ │
│  │  SQLite   │ │   Redis   │ │  Secrets  │ │   Registry Client   │ │
│  │           │ │           │ │  Manager  │ │                     │ │
│  │ - Catalog │ │ - Queue   │ │           │ │ - Label inspection  │ │
│  │ - History │ │ - Pub/Sub │ │ - Encrypt │ │ - Version check     │ │
│  │ - Models  │ │ - State   │ │ - Version │ │                     │ │
│  └───────────┘ └───────────┘ └───────────┘ └─────────────────────┘ │
│                                                                     │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               │ Docker API
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
         ▼                     ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ agent-postgres  │  │   agent-fs      │  │  agent-rclone   │
│                 │  │                 │  │                 │
│ roles:          │  │ roles:          │  │ roles:          │
│ - snapshot      │  │ - snapshot      │  │ - transfer      │
│ - restore       │  │ - restore       │  │ - fetch         │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Does NOT do |
|-----------|----------------|-------------|
| **Discovery** | Parse labels, watch containers | Execute backups |
| **Scheduler** | Evaluate cron, trigger jobs | Know backup details |
| **Orchestrator** | Create jobs, manage pipeline | Execute agent operations |
| **Agent Registry** | Track agent capabilities | Pull images (lazy) |
| **Redis** | Queue, pub/sub, ephemeral state | Persist long-term data |
| **SQLite** | Catalog, history, models | Handle real-time state |
| **Secrets Manager** | Encrypt, version, resolve | Store unencrypted secrets |
| **Registry Client** | Inspect images without pulling | Execute containers |

### Data Flow

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Schedule │───▶│ Snapshot │───▶│ Validate │───▶│ Transfer │
│ Trigger  │    │  Agent   │    │  (local) │    │  Agent   │
└──────────┘    └────┬─────┘    └──────────┘    └────┬─────┘
                     │                               │
                     ▼                               ▼
                ┌──────────┐                   ┌──────────┐
                │ Staging  │                   │ Storage  │
                │ Volume   │                   │ Backend  │
                └──────────┘                   └──────────┘
```

**Detailed flow**:

1. **Schedule triggers** → Scheduler evaluates cron expressions against current time
2. **Job created** → Orchestrator creates job in Redis, resolves configuration
3. **Snapshot agent spawned** → Docker API creates container with target's network
4. **Agent executes** → pg_dump/tar/etc writes to staging volume
5. **Agent reports** → Status written to Redis, controller notified via pub/sub
6. **Validation runs** → Controller checks output (structure, checksums)
7. **Transfer agents spawned** → One per storage destination
8. **Transfer executes** → rclone/s3 uploads from staging
9. **Post-transfer validation** → Checksum verification where supported
10. **Cleanup** → Staging cleared, retention applied, catalog updated

### Organisation Structure

```
github.com/baqupio/
├── baqup                  # Controller (Python)
├── agent-postgres         # Database agents
├── agent-mariadb
├── agent-mongo
├── agent-redis
├── agent-sqlite
├── agent-fs               # Filesystem agent
├── agent-rclone           # Transfer agents
├── agent-s3
├── agent-sftp
└── agent-spec             # Agent specification/contract documentation
```

**Why separate repositories?**:

1. **Independent versioning**: Update agent-postgres without touching others
2. **Independent CI/CD**: Each agent builds on its own schedule
3. **Clear ownership**: Contributors can focus on specific agents
4. **Size optimization**: Each image contains only what it needs
5. **Custom agents**: Third parties follow the same pattern

### Image Naming Convention

```
ghcr.io/baqupio/baqup                 # Controller
ghcr.io/baqupio/agent-postgres        # Agents follow agent-{type} pattern
ghcr.io/baqupio/agent-mariadb
ghcr.io/baqupio/agent-fs
ghcr.io/baqupio/agent-rclone
```

**Convention enables discovery**: Controller can construct image names from snapshot types without explicit configuration.

---

## Label Schema

### Pattern

```
baqup.{entity}.{instance}.{property}
```

Where:
- `entity`: schedule, retention, storage, snapshot, model, validation
- `instance`: User-defined name (must be unique within entity type)
- `property`: Configuration property (may be nested with dots)

### Design Decision: Type as Property

We evaluated two approaches for identifying the agent type:

**Option A: Type in path** (rejected)
```yaml
- "baqup.snapshot.postgres.main-db.port=5432"
```

**Option B: Type as property** (chosen)
```yaml
- "baqup.snapshot.main-db.type=postgres"
- "baqup.snapshot.main-db.port=5432"
```

| Criterion | Type in Path | Type as Property |
|-----------|--------------|------------------|
| Parsing complexity | Need to know all types to parse | Simple regex: `baqup\.(\w+)\.(\w+)\.(.+)` |
| Instance uniqueness | `postgres.main` and `mariadb.main` could coexist | `main` is globally unique per entity |
| Custom types | Parser must handle unknown path segments | Works naturally - type is just a value |
| Visual grouping | All postgres labels cluster | Grouped by instance name |

**Type as property wins** because:
- Simpler, more predictable parsing
- Instance names are globally unique (no collision risk)
- Custom agents work without parser changes
- Matches how other properties are specified

### Complete Label Reference

#### Global

```yaml
- "baqup.enabled=true"    # Required to enable backups on this container
```

**Rationale**: Explicit opt-in prevents accidental backups. A container without `baqup.enabled=true` is completely ignored.

#### Schedules

```yaml
- "baqup.schedule.{name}.cron=<expression>"
- "baqup.schedule.{name}.timezone=<tz>"       # Optional, default UTC
```

**Examples**:
```yaml
- "baqup.schedule.hourly.cron=0 * * * *"
- "baqup.schedule.daily.cron=0 3 * * *"
- "baqup.schedule.weekly.cron=0 4 * * 0"
- "baqup.schedule.business-hours.cron=0 9-17 * * 1-5"
- "baqup.schedule.business-hours.timezone=America/New_York"
```

**Note**: Schedule does NOT include retention. Retention is a separate first-class entity because you may want different retention per storage location for the same schedule.

#### Retention

```yaml
- "baqup.retention.{name}.count=<n>"          # Keep N most recent backups
- "baqup.retention.{name}.days=<n>"           # Keep backups for N days (alternative)
```

**Examples**:
```yaml
- "baqup.retention.critical.count=48"    # Keep 48 hourly backups (2 days)
- "baqup.retention.standard.count=7"     # Keep 7 backups
- "baqup.retention.archive.count=4"      # Keep 4 backups
- "baqup.retention.monthly.days=365"     # Keep for a year
```

**Note**: `count` and `days` are mutually exclusive. If both specified, `count` takes precedence.

#### Storage

```yaml
# Core properties (all storage types)
- "baqup.storage.{name}.type=<agent>"         # rclone, s3, sftp, etc.
- "baqup.storage.{name}.retry.count=<n>"
- "baqup.storage.{name}.retry.backoff=<type>" # linear, exponential, fixed
- "baqup.storage.{name}.retry.initial_delay=<seconds>"
- "baqup.storage.{name}.retry.max_delay=<seconds>"
- "baqup.storage.{name}.verify=<mode>"        # auto, always, never
- "baqup.storage.{name}.window.start=<HH:MM>"
- "baqup.storage.{name}.window.end=<HH:MM>"
- "baqup.storage.{name}.bandwidth=<limit>"    # e.g., 10M, 1G

# Validation overrides (per-storage)
- "baqup.storage.{name}.validation.post_transfer=<level>"
- "baqup.storage.{name}.validation.scheduled=<level>"

# Type-specific properties
# rclone
- "baqup.storage.{name}.remote=<rclone-remote>"

# s3
- "baqup.storage.{name}.bucket=<bucket>"
- "baqup.storage.{name}.region=<region>"
- "baqup.storage.{name}.endpoint=<url>"       # For S3-compatible services
- "baqup.storage.{name}.access_key_env=<var>"
- "baqup.storage.{name}.secret_key_env=<var>"

# sftp
- "baqup.storage.{name}.host=<host>"
- "baqup.storage.{name}.port=<port>"
- "baqup.storage.{name}.username=<user>"
- "baqup.storage.{name}.key_file=<path>"
```

**Storage separated from transfer**: Earlier designs embedded storage config in a "transfer" entity. We separated them because storage is a *destination* (noun) while transfer is an *action* (verb). This enables:
- Defining a storage once, using from multiple snapshots
- Storage-level retry/window settings independent of what's being transferred
- Per-storage validation overrides

#### Snapshots

```yaml
# Core properties
- "baqup.snapshot.{name}.type=<agent>"        # postgres, mariadb, fs, etc.
- "baqup.snapshot.{name}.schedule=<n>"     # Reference to schedule
- "baqup.snapshot.{name}.storage=<list>"      # Comma-separated storage names
- "baqup.snapshot.{name}.enabled=<bool>"      # Toggle individual snapshot
- "baqup.snapshot.{name}.compress=<bool>"     # Enable compression (default: true)
- "baqup.snapshot.{name}.pre_exec=<command>"  # Pre-snapshot hook

# Retention (using .default pattern)
- "baqup.snapshot.{name}.retention.default=<policy>"     # Applied to all storage
- "baqup.snapshot.{name}.retention.{storage}=<policy>"   # Per-storage override

# Validation settings
- "baqup.snapshot.{name}.validation.post_snapshot=<level>"
- "baqup.snapshot.{name}.validation.post_transfer=<level>"
- "baqup.snapshot.{name}.validation.scheduled=<level>"
- "baqup.snapshot.{name}.validation.scheduled.cron=<expression>"
- "baqup.snapshot.{name}.validation.scheduled.scope=<scope>"
- "baqup.snapshot.{name}.validation.on_failure=<action>"

# Restore settings
- "baqup.snapshot.{name}.restore.target=<target>"
- "baqup.snapshot.{name}.restore.capture=<bool>"
- "baqup.snapshot.{name}.restore.model=<n>"
- "baqup.snapshot.{name}.restore.sources=<list>"  # Priority order for restore

# Database-specific properties
- "baqup.snapshot.{name}.host=<host>"
- "baqup.snapshot.{name}.port=<port>"
- "baqup.snapshot.{name}.username=<user>"
- "baqup.snapshot.{name}.password=<literal>"             # Not recommended
- "baqup.snapshot.{name}.password_env=<var>"             # Reference container env
- "baqup.snapshot.{name}.password_file=<path>"           # Reference secret file
- "baqup.snapshot.{name}.databases=<list>"               # all, or comma-separated

# Filesystem-specific properties
- "baqup.snapshot.{name}.path=<path>"
- "baqup.snapshot.{name}.exclude=<patterns>"             # Comma-separated
```

#### Models

Models define restore environment templates for database validation and recovery:

```yaml
- "baqup.model.{name}.image=<image>"
- "baqup.model.{name}.env.{var}=<value>"
- "baqup.model.{name}.tmpfs=<paths>"          # Comma-separated list
- "baqup.model.{name}.memory=<limit>"
- "baqup.model.{name}.network=<network>"
- "baqup.model.{name}.ephemeral=<bool>"       # Auto-remove after use
- "baqup.model.{name}.cmd=<command>"
- "baqup.model.{name}.entrypoint=<entrypoint>"
```

**Why models exist**: Database restores require compatible environments. You can't restore a PostgreSQL 15 dump into a PostgreSQL 12 container. Models let you define test environments that match (or approximate) production for validation restore-tests.

#### Validation Policies

Standalone validation policies for reuse:

```yaml
- "baqup.validation.{name}.level=<level>"     # exists, checksum, integrity, structure, restore-test
- "baqup.validation.{name}.schedule=<n>"   # When to run
- "baqup.validation.{name}.scope=<scope>"     # latest, all, count:N
- "baqup.validation.{name}.on_failure=<action>"  # alert, alert+retry, quarantine
```

---

## Core Concepts

### Schedules

Schedules define when backups run. They are named entities that can be referenced by multiple snapshots.

```yaml
labels:
  - "baqup.schedule.hourly.cron=0 * * * *"
  - "baqup.schedule.daily.cron=0 3 * * *"
  - "baqup.schedule.weekly.cron=0 4 * * 0"
```

**Controller defaults**: If a container doesn't define any schedules, it uses the controller's default schedules. If a snapshot doesn't specify a schedule, it uses `daily` by default.

**Cron expression format**: Standard 5-field cron (`minute hour day month weekday`).

### Retention

Retention defines how long backups are kept. Crucially, retention is separated from schedules to allow per-storage policies.

```yaml
labels:
  - "baqup.retention.critical.count=30"
  - "baqup.retention.standard.count=7"
  - "baqup.retention.minimal.count=3"
```

**The `.default` pattern for retention**:

```yaml
labels:
  - "baqup.snapshot.main-db.retention.default=critical"    # All storage locations
  - "baqup.snapshot.main-db.retention.offsite=minimal"     # Override for expensive S3
```

This allows:
- Keep 30 backups on local NAS (lots of space, fast recovery)
- Keep 30 backups on Google Drive (default applies)
- Keep only 3 backups on S3 Glacier (expensive)

### Storage

Storage defines where backups are sent. Storage is separated from the snapshot definition to enable:
- Reuse across multiple snapshots
- Independent configuration (retry, windows, bandwidth)
- Per-storage validation settings

```yaml
labels:
  # Local NAS - fast, no restrictions
  - "baqup.storage.nas.type=rclone"
  - "baqup.storage.nas.remote=nas:/backups"
  
  # Google Drive - limited bandwidth
  - "baqup.storage.gdrive.type=rclone"
  - "baqup.storage.gdrive.remote=gdrive:backups"
  - "baqup.storage.gdrive.bandwidth=10M"
  
  # Offsite S3 - expensive, transfer at night only
  - "baqup.storage.offsite.type=s3"
  - "baqup.storage.offsite.bucket=disaster-recovery"
  - "baqup.storage.offsite.region=us-east-1"
  - "baqup.storage.offsite.window.start=02:00"
  - "baqup.storage.offsite.window.end=06:00"
  - "baqup.storage.offsite.validation.post_transfer=none"  # Skip - too expensive
```

### Snapshots

Snapshots define what to backup, tying together schedules, storage, and retention:

```yaml
labels:
  - "baqup.snapshot.main-db.type=postgres"
  - "baqup.snapshot.main-db.username=postgres"
  - "baqup.snapshot.main-db.password_env=POSTGRES_PASSWORD"
  - "baqup.snapshot.main-db.schedule=hourly"
  - "baqup.snapshot.main-db.storage=nas,gdrive,offsite"
  - "baqup.snapshot.main-db.retention.default=critical"
  - "baqup.snapshot.main-db.retention.offsite=minimal"
```

**Multiple snapshots per container**: A single container can define multiple snapshots:

```yaml
labels:
  # Database snapshot
  - "baqup.snapshot.db.type=postgres"
  - "baqup.snapshot.db.schedule=hourly"
  
  # Config files snapshot
  - "baqup.snapshot.config.type=fs"
  - "baqup.snapshot.config.path=/config"
  - "baqup.snapshot.config.schedule=daily"
```

---

## Agents

### Design Principles

1. **Single image, multiple roles**: `agent-postgres` handles snapshot, restore, and validation
2. **Stateless**: All configuration via environment variables
3. **Self-describing**: Docker image labels declare capabilities
4. **Contract-based**: Follow agent specification for interoperability

### Agent Image Labels

Every agent image must include these labels:

```dockerfile
LABEL io.baqup.agent="true"
LABEL io.baqup.type="postgres"
LABEL io.baqup.version="1.0.0"
LABEL io.baqup.roles="snapshot,restore"
LABEL io.baqup.schema='{"..."}' # Injected at build time
```

The `io.baqup.schema` label contains a JSON schema defining the agent's configuration contract (required/optional variables, types, defaults, constraints). This is injected at build time:

```bash
docker build \
  --label "io.baqup.schema=$(cat agent-schema.json | jq -c .)" \
  -t ghcr.io/baqupio/agent-postgres:latest .
```

**Why image labels?**: The controller can inspect image labels via the registry API *without pulling the image*. This enables:
- Discovery of available agents
- Capability checking before job creation
- Version comparison for updates
- Schema-driven configuration validation

### Supported Roles

| Role | Purpose | Example Agents |
|------|---------|----------------|
| `snapshot` | Extract data from source | postgres, mariadb, mongo, redis, sqlite, fs |
| `restore` | Restore data to target | postgres, mariadb, mongo, redis, sqlite, fs |
| `transfer` | Upload to storage | rclone, s3, sftp |
| `fetch` | Download from storage | rclone, s3, sftp |

**Note**: Validation is typically handled by the snapshot agent (same tools needed to validate a SQL dump as to create one).

### Agent Discovery

The controller discovers agents through three mechanisms (in priority order):

1. **Explicit configuration**: User specifies exact images
2. **Convention**: Assume `ghcr.io/baqupio/agent-{type}:latest` exists
3. **Registry inspection**: Query GHCR API for image labels

```yaml
# Controller config
agents:
  registry: ghcr.io
  namespace: baqupio
  discovery: auto  # auto | explicit
  pull_policy: lazy  # lazy | startup | always
  
  custom:
    # Add custom agent
    oracle:
      image: mycompany/agent-oracle:latest
      auth:
        username_env: REGISTRY_USER
        password_env: REGISTRY_PASS
    
    # Override official agent
    postgres:
      image: mycompany/agent-postgres:custom
    
    # Local-only agent (never pull)
    experimental:
      image: my-agent:dev
      pull_policy: never
```

### Agent Execution

Controller spawns agents via Docker API:

```python
docker_client.containers.run(
    image="ghcr.io/baqupio/agent-postgres:latest",
    command=["snapshot"],
    environment={
        "BAQUP_JOB_ID": job.id,
        "BAQUP_OUTPUT_DIR": "/output",
        "BAQUP_REDIS_URL": "redis://redis:6379",
        "BAQUP_PG_HOST": "localhost",  # Via network_mode, localhost IS the target
        "BAQUP_PG_PORT": "5432",
        "BAQUP_PG_USER": "postgres",
        "BAQUP_PG_PASSWORD": resolved_password,
        "BAQUP_COMPRESS": "true",
    },
    network_mode=f"container:{target_container_id}",  # Share network namespace
    volumes={
        staging_volume: {"bind": "/output", "mode": "rw"},
    },
    remove=True,  # Auto-cleanup on exit
)
```

**Key insight**: `network_mode=container:{id}` attaches the agent to the target container's network namespace. This means `localhost` inside the agent IS the database container. No network configuration needed.

### Agent Contract

All agents must:

1. **Read configuration from environment variables** (prefixed with `BAQUP_`)
2. **Write output to `$BAQUP_STAGING_DIR`** using atomic completion pattern
3. **Report status to Redis on completion**
4. **Use standard exit codes** (see Appendix C)

See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for the complete agent specification.

### Agent Lifecycle

Agents follow a defined state machine:

```
INITIALIZING → RUNNING → COMPLETING → TERMINATED
     │            │           │
     └──► FAILED ◄┴───────────┘
```

| State | Description | Timeout |
|-------|-------------|---------|
| `INITIALIZING` | Config validation, bus connection | 30s |
| `RUNNING` | Active work, heartbeats every 10s | Job timeout |
| `COMPLETING` | Writing artifacts, checksums | 60s |
| `TERMINATED` | Exit 0, status reported | - |
| `FAILED` | Exit non-zero, error reported | - |

**Heartbeat**: Agents emit heartbeats every 10 seconds with an `intent` field signalling expected duration of long operations. Controller presumes agent dead after 3 missed heartbeats (30s silence).

### Runtime Commands

Agents support introspection commands:

```bash
# Get agent info and capabilities
docker run ghcr.io/baqupio/agent-postgres:latest --info

# Validate configuration without executing
docker run -e BAQUP_SOURCE_HOST=db ghcr.io/baqupio/agent-postgres:latest --validate
```

**Status report structure** (written to Redis):

```json
{
  "job_id": "abc123",
  "agent": "agent-postgres",
  "role": "snapshot",
  "status": "success",
  "started_at": "2024-01-15T03:00:00Z",
  "completed_at": "2024-01-15T03:00:47Z",
  "files": [
    {
      "path": "/staging/abc123/20240115T030000Z.sql.zst",
      "size_bytes": 1048576,
      "checksum_sha256": "abc123..."
    }
  ],
  "error": null,
  "metrics": {
    "duration_seconds": 47.2,
    "bytes_processed": 52428800
  },
  "warnings_summary": {
    "PERMISSION_DENIED": 0,
    "SYMLINK_SKIPPED": 0
  }
}
```

### Agent Base Images

| Agent | Recommended Base | Rationale |
|-------|------------------|-----------|
| agent-postgres | postgres:XX-alpine | Guaranteed compatible pg_dump |
| agent-mariadb | mariadb:XX | Compatible mariadb-dump |
| agent-mongo | mongo:XX | Compatible mongodump |
| agent-redis | redis:XX-alpine | Compatible redis-cli |
| agent-fs | alpine:latest | Just needs tar, gzip |
| agent-rclone | rclone/rclone:latest | Official, always current |
| agent-s3 | alpine + aws-cli | Minimal |

**Version matching matters**: Using the official postgres:15 image for agent-postgres guarantees the pg_dump version matches the server version, avoiding compatibility issues.

---

## Communication & Storage

### Redis

Redis serves as the communication backbone for job coordination:

| Purpose | Redis Structure | TTL |
|---------|-----------------|-----|
| Job queue | List: `baqup:jobs:pending:{agent_type}` | None |
| Active jobs | Set: `baqup:jobs:running` | None |
| Job details | Hash: `baqup:job:{id}` | 24h after completion |
| Events | Pub/Sub: `baqup:events` | N/A |
| Agent heartbeat | Hash: `baqup:agent:{id}` | 60s (refresh) |

**Why Redis over filesystem?**:

The initial design used a shared filesystem for job status. This was rejected because:

| Approach | Single Node | Swarm | Kubernetes |
|----------|-------------|-------|------------|
| Filesystem | Works | Requires shared storage | Complex (PVCs) |
| **Redis** | Works | Works | Works (single service) |

Redis provides:
- Atomic operations (LPUSH/BRPOP for reliable queuing)
- Pub/Sub for real-time event notification
- TTL for automatic cleanup
- Works identically across deployment models

### Job Lifecycle in Redis

```python
# 1. Controller creates job
HSET baqup:job:{id} status pending container xyz ...
LPUSH baqup:jobs:pending:postgres {id}

# 2. Agent picks up job (blocking pop)
BRPOP baqup:jobs:pending:postgres 30 → {id}
HSET baqup:job:{id} status running started_at ...
SADD baqup:jobs:running {id}

# 3. Agent completes
HSET baqup:job:{id} status success files [...] ...
SREM baqup:jobs:running {id}
PUBLISH baqup:events {type: "job_completed", id: ...}
EXPIRE baqup:job:{id} 86400  # 24h TTL

# 4. Controller receives event
SUBSCRIBE baqup:events
→ Triggers next pipeline stage
```

### SQLite

SQLite stores persistent metadata that must survive Redis restarts:

| Table | Purpose |
|-------|---------|
| `backups` | Backup catalog with captured configs |
| `validations` | Validation history |
| `models` | Restore environment templates |

```sql
CREATE TABLE backups (
    id TEXT PRIMARY KEY,
    container_id TEXT NOT NULL,
    container_name TEXT NOT NULL,
    snapshot_instance TEXT NOT NULL,
    snapshot_type TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    size_bytes INTEGER,
    checksum TEXT,
    storage_locations JSON,     -- ["nas", "gdrive", "offsite"]
    captured_config JSON,       -- Container config at backup time
    secret_ref TEXT,            -- Reference to versioned secret
    metadata JSON
);

CREATE TABLE validations (
    id TEXT PRIMARY KEY,
    backup_id TEXT NOT NULL REFERENCES backups(id),
    storage TEXT NOT NULL,
    level TEXT NOT NULL,
    status TEXT NOT NULL,       -- passed, failed, skipped
    checks JSON,                -- {exists: true, checksum: true, ...}
    error TEXT,
    validated_at DATETIME NOT NULL
);

CREATE TABLE models (
    name TEXT PRIMARY KEY,
    config JSON NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

### Staging Volume

Local temporary storage before transfer:

```
/var/lib/baqup/staging/
├── {job_id}/
│   ├── .staging/              # Work in progress (invisible to controller)
│   │   ├── data.tar.zst
│   │   └── manifest.json
│   ├── data.tar.zst           # Final artifact (atomic move from .staging/)
│   └── manifest.json          # Completion signal
```

**Atomic completion pattern**: Agents write artifacts to `.staging/` subdirectory, then atomically move to parent when complete. The presence of `manifest.json` in the parent directory signals completion. See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for manifest schema and validation rules.

**Controller view** (organised by container after transfer):

```
/var/lib/baqup/staging/
├── nextcloud/
│   ├── main-db/
│   │   └── 20240115T030000Z.sql.zst
│   └── config/
│       └── 20240115T030000Z.tar.zst
```

**Timestamp format**: `20240115T030000Z` (compact ISO8601)

- No colons (filesystem safe on all platforms)
- UTC timezone (Z suffix)
- Sortable chronologically

---

## Configuration Resolution

### Priority Order

When the same target is defined in both controller config and container labels:

```
1. Container labels        → Highest priority (if label-wins mode)
2. Controller config       → Fallback for unlabelled containers
3. Controller defaults     → Final fallback for missing properties
```

### Conflict Resolution Modes

```yaml
# Controller config
discovery:
  conflict_resolution: label-wins  # label-wins | config-wins | error | merge
```

| Mode | Behaviour | Use Case |
|------|-----------|----------|
| `label-wins` | Labels override config | Default - container owner decides |
| `config-wins` | Config overrides labels | Central control, labels as hints |
| `error` | Fail if both define same target | Strict environments |
| `merge` | Deep merge, labels win per-property | Maximum flexibility |

**Default is `label-wins`** because:
- Labels are "closest to the workload"
- Container owner understands their backup needs best
- Config provides defaults for containers that don't specify

### Merge Mode Detail

When `conflict_resolution: merge`:

```yaml
# Controller config
targets:
  webapp:
    snapshot:
      db:
        type: postgres
        port: 5432
        username: admin
        schedule: weekly
```

```yaml
# Container labels
labels:
  - "baqup.snapshot.db.schedule=daily"  # Override
  - "baqup.snapshot.db.databases=app"   # Add new
```

```python
# Merged result
{
    "type": "postgres",      # From config
    "port": 5432,            # From config
    "username": "admin",     # From config
    "schedule": "daily",     # From labels (override)
    "databases": "app"       # From labels (new)
}
```

### Property Resolution Chain

```python
def resolve_property(snapshot, property_name):
    # 1. Explicit on snapshot (from labels or config)
    if property_name in snapshot.properties:
        return snapshot.properties[property_name]
    
    # 2. Controller defaults for this snapshot type
    type_defaults = controller.defaults.get(snapshot.type, {})
    if property_name in type_defaults:
        return type_defaults[property_name]
    
    # 3. Global controller defaults
    if property_name in controller.defaults.global:
        return controller.defaults.global[property_name]
    
    # 4. Built-in defaults
    return BUILT_IN_DEFAULTS.get(property_name)
```

### Credential Resolution

Three methods for providing credentials, checked in order:

```yaml
# Method 1: Direct value (NOT RECOMMENDED - visible in docker inspect)
- "baqup.snapshot.main-db.password=literal"

# Method 2: Container environment variable (RECOMMENDED)
- "baqup.snapshot.main-db.password_env=POSTGRES_PASSWORD"

# Method 3: Secret file (for Docker secrets)
- "baqup.snapshot.main-db.password_file=/run/secrets/db_pass"
```

**Resolution logic**:

```python
def resolve_credential(snapshot, container):
    # Direct value (checked first for backwards compatibility)
    if snapshot.password:
        log.warning("Direct password in labels is not recommended")
        return snapshot.password
    
    # Environment variable reference
    if snapshot.password_env:
        # Inspect container to get env value
        env = container.attrs['Config']['Env']
        for item in env:
            key, _, value = item.partition('=')
            if key == snapshot.password_env:
                return value
        raise CredentialError(f"Env var {snapshot.password_env} not found")
    
    # Secret file reference
    if snapshot.password_file:
        # Exec into container to read file
        result = container.exec_run(f"cat {snapshot.password_file}")
        if result.exit_code != 0:
            raise CredentialError(f"Failed to read {snapshot.password_file}")
        return result.output.decode().strip()
    
    raise CredentialError("No credential method specified")
```

**Controller→Agent handoff**: The controller resolves credentials from labels and passes them to agents as `BAQUP_SECRET_*` environment variables. Agents treat all `BAQUP_SECRET_*` variables as sensitive (never logged, cleared after use). See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for agent-side secret handling requirements.

---

## Pipeline & Workflows

### Snapshot Pipeline

```
Schedule triggers
        │
        ▼
┌───────────────────┐
│ 1. Create job     │
│    - Resolve config│
│    - Store in Redis│
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ 2. Capture secrets│
│    - Resolve creds │
│    - Encrypt       │
│    - Version       │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ 3. Capture config │
│    - Container env │
│    - Image info    │
│    - Networks      │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ 4. Spawn snapshot │
│    agent          │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐     ┌───────────────────┐
│ 5. Post-snapshot  │────▶│ Fail: Alert,      │
│    validation     │     │ retry snapshot    │
└─────────┬─────────┘     └───────────────────┘
          │ Pass
          ▼
┌───────────────────┐
│ 6. For each       │
│    storage:       │
│    spawn transfer │
└─────────┬─────────┘
          │
          ▼ (per storage)
┌───────────────────┐     ┌───────────────────┐
│ 7. Post-transfer  │────▶│ Fail: Alert,      │
│    validation     │     │ retry transfer    │
└─────────┬─────────┘     │ continue others   │
          │ Pass          └───────────────────┘
          ▼
┌───────────────────┐
│ 8. Finalize       │
│    - Update catalog│
│    - Cleanup staging│
│    - Apply retention│
└───────────────────┘
```

### Transfer Behaviour

**Failure handling**: Always continue to other storage locations. Each storage is independent.

```yaml
defaults:
  transfer:
    retry:
      count: 3
      backoff: exponential  # 5s, 10s, 20s, 40s...
      initial_delay: 5
      max_delay: 300
    on_failure: continue  # Continue to other storage
```

**Transfer windows**: Per-storage time restrictions for expensive/metered connections:

```yaml
- "baqup.storage.offsite.window.start=02:00"
- "baqup.storage.offsite.window.end=06:00"
```

Jobs outside window are queued in Redis until window opens:

```python
def schedule_transfer(job, storage):
    if storage.window and not in_window(storage.window, now()):
        # Calculate next window start
        next_window = calculate_next_window(storage.window)
        job.scheduled_at = next_window
        redis.zadd('baqup:transfers:scheduled', {job.id: next_window.timestamp()})
    else:
        redis.lpush(f'baqup:jobs:pending:{storage.type}', job.id)
```

**Bandwidth limits**: Per-storage throttling:

```yaml
- "baqup.storage.offsite.bandwidth=10M"  # 10 MB/s
```

Passed to transfer agent as environment variable, agent implements throttling.

### Multi-Storage Execution Example

```yaml
labels:
  - "baqup.snapshot.main-db.storage=nas,gdrive,offsite"
```

```
Snapshot completes at 14:00
        │
        ├──▶ nas: no window → immediate transfer
        │         └── success ✓
        │
        ├──▶ gdrive: no window → immediate transfer
        │         ├── attempt 1: fail (network error)
        │         ├── wait 5s
        │         ├── attempt 2: fail
        │         ├── wait 10s
        │         ├── attempt 3: fail
        │         ├── wait 20s
        │         ├── attempt 4 (max): fail
        │         └── ALERT raised, continue to next storage
        │
        └──▶ offsite: window 02:00-06:00 → queued
                  └── will transfer at 02:00 next day

Final result:
- nas: ✓ transferred
- gdrive: ✗ failed (alert raised)
- offsite: ⏳ scheduled for 02:00
```

---

## Validation

### The Problem Validation Solves

A backup that exists but can't be restored is worthless. Validation catches:

| Problem | How Validation Catches It |
|---------|---------------------------|
| Corrupted during write | Checksum mismatch |
| Incomplete upload | Size mismatch, missing files |
| Storage failure | Connection error, file not found |
| Backup file corrupted | Decompression failure |
| Database dump malformed | SQL parse error |
| Missing encryption key | Decryption failure |

### Stage-Aware Validation

Validation occurs at different pipeline stages with appropriate depth:

| Stage | Trigger | Default Level | Rationale |
|-------|---------|---------------|-----------|
| `post_snapshot` | After snapshot, before transfer | `structure` | Don't waste bandwidth transferring garbage |
| `post_transfer` | After each storage transfer | `auto` | Confirm transfer integrity |
| `scheduled` | Periodic cron | `checksum` | Catch bit-rot over time |

### Validation Levels

| Level | What It Checks | Cost | Confidence |
|-------|----------------|------|------------|
| `none` | Skip validation | Free | None |
| `exists` | File present in storage | API call | Low |
| `checksum` | Hash matches original | Depends on storage | Medium |
| `integrity` | File readable, decompresses | Download + decompress | Medium-High |
| `structure` | Format valid (SQL parseable, tar OK) | Download + parse | High |
| `restore-test` | Actually restore to temp environment | Full restore | Highest |

### Auto Mode for Post-Transfer

`post_transfer=auto` intelligently selects validation level:

```python
def auto_validation_level(storage):
    # Remote supports checksums? Use them (cheap)
    if storage.supports_remote_checksum():
        return "checksum"
    
    # Local/NAS? Full structure check is still cheap
    elif storage.is_local():
        return "structure"
    
    # Remote without checksum support? Skip (too expensive)
    else:
        return "none"
```

**Storage checksum support**:

| Storage Type | Checksum Method | Cost |
|--------------|-----------------|------|
| S3 | ETag (MD5) in response header | Free |
| GCS | MD5/CRC32C in metadata | Free |
| Azure Blob | Content-MD5 header | Free |
| rclone remotes | `--checksum` flag | Varies |
| SFTP | No native support | Must download |

### Validation Configuration

```yaml
labels:
  # Per-snapshot validation
  - "baqup.snapshot.main-db.validation.post_snapshot=structure"
  - "baqup.snapshot.main-db.validation.post_transfer=auto"
  - "baqup.snapshot.main-db.validation.scheduled=checksum"
  - "baqup.snapshot.main-db.validation.scheduled.cron=0 4 * * 0"
  - "baqup.snapshot.main-db.validation.scheduled.scope=latest"
  - "baqup.snapshot.main-db.validation.on_failure=alert"
  
  # Per-storage override (for expensive storage)
  - "baqup.storage.glacier.validation.post_transfer=none"
  - "baqup.storage.glacier.validation.scheduled=exists"
  - "baqup.storage.glacier.validation.scheduled.cron=0 0 1 * *"  # Monthly
```

### Scheduled Validation Scope

| Scope | Behaviour |
|-------|-----------|
| `latest` | Only validate most recent backup |
| `all` | Validate all backups within retention |
| `count:N` | Validate N most recent |

### Failure Actions

| Action | Behaviour |
|--------|-----------|
| `alert` | Send notification, take no other action |
| `alert+retry` | Send notification, retry validation once |
| `quarantine` | Mark backup as suspect, exclude from restore candidates |

**Quarantine behavior**: Quarantined backups are:
- Flagged in the catalog database
- Excluded from `restore.sources` resolution
- Still retained (not deleted) for manual investigation
- Visible in UI with warning indicator

---

## Restore & Models

### The Database Restore Problem

Unlike filesystem restores (just extract tar), database restores require:

1. **Compatible database version**: pg_restore from PG15 may not work in PG12
2. **Required extensions**: If backup used PostGIS, restore target needs PostGIS
3. **Proper configuration**: Character sets, collation, etc.
4. **Valid credentials**: Different from production in test environments

### Restore Targets

| Target | Description | Risk Level | Use Case |
|--------|-------------|------------|----------|
| `mirror` | Auto-created from captured config | Low | Default, safe testing |
| `model:{name}` | User-defined template | Low | Validation, controlled tests |
| `container:{name}` | Existing container | Medium | Standby servers |
| `source` | Original container | **High** | Production recovery |

### Configuration Capture

At snapshot time, baqup captures the source container's configuration:

```json
{
  "captured": {
    "image": "postgres:15.4",
    "image_id": "sha256:abc123def456...",
    "env": {
      "POSTGRES_USER": "app",
      "POSTGRES_DB": "mydb",
      "POSTGRES_PASSWORD": "${SECRET:nextcloud/main-db/20240115T030000Z-v1}"
    },
    "cmd": null,
    "entrypoint": null,
    "networks": ["app_network"],
    "labels": {
      "com.docker.compose.project": "nextcloud"
    },
    "captured_at": "2024-01-15T03:00:00Z"
  },
  "overrides": {
    // User-specified overrides (if any)
  }
}
```

**Note**: Secrets are NOT stored in captured config. Instead, a reference to the versioned encrypted secret is stored.

### Models

Models are user-defined restore environment templates:

```yaml
labels:
  - "baqup.model.pg-test.image=postgres:15"
  - "baqup.model.pg-test.env.POSTGRES_USER=test"
  - "baqup.model.pg-test.env.POSTGRES_PASSWORD=test"
  - "baqup.model.pg-test.env.POSTGRES_DB=restore_test"
  - "baqup.model.pg-test.tmpfs=/var/lib/postgresql/data,/tmp"
  - "baqup.model.pg-test.memory=512m"
  - "baqup.model.pg-test.network=baqup-test"
  - "baqup.model.pg-test.ephemeral=true"
```

**Model properties**:

| Property | Description |
|----------|-------------|
| `image` | Container image (required) |
| `env.*` | Environment variables |
| `tmpfs` | Comma-separated tmpfs mounts (fast, ephemeral) |
| `memory` | Memory limit |
| `network` | Network to attach |
| `ephemeral` | Auto-remove after use |
| `cmd` | Override command |
| `entrypoint` | Override entrypoint |

**Why tmpfs?**: For validation restore-tests, we don't need persistence. tmpfs provides:
- Faster I/O (RAM-backed)
- Automatic cleanup
- No disk space concerns

### Restore Flow

```
Restore request: main-db from 20240115T030000Z
        │
        ▼
┌───────────────────────────────────────┐
│ 1. Resolve target                     │
│                                       │
│    target=mirror?                     │
│      → Load captured config           │
│      → Apply any overrides            │
│                                       │
│    target=model:pg-test?              │
│      → Load model definition          │
│                                       │
│    target=container:standby?          │
│      → Find existing container        │
│                                       │
│    target=source?                     │
│      → REQUIRE CONFIRMATION           │
│      → Find original container        │
└─────────────────┬─────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────┐
│ 2. Create/connect target container    │
│                                       │
│    mirror/model:                      │
│      → docker.containers.run(config)  │
│      → Wait for healthy               │
│                                       │
│    container/source:                  │
│      → docker.containers.get(name)    │
│      → Verify accessible              │
└─────────────────┬─────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────┐
│ 3. Resolve secrets                    │
│                                       │
│    → Load secret reference from backup│
│    → Decrypt versioned secret         │
│    → Inject into restore environment  │
└─────────────────┬─────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────┐
│ 4. Fetch backup                       │
│                                       │
│    → Try sources in priority order    │
│    → Download to staging              │
│    → Verify checksum                  │
└─────────────────┬─────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────┐
│ 5. Spawn restore agent                │
│                                       │
│    → Mount staging volume             │
│    → Attach to target network         │
│    → Execute pg_restore / tar extract │
└─────────────────┬─────────────────────┘
                  │
                  ▼
┌───────────────────────────────────────┐
│ 6. Post-restore actions               │
│                                       │
│    ephemeral model:                   │
│      → Remove container               │
│                                       │
│    persistent restore:                │
│      → Keep running                   │
│      → Notify user                    │
└───────────────────────────────────────┘
```

### Restore Sources

Priority order for fetching backups during restore:

```yaml
- "baqup.snapshot.main-db.restore.sources=nas,gdrive,offsite"
```

Resolution:
1. Try `nas` - if available and file exists, use it
2. If `nas` fails, try `gdrive`
3. If `gdrive` fails, try `offsite`
4. If all fail, abort with error

**Default**: If `restore.sources` not specified, defaults to reverse of `storage` list (assumes local storage is listed first and is fastest for recovery).

---

## Security & Secrets

### Principles

1. **Never store secrets in plain text**: All secrets encrypted at rest
2. **Reference secrets, don't embed**: Captured configs contain references, not values
3. **Version secrets**: Each backup references a specific secret version
4. **Support rotation**: Password changes create new versions without breaking old restores

### Secret Versioning

The key insight: **backups and secrets must be versioned together**.

If you back up a database on Monday, change the password on Tuesday, and try to restore Wednesday using Monday's backup, you need Monday's password - not Tuesday's.

### Secret Storage Structure

```
/storage/{storage}/baqup-secrets/
├── {container}/
│   └── {snapshot}/
│       └── {timestamp}-v{version}.enc

# Example
/storage/nas/baqup-secrets/
├── nextcloud/
│   └── main-db/
│       ├── 20240115T030000Z-v1.enc   # Initial password
│       ├── 20240116T030000Z-v1.enc   # Same password
│       ├── 20240117T030000Z-v1.enc   # Same password
│       └── 20240118T030000Z-v2.enc   # Password rotated
```

**Version increment**: A new version is created when the resolved secret value changes from the previous backup.

### Secret References in Captured Config

```json
{
  "env": {
    "POSTGRES_USER": "app",
    "POSTGRES_PASSWORD": "${SECRET:nextcloud/main-db/20240115T030000Z-v1}"
  }
}
```

The reference format: `${SECRET:{container}/{snapshot}/{timestamp}-v{version}}`

### Encryption

Secrets encrypted with a controller-managed key:

```yaml
# Controller config
secrets:
  encryption_key_env: BAQUP_SECRET_KEY
  algorithm: AES-256-GCM  # Default
  storage: default  # Store in default storage backend
```

**Key management options**:
- Environment variable (simple, for homelabs)
- Docker secret mount
- External secret manager (Vault, AWS Secrets Manager) - roadmap item

### Secret Resolution Flow

```python
def resolve_secret_for_restore(backup, secret_ref):
    # Parse reference
    # ${SECRET:nextcloud/main-db/20240115T030000Z-v1}
    container, snapshot, timestamp_version = parse_secret_ref(secret_ref)
    
    # Locate encrypted secret file
    secret_path = f"baqup-secrets/{container}/{snapshot}/{timestamp_version}.enc"
    
    # Download from storage
    encrypted_data = storage.download(secret_path)
    
    # Decrypt with controller key
    key = get_encryption_key()
    decrypted = decrypt(encrypted_data, key)
    
    return json.loads(decrypted)
```

---

## Authentication & Authorization (v2)

This section documents the planned authentication and authorization system for v2. v1 includes preparatory structures (audit log, `initiated_by` metadata) but does not enforce authentication.

### Design Principles

1. **OIDC-first**: Support standard OIDC/OAuth2 providers (Authentik, Keycloak, Okta, Azure AD)
2. **Scope-based authorization**: OAuth2 scopes for API access control
3. **Claims-to-roles mapping**: Flexible mapping from IdP claims to baqup roles
4. **Audit everything**: All operations logged with actor identity

### OAuth2 Flows

| Use Case | Flow | Notes |
|----------|------|-------|
| Web UI login | Authorization Code + PKCE | Standard browser flow |
| CLI authentication | Device Authorization | `baqup login` command |
| API automation | Client Credentials | Service accounts |
| API keys | N/A (baqup-issued) | Fallback for simple integrations |

### OAuth2 Scopes

```
baqup:read              # View backups, schedules, status
baqup:write             # Trigger backups, modify schedules
baqup:restore           # Perform restores (high-risk)
baqup:admin             # Manage storage, agents, system config
baqup:audit:read        # View audit logs
```

### Default Roles

| Role | Scopes | Use Case |
|------|--------|----------|
| `baqup-viewer` | `baqup:read` | View-only access |
| `baqup-operator` | `baqup:read`, `baqup:write` | Day-to-day operations |
| `baqup-restorer` | `baqup:read`, `baqup:restore` | Disaster recovery |
| `baqup-admin` | All scopes | Full control |

### OIDC Claims Mapping

baqup expects custom claims in the JWT token. Example Authentik property mapping:

```python
# Authentik property mapping expression
return {
    "baqup_roles": [g.name for g in user.ak_groups.all() if g.name.startswith("baqup-")],
    "baqup_scopes": user.attributes.get("baqup_scopes", [])
}
```

Resulting JWT claims:

```json
{
  "sub": "user-uuid",
  "email": "admin@example.com",
  "baqup_roles": ["baqup-admin"],
  "baqup_scopes": ["baqup:read", "baqup:write", "baqup:restore", "baqup:admin"]
}
```

### v1 Preparation: Job Metadata

All jobs include an `initiated_by` field for audit trail:

```json
{
  "id": "abc123",
  "initiated_by": {
    "type": "system",
    "identity": null,
    "display_name": null,
    "scopes": ["*"],
    "ip_address": null
  }
}
```

| Field | v1 Value | v2 Value |
|-------|----------|----------|
| `type` | `"system"` | `"user"`, `"api_key"`, `"service"` |
| `identity` | `null` | OIDC `sub` claim or API key ID |
| `display_name` | `null` | Human-readable name from IdP |
| `scopes` | `["*"]` | Scopes at time of request |
| `ip_address` | `null` | Client IP for audit |

### v1 Preparation: Audit Log

All operations are logged to SQLite for v2 querying:

```sql
CREATE TABLE audit_log (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,

    -- Actor
    actor_type TEXT NOT NULL,        -- system, user, api_key, service
    actor_identity TEXT,             -- OIDC sub, API key ID
    actor_display_name TEXT,
    actor_ip TEXT,

    -- Action
    action TEXT NOT NULL,            -- backup.triggered, restore.started, etc.
    resource_type TEXT,              -- backup, schedule, storage, container
    resource_id TEXT,

    -- Context
    request_id TEXT,                 -- Correlation ID
    scopes_used TEXT,                -- JSON array of scopes

    -- Outcome
    outcome TEXT NOT NULL,           -- success, failure, denied
    error_code TEXT,
    error_message TEXT,

    -- Metadata
    metadata JSON                    -- Action-specific details
);

CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_actor ON audit_log(actor_identity);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
```

**Action types:**

| Action | Description |
|--------|-------------|
| `backup.triggered` | Manual or scheduled backup started |
| `backup.completed` | Backup finished (success or failure) |
| `restore.started` | Restore operation initiated |
| `restore.completed` | Restore finished |
| `config.updated` | Schedule, storage, or retention changed |
| `auth.login` | User authenticated (v2) |
| `auth.logout` | User logged out (v2) |
| `auth.denied` | Access denied (v2) |

---

## Implementation Considerations

### Potential Friction Points

Based on design analysis, these areas require careful implementation:

#### Label Parsing

**Challenge**: Docker labels are strings. Complex values need escaping.

```yaml
# Problematic cases
- "baqup.snapshot.db.exclude=*.log,cache/*,temp*"  # Commas in value
- "baqup.snapshot.db.pre_exec=echo 'hello world'"  # Quotes in value
```

**Recommendation**:
- Document escaping rules clearly
- Use comma as list separator (most intuitive)
- For complex values, consider JSON encoding: `baqup.snapshot.db.exclude=["*.log","cache/*"]`

#### Agent Spawning

**Challenge**: Network attachment, volume permissions, cleanup on failure.

**Network attachment**:
```python
# Correct: Share network namespace with target
network_mode=f"container:{target_container_id}"

# Wrong: Attach to same network (different IP, may not work)
networks=["app_network"]
```

**Volume permissions**: Agent runs as root by default. If target data has specific ownership:
```python
# May need to match user
user=f"{target_uid}:{target_gid}"
```

**Cleanup on failure**: Agent container may not auto-remove if it crashes:
```python
try:
    container = docker.containers.run(..., remove=True)
except Exception:
    # Ensure cleanup
    try:
        container.remove(force=True)
    except:
        pass
```

#### Secret Versioning

**Challenge**: Detecting secret changes, migration of existing systems.

**Detection logic**:
```python
def should_increment_version(current_secret, previous_secret):
    if previous_secret is None:
        return False  # First backup, v1
    return current_secret != previous_secret
```

**Migration**: For existing backups without versioned secrets:
- Treat as v0 (unversioned)
- Require manual secret input for restore
- Document migration path

#### Redis Reliability

**Challenge**: Connection handling, reconnection, message loss.

**Recommendations**:
```python
class BaqupRedis:
    def __init__(self, url, retry_count=3, retry_delay=1):
        self.url = url
        self.retry_count = retry_count
        self.retry_delay = retry_delay
        self._connect()
    
    def _connect(self):
        self.redis = redis.from_url(self.url)
        # Verify connection
        self.redis.ping()
    
    def execute_with_retry(self, func, *args):
        for attempt in range(self.retry_count):
            try:
                return func(*args)
            except redis.ConnectionError:
                if attempt == self.retry_count - 1:
                    raise
                time.sleep(self.retry_delay * (attempt + 1))
                self._connect()
```

**Message loss**: Pub/Sub messages are fire-and-forget. For critical events:
- Use Redis Streams (roadmap) instead of Pub/Sub
- Or poll job status as backup to event subscription

#### Validation Pipeline

**Challenge**: Blocking vs async, timeout handling.

**Recommendation**: Validation should be synchronous within the pipeline:

```python
# Post-snapshot validation BLOCKS transfer
result = validate(backup, level="structure")
if not result.passed:
    alert(result)
    if should_retry(result):
        return retry_snapshot()
    return fail_job()

# Only proceed to transfer if validation passed
for storage in storages:
    transfer(backup, storage)
```

**Timeout handling**:
```python
def validate_with_timeout(backup, level, timeout=300):
    try:
        with timeout_context(timeout):
            return validate(backup, level)
    except TimeoutError:
        return ValidationResult(
            passed=False,
            error="Validation timed out"
        )
```

---

## Examples

### Minimal Configuration

The simplest possible baqup setup:

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret
    labels:
      - "baqup.enabled=true"
      - "baqup.snapshot.db.type=postgres"
      - "baqup.snapshot.db.username=postgres"
      - "baqup.snapshot.db.password_env=POSTGRES_PASSWORD"
```

**What happens with defaults**:
- Schedule: daily at 03:00 (controller default)
- Retention: 7 backups (controller default)
- Storage: whatever controller config specifies
- Validation: structure pre-transfer, auto post-transfer
- Compression: enabled

### Full Production Configuration

Complete example with all features:

```yaml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    labels:
      # Enable baqup
      - "baqup.enabled=true"
      
      # Schedules
      - "baqup.schedule.hourly.cron=0 * * * *"
      - "baqup.schedule.daily.cron=0 3 * * *"
      - "baqup.schedule.weekly.cron=0 4 * * 0"
      
      # Retention policies
      - "baqup.retention.critical.count=48"      # 2 days of hourly
      - "baqup.retention.standard.count=7"       # 1 week of daily
      - "baqup.retention.archive.count=4"        # 1 month of weekly
      
      # Storage backends
      - "baqup.storage.nas.type=rclone"
      - "baqup.storage.nas.remote=nas:/backups"
      
      - "baqup.storage.gdrive.type=rclone"
      - "baqup.storage.gdrive.remote=gdrive:backups"
      - "baqup.storage.gdrive.bandwidth=10M"
      
      - "baqup.storage.offsite.type=s3"
      - "baqup.storage.offsite.bucket=disaster-recovery"
      - "baqup.storage.offsite.region=us-east-1"
      - "baqup.storage.offsite.window.start=02:00"
      - "baqup.storage.offsite.window.end=06:00"
      - "baqup.storage.offsite.retry.count=5"
      - "baqup.storage.offsite.validation.post_transfer=none"
      
      # Critical database - hourly, triple redundancy
      - "baqup.snapshot.main-db.type=postgres"
      - "baqup.snapshot.main-db.username=postgres"
      - "baqup.snapshot.main-db.password_env=POSTGRES_PASSWORD"
      - "baqup.snapshot.main-db.databases=all"
      - "baqup.snapshot.main-db.compress=true"
      - "baqup.snapshot.main-db.schedule=hourly"
      - "baqup.snapshot.main-db.storage=nas,gdrive,offsite"
      - "baqup.snapshot.main-db.retention.default=critical"
      - "baqup.snapshot.main-db.retention.offsite=archive"
      - "baqup.snapshot.main-db.validation.post_snapshot=structure"
      - "baqup.snapshot.main-db.validation.scheduled=restore-test"
      - "baqup.snapshot.main-db.validation.scheduled.cron=0 5 * * 0"
      - "baqup.snapshot.main-db.validation.on_failure=alert"
      - "baqup.snapshot.main-db.restore.capture=true"
      - "baqup.snapshot.main-db.restore.model=pg-test"
      - "baqup.snapshot.main-db.restore.sources=nas,gdrive,offsite"
      
      # Model for validation restore-tests
      - "baqup.model.pg-test.image=postgres:15"
      - "baqup.model.pg-test.env.POSTGRES_USER=test"
      - "baqup.model.pg-test.env.POSTGRES_PASSWORD=test"
      - "baqup.model.pg-test.env.POSTGRES_DB=restore_test"
      - "baqup.model.pg-test.tmpfs=/var/lib/postgresql/data,/tmp"
      - "baqup.model.pg-test.memory=512m"
      - "baqup.model.pg-test.ephemeral=true"
```

### Multi-Service Application Stack

```yaml
services:
  app:
    image: myapp:latest
    labels:
      - "baqup.enabled=true"
      
      # Config files - daily backup
      - "baqup.snapshot.config.type=fs"
      - "baqup.snapshot.config.path=/config"
      - "baqup.snapshot.config.exclude=*.log,*.tmp,cache/*"
      - "baqup.snapshot.config.schedule=daily"
      - "baqup.snapshot.config.storage=nas,gdrive"
      
  postgres:
    image: postgres:15
    labels:
      - "baqup.enabled=true"
      
      # Database - hourly backup
      - "baqup.snapshot.db.type=postgres"
      - "baqup.snapshot.db.username=postgres"
      - "baqup.snapshot.db.password_env=POSTGRES_PASSWORD"
      - "baqup.snapshot.db.schedule=hourly"
      - "baqup.snapshot.db.storage=nas,gdrive,offsite"
      
  redis:
    image: redis:7
    labels:
      - "baqup.enabled=true"
      
      # Cache - daily backup (less critical)
      - "baqup.snapshot.cache.type=redis"
      - "baqup.snapshot.cache.schedule=daily"
      - "baqup.snapshot.cache.storage=nas"
      - "baqup.snapshot.cache.retention.default=minimal"
```

### Custom Agent Integration

Using a custom agent for unsupported database:

```yaml
# Controller config (config.yml)
agents:
  custom:
    oracle:
      image: mycompany/agent-oracle:2.1.0
      auth:
        username_env: REGISTRY_USER
        password_env: REGISTRY_PASS
```

```yaml
# docker-compose.yml
services:
  oracle:
    image: oracle/database:19c
    labels:
      - "baqup.enabled=true"
      - "baqup.snapshot.main.type=oracle"
      - "baqup.snapshot.main.connection_string_env=ORACLE_CONN"
      - "baqup.snapshot.main.schedule=daily"
      - "baqup.snapshot.main.storage=nas"
```

The custom agent must follow the agent contract (environment variables, output format, Redis reporting).

---

## Roadmap

### Version 1.0 (Current Scope)

Core functionality for production use:

- [x] Label-based autodiscovery
- [x] Core agents: postgres, mariadb, mongo, redis, sqlite, fs
- [x] Transfer agents: rclone, s3
- [x] Redis communication backbone
- [x] SQLite catalog
- [x] Stage-aware validation
- [x] Capture-based restore
- [x] Model definitions
- [x] Secret versioning
- [x] Multi-storage with per-storage retention
- [x] Transfer windows and bandwidth limits
- [x] Configurable conflict resolution

**v2 Preparation (included in v1.0):**
- [ ] `initiated_by` field in job metadata
- [ ] Audit log table (write-only, queryable in v2)
- [ ] Extensible config structure for auth/audit

### Version 1.1

Observability and control:

- [ ] Web UI (Traefik-style dashboard)
- [ ] REST API for state/control (with OAuth2 scope annotations)
- [ ] Prometheus metrics endpoint
- [ ] Healthcheck endpoints
- [ ] Webhook notifications
- [ ] Audit log REST API

### Version 2.0

Enterprise features:

**Authentication & Authorization:**
- [ ] OIDC/OAuth2 integration (Authentik, Keycloak, Okta, Azure AD)
- [ ] Role-based access control (RBAC) with OIDC claims
- [ ] API key authentication for automation
- [ ] OAuth2 scope enforcement on REST API

**Infrastructure:**
- [ ] Redis AUTH support
- [ ] Redis TLS
- [ ] Redis Cluster support

**Operations:**
- [ ] tmpfs with mount options (size, mode)
- [ ] Parallel agent execution
- [ ] Dependency ordering / backup groups
- [ ] Grandfather-father-son retention
- [ ] Backup encryption at rest
- [ ] Custom validation scripts
- [ ] Audit log UI and querying

### Future Considerations

Long-term possibilities:

- [ ] Kubernetes CRD provider (alternative to labels)
- [ ] Multi-instance controller coordination
- [ ] Sidecar health checks before backup
- [ ] Incremental backups (for supported types)
- [ ] Point-in-time recovery (WAL shipping for Postgres)
- [ ] Cloud-native secret management (Vault, AWS Secrets Manager)
- [ ] Backup deduplication
- [ ] Cross-site replication

---

## Appendices

### Appendix A: Controller Configuration Reference

Complete controller configuration file:

```yaml
# /etc/baqup/config.yml

redis:
  url: redis://localhost:6379
  db: 0
  prefix: baqup
  # Future: auth, tls, cluster

database:
  path: /var/lib/baqup/catalog.db

secrets:
  encryption_key_env: BAQUP_SECRET_KEY
  storage: default  # Store secrets in default storage

agents:
  registry: ghcr.io
  namespace: baqupio
  discovery: auto          # auto | explicit
  pull_policy: lazy        # lazy | startup | always
  refresh_interval: 86400  # Seconds between registry checks
  
  custom: {}               # Custom agent definitions

discovery:
  enabled: true
  label_prefix: baqup
  conflict_resolution: label-wins  # label-wins | config-wins | error | merge
  poll_interval: 60               # Seconds between Docker API polls

staging:
  path: /var/lib/baqup/staging
  cleanup_after_upload: true

defaults:
  schedules:
    daily:
      cron: "0 3 * * *"
  
  retention:
    default:
      count: 7
  
  snapshot:
    schedule: daily
    compress: true
    validation:
      post_snapshot: structure
      post_transfer: auto
      scheduled: checksum
      scheduled_cron: "0 4 * * 0"
      on_failure: alert
  
  transfer:
    retry:
      count: 3
      backoff: exponential
      initial_delay: 5
      max_delay: 300
    on_failure: continue
    verify: auto

notifications:
  on_failure:
    - ntfy://ntfy.sh/my-alerts
  on_success: false

logging:
  level: info             # debug | info | warning | error
  format: json            # json | text

# v1: Auth disabled, structure ready for v2
auth:
  enabled: false
  # v2: OIDC configuration
  # oidc:
  #   issuer: https://auth.example.com/application/o/baqup/
  #   client_id: baqup-controller
  #   client_secret_env: BAQUP_OIDC_CLIENT_SECRET
  #   scopes: [openid, profile, email, baqup]
  #   claims:
  #     roles: baqup_roles
  #     scopes: baqup_scopes
  # v2: API keys
  # api_keys:
  #   enabled: true
  #   storage: sqlite

# v1: Audit enabled by default
audit:
  enabled: true
  retention_days: 90
  # v2: External sink
  # sink:
  #   type: webhook
  #   url: https://siem.example.com/ingest
```

### Appendix B: Agent Environment Variables

Environment variables follow a structured prefix convention. See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for complete reference.

#### Variable Prefix Categories

| Prefix | Purpose |
|--------|---------|
| `BAQUP_JOB_*` | Job context (ID, timeout, retry settings) |
| `BAQUP_AGENT_*` | Agent behaviour (heartbeat, logging, parallelism) |
| `BAQUP_STAGING_*` | Staging area location |
| `BAQUP_ARTIFACT_*` | Backup file format (compression, checksums) |
| `BAQUP_BUS_*` | Communication (Redis URL, key prefix) |
| `BAQUP_SOURCE_*` | Data origin (path, host, excludes) |
| `BAQUP_TARGET_*` | Data destination (for restore operations) |
| `BAQUP_SECRET_*` | Sensitive values (never logged) |

#### Common Variables (All Agents)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_JOB_ID` | Yes | - | Unique job identifier (UUID) |
| `BAQUP_JOB_TIMEOUT` | No | 3600 | Maximum job duration in seconds |
| `BAQUP_STAGING_DIR` | Yes | - | Output directory for backup artifacts |
| `BAQUP_BUS_URL` | Yes | - | Redis connection URL |
| `BAQUP_ARTIFACT_COMPRESS` | No | true | Enable compression |
| `BAQUP_ARTIFACT_COMPRESS_ALGORITHM` | No | zstd | zstd, gzip, lz4, none |
| `BAQUP_ARTIFACT_CHECKSUM_ALGORITHM` | No | sha256 | sha256, sha512, blake3 |
| `BAQUP_AGENT_LOG_LEVEL` | No | info | trace, debug, info, warn, error, fatal |
| `BAQUP_AGENT_HEARTBEAT_INTERVAL` | No | 10 | Seconds between heartbeats |

#### Snapshot: postgres

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_HOST` | Yes | - | Database host |
| `BAQUP_SOURCE_PORT` | No | 5432 | Database port |
| `BAQUP_SECRET_DB_USER` | Yes | - | Database user |
| `BAQUP_SECRET_DB_PASSWORD` | Yes | - | Database password |
| `BAQUP_SOURCE_DATABASES` | No | all | Databases to dump (comma-separated or "all") |

#### Snapshot: mariadb / mysql

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_HOST` | Yes | - | Database host |
| `BAQUP_SOURCE_PORT` | No | 3306 | Database port |
| `BAQUP_SECRET_DB_USER` | Yes | - | Database user |
| `BAQUP_SECRET_DB_PASSWORD` | Yes | - | Database password |
| `BAQUP_SOURCE_DATABASES` | No | all | Databases to dump |

#### Snapshot: mongo

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_HOST` | Yes | - | Database host |
| `BAQUP_SOURCE_PORT` | No | 27017 | Database port |
| `BAQUP_SECRET_DB_USER` | No | - | Database user |
| `BAQUP_SECRET_DB_PASSWORD` | No | - | Database password |
| `BAQUP_SOURCE_AUTH_DB` | No | admin | Authentication database |

#### Snapshot: redis

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_HOST` | Yes | - | Redis host |
| `BAQUP_SOURCE_PORT` | No | 6379 | Redis port |
| `BAQUP_SECRET_DB_PASSWORD` | No | - | Redis password |

#### Snapshot: fs

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_PATH` | Yes | - | Path to backup |
| `BAQUP_SOURCE_EXCLUDE` | No | - | Exclude patterns (comma-separated) |
| `BAQUP_SOURCE_FOLLOW_SYMLINKS` | No | shallow | follow, skip, fail, shallow |

#### Transfer: rclone

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_PATH` | Yes | - | Local source directory |
| `BAQUP_TARGET_REMOTE` | Yes | - | Rclone remote path |
| `BAQUP_TARGET_CONFIG` | No | - | Path to rclone.conf |
| `BAQUP_TARGET_BANDWIDTH` | No | - | Bandwidth limit |

#### Transfer: s3

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BAQUP_SOURCE_PATH` | Yes | - | Local source directory |
| `BAQUP_TARGET_BUCKET` | Yes | - | S3 bucket name |
| `BAQUP_TARGET_PREFIX` | No | - | Key prefix |
| `BAQUP_TARGET_REGION` | No | us-east-1 | AWS region |
| `BAQUP_TARGET_ENDPOINT` | No | - | Custom endpoint (for S3-compatible) |
| `BAQUP_SECRET_S3_ACCESS_KEY` | Yes | - | AWS access key |
| `BAQUP_SECRET_S3_SECRET_KEY` | Yes | - | AWS secret key |

### Appendix C: Exit Codes

Agents use sysexits.h-inspired exit codes for precise error classification:

| Code | Meaning | Retriable | Controller Action |
|------|---------|-----------|-------------------|
| 0 | Success | - | Continue pipeline |
| 1 | General failure | Maybe | Evaluate error details |
| 64 | Usage/config error | No | Abort, alert (fix config) |
| 65 | Data error | No | Abort, alert |
| 69 | Resource unavailable | Yes | Retry with backoff |
| 70 | Internal error | Maybe | Evaluate, possibly retry |
| 73 | Can't create output | Yes | Retry (may be transient) |
| 74 | I/O error | Yes | Retry with backoff |
| 75 | Completed but unreported | Yes | Check staging for artifacts |
| 76 | Partial failure | Configurable | Per `partial_failure_mode` setting |

See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for full error taxonomy and categories.

### Appendix D: Redis Key Reference

#### Controller Keys

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `baqup:job:{id}` | Hash | 24h | Job details and configuration |
| `baqup:jobs:pending:{type}` | List | - | Job queue per agent type |
| `baqup:jobs:running` | Set | - | Currently executing jobs |
| `baqup:transfers:scheduled` | Sorted Set | - | Transfer jobs waiting for window |

#### Agent Keys

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `baqup:heartbeat:{agent_id}` | Hash | 30s | Agent liveness with intent |
| `baqup:progress:{job_id}` | Hash | - | Real-time progress (overwritten) |
| `baqup:status:{job_id}` | Hash | 24h | Final job status (persisted) |
| `baqup:events:{job_id}` | Stream | 24h | Per-job event log |
| `baqup:notify` | Pub/Sub | - | Live notifications |

See [AGENT-CONTRACT-SPEC.md](AGENT-CONTRACT-SPEC.md) for message structures and communication protocol details.

---

*Document version: 1.1.0-draft*
*Last updated: 2025-12-03*
*Authors: baqup design team*
