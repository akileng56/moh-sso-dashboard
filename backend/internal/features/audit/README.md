# Audit read host adapters

The dashboard pins
`github.com/mohuganda/moh-sso-dashboard/backend/modules/audit v0.1.0` without
local replacements. Its immutable release tag is `backend/modules/audit/v0.1.0`.
The compiled `audit` inventory reports the selected dependency version.

## Responsibilities

`package.go` aliases the public DTOs and read-only repository, service and cache
contracts, and delegates construction, cached reads and export formatting to the
published package. The retained host service, cache orchestration and export
implementations have been removed.

`repository_postgres.go` and `mapper_postgres.go` adapt SQLC queries and records;
`mapper.go` builds host HTTP response envelopes. Bootstrap supplies the optional
Redis cache, preserving a nil interface when Redis is unavailable.
`handler.go` and `routes.go` retain authentication, `audit:read` permissions,
query validation, export limits, headers and streaming error handling.

Core event recording and audit migrations remain host-owned.
`DISABLED_MODULES=audit` removes the read routes on restart while event recording
continues. Audit reads own no background worker.

## Read and export contracts

The package's `NewCachedReader` uses typed responses, the existing query identity,
the `audit_logs:` prefix and a 20-second TTL. Cache errors fall back to the
repository; write failures do not fail successful reads, and repository errors
are never cached. Authentication and permissions run before cache lookup.

`BuildExport` prepares UTC ranges, record counts and SHA-256 hashes of JSON items
before a response starts. `WriteCSV` preserves the eight public columns and
propagates encoding, write and flush errors. The host sets the CSV manifest
header before streaming and never appends JSON to a partial CSV response.

## Verification

Portable implementation and external-consumer tests live in the
[standalone package](../../../modules/audit/README.md). Host tests exercise the
selected dependency through SQLC normalization, cached HTTP responses, JSON/CSV
exports, large response headers, authorization and registry disablement.

```sh
go test -race ./internal/features/audit ./internal/api ./internal/bootstrap
```

The authenticated distribution fixture also checks the compiled package version,
administrator/viewer access, cache-hit response parity, export manifests,
disablement and continued core event recording.
