# Metrics Go module

This module provides the dashboard metrics service, HTTP handlers, response
types and repository port. It has its own `go.mod` and can be consumed from a
different Go repository. Version v0.2.0 owns the HTTP entry point and PostgreSQL reads. Authentication,
permission resolution, audit logging and application lifecycle remain shared
platform capabilities.

## Use

```sh
go get github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics@v0.2.0
```

Implement `metrics.Repository` using the public records in `metrics/model`.
Construct the service with `metrics.NewService(repository)` and the handler
with `metrics.NewHandler(service)`. Register routes with
`metrics.RegisterAdminRoutes(adminGroup, handler, authorizationMiddleware)`.
The authorization middleware is required. The host should also apply its
authentication, audit and rate limiting middleware to the parent group.

The supplied admin group receives the `/metrics` prefix and all 19 existing
endpoints. Success and error envelopes preserve the portal API format. Hosts
can set the Gin key `request_id` to include `meta.requestId` in responses.

`metrics.Version()` reports the version selected by the consuming binary.
Local filesystem and workspace replacements report `dev`.

## Database ownership

Metrics reads the host's users, client, sessions and audit data through its
repository port. The module owns no database tables or migrations. In the
SSO-DASHBOARD distribution, core migrations run before module initialization.
The module owns its 32 read queries in
`internal/postgres` and provides `NewRepository(*sql.DB) (Repository, error)`.
The caller owns the database pool and its lifetime. No production schema,
migration history or worker behavior changes.

## Develop and verify

Run within this directory.

```sh
GOWORK=off go test -race ./...
GOWORK=off go vet ./...
```

For local dashboard integration, run from `backend`.

```sh
go work init . ./modules/metrics
go test ./...
```

If a workspace already exists, add the module using
`go work use ./modules/metrics`. The distribution dependency remains pinned in
`backend/go.mod`. Use `GOWORK=off` to test the published artifact.

## Publish a version

Versions use repository tags with the module directory prefix. For example,
`backend/modules/metrics/v0.1.0` publishes version `v0.1.0` at this module path.
Commit and verify the module source before creating the tag. Verify a clean
consumer can download it before updating the distribution dependency.

The repository can host several independently versioned Go modules. Each module
must keep public APIs independent of other repositories' `internal` packages.

## Self-contained module

Version v0.2.0 provides:

- `NewModule(Dependencies)`: a complete HTTP module with all 19 GET endpoints.
- `Register(Registrar, Dependencies)` and `Patterns()`: standard HTTP integration,
  compatible with ServeMux and the platform's generic registrar.
- `OpenAPI()`: package-owned operations, query/path parameters and DTO schemas.
- `NewRepository(*sql.DB)`: package-owned PostgreSQL reads using public model values.

Supply exactly one of `Database` or `Repository` and both `RequireAdmin` and
`AuthorizeAny` callbacks. They must read verified platform identity and run
before any parameter validation or query. Permission remains `metrics:read`.
Errors exposing `StatusCode() int` preserve 401/403; unclassified guard failures
return a safe 503. Raw provider/database messages are never returned.
The module does not parse credentials or establish identity from caller headers.
`RequestID` supplies the shared trace identifier. Default mount is
`/admin/metrics` (the outer platform registrar adds `/api/v1`); `BasePath`
supports alternate mounts.

The module retains its existing Gin and UUID dependencies. PostgreSQL array
scanning adds the reviewed `github.com/lib/pq v1.10.9` dependency. The new
standard HTTP entry point hides Gin from host integration. Existing
`RegisterAdminRoutes` / `NewHandler` APIs remain compatible with v0.1.0 hosts.

Date parsing, default client limit, nullable DTOs and empty collections retain
existing behavior. The overview endpoint retains its existing best-effort
aggregation: failed secondary reads can yield partial/zero results with 200.
Individual metric endpoints still return safe 500 errors on query failures.
Improving that overview error contract is separate work.

From the repository root:

```sh
python3 infra/scripts/test-metrics-package.py --postgres
```

Verification copies only this module into an external directory, disables Go
workspaces and downloads, verifies installed third-party checksums, checks the
dependency boundary, and runs build/vet/uncached race tests. A separate consumer
binary runs in an empty working directory. The optional PostgreSQL fixture
creates/removes its own container and a SELECT-only account; it exercises all
32 repository reads and 19 routes, NULLs, arrays, parameter binding, canceled
contexts and caller-owned pool lifetime. It never connects to an existing
operator database. The fixture schema is test data, not a production migration.

Hosts should import the published module directly, wire shared admin/permission
capabilities, register its patterns and merge its OpenAPI schemas. Disabling it
must omit its construction and routes. Identity, audit recording and primary
schema migrations remain platform responsibilities.
