# Storage Controller Configurations

## Overview

The storage controller in Neon has a comprehensive set of configuration options organized into multiple categories. These configurations control node availability, tenant sharding, safekeeper settings, API protocols, and cluster management.

## Configuration Structure

- **Location:** `control_plane/src/local_env.rs` (lines 178-259)
- **Struct:** `NeonStorageControllerConf`
- **Format:** TOML configuration files
- **Deserialization:** Uses `#[serde(default)]` for optional fields

## Configuration Categories

### 1. Node Health & Monitoring

Controls how the storage controller monitors and manages node availability.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `max_offline` | Duration | 10s | Heartbeat timeout before marking a node offline. Controls how long the storage controller waits before considering a pageserver unavailable. |
| `max_warming_up` | Duration | 30s | Maximum duration for a node to be in warming up state. |
| `heartbeat_interval` | Duration | 1s | Interval between heartbeat checks to pageservers. |
| `long_reconcile_threshold` | Optional Duration | None | Threshold for flagging reconciliation operations as taking too long. |

### 2. Tenant Sharding & Splitting Configuration

Manages automatic tenant shard splitting and distribution.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `split_threshold` | Optional u64 | None | Byte threshold that triggers automatic tenant shard splitting. |
| `max_split_shards` | Optional u8 | None | Maximum number of shards a tenant can be split into. |
| `initial_split_threshold` | Optional u64 | None | Initial/starting threshold for shard splitting (may differ from ongoing threshold). |
| `initial_split_shards` | Optional u8 | None | Number of shards to create a new tenant with initially. |
| `max_secondary_lag_bytes` | Optional u64 | None | Maximum allowed lag (in bytes) for secondary replica locations. Related to secondary placement strategy. |
| `shard_split_request_timeout` | Optional Duration | None | Timeout duration for shard split requests. |

### 3. Safekeeper Configuration

Configures safekeeper integration and timeline management.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `timeline_safekeeper_count` | Optional usize | None | Number of safekeepers to use for timeline storage. If not specified, defaults to the total number of configured safekeepers. Used in test environments with fewer than 3 safekeepers. |
| `timelines_onto_safekeepers` | bool | true | Enable placing timelines onto safekeepers. Passed as `--timelines-onto-safekeepers` flag. |

### 4. API & Protocol Configuration

Controls which protocols are used for inter-service communication.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `use_https_pageserver_api` | bool | false | Use HTTPS when communicating with pageservers. Passed as `--use-https-pageserver-api` flag. |
| `use_https_safekeeper_api` | bool | false | Use HTTPS when communicating with safekeepers. Passed as `--use-https-safekeeper-api` flag. |
| `use_local_compute_notifications` | bool | true | Use local compute notifications (vs. remote control plane hooks). |

### 5. Cluster Management

Manages multi-node storage controller deployments and coordination.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `start_as_candidate` | bool | false | Whether to start as a candidate node in multi-node storage controller setups. Passed as `--start-as-candidate` flag. |
| `database_url` | Optional SocketAddr | None | Database URL used when running multiple storage controller instances. Required when using `--base-port` argument. |

### 6. Advanced Features

Additional configuration options for specialized behaviors.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `kick_secondary_downloads` | Optional bool | None | Control behavior for secondary replica downloads. Passed as `--kick-secondary-downloads={value}` flag. |
| `posthog_config` | Optional PostHogConfig | None | PostHog analytics configuration. Passed via `POSTHOG_CONFIG` environment variable as JSON. |

## Configuration Sources

### 1. TOML Configuration File
- Location: `.neon/config`
- Format: Standard TOML format
- All fields are optional due to `#[serde(default)]`
- Deserialized via `OnDiskConfig` struct

### 2. Environment Variables
- `POSTHOG_CONFIG` - PostHog analytics configuration (JSON format)
- `LD_LIBRARY_PATH` - Library search paths (Unix/Linux)
- `DYLD_LIBRARY_PATH` - Library search paths (macOS)

### 3. Command-Line Arguments
- Arguments built from config when starting via neon_local
- Examples: `--max-offline-interval`, `--heartbeat-interval`, `--split-threshold`

## Default Values Summary

| Option | Default |
|--------|---------|
| `max_offline` | 10 seconds |
| `max_warming_up` | 30 seconds |
| `heartbeat_interval` | 1 second (1000ms) |
| `start_as_candidate` | false |
| `use_https_pageserver_api` | false |
| `timelines_onto_safekeepers` | true |
| `use_https_safekeeper_api` | false |
| `use_local_compute_notifications` | true |
| All optional fields | None |

## Configuration Example

```toml
[storage_controller]
max_offline = "10s"
max_warming_up = "30s"
heartbeat_interval = "1s"
split_threshold = 1099511627776  # 1 TB in bytes
max_split_shards = 8
initial_split_shards = 4
start_as_candidate = false
use_https_pageserver_api = false
timelines_onto_safekeepers = true
use_local_compute_notifications = true
```

## Related Files

- **Startup Logic:** `control_plane/src/storage_controller.rs` (lines 527-676)
  - Shows how configuration is converted to command-line arguments
  - Includes JWT token handling and authentication setup

- **Documentation:** `docs/storage_controller.md`
  - High-level documentation of storage controller architecture
  - Deployment guidance and hook configurations

- **Configuration Definition:** `control_plane/src/local_env.rs` (lines 178-259)
  - Main configuration struct definition

## Use Cases

These configurations allow fine-tuning of storage controller behavior for different deployment scenarios:

- **Local Development:** Use defaults via neon_local with minimal overrides
- **Testing:** Adjust timeouts and thresholds for test environments
- **Production:** Configure for multi-node deployments with HTTPS, custom safekeeping strategies, and optimized shard splitting
- **High Availability:** Set up clustering with `start_as_candidate` and `database_url` configurations

## Duration Format

All duration fields use `humantime_serde` for deserialization, supporting human-readable formats:
- `10s` - 10 seconds
- `1m` - 1 minute
- `30ms` - 30 milliseconds
- `1h` - 1 hour
