# Platform core boundaries

The platform core is the mandatory foundation used by optional applications.
Separately packaged capabilities can still be mandatory. The module inventory
and `DISABLED_MODULES` describe optional compiled applications, not a switch for
authentication, authorization or governance.

## Current implementation: phase 1

`backend/core` owns framework-independent module inventory, composition
validation, initialization, cancellation, rollback and cleanup. It uses only
the standard library. The existing `backend/module` package is a Gin adapter
that delegates lifecycle management to the core and retains route callbacks.
Existing bootstrap wiring and compiled-module APIs remain compatible. The host
inventory view retains the `module.Info` schema name while deriving its fields
from `core.ModuleInfo`, keeping generated client contracts stable.

| Contract | Responsibility |
| --- | --- |
| `core.ModuleInfo` | Existing ID/version/enabled JSON inventory |
| `core.ModuleDefinition` | Explicit dependency order and readiness/initialization callback |
| `core.Stop` | Cooperative cleanup using the caller's shutdown context |
| `core.Runtime`, `core.NewRuntime` | Validate and run the compiled-module lifecycle |

Initialization runs once and skips disabled definitions. Dependencies must
precede consumers; enabled consumers cannot require disabled dependencies.
Failure cancels the shared module context and cleans up partial acquisitions
in reverse order, using the existing independent 15-second rollback context.
Normal shutdown cancels module contexts before reverse cleanup and returns
joined errors. Cleanup is idempotent, including concurrent calls. Callbacks
must honor cancellation/deadlines; the runtime does not forcibly terminate
callbacks that ignore them.

The host still owns routing, authentication, authorization, health-context
selection and infrastructure. No security decisions have moved into a new
policy implementation in this phase. The core runtime runs trusted compiled
code; it is not a sandbox or runtime code downloader.

## Mandatory capability ownership

| Capability | Existing host owner | Target portable boundary |
| --- | --- | --- |
| Identity and sessions | `auth`, `authsession`, `sessions`, Keycloak/session adapters | Authenticated identity, session lifecycle and revocation contracts; tokens/cookies stay with trusted host adapters |
| Authorization | `authz`, `rbac`, `system_rbac` | Shared permission/system policy and evaluation contracts |
| Application governance | Application registry/manifests in `rbac`, `platform` APIs | Application identity, reviewed lifecycle, launch/access metadata and required capabilities |
| Module runtime | `core` with `module` HTTP adapter | Implemented in phase 1 |
| Context authorization | `health_context` | Selected context, assignments and scope evaluation; geography sources remain adapters |
| Audit recording | Shared audit service and primary persistence | Record required security/operational events independently of the optional Audit reader |
| Operations and bootstrap | `platform`, `bootstrap`, configuration/logging/readiness | Required capability health, version/request context, resource startup/shutdown |
| Frontend shell/contracts | Shell plus shared auth/API/config/state/types/UI/microfrontend packages | Authenticated context, application loading, navigation and extension contracts |

The current feature catalogue already marks the seven host features
`auth`, `authsession`, `sessions`, `rbac`, `system_rbac`,
`health_context` and `platform` as core. This classification is a host
ownership designation, not a claim that all their implementations are portable.
Those feature IDs cannot be disabled through the optional-module setting.

## Shared contracts versus optional applications

| Area | Required boundary | Optional implementation/application |
| --- | --- | --- |
| Users | Current identity and user-directory contracts | User administration, bulk imports and their UI |
| Clients | Application identity, governed registry and launch/access metadata | Keycloak client provisioning and administration UI |
| Notifications/email | Event/delivery contracts for consumers | Outbox workers, channel providers and administration UI |
| Health context | Assignment and selected-scope policy | Observatory replication and warehouse/geography source adapters |
| Audit/metrics | Required event recording and basic operational telemetry | Audit browsing/export and dashboard analytics |

Announcements, Administrative Units, GeoJSON, Report Scheduler, Visualizer,
Surveillance, DQA, Documents, Document Templates and Storage Locations remain
optional capabilities. The default distribution may ship them enabled.
Packaging and enablement are separate decisions.

## Dependency rules

- Portable core may import the standard library and its own subpackages.
- HTTP/Keycloak/SQL/Redis/SMTP/Observatory adapters depend on core contracts.
- The module HTTP adapter receives feature behavior through composition;
  it does not import host features or published application packages.
- Optional applications depend on the contracts they need. They do not provide
  the mandatory identity or governance authority for the platform.
- Provider credentials, raw session tokens, database connections and route
  objects belong in host adapters, not portable inventory/lifecycle contracts.

Architecture tests enforce the core import boundary and prevent feature imports
in the HTTP module adapter. CI also copies the entire core tree into an isolated
temporary module and checks build/vet/uncached race tests with an empty module
cache, workspaces/downloads disabled and no third-party dependencies.

## Remaining bootstrap work

The current distribution is not yet a minimal-core boot profile. Bootstrap still
opens Remote, DWH and DQA databases, starts the Observatory integration, creates
storage/delivery adapters and starts legacy workers. The platform operations
readiness endpoint also currently includes email/notification worker health.
Phase 1 preserves that behavior; changing disablement does not detach all of
these dependencies.

The target minimum host uses its configured identity provider, primary
persistence and session store without requiring warehouse, DQA, Observatory,
SMTP/SMS or document storage integrations. Feature readiness must be reported
separately from mandatory-core readiness. Optional workers must be cancelled and
joined before their resources close. Existing schema/migration compatibility
must be reviewed explicitly during that transition.

## Next phases

1. Separate identity/session and access-evaluation contracts behind existing
   trusted adapters, preserving security policy and token handling.
2. Separate application-governance and contextual-policy boundaries from
   provisioning, warehouse and Observatory integration.
3. Add an explicit core-only boot profile and per-feature adapter/worker wiring;
   verify startup/readiness/shutdown with optional providers absent.
4. Align frontend shell/extension contracts and verify an optional application
   against the core, then decide independent release/version boundaries.

Document, template and storage-location extraction remains deferred.

## Verification

```sh
python3 infra/scripts/test-platform-core.py
go -C backend test -race ./core ./module ./internal/api ./internal/bootstrap
make -C backend modules-check contracts-check
```
