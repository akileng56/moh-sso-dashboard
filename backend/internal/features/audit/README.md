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
`repository_postgres.go` and `mapper_postgres.go`. Portable response and export
manifest helpers remain in `mapper.go`.

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

The feature remains registry-hosted and reports the backend build version.
The next phases separate portable read/export code, prepare and test a standalone
module, then publish and pin a release with thin host adapters. No Audit module
version has been published in this phase.
