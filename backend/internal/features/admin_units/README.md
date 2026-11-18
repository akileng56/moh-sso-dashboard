# Administrative Units extraction

Phase 1 isolates the portable contracts and behavior. The feature remains
host-wired and has not yet entered the module registry or been independently
published.

| Source | Responsibility |
| --- | --- |
| `models.go` | Five feature-owned organizational unit, facility and tree models; existing pointer/null and JSON semantics |
| `repository.go` | Nine warehouse lookup contracts using only feature models and context |
| `service.go` | Existing list delegation and repository-backed hierarchy service |
| `hierarchy.go` | Existing hierarchy building, fallback matching and deduplication algorithm |
| `visibility.go` | Portable visibility reader/value contracts, DWH alias resolution and row filtering |
| `repository_postgres.go` | Existing DWH SQL operations; host database ownership |
| `visibility_host.go` | Health-context adapter, typed Gin context selection and compatibility constructor |
| `handler.go`, `mapper.go`, `routes.go` | HTTP behavior, response mapping, error codes and the existing nine routes |

The core has only standard-library imports. Model fields/JSON tags and hierarchy algorithm remain unchanged. SQL operations
are unchanged apart from trailing whitespace cleanup in the moved queries. The repository uses the
existing warehouse connection and `dwh.dim_org_hierarchy`; extraction does not
introduce a new schema or migration.

`NewHandlerWithPorts` injects the service and `VisibilityReader`. Bootstrap
explicitly supplies the warehouse repository and health-context adapter.
`NewHandler` remains as a compatibility constructor. The health-context adapter
copies only node code and alias namespace/external ID. It forwards contexts,
scope IDs and descendant flags, and retains provider error causes.

The host authenticates/authorizes context selection. An absent selection
retains existing unscoped behavior. A present invalid/zero UUID or missing
reader is forbidden. Exactly `UG` with descendant scope retains the existing
national exception. Every other selected scope requires nonempty
`dwh-org-unit` aliases. Invalid/unmapped/failed selected scopes return the
existing 403 error codes without data. Full/facility matching retains ancestor
identifiers; simple responses match their own unit ID. Hierarchy construction
still operates on the filtered rows and preserves existing fallback, pruning,
deduplication, ordering and null-versus-empty children behavior.

## Verification

```sh
python3 infra/scripts/test-admin-units-core.py
go -C backend test -race ./internal/features/admin_units ./internal/bootstrap ./internal/api
make -C backend modules-check contracts-check
```

The portable verifier copies only models, contracts, service, hierarchy,
visibility and core tests into a temporary module with an empty cache,
workspaces/downloads disabled and no third-party dependencies. It checks the
dependency graph, build, vet, uncached race tests and unchanged metadata.
CI runs this check.

Portable regressions cover alias namespace/whitespace/deduplication, exact
national exception and node-only behavior, provider errors, context and optional
filter values, matching/empty shapes and hierarchy fallback/deduplication.
Host tests cover all nine HTTP routes across unselected, restricted, national,
invalid, missing-adapter, unmapped, forbidden and failed lookup contexts, plus
health-context value conversion and error and request-context forwarding.

Next: registry integration and disablement/permission checks, then standalone
package verification and finally version publication/pinning with host adapters.
