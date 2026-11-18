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

## Host ownership

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
