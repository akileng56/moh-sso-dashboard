# Administrative Units Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units`.
Requires Go 1.25.4. The package uses only the standard library.

This package is prepared for its first standalone release. No version has
been published or selected by the dashboard yet. Local replacements report
`dev`. The dashboard retains the matching core and reports its host build
version until release verification and pinning.

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

## Host ownership

The package has no Gin/Keycloak/SQLC/PostgreSQL/cache dependencies, database
adapter, migrations, infrastructure resources or background worker.
Warehouse connections/queries, health-context selection authorization, HTTP
handlers/DTOs, system/read permissions, registry wiring and lifecycle remain
in the dashboard. Importing this library does not read a warehouse, migrate a
schema, register routes or alter runtime module availability.

## Verification

From the repository root:

```sh
python3 infra/scripts/test-admin-units-package.py
make -C backend modules-check
```

The verifier copies only this package outside the host tree and uses an empty
module cache with workspaces/downloads disabled. It verifies exact dependency
graphs, build/vet/uncached race tests and unchanged metadata. It then copies
`testdata/consumer/main.go` into a separate module and builds, vets and runs it
with race instrumentation. The consumer uses public APIs for all nine lookup
contracts, nullable filters, hierarchy construction, visibility, errors and
module version reporting.

Until the release/pinning phase, the verifier also rejects drift between the
retained portable host files and the extracted core. CI runs this verification.
