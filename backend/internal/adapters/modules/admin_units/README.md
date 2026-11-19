# Administrative Units host adapters

The dashboard selects
`github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units v0.1.0`
with reviewed checksums and no local replacement. The release tag is
`backend/modules/admin-units/v0.1.0`; runtime inventory reports the selected
dependency version.

| Source | Host responsibility |
| --- | --- |
| `contracts.go` | Compatibility type aliases, shared forbidden sentinel and delegates to the published service, hierarchy and visibility functions |
| `repository_postgres.go` | Existing DWH connection/query contract; SQL operations remain unchanged |
| `visibility_host.go` | Adapt shared health-context calls, validate typed selected context values, preserve causes and compatibility construction |
| `handler.go`, `mapper.go`, `routes.go` | HTTP responses, filtering orchestration, permissions and nine existing route contracts |

No retained model, repository interface, service implementation, hierarchy
algorithm or visibility policy remains in the host. The HTTP handler converts
already-filtered full rows and delegates hierarchy construction to the selected
package. The host still authenticates/authorizes requests and context selection;
the library does not grant access or infer missing ancestors.

All nine `/api/v1/visualizer/adminunits/*` routes are composed as `admin-units`.
The registry retains Data Statistics system access, validated health-context
selection and the four alternative read permissions. Enabled initialization
requires both handlers. `DISABLED_MODULES=admin-units` omits all nine routes after
restart, preserving unrelated Visualizer endpoints, shared health-context
services and warehouse data. The module owns no worker or migration.

## Verification

```sh
python3 infra/scripts/test-admin-units-package.py
python3 infra/scripts/test-admin-units-release.py
go -C backend test -race ./internal/adapters/modules/admin_units ./internal/features/health_context ./internal/bootstrap ./internal/api
make -C backend modules-check contracts-check
```

The package verifier uses an isolated empty cache with downloads/workspaces
disabled. The release verifier intentionally downloads v0.1.0 into a fresh cache,
checks the host sums and exact dependency graph, rejects local replacements and
builds/vets/runs an independent race-instrumented consumer that must report
v0.1.0. It also compares all shipped Go source files to the selected release.

Portable hierarchy and visibility regressions live in the standalone module.
Host tests cover type/sentinel identity, delegation, SQL-facing service/error
contracts, all nine HTTP routes, adapter value/cause conversion and selection
validation. Registry tests retain route parity, inventory, all accepted read
permissions, system/context failures, scoped results and complete disablement.
The real Keycloak/PostgreSQL fixture checks selected national context, nine
warehouse reads, inventory v0.1.0, disablement/re-enablement, data parity and
backup/restore in its disposable database.
