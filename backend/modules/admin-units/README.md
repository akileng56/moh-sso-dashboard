# Administrative Units Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units`.
Requires Go 1.25.4. The package uses only the standard library.

Version `v0.1.0` is published under `backend/modules/admin-units/v0.1.0`.
The dashboard selects this immutable dependency with reviewed checksums and
host-owned adapters. The working tree additionally contains an unreleased
self-contained HTTP/PostgreSQL candidate; the dashboard still selects v0.1.0.
Local replacements report `dev`.

## Public API

| API | Responsibility |
| --- | --- |
| `Repository` | Nine context-aware warehouse lookup contracts using package-owned values |
| `Service`, `NewService` | Existing list delegation and repository-backed hierarchy building |
| `BuildHierarchyTree` | Build the display hierarchy from caller-supplied, already scoped rows |
| `OrgUnit`, `OrgUnitFull`, `OrgUnitSimple`, `Facility`, `TreeNode` | Existing pointer/null, row version and JSON model semantics |
| `VisibilityReader`, `VisibilityNode`, `VisibilityAlias` | Minimal node-code and alias lookup contracts |
| `ResolveVisibility` | Resolve mapped DWH IDs and whether filtering is required |
| `FilterOrgUnits`, `FilterFacilities`, `FilterSimpleOrgUnits` | Retain existing ancestor/direct-ID matching and output shape |
| `ErrScopeForbidden` | Missing/invalid selected visibility dependency classification |
| `ModulePath`, `Version` | Selected dependency identity; local/unversioned source reports `dev` |

Hosts authenticate and authorize callers and validate selected context IDs.
This library does not grant access or choose the active context. The host
decides whether its policy permits an unselected request; calling
`ResolveVisibility` with an empty selected ID or missing reader fails closed.

For a valid selected context, handle errors before exposing data. Apply the
filter only when the returned restricted flag is true. Exactly `UG` with
descendant scope retains the existing unrestricted national exception.
Every other selected scope requires nonempty `dwh-org-unit` aliases. The
resolver retains namespace filtering, whitespace trimming and deduplication.
Provider error causes propagate.

Full/facility filtering retains matching against ancestor identifiers; simple
rows match their own unit IDs. Filtering with an empty allowed set returns an
empty slice. The hierarchy builder preserves existing order, UID/name fallback,
deduplication, missing-parent/subcounty pruning and null children on leaves.
It does not fetch missing ancestors or infer additional visibility.

`Service.GetHierarchy` builds from the repository's `ListHierarchy` result.
The host remains responsible for scoping that result. HTTP adapters that
already fetch/filter full rows can convert them to `OrgUnit` and call
`BuildHierarchyTree` without duplicating the algorithm.

## Self-contained candidate

The unreleased candidate owns all nine existing HTTP routes under
`/visualizer/adminunits`, their DTOs, safe response envelopes, OpenAPI schemas,
PostgreSQL lookup SQL and visibility filtering before hierarchy construction.
`NewModule(Dependencies)` prepares the module; `Register(Registrar, Dependencies)`
mounts its `Patterns()` in a standard-library ServeMux or generic host registrar.
`OpenAPI()` returns the module-owned operation and component schema document.

Supply exactly one of `Database *sql.DB` or `Repository`. `NewRepository(*sql.DB)`
reads the existing externally managed `dwh.dim_org_hierarchy` table, excluding
sentinel/non-current rows and binding optional filters as SQL parameters. The
caller opens the pool, registers its PostgreSQL driver and closes the pool.
No schema migrations, cache, filesystem assets or workers are required.
The package still depends only on the standard library.

The platform supplies verified `RequireSystem`, `AuthorizeAny` and
`SelectContext` capabilities. They run in that order before body decoding or
warehouse access. System access remains `data-statistics` with any of
`report_browser:read`, `documents:read`, `data_quality:read` or `surveillance:read`.
`SelectContext` must enforce selection requirements, membership and expiry.
Only an explicit empty ID with no error permits legacy unscoped access.
The package never interprets caller headers as verified identity or context.

Shared `ContextNodeCode` and `ContextAliases` callbacks return strings and a
namespace-to-ID-list map. The module owns conversion, namespace filtering,
national scope rules and hierarchy matching; the host needs no module-specific
health-context model adapter. Guard errors may expose `StatusCode() int` and
optional `ErrorCode() string` / `PublicMessage() string` methods for trusted
public error metadata. Raw provider/database error messages are never returned.
Unclassified guard failures fail closed with 503.

POST filters preserve omitted/null/empty values. Malformed bodies now stop with
400 before any query (the old Gin adapter could continue querying after BindJSON
had committed 400). Multiple JSON values are rejected. The default body limit
is 1 MiB, configurable with `MaxBodyBytes`, with oversized bodies returning 413.
Existing nullable response fields and hierarchy leaf semantics are preserved;
empty top-level collections serialize as arrays.

## Integration status

This phase does not publish a tag or change the dashboard pin or adapter.
The next phase must publish v0.2.0, verify its download/checksums, pin it, then
remove the host HTTP/SQL/context adapter. Generic platform route registration
must support multiple patterns and API generation must merge package-owned
component schemas as well as operations. Module-disablement and real
Keycloak/full/core checks belong to that integration phase.

## Verification

From the repository root:

```sh
python3 infra/scripts/test-admin-units-package.py
python3 infra/scripts/test-admin-units-postgres.py
make -C backend modules-check
```

The verifier copies only this package outside the host tree and uses an empty
module cache with workspaces/downloads disabled. It verifies exact dependency
graphs, build/vet/uncached race tests and unchanged metadata. It then copies
`testdata/consumer/main.go` into a separate module and builds, vets and runs it
with race instrumentation. The consumer uses public APIs for all nine lookup
contracts, nullable filters, hierarchy construction, visibility, errors and
module version reporting.

The HTTP consumer is compiled separately and run from an empty directory with
downloads disabled. The PostgreSQL verifier creates and removes its own Docker
container, seeds only a disposable warehouse and uses a SELECT-only account.
It checks all nine routes, every repository query, NULLs, optional filters,
parameter binding, cancellation and caller-owned pool lifetime. Its external
consumer alone imports the reviewed PostgreSQL driver; the module has no driver
dependency. CI runs both verifiers.

The v0.1.0 release checker remains for the published version. Its source-parity
check will intentionally differ from this unreleased candidate until the next
release phase updates the version and expected artifact. To verify the published download
using a fresh cache and no replacement, run
`python3 infra/scripts/test-admin-units-release.py`. This explicit release
check requires network access and validates checksums, selected dependency
version, source parity and an independent public consumer.
