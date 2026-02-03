
## PGXN

Compute nodes → Need pgxn (to communicate with pageserver, stream WAL, etc.)
Pageserver → Pure Rust, no pgxn needed
WAL Redo Processes → Ephemeral PostgreSQL instances spawned on-demand to replay WAL; they use pgxn extensions but are temporary, not part of the persistent pageserver deployment

## Storage Scrubber

The Storage Scrubber is a critical maintenance tool that manages and validates storage integrity across remote cloud storage (S3, Azure, GCS). Think of it as a cleanup and validation utility for tenant data stored in the cloud

FindGarbage	Scan cloud storage and identify unreferenced objects (deleted tenants that still have S3 files)
PurgeGarbage	Safely delete the garbage objects found by FindGarbage
ScanMetadata	Validate all metadata files for corruption or consistency issues
PageserverPhysicalGC	Remove old layer files and stale index files (with 3 safety modes: DryRun, IndicesOnly, Full)
TenantSnapshot	Download/backup an entire tenant's data locally for disaster recovery
FindLargeObjects	Identify large storage objects for cost optimization analysis
CronJob	Run periodic maintenance tasks automatically

storage_scrubber is a separate, independent binary that runs as a CLI tool for maintenance tasks

## compute_tools

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
