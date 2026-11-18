# Administrative Units extraction

The feature is registry-hosted as `admin-units`, reporting the host build
version pending standalone publication and pinning. All nine existing routes are registered
together. Enabled initialization requires the Administrative Units and shared
health-context handlers. Disabled initialization skips both requirements.
The portable contracts and behavior were isolated in phase 1.

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

Phase 2 moves registration into the module registry. Legacy registration runs
only when no registry is supplied. The registry and legacy paths share the
same system/health-context and Visualizer permission group helpers; enabling
the module cannot skip those guards. `DISABLED_MODULES=admin-units` removes
all nine routes after restart, preserving unrelated Visualizer endpoints,
shared health-context services and existing warehouse data. There are no
module-owned workers.

Registry regressions cover exact legacy/registry route parity, production
router disablement for all nine endpoints, unchanged unrelated Visualizer
routes, initialization/inventory, all four alternative read permissions,
system denial, missing/invalid/inaccessible context selection, provider
failures and scoped row filtering. Disabled requests touch no module or
health-context dependencies. The authenticated distribution fixture seeds only
its disposable warehouse dimension and exercises real Keycloak authentication,
selected national context, lookups, disablement and re-enablement.

Phase 3 adds `backend/modules/admin-units` with the unchanged portable core,
standard-library-only dependencies, public hierarchy construction and version
reporting. The dashboard retains its implementation until a reviewed release
is selected. The package verifier rejects drift in the retained core files,
builds/vets/race-tests outside the host tree with an empty cache and downloads/
workspaces disabled, then builds and runs an independent public-API consumer.
CI and automatic `modules-check` discovery include the package.

Next: publish v0.1.0, verify its download/checksums, pin it and replace the
retained core with package aliases/adapters.
