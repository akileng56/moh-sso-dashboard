# Audit read feature

This feature provides audit log browsing, cursor pagination, action discovery,
metrics and CSV/JSON export. Its existing registry ID is `audit`.
`DISABLED_MODULES=audit` omits its routes on restart. Core event recording and
audit schema migrations remain owned by the host and continue while read APIs
are disabled. The feature owns no background worker.

## Extraction phase 1: dependency contracts

`Repository` in `repository.go` is the read-only query boundary. It accepts
feature-owned input types and returns feature-owned response types from `dto.go`;
it exposes no SQLC-generated types or event-writing methods. `Service` uses that
port. The concrete SQLC repository and generated-row translations now live in
`repository_postgres.go` and `mapper_postgres.go`. Portable response helpers remain in `mapper.go`; export preparation and CSV
formatting are isolated in `export.go` as described below.

`Cache` in `cache.go` exposes only `Get` and `Set`. Handlers no longer import the
host Redis implementation. The bootstrap composition root supplies its existing
Redis cache through that interface, preserving a genuinely nil interface when
the optional concrete pointer is nil. Other cache implementations can implement
the same decoding and TTL contract.

Audit list caching remains best effort with its existing query key and 20-second
TTL. Hits must decode successfully; lookup/decoding errors fall back to the
repository, and cache write errors do not fail successful reads. Repository
failures are not cached. Host authentication and `audit:read` permission checks
run before cache access. Existing export rate limits, response envelopes, JSON
metadata, cursor fields, CSV quoting and manifest hashes are preserved.

Regression checks cover those cache paths, filtered/cursor inputs, denial before
cache lookup, JSON/CSV exports and SQLC response normalization. Existing registry
tests cover read-route permissions and disablement independently of core event
recording. Run:

```sh
go test -race ./internal/features/audit ./internal/api ./internal/bootstrap
```

## Extraction phase 2: portable reads and exports

`NewCachedReader` in `read.go` serves typed `AuditLogListResponse` data through
the service/cache ports. It has no
Gin, Redis or SQLC dependency. The HTTP handler parses and validates filters and
cursors before invoking it, and retains the existing raw-query cache identity,
`audit_logs:` prefix and 20-second TTL. Non-HTTP callers supply an identity
covering their complete query; an empty identity bypasses caching.

`BuildExport` in `export.go` builds the public JSON export and manifest directly
from typed audit records. It derives the count from the records, preserves UTC
ranges and the SHA-256 hash of the JSON items, and returns serialization errors
before an HTTP response starts. `WriteCSV` writes the existing columns to an
`io.Writer` and returns encoding, write and flush errors.

The HTTP adapter owns authentication/permissions, query validation, envelopes,
content headers and handling of partial response failures. The CSV manifest
header is set before body writes, so it is present even when an export exceeds
the CSV buffer. A streaming failure is recorded on the Gin context and aborts
the handler without appending JSON to a partial CSV response. Export rate limits
remain in the host route registrar.

Tests cover query cache isolation, context/error propagation, empty cache
identities, unchanged JSON/CSV fields, manifest hashes, writer/flush failures,
unsupported metadata, large-export headers and validation before cache lookup.
The generated OpenAPI list success schema now consistently names the typed audit
list response for both warm and cold cache paths; routes and payload fields are
unchanged.

The feature remains registry-hosted and reports the backend build version.
The standalone release candidate now lives in
[`backend/modules/audit`](../../../modules/audit/README.md), with its own dependency
metadata, public APIs, portable tests and an external-consumer test. SQLC/cache
implementations, HTTP permissions, event recording and schema migrations remain
host-owned. The next phase publishes and pins a release, then replaces the
retained portable host implementation with adapters. No Audit module version
has been published yet.
