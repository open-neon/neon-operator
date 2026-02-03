
# PGXN

Compute nodes → Need pgxn (to communicate with pageserver, stream WAL, etc.)
Pageserver → Pure Rust, no pgxn needed
WAL Redo Processes → Ephemeral PostgreSQL instances spawned on-demand to replay WAL; they use pgxn extensions but are temporary, not part of the persistent pageserver deployment

# Storage Scrubber

The Storage Scrubber is a critical maintenance tool that manages and validates storage integrity across remote cloud storage (S3, Azure, GCS). Think of it as a cleanup and validation utility for tenant data stored in the cloud

FindGarbage	Scan cloud storage and identify unreferenced objects (deleted tenants that still have S3 files)
PurgeGarbage	Safely delete the garbage objects found by FindGarbage
ScanMetadata	Validate all metadata files for corruption or consistency issues
PageserverPhysicalGC	Remove old layer files and stale index files (with 3 safety modes: DryRun, IndicesOnly, Full)
TenantSnapshot	Download/backup an entire tenant's data locally for disaster recovery
FindLargeObjects	Identify large storage objects for cost optimization analysis
CronJob	Run periodic maintenance tasks automatically

storage_scrubber is a separate, independent binary that runs as a CLI tool for maintenance tasks

# compute_tools

compute_ctl - Postgres wrapper binary intended to run as a Docker entrypoint or systemd ExecStart option

Run as the container's entrypoint/init process - replacing or wrapping the direct postgres command
Manage the PostgreSQL process - starting, stopping, monitoring
Handle configuration - applying compute specifications, extensions, etc.
Monitor and report metrics - collecting compute node telemetry

Container:
├── PostgreSQL binary/executable
├── compute_ctl (from compute_tools)
└── Other dependencies
    ├── Extension server
    ├── Configuration utilities
    └── Monitoring tools

fast-import - Fast database import utility

Uses: extension_server module
Purpose: Prototype for dumping remote Postgres and uploading PGDATA to object storage

# storage controller 

The **Storage Controller** is Neon's centralized orchestration system that manages placement and lifecycle of tenant data across distributed storage infrastructure (pageservers and safekeepers).

## Core Responsibility

**Translate high-level requirements into infrastructure decisions:**

```
User Request: "Create tenant with 2 pageserver replicas"
    ↓
Storage Controller:
  - Selects which pageservers to use (based on placement policy, load, availability zones)
  - Assigns tenants to selected pageservers
  - Monitors and reconciles state continuously
  - Ensures data safety through generation numbers
```

## What It Manages

```
Storage Controller
├── Tenants
│   └── Shards (distributed across pageservers)
│       ├── Generation (monotonic version for safety)
│       ├── Placement intent (which node should be primary, which secondaries)
│       └── Observed state (what's actually running)
├── Pageservers
│   ├── Health status
│   ├── Capacity and placement decisions
│   └── Node location (availability zone)
└── Safekeepers
    ├── WAL archival assignments
    ├── Health status
    └── Scheduling policy
```

## 7 Main Responsibilities

1. **Shard Placement** - Decide which pageserver stores which tenant shard
2. **Auto-Scaling** - Split shards when they get too large
3. **Node Management** - Register, monitor, drain, and rebalance across nodes
4. **Reconciliation Loop** - Continuously ensure actual state matches desired state
5. **Data Safety** - Use generation numbers to prevent split-brain scenarios
6. **Compute Notifications** - Tell PostgreSQL endpoints where their data is stored
7. **Metadata Validation** - Monitor storage health, detect corruption

## How It Works

### Reconciliation Loop (Core Mechanism)

```
┌────────────────────────────────────┐
│ Continuously (every ~1 second):    │
│                                    │
│ 1. Check all tenant shards         │
│ 2. If intent != observed:          │
│    Spawn reconciler task           │
│ 3. Reconciler makes API calls to   │
│    pageservers to update config    │
│ 4. Notify compute nodes of changes │
│ 5. Update observed state           │
│ 6. Repeat                          │
└────────────────────────────────────┘
```

## Key Interactions

### With Pageservers

- Sends location config: "Attach tenant X with generation 5"
- Receives: heartbeats, metadata health reports
- Validates generation before deletion (prevents data loss)

### With Safekeepers

- Assigns which safekeepers should store WAL for each timeline
- Monitors their availability

### With Compute Nodes

- Notifies: "Tenant X is on pageserver-5"
- Allows compute to route requests to correct storage

## Safety Features

### Generation Numbers (Most Important)

- Each tenant shard has a version number (generation)
- Incremented before each attach operation
- Only ONE pageserver can attach a shard at a time
- Prevents split-brain if pageserver crashes

### Idempotent Reconciliation

- Safe to retry failed operations
- Handles transient failures automatically

### Persistent State

- Stores critical metadata in PostgreSQL database
- Survives controller restarts
- Enables recovery after failures

## Example: Tenant Migration

```
Initial State: Tenant on pageserver-3
    ↓
User command: Migrate to pageserver-5
    ↓
Storage Controller:
  1. Increment generation (5 → 6)
  2. Send detach to pageserver-3
  3. Send attach with gen=6 to pageserver-5
  4. Notify compute: tenant now on pageserver-5
  5. Validate pageserver-3 deleted correctly
    ↓
Final State: Tenant on pageserver-5
```

## Key Operations

| Operation | Purpose |
|-----------|---------|
| **Auto-Split** | When shard > threshold size, automatically split into multiple shards |
| **Manual Split** | Admin-requested shard split |
| **Migration** | Move shard to different pageserver |
| **Drain** | Gracefully remove shards before node shutdown |
| **Fill** | Restore shards after node comes back online |

## Core Modules

| Module | Lines | Purpose |
|--------|-------|---------|
| **service.rs** | 10,544 | Main orchestration engine |
| **tenant_shard.rs** | 3,220 | Shard state machine |
| **reconciler.rs** | 1,284 | Applies intent to infrastructure |
| **scheduler.rs** | 1,621 | Placement algorithm |
| **http.rs** | 2,710 | REST API endpoints |
| **persistence.rs** | 2,726 | Database persistence |
| **heartbeater.rs** | 448 | Health monitoring |

## Core Data Structures

### Tenant Shard State

```rust
pub struct TenantShard {
    tenant_shard_id: TenantShardId,
    policy: PlacementPolicy,           // High-level intent (Attached, Secondary, Detached)
    intent: IntentState,               // Low-level target state (which node attached/secondary)
    observed: ObservedState,           // What's actually running on infrastructure
    generation: Option<Generation>,    // Monotonic version for consistency
    sequence: Sequence,                // Coordination number for reconcilers
    config: TenantConfig,              // Opaque tenant settings
    splitting: SplitState,             // In-flight shard split state
    importing: TimelineImportState,    // Timeline import tracking
}
```

### Key State Distinctions

- **PlacementPolicy**: High-level intent (e.g., "attached to 1 node + 2 secondaries")
- **IntentState**: Concrete assignment (e.g., "attached to pageserver-5, secondaries on pageserver-3,7")
- **ObservedState**: What's actually configured on infrastructure right now
- **Generation**: Monotonically-increasing number; prevents split-brain scenarios

## APIs Exposed

### Tenant Management

- `POST /api/v1/tenant` - Create tenant
- `DELETE /api/v1/tenant/{tenant_id}` - Delete tenant
- `PATCH /api/v1/tenant/{tenant_id}/config` - Update tenant config
- `GET /api/v1/tenant/{tenant_id}/shards` - Describe tenant shards
- `GET /api/v1/tenant/{tenant_id}/locate` - Locate where tenant is stored
- `POST /api/v1/tenant/{tenant_id}/shard/migrate` - Migrate shard to different node
- `POST /api/v1/tenant/{tenant_id}/shard/split` - Split shard into multiple shards

### Timeline Management

- `POST /api/v1/tenant/{tenant_id}/timeline` - Create timeline
- `DELETE /api/v1/tenant/{tenant_id}/timeline/{timeline_id}` - Delete timeline
- `POST /api/v1/tenant/{tenant_id}/timeline/{timeline_id}/safekeeper-migrate` - Move timeline's safekeepers

### Node Management

- `POST /api/v1/nodes` - Register new node
- `DELETE /api/v1/nodes/{node_id}` - Remove node
- `POST /api/v1/nodes/{node_id}/drain` - Drain node for maintenance
- `POST /api/v1/nodes/{node_id}/fill` - Fill node after maintenance

### Pageserver Integration (Upcall APIs)

- `POST /api/v1/re-attach` - Pageserver calls on startup to get tenants to attach
- `POST /api/v1/validate` - Pageserver confirms generation ownership before deletion
- `POST /api/v1/metadata-health-update` - Metadata corruption reports

### Operational APIs

- `GET /api/v1/status` - Health status
- `GET /api/v1/metrics` - Prometheus metrics
- `POST /api/v1/reconcile-all` - Force reconciliation of all shards

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────┐
│                  Storage Controller                       │
│  (Orchestrates placement, reconciliation, coordination)   │
└──────────┬───────────────────────┬──────────────────────┘
           │                       │
    ┌──────▼──────┐        ┌──────▼──────┐
    │  Pageserver │        │  Safekeeper │
    │  (2 replicas)        │  (3 replicas)
    └──────┬──────┘        └──────┬──────┘
           │                      │
           │  Notify of change    │
           │  (AttachHook)        │ Store WAL
           │                      │
    ┌──────▼─────────────────────▼────┐
    │   Compute (PostgreSQL endpoints) │
    │   (receives attachment info)     │
    └─────────────────────────────────┘
```

## Production vs Development

| Aspect | Storage Controller | Local Control Plane |
|--------|-------------------|-------------------|
| **Scope** | Production distributed system | Local development environment |
| **Nodes** | Manages 10s-100s of physical pageservers | Manages local processes |
| **Leadership** | Multi-instance with leadership election | Single process |
| **Database** | Requires external PostgreSQL | Uses embedded/local data |
| **Persistence** | Durable state in database (critical) | Ephemeral, JSON file backups |
| **Failure Handling** | Handles node failures, network partitions | No failure handling |
| **API** | Full production HTTP API | Minimal test/dev API |
| **Scaling** | Designed for thousands of tenants | Designed for single test case |
| **Features** | Shard splits, migrations, rebalancing | Basic tenant/timeline CRUD |

## Key Files

- **Main Entry**: `storage_controller/src/main.rs` (645 lines)
- **Service Core**: `storage_controller/src/service.rs` (10,544 lines)
- **Tenant Shard State**: `storage_controller/src/tenant_shard.rs` (3,220 lines)
- **Reconciliation**: `storage_controller/src/reconciler.rs` (1,284 lines)
- **Placement**: `storage_controller/src/scheduler.rs` (1,621 lines)
- **Database**: `storage_controller/src/persistence.rs` (2,726 lines)
- **HTTP API**: `storage_controller/src/http.rs` (2,710 lines)

## Summary

The **Storage Controller is the brain of Neon's distributed storage system**—it makes all the decisions about where tenant data lives, ensures data safety, and automatically adapts to failures and changes.

Key characteristics:
- **Distributed**: Handles coordination across many nodes
- **Self-healing**: Continuously reconciles state and fixes discrepancies
- **Safe**: Generation numbers prevent data loss and split-brain scenarios
- **Scalable**: Designed to manage thousands of tenant shards
- **Observable**: Exposes metrics and health status for monitoring

- **Sharding**: A single timeline must not be sharded among multiple pageserver