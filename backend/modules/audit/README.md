# Audit read Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/audit`.
Requires Go 1.25.4.

Version `v0.1.0` is the first standalone Audit read release. Its Go module tag
is `backend/modules/audit/v0.1.0`. Hosts select the immutable dependency version
and supply their repository/cache implementations and HTTP authorization.

## Public API

| API | Responsibility |
| --- | --- |
| `Repository` | Read-only log listing/detail, action discovery, metrics and export queries |
| `Service`, `NewService` | Query service over the supplied repository |
| `Cache` | Optional Get/Set cache interface; the host supplies serialization/storage |
| `CachedReader`, `NewCachedReader` | Typed list caching with a 20-second TTL and best-effort cache failures |
| `BuildExport` | Public export response, record count, UTC range and JSON-item SHA-256 manifest |
| `WriteCSV` | Existing audit columns written to an `io.Writer`; returns encoding/write/flush errors |
| Input/response types | Public records, filters, cursors, metrics, export responses and manifests |
| `ModulePath`, `Version` | Selected consumer dependency version; local/unversioned builds report `dev` |

The caller authenticates and authorizes access, validates query parameters and
provides a context. For cached reads, the caller supplies an identity covering the
complete query, filters, cursor and applicable scope. The dashboard supplies its
existing raw-query identity. An empty identity bypasses caching. Cache errors fall
back to the repository; repository failures are not cached.

Call `BuildExport` before starting an HTTP response so invalid metadata fails
before streaming. The manifest hashes the JSON items, not the CSV bytes. The host
sets content and manifest headers before calling `WriteCSV`, handles partial
writer failures and retains export rate limits and permission checks.

## Published v0.1.0 host ownership

The package has no database adapter, SQL migration files, event writer, background
worker or owned storage resource. Implement `Repository` and optionally `Cache`
using the host platform. The dashboard retains its SQLC repository and row-mapping
adapters, Redis connection, HTTP handlers, authorization and registry wiring.

Core audit-event recording and audit schema migrations remain in the dashboard.
Disabling the read module must continue to preserve event recording. Packaging
or upgrading this library neither runs migrations nor changes database history.

## Verification

From this directory:

```sh
GOWORK=off go test -race -mod=readonly ./...
GOWORK=off go vet -mod=readonly ./...
```

From the repository root:

```sh
python3 infra/scripts/test-audit-package.py
make -C backend modules-check
```

The script copies only this module outside the host tree, disables Go workspaces,
checks production and test dependency graphs, verifies module checksums and runs
build/vet/race tests with read-only dependency metadata. Tests cover typed cache
behavior, context/error propagation, query identities, export hashes and CSV
encoding/write/flush failures. An external-package consumer test exercises public
ports and the read-to-export pipeline without host code.

Host SQLC mapping, HTTP authorization, large-response headers and module
disablement remain covered by dashboard tests. To integrate a release, verify a
clean download, pin it in host dependency metadata, supply adapters and report
`Version()` in the compiled inventory. Production builds use the selected
dependency without local workspace replacements. Event recording and migrations
remain host-owned.

## Self-contained v0.2.0

Version v0.2.0 provides the self-contained HTTP and PostgreSQL implementation.
Release tag: `backend/modules/audit/v0.2.0`.

`NewModule(Dependencies)` / `Register(Registrar, Dependencies)` provide the seven
existing GET endpoints under `/admin/audit-logs` (a host registrar adds `/api/v1`).
`Patterns()` returns detached method/path declarations; `OpenAPI()` owns the
operation, filter, DTO and CSV/JSON export contracts. `BasePath` supports
alternate mounts. The package uses standard HTTP plus its existing UUID
dependency, with no Gin, SQLC, Redis or PostgreSQL-driver dependency.

Supply exactly one of `Database *sql.DB` or `Repository`.
`NewRepository(*sql.DB) (Repository, error)` owns all seven read queries,
keyset pagination and row conversion in the package. SQL NULL metadata and
nullable UUIDs retain their API semantics. Caller-owned pools are never closed;
the package never migrates tables or records events. Core migrations and
continuous audit recording remain platform responsibilities.

`RequireAdmin` and `AuthorizeAny` must inspect verified platform identity.
Both administrator access and `audit:read` run before query validation, cache
lookup or SQL. Raw headers do not establish identity. The new HTTP entry point
requires the explicit administrator gate as well as its existing read permission.

`LimitExport` is also required. It receives the package-owned `audit-export`
policy, limit 10 and one-minute window, and must enforce a per-verified-user
limit. It returns standard HTTP headers and an error; rate headers are applied
to both allowed and denied exports. Errors implementing `StatusCode() int`
preserve 401/403/429, while unclassified failures return a safe 503.
Optional `RetryAfter() time.Duration` sets a rounded-up `Retry-After` header.
Provider errors are never exposed verbatim. A callback returning nil explicitly
permits the request; omitting the capability is invalid.

Optional `Cache` keeps existing raw-query cache keys and 20-second best-effort
caching. Authorization always precedes cache lookup. `RequestID` supplies shared
trace metadata. Optional `StreamError` reports CSV write/flush failures without
appending a JSON error to a partial CSV response. Export filenames and manifest
headers precede body writes, including exports larger than CSV's buffer.
JSON manifests hash the JSON items; CSV headers carry that same integrity hash.
Unsupported export format values preserve the existing CSV fallback.

The real PostgreSQL suite uncovered the retained failed-login-series scan bug:
the query returns a timestamp while its old internal row declared an integer.
Version v0.2.0 scans it as a timestamp and returns the public numeric day bucket
as Unix milliseconds. Consumers receive this correction when selecting v0.2.0.

Verify from the repository root:

```sh
python3 infra/scripts/test-audit-package.py
python3 infra/scripts/test-audit-http-package.py --postgres
```

The HTTP verifier copies the package outside the dashboard, disables workspaces
and downloads, checks dependency boundaries and metadata, and runs build/vet/
uncached race tests. A separate HTTP binary runs in an empty directory.
Its SELECT-only PostgreSQL fixture verifies all seven queries/routes, tied-time
keyset pagination, NULL JSON/UUID values, bound filters, canceled contexts, pool
lifetime and CSV/JSON manifest integrity. The PostgreSQL driver belongs only to
that external consumer.

Hosts import the published module directly and connect shared verified admin,
permission, rate-limit, request-ID and stream-error capabilities. Disabling the
read module must omit its routes while core event recording continues.
