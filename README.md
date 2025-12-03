<p align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="docs/content/assets/img/baqup.logo-dark.svg">
      <source media="(prefers-color-scheme: light)" srcset="docs/content/assets/img/baqup.logo.svg">
      <img alt="baqup" title="baqup" src="docs/content/assets/img/baqup.logo.svg" width="30%">
    </picture>
</p>

<p align="center">
  <strong>Container-native backup orchestration. Zero-config simplicity. Full control when needed.</strong>
</p>

> [!CAUTION]
> **Pre-alpha software.** This project is in early development. Everything below describes the intended design—not current functionality. APIs will change. Here be dragons.

> [!NOTE]
> **Standing on the shoulders of giants.** baqup is built on a fork of [Traefik](https://github.com/traefik/traefik), the excellent cloud-native reverse proxy. We're deeply grateful to the Traefik team and community for creating such a solid foundation for container-native infrastructure tooling. The provider discovery, label-based configuration, and dynamic reconfiguration patterns that make baqup possible all trace their lineage to Traefik's pioneering work.

<p align="center">
  <a href="#how-it-works">How It Works</a> •
  <a href="DESIGN.md">Design</a> •
  <a href="AGENT-CONTRACT-SPEC.md">Agent Spec</a> •
  <a href="GOVERNANCE.md">Governance</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-pre--alpha-orange.svg" alt="Status: Pre-alpha">
  <a href="https://github.com/baqupio/baqup/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-Fair%20Source-blue.svg" alt="License"></a>
  <a href="https://github.com/traefik/traefik"><img src="https://img.shields.io/badge/built%20on-Traefik-24a1c1.svg" alt="Built on Traefik"></a>
</p>

---

## What is baqup?

baqup is a container-native backup orchestration system. It discovers your workloads, backs them up automatically, and verifies they're actually restorable—not just present.

```yaml
# Add labels to your container. That's it.
services:
  postgres:
    image: postgres:16
    labels:
      - "baqup.enabled=true"
      - "baqup.snapshot.db.type=postgres"
      - "baqup.snapshot.db.password_env=POSTGRES_PASSWORD"
```

baqup handles the rest: agent selection, snapshot coordination, multi-destination shipping, retention policies, and restore verification.

## Why baqup?

**The problem**: Backup tooling for containers is either too simple (volume tarballs that corrupt databases) or too complex (enterprise solutions requiring dedicated teams).

**baqup's approach**:

| Traditional Backup | baqup |
|-------------------|-------|
| Configure each backup job manually | Auto-discovers workloads via labels |
| Generic volume snapshots | Application-aware agents (pg_dump, mysqldump, etc.) |
| Hope backups work | Validates restorability automatically |
| Single destination | Multi-destination in parallel |
| Per-node pricing | Free for small teams, flat rate for enterprises |

## Features

- **Zero-config discovery** — Labels on containers, automatic orchestration
- **Application-aware agents** — Purpose-built for PostgreSQL, MariaDB, filesystem, and more
- **Multi-destination** — Single snapshot to S3, B2, local storage, or any rclone target
- **Restore verification** — Proves backups are restorable, not just present
- **Encryption at rest** — AES-256-GCM or ChaCha20-Poly1305
- **Checksums everywhere** — SHA-256/SHA-512/BLAKE3 on all artifacts
- **Clean failure modes** — Partial failures reported clearly, successful artifacts preserved

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                         CONTROLLER                               │
│  Discovers workloads • Schedules jobs • Spawns agents           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
           ┌───────────────────┼───────────────────┐
           │                   │                   │
           ▼                   ▼                   ▼
    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
    │ agent-postgres│    │ agent-mariadb│    │  agent-fs   │
    │             │     │             │     │             │
    │ pg_dump     │     │ mysqldump   │     │ tar + zstd  │
    │ streaming   │     │ consistent  │     │ incremental │
    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │   agent-rclone      │
                    │                     │
                    │  S3 / B2 / GCS /    │
                    │  Azure / Local      │
                    └─────────────────────┘
```

**Controller**: Watches your orchestrator (Docker, Swarm, Kubernetes), discovers labelled workloads, schedules backup jobs, spawns short-lived agent containers.

**Agents**: Stateless containers that execute a single backup or restore operation. Each agent type understands its data source—`agent-postgres` uses `pg_dump` with proper transaction isolation, not naive volume copies.

**Manifests**: Every backup produces a manifest with checksums. No manifest = incomplete backup = discarded automatically.

## Quickstart

> [!NOTE]
> **Aspirational.** This shows the intended UX. Implementation in progress.

### Docker Compose

```yaml
services:
  baqup:
    image: ghcr.io/baqupio/baqup:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./backups:/backups
    environment:
      BAQUP_STAGING_DIR: /backups

  postgres:
    image: postgres:16
    labels:
      - "baqup.enabled=true"
      - "baqup.snapshot.db.type=postgres"
      - "baqup.snapshot.db.username=postgres"
      - "baqup.snapshot.db.password_env=POSTGRES_PASSWORD"
      - "baqup.snapshot.db.schedule=daily"
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
```

```bash
docker compose up -d
```

That's it. baqup discovers the labelled container, uses controller defaults for schedule (`daily` = 3 AM) and retention (7 backups), and handles the rest.

### Manual Backup

```bash
# Trigger immediate backup
docker exec baqup baqup backup postgres

# List backups
docker exec baqup baqup list

# Restore
docker exec baqup baqup restore postgres --target postgres-restore
```

## Agents

| Agent | Data Source | Method |
|-------|-------------|--------|
| `agent-postgres` | PostgreSQL | `pg_dump` with consistent snapshots |
| `agent-mariadb` | MariaDB/MySQL | `mysqldump` or `mariadb-dump` |
| `agent-fs` | Filesystem | Streaming tar with zstd compression |
| `agent-rclone` | Remote storage | Any rclone-supported backend |

Each agent implements the [Agent Contract Specification](AGENT-CONTRACT-SPEC.md)—a detailed protocol for lifecycle management, error handling, and artifact production.

## Configuration

baqup follows a hierarchy: **container labels → controller config → CLI flags**.

### Container Labels

```yaml
labels:
  # Global
  - "baqup.enabled=true"                                    # Enable baqup for this container

  # Schedules (reusable, named)
  - "baqup.schedule.hourly.cron=0 * * * *"
  - "baqup.schedule.daily.cron=0 3 * * *"

  # Retention policies (reusable, named)
  - "baqup.retention.critical.count=48"                     # Keep 48 backups
  - "baqup.retention.standard.count=7"                      # Keep 7 backups

  # Storage backends (reusable, named)
  - "baqup.storage.nas.type=rclone"
  - "baqup.storage.nas.remote=nas:/backups"

  # Snapshots (what to back up)
  - "baqup.snapshot.db.type=postgres"                       # Agent type
  - "baqup.snapshot.db.username=postgres"
  - "baqup.snapshot.db.password_env=POSTGRES_PASSWORD"      # Reference container env var
  - "baqup.snapshot.db.schedule=daily"                      # Reference named schedule
  - "baqup.snapshot.db.storage=nas"                         # Reference named storage
  - "baqup.snapshot.db.retention.default=standard"          # Reference named retention
```

### Controller Configuration

```yaml
# /etc/baqup/config.yml
redis:
  url: redis://localhost:6379

staging:
  path: /var/lib/baqup/staging

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
```

Labels on containers override controller defaults. See [DESIGN.md](DESIGN.md) for the complete configuration reference.

## Supported Platforms

| Platform | Status |
|----------|--------|
| Docker (standalone) | 🚧 In development |
| Docker Swarm | 🚧 Planned |
| Kubernetes | 🚧 Planned |
| Podman | 🚧 Planned |

## Documentation

Documentation is being developed alongside the code. For now:

- **[Design Document](DESIGN.md)** — Comprehensive architecture, label schema, and rationale
- **[Agent Contract Spec](AGENT-CONTRACT-SPEC.md)** — Technical specification for agent implementation

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Support

This is pre-alpha software. Expect rough edges.

- **Issues**: [GitHub Issues](https://github.com/baqupio/baqup/issues)
- **Discussions**: [GitHub Discussions](https://github.com/baqupio/baqup/discussions)

## Repositories

Planned repository structure:

| Repository | Purpose |
|------------|---------|
| `baqup` | Controller and CLI |
| `agent-postgres` | PostgreSQL backup agent |
| `agent-mariadb` | MariaDB/MySQL backup agent |
| `agent-fs` | Filesystem backup agent |
| `agent-rclone` | Remote storage shipping agent |

## Credits

baqup exists because of the remarkable work done by the [Traefik](https://traefik.io) team.

The core discovery and orchestration engine is forked from Traefik, repurposing its battle-tested provider system (Docker, Swarm, Kubernetes) from routing HTTP traffic to orchestrating backup agents. Traefik's architecture—dynamic configuration via labels, automatic service discovery, clean provider abstractions—turns out to be exactly what backup orchestration needs.

We encourage you to check out [Traefik](https://github.com/traefik/traefik) if you need a reverse proxy. It's excellent software built by excellent people.

---

<p align="center">
  <sub>Built with care for homelabbers, startups, and enterprises who actually want their backups to work.</sub>
</p>