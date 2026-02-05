# Neon Scaling Guide

Complete guide to scaling strategy in Neon's distributed PostgreSQL architecture.

## Table of Contents

1. [Quick Summary](#quick-summary)
2. [Pageserver Scaling](#pageserver-scaling)
3. [Compute Scaling](#compute-scaling)
4. [Storage Monitoring](#storage-monitoring)
5. [Scaling Strategies](#scaling-strategies)
6. [Limitations](#limitations)
7. [Decision Tree](#decision-tree)

---

## Quick Summary

```
┌─────────────────────────────────────────────────────────────┐
│ SCALING IN NEON: TWO INDEPENDENT SYSTEMS                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ PAGESERVER SCALING (Infrastructure)                        │
│  • Triggered: Data size growth                             │
│  • Managed: Automatic (Storage Controller)                 │
│  • Mechanism: Auto-split when threshold exceeded           │
│  • Scaling factor: Data size in bytes                       │
│  • Scaling trigger: split_threshold (e.g., 10GB/shard)    │
│                                                             │
│ COMPUTE SCALING (User-facing)                              │
│  • Triggered: Connection/workload demand                   │
│  • Managed: User/Application-driven                        │
│  • Mechanism: Create new endpoints or promote replicas     │
│  • Scaling factor: Connections, QPS, workload              │
│  • Scaling trigger: User decision or autoscaling rules     │
│                                                             │
│ They scale INDEPENDENTLY:                                  │
│  ✓ Pageserver scales automatically by data                 │
│  ✓ Compute scales manually by user                         │
│  ✓ One doesn't affect the other                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Pageserver Scaling

### Overview

Pageserver scaling is **automatic and data-driven**. The Storage Controller monitors tenant data sizes and automatically splits shards when thresholds are exceeded.

### How It Works

#### 1. **Auto-Split Triggers**

Storage Controller runs `autosplit_tenants()` every 20 seconds:

```
Two independent triggers:

INITIAL SPLIT (Unsharded → Sharded)
  Condition: Single-shard tenant > initial_split_threshold
  Action: Split into initial_split_shards
  Example: 150GB unsharded tenant → split into 4 shards

SIZE-BASED SPLIT (Multi-shard → More Shards)
  Condition: per_shard_size > split_threshold
  Action: Split into enough shards so each ≤ threshold
  Example: 100GB on 4 shards (25GB each) → split into 16
           (because 25GB > 10GB threshold)
```

#### 2. **Configuration Parameters**

```bash
# Size-based splitting
--split_threshold <BYTES>
  When per-shard size exceeds this, trigger split
  Default: None (disabled)
  Example: --split_threshold 10737418240  (10GB)
  Recommendation: 10GB-50GB depending on workload

--max_split_shards <COUNT>
  Maximum shards via auto-split (safety limit)
  Default: 16
  Recommendation: 16-32 for most deployments

# Initial splitting (unsharded → sharded)
--initial_split_threshold <BYTES>
  When unsharded tenant exceeds this, split it
  Default: None (disabled)
  Example: --initial_split_threshold 107374182400  (100GB)
  Recommendation: 100GB+ for better throughput

--initial_split_shards <COUNT>
  Target shards for initial split
  Default: 2
  Recommendation: 4-8 for ingestion-heavy workloads
```

### Scaling Timeline Example

```
Day 1: New tenant created
  Size: 5GB
  Shards: 1 shard (0/1)
  Status: Auto-split disabled (below threshold)

Week 1: User adds data
  Size: 50GB
  Shards: 1 shard (0/1)
  Status: Auto-split disabled (below 100GB threshold)

Month 1: Steady growth
  Size: 150GB
  Shards: 1 shard (0/1)
  Action: INITIAL SPLIT TRIGGERED! (150GB > 100GB)
    ↓
  New shards: 4 shards (0-3/4)
    ↓
  Distribution:
    Shard 0/4: 37.5GB on Pageserver A
    Shard 1/4: 37.5GB on Pageserver B
    Shard 2/4: 37.5GB on Pageserver C
    Shard 3/4: 37.5GB on Pageserver D

Month 3: Continued growth
  Size: 160GB on 4 shards
  Per-shard: 40GB each
  Status: OK (40GB < 50GB threshold)

Month 6: Heavy growth
  Size: 250GB on 4 shards
  Per-shard: 62.5GB each
  Action: SIZE-BASED SPLIT TRIGGERED! (62.5GB > 50GB)
    ↓
  Calculation: 250GB / 50GB = 5 shards needed
             Round up to power of 2 → 8 shards
    ↓
  New shards: 8 shards (0-7/8)
    ↓
  Distribution: 250GB / 8 = 31.25GB per shard ✓
```

### When to Adjust Thresholds

```
INCREASE threshold if:
  • Pageservers have plenty of disk space
  • You want fewer, larger shards
  • You want less frequent splits

DECREASE threshold if:
  • Want smaller shards for parallel efficiency
  • Expect rapid growth
  • Have many pageservers to utilize

DISABLE autosplit if:
  • Set split_threshold = 0
  • Manual management preferred
  • Testing specific configurations
```

### Important Notes

```
✓ Auto-split is non-blocking
  • Storage Controller continues normal operations
  • Only one shard splits per 20-second cycle
  • Large backlogs take time to process

✓ Split is safe
  • Persisted to database before execution
  • Idempotent - safe to retry
  • No data loss

✗ Cannot merge shards
  • If data shrinks, shards remain
  • Slight overhead but fully functional
  • Workaround: delete and recreate tenant

✗ Based on DATA SIZE only
  • NOT CPU usage
  • NOT memory usage
  • NOT request rate
  • This is architectural design choice
```

---

## Compute Scaling

### Overview

Compute scaling is **user-driven and workload-based**. Users create additional compute endpoints to handle more connections, workload distribution, or failover scenarios.

### Compute Modes

Neon supports three compute modes per timeline:

```
PRIMARY (read-write)
  • One per timeline
  • Accepts INSERT/UPDATE/DELETE
  • Generates WAL records
  • Sent to safekeepers
  • Only this mode can write

REPLICA (read-only, hot standby)
  • Multiple allowed per timeline
  • Read-only (SELECT only)
  • Follows primary's WAL in real-time
  • Can be promoted to primary
  • Can be used for load balancing reads

STATIC (read-only, frozen at LSN)
  • Multiple allowed at different LSNs
  • Read-only (SELECT only)
  • Pinned at specific LSN
  • Does NOT follow primary updates
  • Useful for point-in-time recovery/analysis
```

### Scaling Strategies

#### 1. **Vertical Scaling (Increase Resources)**

```
Strategy: Increase CPU/memory of single compute

When to use:
  • Limited connections but high CPU usage per connection
  • Complex queries needing more memory
  • Temporary spike handling

Approach:
  compute_ctl \
    --pgdata /var/db/postgres \
    --config config.json \
    --resources '{"cpu": 8, "memory": 32GB}'

Limitations:
  • Eventually hits hardware limits
  • Can't exceed physical capacity
  • No automatic distribution
  • Single point of failure
```

#### 2. **Horizontal Scaling (Multiple Replicas)**

```
Strategy: Add read-only replica computes

When to use:
  • High read query volume
  • Load balancing reads across instances
  • HA/failover capability needed
  • Separate workload types

Architecture:
  ┌─ Compute 1 (Primary)
  │   └─ INSERT/UPDATE/DELETE
  │
  ├─ Compute 2 (Replica)
  │   └─ SELECT queries (50% of read load)
  │
  └─ Compute 3 (Replica)
      └─ SELECT queries (50% of read load)

All connect to: Same pageserver shards

Implementation:
  1. Create replica endpoint via Control Plane
  2. Configure load balancer (nginx, HAProxy, etc.)
  3. Route reads to replicas, writes to primary
  4. Add more replicas as needed

Scaling:
  Add as many read replicas as needed
  Each replica independently handles reads
  No coordination required
```

#### 3. **Connection Pooling**

```
Strategy: Multiplex connections through pooler

When to use:
  • Many clients, limited compute resources
  • Reduce connection overhead
  • Cost optimization

Tool: PgBouncer

Architecture:
  Clients → PgBouncer → Compute (1 backend per pool)

  Multiplexes:
    100 client connections → 10 backend connections
    Reduces resource usage

Configuration:
  pgbouncer.ini:
    [databases]
    dbname = host=compute:5432

    [pgbouncer]
    pool_mode = transaction
    max_client_conn = 10000
    default_pool_size = 25
```

#### 4. **Branching for Isolation**

```
Strategy: Create compute on different timeline/branch

When to use:
  • Development/testing without affecting production
  • Different workload types (analytics vs transactional)
  • Isolated environment

Architecture:
  Production Timeline (main)
    └─ Compute Primary (production workload)

  Development Timeline (branch from main)
    └─ Compute Primary (dev workload)

  Analytics Timeline (branch from main)
    └─ Static Compute at specific LSN (reporting)

Each timeline:
  • Independent data (branched at point in time)
  • Own primary compute
  • Separate from other timelines
```

#### 5. **Autoscaling (Kubernetes)**

```
Strategy: Automatic scaling based on metrics

When to use:
  • Cloud-native deployments
  • Variable workload patterns
  • Need for automatic elasticity

Implementation (Kubernetes HPA):
  apiVersion: autoscaling/v2
  kind: HorizontalPodAutoscaler
  metadata:
    name: compute-autoscaler
  spec:
    scaleTargetRef:
      apiVersion: apps/v1
      kind: Deployment
      name: compute-replicas
    minReplicas: 1
    maxReplicas: 10
    metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: postgresql_connections
        target:
          type: AverageValue
          averageValue: "1000"

Behavior:
  • CPU > 70% → Add replicas
  • CPU < 30% → Remove replicas
  • Connections > 1000/replica → Add replicas
```

### Multi-Compute Architecture Example

```
┌───────────────────────────────────────────────────────┐
│ Load Balancer (pgBouncer or nginx)                    │
└────────┬──────────────────────────────────────────────┘
         │
    ┌────┴────┬──────────┬──────────┬──────────┐
    │          │          │          │          │
    ▼          ▼          ▼          ▼          ▼
Compute 1   Compute 2   Compute 3   Compute 4  Compute 5
PRIMARY     REPLICA     REPLICA     REPLICA    STATIC@LSN
(RW)        (RO)        (RO)        (RO)       (RO)
            │           │           │          │
            └────────────┴───────────┴──────────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
          Shard 0/4  Shard 1/4  Shard 2/4  Shard 3/4
          (PS-A)     (PS-B)     (PS-C)     (PS-D)

Write pattern:
  INSERT/UPDATE/DELETE → Compute 1 (Primary only)
    ↓
  Pageserver Shards 0-3
    ↓
  WAL to Safekeepers
    ↓
  Replicated to Compute 2-4

Read pattern:
  SELECT → Load balancer routes to:
    • Compute 2-4 (replicas, 80%)
    • Compute 1 (primary, 20%, if needed)
```

---

## Storage Monitoring

### How Storage Controller Knows Tenant Size

Storage Controller queries pageservers every 20 seconds:

```
Timeline of Size Tracking:

1. User writes data
   ↓
2. Pageserver receives pages
   ↓
3. Pageserver updates metrics:
   - current_logical_size_gauge
   - resident_physical_size_gauge
   - remote_physical_size_gauge
   ↓
4. Storage Controller heartbeat:
   GET /top_tenants endpoint on pageserver
   ↓
5. Pageserver responds:
   {
     "max_logical_size": 150000000,  (150GB)
     "resident_size": 25000000,
     "physical_size": 30000000
   }
   ↓
6. Storage Controller stores in database:
   UPDATE tenant_shards
   SET max_logical_size = 150000000
   WHERE tenant_shard_id = 'abc123/0001'
   ↓
7. Auto-split decision:
   if 150GB > initial_split_threshold (100GB):
     trigger split
```

### Key Metrics

```
Essential Metrics (for scaling):
  • max_logical_size: Largest timeline size across all shards
  • disk_usage_bytes: Physical disk space used
  • free_space_bytes: Available disk space
  • shard_count: Number of shards on pageserver

Calculation:
  per_shard_size = max_logical_size / shard_count

  If per_shard_size > split_threshold:
    → Split needed

Monitoring:
  • Query pageserver /utilization endpoint
  • Check storage_controller database
  • Monitor disk space trends
  • Plan capacity based on growth rate
```

---

## Scaling Strategies

### For Different Use Cases

#### High Write Throughput

```
Problem: Heavy INSERT/UPDATE/DELETE workload

Solution:
  1. Pageserver scaling (automatic)
     • Auto-split when data size grows
     • Distributes storage load

  2. Connection optimization
     • Use connection pooling
     • Tune batch sizes
     • Optimize transaction rates

  3. Configuration:
     --initial_split_threshold 50GB    (split early)
     --initial_split_shards 8          (many shards)
     --split_threshold 5GB             (aggressive)
```

#### High Read Throughput

```
Problem: Heavy SELECT workload

Solution:
  1. Add read replicas
     • Create 2-10 read-only replicas
     • Load balance reads across them
     • Primary handles writes

  2. Caching strategy
     • Client-side caching
     • Redis cache layer
     • Optimize query patterns

  3. Compute configuration:
     • Multiple replica endpoints
     • Load balancer distribution
     • Connection pooling
```

#### Rapid Data Growth

```
Problem: Data growing very fast (ingestion phase)

Solution:
  1. Aggressive pageserver splitting
     --initial_split_threshold 50GB    (low threshold)
     --initial_split_shards 8-16       (many shards)
     --split_threshold 5-10GB          (frequent splits)

  2. Pre-provision pageservers
     • Add spare capacity
     • Plan for exponential growth

  3. Monitor closely
     • Track shard count trend
     • Monitor disk usage
     • Plan ahead for capacity
```

#### Cost Optimization

```
Problem: Want to minimize infrastructure cost

Solution:
  1. Pageserver optimization
     --split_threshold 50GB      (fewer shards)
     --max_split_shards 8        (limit shards)
     Higher threshold → fewer pageservers needed

  2. Compute optimization
     • Single primary compute
     • Add replicas only if needed
     • Use connection pooling
     • Size appropriately (not over-provision)

  3. Storage optimization
     • Monitor unused databases
     • Archive old data
     • Clean up branches
```

#### High Availability

```
Problem: Need failover capability

Solution:
  1. Pageserver redundancy
     • Multiple pageservers (automatic sharding)
     • Safekeepers for WAL (automatic)
     • Storage Controller leadership election

  2. Compute redundancy
     • Multiple replica computes
     • Can promote replica → primary
     • Load balancer health checks

  3. Monitoring
     • Health checks on all components
     • Alerting on failures
     • Automatic or manual failover
```

---

## Limitations

### Pageserver Scaling

```
✗ CANNOT merge shards
  Problem: Data shrinks, extra shards remain
  Solution: Delete and recreate tenant if critical

✗ Based on DATA SIZE only
  Problem: CPU hotspot not auto-handled
  Solution: Optimize queries, manual intervention

✗ Scaling is per-tenant
  Problem: No cross-tenant load balancing
  Solution: Federation at application level
```

### Compute Scaling

```
✗ NO multiple write replicas
  Problem: Can't have 2+ primaries
  Reason: PostgreSQL limitation, conflict resolution
  Solution: Use application-level federation

✗ Replicas are read-only
  Problem: Can't distribute write load
  Solution: Accept write concentration on primary

✗ NO automatic write distribution
  Problem: Single primary handles all writes
  Solution: Optimize primary, use connection pooling

✗ NO automatic failover
  Problem: Manual promotion of replica needed
  Solution: Implement automated tools/monitoring
```

### Architecture Constraints

```
✗ Sharding is data-based, not load-based
  Problem: CPU hotspots not auto-distributed
  Reason: Deterministic key mapping required
  Solution: Optimize queries, manual sharding

✗ Compute and pageserver scales independently
  Problem: Can't use metrics from one to scale other
  Reason: They address different concerns
  Solution: Monitor and scale each independently
```

---

## Decision Tree

### When to Add Pageservers

```
Is disk usage > 80%?
  YES → Add pageserver (auto-split will trigger)
  NO → Go to next question

Is shard count approaching max_split_shards?
  YES → Add pageserver (to accommodate new shards)
  NO → Go to next question

Forecasted data growth in next 3 months?
  YES → Add pageserver proactively
  NO → Monitor, no action needed yet

RESULT: You probably need more pageservers
```

### When to Add Compute

```
Are reads per second high (>1000 QPS)?
  YES → Add read replicas
  NO → Go to next question

Connection errors or timeouts?
  YES → Add compute or connection pooling
  NO → Go to next question

CPU usage > 80%?
  YES → Increase CPU or optimize queries
  NO → Go to next question

RESULT: You probably don't need more compute
```

### Scaling Priority

```
PRIORITY 1: Pageserver Capacity
  Monitor: Disk usage %
  Trigger: > 80%
  Action: Add pageserver

PRIORITY 2: Compute Connections
  Monitor: Active connections, timeouts
  Trigger: Errors, timeout threshold exceeded
  Action: Add compute or enable pooling

PRIORITY 3: Query Performance
  Monitor: Query latency, slow queries
  Trigger: > SLA threshold
  Action: Optimize, add replicas, increase resources

PRIORITY 4: Redundancy
  Monitor: Failure scenarios
  Trigger: Planned testing
  Action: Add replicas, implement failover
```

---

## Checklists

### Before Scaling Pageservers

```
□ Verified disk usage via monitoring dashboard
□ Confirmed threshold exceeded (>80%)
□ Checked shard count trend
□ Reviewed storage controller logs
□ Planned capacity for next 6 months
□ Have available hardware/infrastructure
□ Tested auto-split in staging
```

### Before Adding Compute

```
□ Verified connection metrics
□ Checked CPU/memory usage of existing compute
□ Reviewed slow query logs
□ Considered connection pooling first
□ Have capacity for new instance
□ Configured load balancer
□ Tested failover scenario
```

### Storage Controller Configuration

```
□ Set split_threshold (e.g., 10GB)
□ Set max_split_shards (e.g., 16-32)
□ Set initial_split_threshold (e.g., 100GB)
□ Set initial_split_shards (e.g., 4-8)
□ Configure heartbeat_interval (typically 5s)
□ Set up monitoring/alerts
□ Document thresholds and reasoning
□ Plan review schedule (quarterly)
```

---

## Summary

```
┌─────────────────────────────────────────────────────┐
│ SCALING SUMMARY                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│ PAGESERVER (Automatic)                              │
│  ✓ Triggered: Data size > threshold                │
│  ✓ Managed: Storage Controller                     │
│  ✓ Scaling: Every 20 seconds check                 │
│  ✓ Action: Auto-split shards                       │
│  ✓ Config: split_threshold, max_split_shards       │
│                                                     │
│ COMPUTE (Manual)                                    │
│  ✓ Triggered: User/app demand                      │
│  ✓ Managed: User/ops team                          │
│  ✓ Scaling: Add replicas, branches                 │
│  ✓ Action: Create endpoints, promote replicas      │
│  ✓ Config: Endpoint creation, load balancing       │
│                                                     │
│ KEY INSIGHTS                                        │
│  • Pageserver scales ≠ Compute scales               │
│  • Data size drives infrastructure                  │
│  • Connection load drives compute                  │
│  • Monitor separately, act independently           │
│  • NO multiple write replicas (PostgreSQL limit)   │
│  • Multiple read replicas for HA/load balancing    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## References

- Neon RFCs:
  - `docs/rfcs/031-sharding-static.md` - Sharding architecture
  - `docs/rfcs/032-shard-splitting.md` - Shard split design
  - `docs/rfcs/041-sharded-ingest.md` - WAL handling with shards
