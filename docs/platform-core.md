# Platform core boundaries

The platform core is the mandatory foundation used by optional applications.
Separately packaged capabilities can still be mandatory. The module inventory
and `DISABLED_MODULES` describe optional compiled applications, not a switch for
authentication, authorization or governance.

## Current implementation: phases 1–3

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

The host still owns routing, credential verification, policy/provider selection,
health-context selection middleware and infrastructure. Phases 2–3 share existing
access predicates and decision rules through portable contracts. The core runtime runs trusted compiled
code; it is not a sandbox or runtime code downloader.

Phase 2 adds portable `Identity`, `SessionStore[T]`, `AuthorizationResolver`,
`Access` and `SystemAccess` contracts. Identity carries verified subject/role
claims rather than credentials. The generic session payload type belongs to the
trusted host: `authsession.Data` still defines and serializes the token bundle
only in the Redis adapter. The host store implements the core interface and the
auth handler consumes it; the not-found sentinel is shared without changing its
existing error text.

The authentication middleware still selects/verifies credentials in the host,
then supplies `core.Identity` to `authz.NewContextForIdentity`. Existing
constructors remain compatible. Host role normalization, defaults and admin/user
flags remain unchanged. The existing DB/static resolver is adapted through
`core.ResolveAuthorization`; provider errors discard partial permission/system
grants and do not trigger a new fallback. Explicit host static-policy selection
remains available where it already existed.

Permission comparisons remain exact, role matching remains case-insensitive
with whitespace normalization, and system matching retains its previous rules.
The existing host `authz.Context`, `SystemAccess` and `ResolvedAccess` types
retain their schema ownership. Adapters preserve all system metadata and
nil/empty collections.

No OIDC verifier, HTTP/cookie parser, Redis client, token payload, credential
fallback, session TTL algorithm or persistence implementation moved into core.
Authorization contracts do not authenticate a caller: the host must verify
credentials and apply its authentication/context guards first.

Phase 3 adds `Application` and a read-only `ApplicationRegistry` contract.
The RBAC adapter preserves registry identity, launch/navigation metadata,
ownership/support information, lifecycle/integration/version, visibility and
configured access roles. This is an administrative registry view, including
inactive registrations; it is not the caller's authorized launcher inventory.
The host's named HTTP DTOs and schema identities stay unchanged.

Core now owns lifecycle status/transition predicates and review-authority
decisions, plus the existing metadata validation rules. The host derives
configuration/review capabilities from its permission catalogue. Approval and
activation require both capabilities; retirement is terminal, and suspension
requires reconfiguration/review before activation. The host still normalizes
requests, validates launch/navigation behavior and enabled access roles, and
serializes lifecycle/configuration changes with audit events in one transaction.
Reviewed manifest import, review tokens and provisioning remain host-owned.

`ContextAuthorization[ID]`, `ContextAccessResolver[ID,E]` and
`ContextDecision[E]` describe combined functional/system/context decisions.
The host supplies its ID and explanation types through generics, keeping UUID
and response-model dependencies outside core. Permission and contextual system
matches remain exact and case-sensitive; absent requirements are optional.
Permission checks precede system checks, which precede scope lookup. Resolver
errors discard partial decisions and preserve their cause. A denied scope
retains its explanation and returns the existing shared forbidden sentinel.

`ContextAssignment[ID]`, `ContextHierarchy[ID]` and
`ResolveContextScope` separate scope decisions from geographic persistence.
Only explicit descendant scope can query ancestry; same-node access requires
no hierarchy query. Assignment ordering, first-match selection, default-National
reason text and group provenance remain unchanged. Adapters reconstruct the
full existing explanation, including requested/assignment nodes and group IDs
and paths. The provider supplies already-effective assignments: expiry, group
membership, defaults and disabled-assignment filtering remain host duties.
The service rejects inactive requested nodes before loading assignments.

Alias resolution and selected-context middleware retain their current
ordering, HTTP errors and legacy handling of requests without a selection.
Observatory replication, warehouse access, provider construction and worker
startup have not changed. This phase does not create a core-only boot profile.

## Mandatory capability ownership

| Capability | Existing host owner | Target portable boundary |
| --- | --- | --- |
| Identity and sessions | `auth`, `authsession`, `sessions`, Keycloak/session adapters | Phase 2 verified-claim and generic session-store contracts; token/cookie verification stays in host adapters |
| Authorization | `authz`, `rbac`, `system_rbac` | Phase 2 resolver/access contracts and predicates; policy catalogues and DB/static implementations stay host-owned |
| Application governance | Application registry/manifests in `rbac`, `platform` APIs | Phase 3 registry descriptor, lifecycle/review predicates and metadata checks; manifest import/provisioning remain host-owned |
| Module runtime | `core` with `module` HTTP adapter | Implemented in phase 1 |
| Context authorization | `health_context` | Phase 3 combined-access and assignment/hierarchy contracts; selection middleware and geography sources remain adapters |
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

1. Add an explicit core-only boot profile and per-feature adapter/worker wiring;
   verify startup/readiness/shutdown with optional providers absent.
2. Align frontend shell/extension contracts and verify an optional application
   against the core, then decide independent release/version boundaries.

Document, template and storage-location extraction remains deferred.

## Verification

```sh
python3 infra/scripts/test-platform-core.py
go -C backend test -race ./core ./module ./internal/architecture ./internal/authz ./internal/features/rbac ./internal/features/health_context ./internal/api ./internal/bootstrap
make -C backend modules-check contracts-check
```
