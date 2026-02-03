# Neon Sharding Guide

Complete guide to understanding how data sharding works in Neon's distributed storage architecture.

## Table of Contents

1. [What is a Shard?](#what-is-a-shard)
2. [Tenant vs Timeline vs Shard](#tenant-vs-timeline-vs-shard)
3. [When Shards are Created](#when-shards-are-created)
4. [Who Creates Shards](#who-creates-shards)
5. [How Routing Works](#how-routing-works)
6. [Transaction Consistency Across Shards](#transaction-consistency-across-shards)
7. [Shard Merging](#shard-merging)

---

## What is a Shard?

A **shard** is a horizontal partition of a tenant's data stored on a specific pageserver.

### Simple Analogy

```
Tenant = "Restaurant Chain" database (e.g., 1TB of data)

If sharded into 4 pieces:
  ├── Shard 0/4 = Data for restaurants A-D (250GB) on Pageserver A
  ├── Shard 1/4 = Data for restaurants E-H (250GB) on Pageserver B
  ├── Shard 2/4 = Data for restaurants I-L (250GB) on Pageserver C
  └── Shard 3/4 = Data for restaurants M-P (250GB) on Pageserver D

Each shard can be stored on a different pageserver
and processed independently.
```

### Technical Definition

A **tenant shard** (`TenantShardId`) is identified by:
- **Tenant ID** - which tenant it belongs to
- **Shard number** - which piece (0, 1, 2, etc.)
- **Shard count** - total pieces (4 in above example)

```rust
pub struct TenantShardId {
    tenant_id: TenantId,
    shard_index: ShardIndex,
}

// Example: TenantShardId {
//   tenant_id: "abc123",
//   shard_index: ShardIndex { number: 1, count: 4 }
// }
```

### Why Shard?

```
Without sharding:
  ┌──────────────────────────────┐
  │ Pageserver A                 │
  │ - 1TB of data                │
  │ - 1000s of requests/sec       │
  │ - Single disk I/O bottleneck  │
  └──────────────────────────────┘

With sharding:
  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
  │ PS-A    │  │ PS-B    │  │ PS-C    │  │ PS-D    │
  │ 250GB   │  │ 250GB   │  │ 250GB   │  │ 250GB   │
  │ 250 req │  │ 250 req │  │ 250 req │  │ 250 req │
  │ /sec    │  │ /sec    │  │ /sec    │  │ /sec    │
  └─────────┘  └─────────┘  └─────────┘  └─────────┘

Benefits:
  ✓ Distributes storage load (1TB → 250GB each)
  ✓ Distributes I/O load (1000 req → 250 each)
  ✓ Enables parallelism in queries
  ✓ Better cache locality
  ✓ Faster layer compaction per shard
```

---

## Tenant vs Timeline vs Shard

### Correct Model

```
Tenant "user123" (4 shards - THIS is what's sharded)
│
├─ Shard 0/4 (on Pageserver A)
│   ├─ Timeline "main" (logical, not sharded)
│   │   └─ Data keys: {0, 4, 8, 12, ...}  (only its portion)
│   ├─ Timeline "dev" (logical, not sharded)
│   │   └─ Data keys: {0, 4, 8, 12, ...}  (only its portion)
│   └─ Timeline "test" (logical, not sharded)
│       └─ Data keys: {0, 4, 8, 12, ...}  (only its portion)
│
├─ Shard 1/4 (on Pageserver B)
│   ├─ Timeline "main" (logical, not sharded)
│   │   └─ Data keys: {1, 5, 9, 13, ...}
│   ├─ Timeline "dev" (logical, not sharded)
│   │   └─ Data keys: {1, 5, 9, 13, ...}
│   └─ Timeline "test" (logical, not sharded)
│       └─ Data keys: {1, 5, 9, 13, ...}
│
... (Shard 2/4 and 3/4 similar)
```

### Key Distinctions

| Concept | Definition | Sharded? |
|---------|-----------|----------|
| **Tenant** | Logical database/project | ✓ YES (into shards) |
| **Timeline** | Database branch (version history) | ✗ NO (logical unit) |
| **Shard** | Physical data partition | N/A (it IS the partition) |

### Timeline is NOT Sharded

A timeline is:
- **A sequence of database states** at successive LSNs
- **A complete version of the database** at each point in time
- **Independent from sharding** - timelines are orthogonal to shards

```
Timeline "main":
  LSN 0 → LSN 1 → LSN 2 → LSN 3 → ... → LSN 5000
  (complete logical database at each LSN)

Timeline "dev" (branched from "main" at LSN 5000):
  LSN 5000 → LSN 5001b → LSN 5002b → LSN 5003b → ...
  (new branch, complete database at each LSN)

The timeline defines:
  - Which WAL records apply
  - How the database evolves
  - The branching hierarchy

Sharding is independent:
  - Data from timeline "main" at LSN 5000 is spread across shards 0-3
  - Data from timeline "dev" at LSN 5002b is spread across shards 0-3
```

### All Shards Handle Same Timelines

Each shard in a tenant manages **all timelines** of that tenant:

```
Timeline "main" exists on:
  ✓ Shard 0/4 (piece of main)
  ✓ Shard 1/4 (piece of main)
  ✓ Shard 2/4 (piece of main)
  ✓ Shard 3/4 (piece of main)

Timeline "dev" exists on:
  ✓ Shard 0/4 (piece of dev)
  ✓ Shard 1/4 (piece of dev)
  ✓ Shard 2/4 (piece of dev)
  ✓ Shard 3/4 (piece of dev)
```

---

## When Shards are Created

### At Tenant Creation (Upfront)

When you create a tenant, you can specify the initial shard count:

```bash
POST /v1/tenant
{
  "shard_parameters": {
    "count": 4,    # Create 4 shards immediately
    "stripe_size": 268435456
  }
}
```

- Default: 1 shard (unsharded)
- All specified shards are created immediately
- Storage Controller persists to database first, then attaches to pageservers

### During Tenant Growth (Automatic)

Storage Controller runs a background task `autosplit_tenants()` that:
- Checks every ~20 seconds
- Monitors tenant sizes
- Automatically splits based on thresholds

### Manually by Admins

Admins can manually trigger splits via API:

```bash
PUT /control/v1/tenant/:tenant_id/shard_split
{
  "new_shard_count": 8
}
```

Requires Admin scope permissions.

---

## Who Creates Shards

**The Storage Controller** is responsible for creating tenant shards. It manages:
- Initial shard creation during tenant creation
- Automatic shard splitting based on size thresholds
- Manual shard splitting via admin API

The **Pageserver** implements the mechanics of splitting but doesn't make split decisions.

### Shard Creation Workflow

```
┌─────────────────────────────────────────────────────┐
│ 1. User/Admin requests tenant creation or split     │
│    (or autosplit task detects growth)               │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│ 2. Storage Controller validates request             │
│    - Check permissions                              │
│    - Calculate shard count                          │
│    - Determine stripe size                          │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│ 3. Persist to Database                              │
│    - Create TenantShardId records for each shard    │
│    - Mark shard state as "Creating"                 │
│    - Database ensures durability                    │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│ 4. Scheduler selects pageservers                    │
│    - For each shard, pick target pageserver         │
│    - Consider load, AZ, placement policy            │
│    - Assign generation numbers                      │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│ 5. Attach shards to pageservers                     │
│    - Send LocationConfig to each pageserver         │
│    - Pageserver creates shard locally               │
│    - Pageserver pulls timeline data                 │
└────────────┬────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────┐
│ 6. Complete                                         │
│    - Mark shards as "Active"                        │
│    - Shards ready to serve queries                  │
└─────────────────────────────────────────────────────┘
```

---

## How Routing Works

### The Hash Function (The Core)

Location: `libs/pageserver_api/src/shard.rs:319`

```rust
pub fn key_to_shard_number(
    count: ShardCount,           // Total shards (e.g., 4)
    stripe_size: ShardStripeSize, // Locality unit (e.g., 2048 pages = 16MB)
    key: &Key,                   // Page key (RelNode, BlockNumber)
) -> ShardNumber {
    // Fast path: unsharded or broadcast keys
    if count < ShardCount(2) || key_is_shard0(key) {
        return ShardNumber(0);
    }

    // Hash the RelNode
    let mut hash = murmurhash32(key.field4);

    // Hash the BlockNumber divided by stripe size
    hash = hash_combine(hash, murmurhash32(key.field6 / stripe_size.0));

    // Return: hash modulo shard count
    ShardNumber((hash % count.0 as u32) as u8)
}
```

### Formula

```
shard = (murmurhash32(relNode) + murmurhash32(blockNum/stripe_size)) % shard_count
```

### Example

```
Table "users" (relNode = 12345)
Block 0      → hash = 42 → shard = 42 % 4 = Shard 2
Block 1      → hash = 43 → shard = 43 % 4 = Shard 3
Block 2048   → hash = 44 → shard = 44 % 4 = Shard 0  (new stripe)
Block 2049   → hash = 45 → shard = 45 % 4 = Shard 1
```

### Routing Happens in Two Places

```
┌──────────────────────┐
│  Client (Application)│
│  PostgreSQL client   │
└──────────┬───────────┘
           │ SQL query
           ▼
┌──────────────────────────────────────────┐
│   Compute Node (PostgreSQL)              │
│   with neon extension                    │
│ ──────────────────────────────────────── │
│                                          │
│  1. PostgreSQL executes query            │
│  2. Needs page from relation 12345,      │
│     block 100                            │
│  3. Storage Manager (neon extension)     │
│     intercepts page request              │
│  4. ROUTES REQUEST → calculates shard    │
│     using same hash function             │
│  5. Sends GetPage to correct pageserver  │
│     shard (pageserver-C)                 │
│                                          │
└──────────┬───────────────────────────────┘
           │ GetPage@LSN(key, shard=2)
           ▼
┌──────────────────────────────────────────┐
│   Pageserver (multiple shards)           │
│ ──────────────────────────────────────── │
│                                          │
│  Shard 0/4 (on PS-A)                    │
│  Shard 1/4 (on PS-B)                    │
│  Shard 2/4 (on PS-C) ← Request arrives │
│  Shard 3/4 (on PS-D)                    │
│                                          │
│  Verifies shard routing                  │
│  Returns page                            │
└──────────┬───────────────────────────────┘
           │ page data
           ▼
┌──────────────────────────────────────────┐
│   Compute Node                           │
│   Returns result to client               │
└──────────────────────────────────────────┘
```

### Why Two-Level Routing?

| Level | Component | Purpose |
|-------|-----------|---------|
| **Compute-side** | neon extension | Decide which pageserver to ask |
| **Pageserver-side** | TenantManager | Verify request is for correct shard, serve page |

**Compute-side routing** answers: "Which pageserver?"
```
shards[shard_num].pageserver_connstring → "pageserver-c:6400"
```

**Pageserver-side routing** answers: "Is this my data? Return it."
```
if is_key_local(key, my_shard) → serve page
```

### Complete Request Flow

```
Compute Node wants to read from "main" timeline:
  GetPage(tenant_id="abc123", timeline_id="timeline1",
          key=Key, lsn=5000)
                           ↓
  Calculate shard: hash(key) % 4 = Shard 2
                           ↓
  Send to Pageserver C (which holds Shard 2/4)
                           ↓
  Pageserver C:
    1. Find tenant "abc123"
    2. Find timeline "timeline1" within that tenant
    3. Find shard 2/4 within that timeline
    4. Return page
                           ↓
  Response: page from "main" timeline, shard 2, LSN 5000
```

---

## Transaction Consistency Across Shards

### The Key Insight

```
WRONG assumption:
  Shard 0 serves LSN 1000
  Shard 1 serves LSN 900
  → Inconsistency! ❌

ACTUAL design:
  Shard 0 serves LSN 1000
  Shard 1 serves LSN 1000
  → Both serve from same point in time ✓
```

### LSN is Global, Not Per-Shard

**LSN (Log Sequence Number) is a single global monotonically increasing value** for the entire database timeline, not per-shard.

- **Single LSN** for all shards in a timeline
- **All shards consume same WAL** at same pace
- **All advance LSN together**

```
Timeline "main":
  LSN: 1000 → 1001 → 1002 → 1003 ...

  Shard 0: LSN 1003 (main)
  Shard 1: LSN 1003 (main)
  Shard 2: LSN 1003 (main)
  Shard 3: LSN 1003 (main)

All shards are at SAME LSN for the timeline ✓
```

### Unified WAL Stream

```
┌────────────────┐
│  Safekeeper    │
│  Single WAL    │
│  Stream        │
└────────┬───────┘
         │ WAL: [Write row1, Write row2, Write row3, ...]
         │ LSN advances: 100→101→102→103...
         │
    ┌────┴─────────────────────────────────┐
    │ All shards subscribe to SAME WAL     │
    │
    ├─→ Shard 0/4
    ├─→ Shard 1/4
    ├─→ Shard 2/4
    └─→ Shard 3/4

All advance LSN together: 100→101→102→103...
```

### Transaction Reads at Single LSN

When compute executes a transaction:

```sql
BEGIN;
  SELECT * FROM users WHERE id = 1;   -- Shard 2
  SELECT * FROM orders WHERE id = 1;  -- Shard 1
  SELECT * FROM products WHERE id = 1; -- Shard 3
COMMIT;
```

All reads happen at **the same LSN** (e.g., LSN = 5000):

```
┌─────────────────────────────────────────┐
│ Transaction reads at LSN 5000           │
├─────────────────────────────────────────┤
│                                         │
│ GetPage(users, block=100, LSN=5000)    │
│   → Shard 2 serves page as of LSN 5000 ✓│
│                                         │
│ GetPage(orders, block=200, LSN=5000)   │
│   → Shard 1 serves page as of LSN 5000 ✓│
│                                         │
│ GetPage(products, block=300, LSN=5000) │
│   → Shard 3 serves page as of LSN 5000 ✓│
│                                         │
│ All pages are from same logical snapshot│
│ → CONSISTENT! ✓                        │
└─────────────────────────────────────────┘
```

### Deterministic Key-to-Shard Mapping

The shard assignment is **deterministic and immutable**:

```
RelNode 12345 + BlockNum 100
  → Hash(12345) + Hash(100/2048)
  → Shard 2 (ALWAYS)

At LSN 1000: Shard 2 has block 100
At LSN 5000: Shard 2 still has block 100
At LSN 10000: Shard 2 still has block 100
```

**Same row, same shard, same column at every LSN** ✓

### ACID Properties Maintained

Neon maintains full ACID properties across sharded pageservers through:

1. **Global, monotonic LSN** - Not per-shard, ensures version alignment
2. **Unified WAL stream** - All shards consume the same WAL at the same pace
3. **Deterministic, immutable key-to-shard mapping** - Same row always on same shard
4. **Version-based reads** - All pages read at same LSN ensures consistency
5. **Metadata replication** - Critical metadata stored on all shards
6. **Quorum-based WAL durability** - Ensures no data loss across failures

| Property | Mechanism | Result |
|----------|-----------|--------|
| **Atomicity (A)** | WAL records are atomic units written by Postgres, replicated to quorum of safekeepers before transaction commits | All shards eventually see same WAL records at same LSN |
| **Consistency (C)** | Deterministic key-to-shard mapping, version-based point-in-time reads via single LSN, metadata replication to all shards | All pages at LSN X are from same logical database state |
| **Isolation (I)** | PostgreSQL handles transaction isolation at compute level, pageserver provides pages at specific LSNs | Since all pages at LSN X are from same instant, isolation preserved |
| **Durability (D)** | WAL written to quorum of safekeepers, archived to S3, all shards receive durable WAL before acknowledging | No data loss |

---

## When to Shard User Data

User data is **automatically sharded based on tenant size thresholds**.

### Two Automatic Sharding Triggers

#### 1. Initial Split (Unsharded → Sharded)

When a single-shard tenant grows too large:

```
Threshold: initial_split_threshold (default: ~100GB per timeline)

Timeline size: 50GB  → No split needed
Timeline size: 100GB → No split needed (at threshold)
Timeline size: 150GB → SPLIT! ✓
```

**What happens:**

```
Before:
  Tenant: 150GB on 1 shard
  Shard 0/1 (Pageserver A): all 150GB

After split:
  Tenant: 150GB on 4 shards
  Shard 0/4 (Pageserver A): 37.5GB
  Shard 1/4 (Pageserver B): 37.5GB
  Shard 2/4 (Pageserver C): 37.5GB
  Shard 3/4 (Pageserver D): 37.5GB
```

#### 2. Size-Based Split (Multi-shard → More shards)

When per-shard data exceeds threshold:

```
Threshold: split_threshold (default: ~10GB per shard)

Tenant: 40GB on 4 shards → Per-shard: 10GB ✓ (OK)
Tenant: 100GB on 4 shards → Per-shard: 25GB > 10GB ✗ (SPLIT!)
```

**What happens:**

```
Before:
  Tenant: 100GB on 4 shards
  Per shard: 25GB > 10GB threshold

Storage Controller calculates:
  100GB / 10GB = 10 shards needed
  Round up to power of 2 → 16 shards

After split:
  Tenant: 100GB on 16 shards
  Per shard: 6.25GB < 10GB ✓ (acceptable)
```

### The Auto-Split Mechanism

**Background task: `autosplit_tenants()`** (runs every ~20 seconds)

```rust
async fn autosplit_tenants() {
    loop {
        // Check all tenants
        let candidates = find_split_candidates().await;

        for tenant in candidates {
            // Case 1: Unsharded tenant > initial_split_threshold
            if tenant.shard_count == 1
               && tenant.max_logical_size > initial_split_threshold {
                split_to(initial_split_shards);
            }

            // Case 2: Multi-shard tenant, per-shard > split_threshold
            if tenant.max_logical_size / tenant.shard_count > split_threshold {
                new_count = compute_split_shards(tenant.max_logical_size);
                split_to(new_count);
            }
        }

        sleep(20s).await;
    }
}
```

### Configuration

Storage Controller config controls when splitting happens:

```toml
[split_threshold]
# Size per shard before splitting
# 0 or commented = disabled (no auto-split)
value_mb = 10240  # 10GB per shard

[initial_split_threshold]
# Size of unsharded tenant before initial split
value_mb = 102400  # 100GB

initial_split_shards = 4    # Split into 4 initially
max_split_shards = 32       # Never exceed 32 shards
```

### Timeline Showing When Sharding Happens

```
Tenant Creation
│
├─ Size: 10GB (1 shard)
│   ├─ Unsharded, small
│   ├─ Auto-split: NO (10GB < 100GB threshold)
│   └─ Manual split: possible but unnecessary
│
├─ Growth to 50GB (1 shard)
│   ├─ Auto-split: NO (50GB < 100GB)
│   └─ All data on single pageserver
│
├─ Growth to 150GB (1 shard)
│   ├─ Auto-split: YES! (150GB > 100GB threshold)
│   │   ↓
│   │   Split into 4 shards
│   │   Shard 0/4: 37.5GB on Pageserver A
│   │   Shard 1/4: 37.5GB on Pageserver B
│   │   Shard 2/4: 37.5GB on Pageserver C
│   │   Shard 3/4: 37.5GB on Pageserver D
│   │
│   └─ Continue from 4 shards
│
├─ Growth to 150GB on 4 shards
│   ├─ Per-shard: 37.5GB
│   ├─ Auto-split: YES (37.5GB > 10GB)
│   │   ↓
│   │   Split into 8 shards (150GB / 10GB = 15 → round to 16)
│   │   Each shard: ~9.4GB < 10GB ✓
│   │
│   └─ Continue from 8 shards
│
└─ Continue until max_split_shards (32)
```

### Example: SaaS Growing Tenant

```
Day 1: New customer signs up
  Database: 1GB
  Shard: 1/1 (on Pageserver A)
  Auto-split: NO

Week 1: Customer adds data
  Database: 50GB
  Shard: 1/1 (on Pageserver A)
  Auto-split: NO

Month 1: Customer is active
  Database: 150GB
  Shard: SPLIT! 1/1 → 4/4

  Before split:
    │ Pageserver A │ (150GB all rows)

  After split:
    │ PS-A │ PS-B │ PS-C │ PS-D │
    │ 37GB │ 37GB │ 37GB │ 37GB │

  Load distributed! Queries can hit 4 servers in parallel

Month 3: Customer grows more
  Database: 200GB on 4 shards → Per-shard: 50GB > 10GB
  Shard: SPLIT! 4/4 → 16/16

  Load: 4 pageservers → 16 pageservers
  Per-shard: 50GB → 12.5GB ✓

Year 1: Massive customer
  Database: 1TB
  Shards: Scaled up as needed
  Read/write load distributed across many pageservers
```

---

## Shard Merging

### Current Status: NOT SUPPORTED ❌

**No, shard merging/combining is NOT supported in Neon.**

From the official Shard Splitting RFC (lines 50-53):

> **Non Goals:**
>
> "The inverse operation (shard merging) is not described in this RFC. This is a lower priority than splitting, because databases grow more often than they shrink, and a database with many shards will still work properly if the stored data shrinks, just with slightly more overhead (e.g. redundant WAL replication)"

### Why No Shard Merging?

#### 1. Lower Priority
- Databases typically GROW, not shrink
- Splitting is more common than merging

#### 2. Still Works Fine
Even if data shrinks, having extra shards doesn't break anything:

```
Example:
  Original: 1TB tenant → Split into 16 shards (62.5GB each)

  Later: Tenant deletes 900GB of data → Now 100GB total

  Current state:
    16 shards × (100GB/16) = 6.25GB per shard

  Problems:
    ✗ Wasted storage (each pageserver has overhead)
    ✗ Redundant WAL replication to 16 places

  But:
    ✓ Database still works correctly
    ✓ Queries still execute properly
    ✓ ACID properties maintained
    ✓ Just "slightly more overhead"
```

#### 3. Complexity
Shard merging is algorithmically harder:

```
Shard SPLIT is simple:
  Parent shard (1/1) → Children (0/2, 1/2)
  Key mapping is stable within new shard config ✓

Shard MERGE is complex:
  Parents (0/16, 1/16, 2/16, ..., 15/16) → Child (0/8, 1/8, 2/8, ..., 7/8)

  Need to:
  1. Recombine layers from multiple shards
  2. Re-encrypt/re-key data
  3. Handle concurrent writes during merge
  4. Deal with compute still reading from old shards
  5. Ensure no data loss

  Much more complex than split! ✗
```

### What Happens If Data Shrinks?

```
Timeline: Tenant with excessive shards

Day 1: Tenant at 1TB
  → Auto-split to 16 shards
  → Per-shard: 62.5GB

Day 30: Customer deletes 90% of data
  → Tenant now 100GB on 16 shards
  → Per-shard: 6.25GB

Current behavior:
  ✗ Still 16 shards (no merging)
  ✗ WAL replicated to 16 pageservers
  ✗ 16 layer compactions happening
  ✓ But database works fine!

Recommendation:
  • If shards are truly wasteful, admin can:
    1. Delete tenant and recreate
    2. Or just leave as-is (slight overhead)
```

### Possible Workaround: Manual Recreation

If you really want fewer shards:

```bash
# Current state: Tenant "abc123" with 16 shards, 100GB

# Option 1: Accept the overhead
  # Leave as-is, slight cost but works fine

# Option 2: Manual recreation (destructive)
  # 1. Dump tenant data
  neon dump tenant abc123 > backup.sql

  # 2. Delete tenant
  neon tenant delete abc123

  # 3. Recreate with fewer shards
  neon tenant create --shard-count 4

  # 4. Restore data
  neon restore backup.sql

  # Trade-off: Downtime, but cleaner shard count
```

### RFC Status

| Operation | Status | Why |
|-----------|--------|-----|
| **Shard Split** | ✅ Implemented | Essential for growth |
| **Shard Merge** | ❌ Not implemented | Lower priority, complex |
| **Auto-split** | ✅ Implemented | Automatic sizing |
| **Auto-merge** | ❌ Not planned | Rarely needed |

---

## Summary

### Key Points

```
┌────────────────────────────────────────────────────┐
│ NEON SHARDING AT A GLANCE                          │
├────────────────────────────────────────────────────┤
│                                                    │
│ What is sharded?                                   │
│  → Tenant data (spread across pageservers)         │
│                                                    │
│ What is NOT sharded?                               │
│  → Timeline (logical branch, complete unit)        │
│  → LSN (global per timeline)                       │
│  → Transactions (consistent via single LSN)        │
│                                                    │
│ When is data sharded?                              │
│  → Automatically when size exceeds threshold       │
│  → Can be manually triggered by admin              │
│  → Disabled by setting thresholds to 0             │
│                                                    │
│ How is routing done?                               │
│  → Deterministic hash function (compute + PS)      │
│  → Compute calculates target shard                 │
│  → Pageserver verifies and serves page             │
│                                                    │
│ Does it break transactions?                        │
│  → NO - all shards at same LSN (global)            │
│  → Unified WAL stream ensures consistency          │
│  → ACID properties fully maintained                │
│                                                    │
│ Can I merge shards if data shrinks?                │
│  → NO - not supported (lower priority)             │
│  → Still works fine with extra shards              │
│  → Slight overhead but fully functional            │
│                                                    │
└────────────────────────────────────────────────────┘
```

### Design Philosophy

```
Neon's shard strategy:

Growth path (common):
  1 shard → 2 shards → 4 shards → 8 shards → 16 shards
           ✓ Implemented ✓ Works great

Shrinking path (rare):
  16 shards → 8 shards → 4 shards
           ✗ Not implemented, not priority
```

### Comparison with Other Components

| Component | Sharded? | Why |
|-----------|----------|-----|
| **Tenant data** | ✅ YES | Distributes storage/I/O load |
| **Timeline** | ❌ NO | Logical unit, complete |
| **Pageserver** | ❌ NO (stores shards) | Infrastructure component |
| **Safekeeper** | ❌ NO (stores WAL) | Infrastructure component |
| **Compute node** | ❌ NO | Client-facing |
| **Storage Controller** | ❌ NO (manages shards) | Infrastructure component |
