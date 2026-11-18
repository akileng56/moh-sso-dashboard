# Modular distribution

## Phase 1 implementation

The backend now has a public module contract in `backend/module`. Composition
lives in `backend/internal/api/modules.go`. Bootstrap constructs and initializes
the registry before starting background workers or serving HTTP traffic.

Metrics is the first feature routed through the registry. Its handlers, SQL,
existing URLs and `metrics:read` permission checks retain their existing behavior.
Its implementation is independently packaged under `backend/modules/metrics`.
Other features still use their existing registration and worker wiring.
The inventory lists only features integrated with the module registry.

Set `DISABLED_MODULES=metrics` in backend configuration to omit metrics routes at
startup. The default empty value enables metrics. Unknown IDs fail startup.
This setting takes effect on restart. Frontend navigation is still managed by the
existing application registry and Keycloak metadata. Manifest synchronization is
scheduled for Phase 3.

`GET /api/v1/admin/modules` returns the normal success envelope with an array of
module IDs, versions and enabled states. The endpoint inherits session
authentication, audit and rate limiting middleware and requires administrator
access.

## Module contract

Each definition supplies identity, version, enabled state and optional dependency
IDs, initialization and route registration callbacks. Dependencies are injected
through constructors and closures. Public signatures use standard library and
Gin types so future module consumers can import the contract without importing
the host's internal packages.

Register definitions in explicit dependency order. The registry validates IDs,
versions, duplicate IDs, missing or incorrectly ordered dependencies and disabled
dependencies before initialization. It preserves this order for startup and route
registration. Disabled definitions appear in inventory but their callbacks are
skipped.

Initialization runs once. An initializer must return only after readiness and
return cleanup even when it fails after partially acquiring resources. Workers
must observe the supplied context. Cleanup must wait for them within its supplied
shutdown context. On failure the registry cancels module contexts, then calls
acquired cleanup callbacks in reverse order with a fresh bounded context. Normal
shutdown also cancels module contexts and runs cleanup once in reverse order,
before bootstrap closes database connections. Failed registries are discarded.

Routes receive protected and admin groups that inherit host middleware. The
admin group is a URL namespace. Each module must apply the appropriate resource
permissions or administrator requirement to its endpoints.

## Repository audit and next phases

| Area | Current integration | Next work |
| --- | --- | --- |
| Backend dependencies | Host module plus independently versioned metrics module | Extract additional domains after their dependency boundaries are reviewed |
| Dependency wiring | Bootstrap builds repositories, services and handlers | Move domain construction behind independent module constructors incrementally |
| Routing | Feature registrars under `internal/features`; metrics, audit reads and GeoJSON use the registry | Migrate additional registrars using the tracked coverage catalogue |
| Workers | Existing bootstrap workers plus a cleanup contract for new modules | Integrate each extracted worker with readiness and shutdown |
| Migrations | Existing golang-migrate runner executes core migrations before metrics initialization | Writing modules will need their own migration ownership and sequencing |
| Frontend | Published npm packages, single-spa, reviewed production pins and dependency lock | Track every new publishable workspace before selecting its release |
| Access and navigation | Keycloak, permission middleware and reviewed manifests in the existing system registry | Preserve authorization and lifecycle governance during each extraction |
| Deployment | Versioned images, immutable deployment bundles and staging promotion | Follow the documented release, compatibility and recovery workflow |

## Phase 2 packaging

Metrics has the public module path
`github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics`. It contains
the service, handlers, route definitions, response DTOs and repository port.
Public repository records live in its `model` package. The module imports no
dashboard implementation packages. Internal host constructors remain available
as compatibility adapters. The host SQL adapter maps its generated records into
the module contract, and the host supplies the existing metrics permission
middleware to every module route.

`backend/go.mod` pins the distribution's module version. The module directory is
separately versioned using `backend/modules/metrics/v0.1.0` style tags in the
existing upstream repository. Go downloads the tagged module during a normal
build, including the dependency-download stage of the existing Dockerfiles.
Metrics inventory reports the selected Go dependency version, independently of
the backend executable version.

For local development, create a Go workspace from `backend` with
`go work init . ./modules/metrics`. If a workspace already exists, use
`go work use ./modules/metrics`. Local workspace files are ignored by Git.
Workspace builds report the metrics version as `dev`. Use `GOWORK=off` to verify
the tagged dependency. There is no production filesystem replacement in the
distribution dependency configuration.

Metrics owns no tables and ships no migrations. Its SQL adapter reads tables
owned by the core. Existing migration files, versions and migration tracking
tables retain their ownership. Core migrations continue to complete before
module initialization, using the existing golang-migrate locking and dirty-state
handling. Module packaging adds no schema upgrade or data rewrite for either a
fresh installation or an existing database. Future modules that own tables must
define a migration stream before they can be extracted.

`make -C backend modules-check` vets and tests the independently packaged module.
The normal backend test command and backend release/PR checks include this step
because `go test ./...` on the host does not traverse nested Go modules.
The API contract generator loads the module's handler source so published
handlers retain their typed response schemas in the generated contract.

To release an updated module, verify its standalone tests and dashboard
integration using a local workspace. Commit the module changes, create the
directory-prefixed semantic version tag and push that tag to the upstream
repository. Verify download from a separate consumer with `GOWORK=off`, then
update the pinned dependency with `go get`, regenerate API contracts and commit
the distribution integration. See `backend/modules/metrics/README.md` for the
public constructor and route contracts.

## Verification

Run from `backend`.

```sh
make modules-check
go test ./module ./internal/api/... ./internal/bootstrap ./internal/config ./internal/architecture
go test -race ./module ./internal/api
(cd modules/metrics && GOWORK=off go test -race ./...)
go run ./cmd/cli openapi generate
go test ./internal/apidocgen
go build ./cmd/server
```

Lifecycle regression tests cover composition validation, partial initialization
rollback, cancellation, disabled callbacks and idempotent cleanup. Router tests
cover inventory authentication and administrator access, disabled metrics routes
and existing metrics permission checks.

Phase 2 verification passed for the full backend suite, standalone module tests,
race checks, vet, Linux amd64 compilation, temporary-workspace integration and
dependency downloads using only `go.mod` and `go.sum`. A separate temporary Go
project downloaded `v0.1.0`, verified its reported version and registered all 19
guarded routes. The published dependency is recorded in the server binary's Go
build metadata. Generated API contract changes are limited to the 19 handler
source paths.

Full Docker image and database-backed runtime checks remain pending because the
local Docker daemon was unavailable during Phase 2 verification.

## Phase 3 manifests

The Systems page now provides administrator-only manifest upload, validation,
preview and import. The implementation uses the current Postgres system registry,
navigation model, access-role tables and lifecycle workflow. A serializable
transaction applies registry metadata, access-role changes and audit records.
Stale previews require review again. Repeated imports are idempotent.

External applications use their deployed URLs. Internal applications reference
an enabled compiled module. Separately hosted single-spa applications use
`/apps/extensions/<application-id>` routes and a browser module entry URL. The
runtime endpoint checks application access and activation before returning that
entry point. The existing shell host supplies the normal application props and
handles mounting and unmounting.

The implemented manifest schema uses the registry's own-client access-role model.
Keycloak client provisioning and user role assignment use the existing explicit
administration workflows. Startup discovery preserves manifest-managed registry
configuration. No additional database migration is required because documents
are stored in existing system metadata. The schema, examples, API contracts and
configuration ownership rules are documented in `examples/manifests/README.md`.

Phase 3 verification passed the full backend suite with isolated Postgres schemas,
transaction and lifecycle race tests, vet, standalone module checks and generated
OpenAPI validation. The backend Docker image built successfully and its version
command ran successfully. Frontend workspace production builds, type checking,
lint and architecture checks passed. Browser coverage includes the manifest
review/import flow on desktop and mobile, dynamic application mount, refresh,
unmount and denied access, plus 50 existing portal page regression checks.

The published-npm production build could not resolve the existing pinned
`@moh-sso/announcements@1.3.1` package. Workspace builds were used for runtime
verification; published-npm deployment verification requires that release to be
available. Live Keycloak provisioning and production deployment were not run.

## Phase 4 distribution packaging

Production now consumes released npm modules (`1.3.2`, with report-scheduler
`0.4.2`) using a committed npm lock and integrity-checked `npm ci` installation.
This resolves the unavailable `1.3.1` pins found in Phase 3. Development continues
to use workspace modules. Backend builds use the published metrics dependency
with `GOWORK=off`, checksum verification and Go build/download caches.

The specialized frontend Dockerfile caches toolchain and published module
installation before copying source. BuildKit keeps the Go and npm download
caches across builds. The distribution Compose entry point reuses the production
gateway and services, builds versioned application images, persists core,
Observatory, Redis and upload data, and supports the existing external Keycloak.
Startup migration behavior remains owned by the existing backend lifecycle.
See `infra/distribution.md` for prerequisites, commands and upgrade guidance.

Phase 4 verification passed registry-backed frontend builds, the production npm
shell type check, lint, architecture checks, 19 module-source tests and 54 browser
tests. Both Docker images built successfully; the server version command and
the frontend production runtime checks passed. A repeated dependency-stage build
reused the toolchain and published-module layers. All 14 Compose/environment
contracts and the existing disposable Postgres migration/registry checks passed.
The new isolated Compose smoke test verified the real frontend image, secret
mounts, database/Redis readiness, module versions and core database persistence
across container recreation. It cleaned up its own containers and volumes.
Authenticated external Keycloak and production deployment were not exercised.

## Phase 5 release and installation verification

The existing GHCR image workflow now builds the frontend with the pinned npm
Dockerfile. Its validation dependency includes the new distribution installation
gate, which builds both images before exercising runtime, authentication and
deployment recovery. Component version/SHA tags and the existing image-pair
deployment bundle remain the release mechanism.

Bundles record the compiled module versions, exact frontend versions and hashes
of `go.mod`, `go.sum` and the production npm lock. Distribution server deployments
use digest-pinned images with no source-build settings. Runtime verification
compares the complete frontend module set with the approved pins. Deployment
snapshots retain their Compose project name, and an existing installation cannot
silently change topology/project. Legacy deployment records remain compatible.

Use `infra/server/deploy.distribution.json.example` for separate staging and
production configurations. The existing trusted-build selector and GitHub
environment deployment workflow promote the same successful main-build bundle
by run ID. The operator guide describes staging-first settings, secret files,
backup hooks, schema compatibility and rollback.

Verification passed 32 deployment/archive contracts, 14 Compose/environment
contracts and 8 trusted-build/promotion tests. Workflow lint and the new backup
hook's shell checks passed. Isolated tests using the real backend and Keycloak
26.0.5 verified fresh migrations, real token/role authorization, module inventory,
manifest import/approval, audit idempotence, startup preservation, graceful
shutdown, restart, and restoration of both databases and uploads. Container
transaction tests passed for legacy and distribution deployments, including
failed-version detection, automatic/operator rollback and data preservation.
Temporary test resources were removed. These are local verification results;
publishing images and deploying to real staging/production require the merged
workflows and configured operator environments and were not triggered here.

## Phase 6 expansion and module workflows

The next bounded expansion integrates audit read APIs and GeoJSON serving into
the compiled registry. Their implementations remain hosted by the backend and
report its build version. Metrics remains independently packaged and versioned.
`DISABLED_MODULES` accepts `metrics`, `audit` and `geojson`; existing route paths,
permissions, export rate limits and response contracts are preserved. Disabling
audit read APIs leaves the host's security/event recording service active.

The source catalogue in `docs/module-catalog.json` tracks all 23 backend features
and 23 publishable frontend packages. [Module coverage](module-coverage.md) is
generated from it and the reviewed production npm pins. Entries distinguish core,
planned, registry-hosted and published stages and record dependencies, current
worker/migration owners and next extraction work. Remaining domains are explicit
incremental work, rather than being reported as independently packaged.

[Developer workflows](module-development.md) cover registry integration, public
ports, worker transfer, schema ownership, standalone Go/npm release verification,
dependency pinning and contract checks. [Operator workflows](module-operations.md)
cover runtime inventory, disablement, affected applications and rollout/recovery.
Report Scheduler is tracked as the next worker-bearing candidate after its
provider/delivery boundaries and worker lifecycle are reviewed.

CI validates catalogue coverage against source directories, publishable
workspaces, production pins/lock and the generated guide. A Go architecture test
checks the catalogue against the actual compiled inventory. Standalone module
checks now discover direct `backend/modules/*/go.mod` packages automatically.

Phase 6 verification passed the full backend suite, vet, standalone module checks
and registry/API/architecture race tests. OpenAPI validation remains current at
366 endpoints, with no public contract changes. Seven catalogue regression tests
and 19 frontend module-source tests passed, along with Compose/environment and
workflow checks. Development npm pins now reference available published releases,
and the shell typechecked against that isolated development installation.
The fresh backend image passed real Keycloak, manifest, backup/restore and restart
checks, including enabled/disabled audit and GeoJSON APIs and continued core audit
recording. Temporary fixture resources were removed. No additional independent
Go package release or live deployment was performed in this expansion step.

## Report Scheduler extraction: lifecycle phase

The scheduler now appears as `report-scheduler` in the compiled module inventory.
Its dependency ports and host adapters were isolated in the first extraction
phase. The registry now owns both its public delivery-download routes and
protected scheduler routes, together with worker initialization and cleanup.
`DISABLED_MODULES=report-scheduler` skips all three.

Initialization requires a handler, worker and successful persistence heartbeat
before any job claims. Cleanup cancels the worker and waits for its active cycle,
subject to the host shutdown deadline. The former bootstrap worker start has been
removed. Existing permissions, download-token checks, rate limits, SQL claims,
retries and migration ownership remain in their current domains.

Verification includes readiness failure, canceled initialization, duplicate
starts, in-flight shutdown, shutdown deadlines, rollback cleanup and route
disablement. The authenticated container fixture also checks scheduler role access,
heartbeat preservation while disabled and a fresh worker when re-enabled.
Standalone packaging, publication and a pinned dependency remain the next phases.

## Report Scheduler extraction: standalone package phase

`backend/modules/report-scheduler` now contains a standalone Go module with public
types, service, worker, dependency ports, Health BI client and PostgreSQL adapters.
HTTP authorization and registry integration remain host responsibilities. Its
dependency graph contains UUID and PostgreSQL libraries and no dashboard imports.

`schema-contract.json` explicitly records host ownership of scheduler migrations
40 through 47 and shared directory/hierarchy reads. No migration files or migration
runner are copied into the package. An isolated check copies only the module out
of the repository, disables Go workspaces, builds/vets it, applies those host-owned
migrations to disposable PostgreSQL and runs race tests. External-package tests
exercise concurrent claims, retries, stale claims and idempotent delivery.

The dashboard continues using its phase-2 implementation until a released
dependency can be selected. The catalogue accurately keeps the runtime stage as
registry-hosted. Phase 4 publishes/verifies an immutable version and replaces the
retained host implementation with package adapters and a reviewed dependency pin.

## Report Scheduler extraction: released dependency integration

Report Scheduler `v0.1.0` is published under the immutable nested module tag
`backend/modules/report-scheduler/v0.1.0`. A clean-cache download verified its
checksums and source commit. The host now selects that version in `backend/go.mod`
and records its checksums in `backend/go.sum`, without a local replace directive.

The retained host scheduling/execution/provider/persistence implementation has
been removed. The feature package supplies aliases and thin constructors, HTTP
handlers and permission/rate-limit adapters. Bootstrap keeps the platform
outbox/user/notification translations. Portable implementation tests live in the
standalone module; host tests exercise HTTP and lifecycle integration against the
selected dependency. Public OpenAPI schema names remain stable across extraction.

The compiled `report-scheduler` inventory reports the selected dependency version,
and deployment bundle metadata includes its exact pin. All schema migrations
remain core-owned. Build a new backend image to consume a release; publishing a
tag does not change a running deployment. The runtime fixture verifies the actual
compiled version, real Keycloak permissions, public/protected route disablement,
stopped heartbeat and worker restart/re-enablement.


## Audit extraction: released dependency integration

Audit `v0.1.0` is published under `backend/modules/audit/v0.1.0`, pointing to
release commit `bd1bc5d095eaf5db66fccc108b0f1686defb897c`. A clean download in
the verification environment passed the official Go signed checksum log.
The verified archive was transferred to the host using a temporary file proxy;
SHA-256 transfer integrity and full release-source parity were checked before
pinning. The dashboard selects this version and its verified checksums in
`backend/go.mod` and `backend/go.sum`, without local replacements. No persistent
proxy configuration or checksum-verification exception was introduced.

Host public types and constructors now delegate to the package. SQLC adapters,
optional Redis, HTTP validation/permissions, export limits, core event recording
and migrations retain host ownership. The compiled inventory reports the selected
package version; deployment bundle metadata records the pin.

The live export check exposed an existing PostgreSQL error: nullable export
filters lacked explicit parameter types. The host SQL query now casts text and
UUID filters consistently with the list query, and its bindings were regenerated
with sqlc v1.30.0. JSON exports are verified with both absent and populated
action filters, and CSV exports retain their existing eight columns. The
published portable package did not require a change.

Verification passed with `GOWORK=off`: full backend tests, targeted
Audit/API/bootstrap race checks, vet, standalone package/external-consumer
checks, module checks, all 366 OpenAPI routes, catalogue checks and seven tests,
and all 32 distribution bundle tests.

The real Keycloak distribution fixture passed against a freshly compiled Linux
binary packaged on an existing local backend runtime image. It verified the
actual Audit version, administrator/viewer permissions, a seeded typed Redis
cache hit, JSON/CSV exports, module disablement, continued core event recording,
backup/restore and restart recovery. The host recorder deliberately invalidates
Audit caches when recording events, so the fixture seeds a distinguishable cached
response rather than assuming two unrestricted reads return identical records.

The standard production Dockerfile subsequently built successfully after
outbound access recovered. Docker downloaded and verified all Go modules and
completed the full multi-stage build without Dockerfile changes. The local
`moh-sso-audit-phase4:backend` image is Linux ARM64, built from clean integration
commit `477680dd287262ee1727a7b7288a4a342bf98ca9`, with image ID
`sha256:ff6d4a3948c03bf881481c28bf7f52f1dda11e8b8dbfa3decbbd1697d57389c7`.
Its OCI revision label matches that source commit.

The additional authenticated fixture run against this exact production image
was stopped before backend readiness because Docker's local disk filled. Its
disposable containers/volumes were cleaned up. The earlier complete fixture
against the freshly compiled integration binary remains the runtime validation;
the production-image fixture must be repeated after sufficient Docker disk
space is available. Only disposable Audit images and selected build-cache
records were removed; existing application volumes were not changed.

The integration branch has been pushed to origin. Building or pushing source
does not publish this image to a registry or deploy a running installation.


## GeoJSON extraction: asset boundary phase

GeoJSON is the next extraction after Audit. It already participates in registry
enablement/disablement and has no worker or migration lifecycle. Phase 1
introduces standard-library-only `Asset` and `AssetReader` contracts and a
host directory adapter; bootstrap injects the adapter into the HTTP handler.
The existing directory constructor remains compatible.

The three public layer names and their filenames remain fixed. Authentication
and the existing any-permission gate run in host middleware before asset
access. The HTTP adapter preserves raw GeoJSON media type, cache/nosniff
headers, safe error envelopes, conditional responses and byte ranges, and
closes streams on every completion path.

The runtime remains registry-hosted and reports the host version. Assets stay
in the host Docker image. Portable layer resolution, standalone verification,
publication and selection of a released Go dependency are subsequent phases.

Phase 1 verification passed the full backend test suite, targeted
GeoJSON/API/bootstrap race checks, vet, standalone module checks, catalogue
validation and its seven tests. OpenAPI validation reports all 366 endpoints
current. Regression coverage includes all three permission alternatives, denial
before storage access, actual bundled files, conditional/range requests and
stream closure on failure.


## GeoJSON extraction: portable service phase

Phase 2 moves the fixed public layer-to-filename mapping into `ResolveLayer`
and portable read orchestration into `Service.Open`. The service accepts the
existing asset-reader contract, preserves modification times and transfers
successful streams to its caller. It closes streams returned on failed or
canceled reads.

Unknown names are classified as `ErrUnknownLayer` before storage access.
Reader failures, missing readers/content and cancellation match `ErrLoadFailed`
with stable error text; original causes and close failures remain discoverable
through `errors.Is/As`. The HTTP adapter translates these errors into the
existing 404/500 envelopes and preserves streaming headers, range/conditional
handling and host authorization. A storage error cannot turn a resolved layer
into an HTTP 404.

The service and its reader contract contain only standard-library types. The
host retains directory access, bundled assets, Docker packaging and permission
middleware. The module remains registry-hosted at the host version; standalone
packaging and release selection are subsequent phases.

Phase 2 verification passed an isolated temporary module containing only the
reader contract, service and portable tests: build, vet and uncached race tests
completed with external downloads disabled. The full backend suite, targeted
GeoJSON/API/bootstrap race tests, vet, existing standalone module checks,
catalogue validation and its seven tests also passed. All 366 OpenAPI routes
remain current.


## GeoJSON extraction: standalone package phase

The standalone release candidate at `backend/modules/geojson` declares
`github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson` and Go 1.25.4.
It copies the portable layer/asset contracts and service without changing their
behavior, and adds dependency-version reporting and external-package tests.
There are no third-party dependencies, map assets, storage/HTTP adapters,
migrations or workers in the module. A checksum file is unnecessary until an
external dependency is introduced.

`infra/scripts/test-geojson-package.py` copies only this module outside the
dashboard with an empty module cache, disables workspaces and dependency
downloads, verifies its dependency graph and runs build/vet/uncached race tests.
It also compiles and runs a separate consumer module, checking that public APIs
and local-replacement version reporting work without host files. All dependency
metadata remains read-only. The PR workflow runs the verifier automatically.

The dashboard still uses its retained phase-2 implementation, remains
registry-hosted and reports the host version. It has no GeoJSON dependency pin
or production replacement. Publishing and verifying an immutable release,
selecting its checksums and replacing retained code with adapters are the next
phase. Host assets, directory access, permissions, HTTP serving and Docker
packaging keep their existing ownership.

Phase 3 verification passed the isolated empty-cache/no-download verifier,
uncached race tests, external-package tests and a separately built consumer.
The consumer verified local replacement versions report `dev`, including Go
build metadata containing `(devel)`. Portable source/tests match the retained
host implementation byte for byte. Existing module checks, the full backend
suite, vet, all 366 OpenAPI routes, catalogue validation and its seven tests,
and workflow actionlint checks also passed.


## GeoJSON extraction: released dependency integration

GeoJSON `v0.1.0` is published under the immutable nested module tag
`backend/modules/geojson/v0.1.0`, pointing to release commit
`f6d681c5db790653f91658be3e44772075e6db1f`. A clean-cache official Go proxy
download passed signed checksum verification, release-commit validation and
complete source parity before selecting the host dependency.

The dashboard selects the released dependency and checksums in `backend/go.mod`
and `backend/go.sum`, without a local replacement. The retained host asset
contracts and portable service implementation have been removed. The feature's
`package.go` provides aliases, error sentinels and thin constructors/resolution
functions. Existing host tests now exercise the selected dependency.

Storage, map files, Docker packaging, HTTP streaming/headers and authorization
remain host adapters. All three existing layer names, safe 404/500 envelopes,
range/conditional responses and permission alternatives retain their contracts.
The compiled inventory reports the selected library version; distribution bundle
metadata records the exact pin. Disabling the module omits its routes without
affecting core event recording or unrelated modules.

The full backend suite, targeted GeoJSON/API/bootstrap race checks, vet, module
checks, all 366 OpenAPI contracts, standalone external-consumer verification,
catalogue validation and its seven tests, and all 32 bundle tests passed with
the released dependency selected.

The standard production Dockerfile built the integration candidate successfully
as `moh-sso-geojson-phase4:backend` (Linux ARM64), image ID
`sha256:f93fdc92e0b8328e508862912e6fe7b04a16c79f885ea634b417c509d0c65f4d`.
This review image was built from the working integration tree with `DIRTY=true`.

The full disposable real-Keycloak fixture passed against that image. It checked
the actual selected GeoJSON version, all three packaged nonempty map layers,
GeoJSON/cache/nosniff headers, byte ranges and conditional 304 responses,
authentication, roleless-user denial and existing realm-user map access.
Default realm-user grants intentionally permit map reads; the fixture uses a
separate roleless principal to test denial without changing host policy.

The same fixture passed Audit cache/export checks, application lifecycle review,
backup/restore, clean shutdown/restart recovery, module disablement, continued
core audit recording and scheduler stop/resume. Disposable fixture containers
and volumes were removed on completion. The library release and source push
do not publish this image to a registry or deploy a running installation.


## Announcements extraction: dependency boundary phase

Announcements is the next selected extraction. Storage Locations, Document
Templates and Documents are deferred and unchanged. Phase 1 introduces
feature-owned user, group-member and client-access models, narrow audience and
attachment-storage interfaces, and a single-operation notification publisher.
The service no longer imports users, RBAC, health-context, Keycloak, shared
notification-service or storage implementation packages directly; compatibility
constructors and identity translations live in host adapters.

Bootstrap composes these ports explicitly. Actor-authorized health-context
resolution and stored-audience expansion remain distinct host operations.
Existing enabled-user filtering, fallback group recipients, email deduplication,
client matching, attachment checksums/orphan cleanup and best-effort delivery
are preserved. Context-aware host calls receive request contexts; legacy user
list/detail calls only reject cancellation before starting because their
underlying signatures lack context support.

The feature remains planned/host-wired in the catalogue. SQLC persistence,
configuration, notification payload types, email normalization, HTTP policy and
host delivery workers are not portable or transferred yet. Subsequent phases
must isolate those models/values, register all public/protected/admin routes and
verify disablement together, then build an independent module before release
publication and host pinning.

Phase 1 verification passed targeted Announcements/bootstrap/API race tests,
the full backend suite, vet, existing independent-module checks, catalogue
validation and its seven tests, and all 366 current OpenAPI endpoints.
Audience/storage contracts also built and passed vet outside the host tree with
only the existing UUID dependency and module downloads disabled. Regression
tests exercised actor scope and contextual errors, user/group translation and
filtering, both client identifiers, cancellation before legacy calls, attachment
checksums and cleanup, notification payload forwarding, and permission denial
before dependency access.


## Announcements extraction: portable models and delivery phase

Phase 2 introduces 20 feature-owned persistence model/input types, local enum
values, service inputs and a 53-method repository contract without SQLC or host
model imports. Standard-library nullable values and UUIDs retain existing
null-versus-zero semantics. Explicit host converters preserve every field,
including unknown enum values and nullable values with nonzero inactive data.

A translating repository adapter wraps the existing SQLC implementation.
Underlying SQL queries, transaction handling and repository method bodies
remain unchanged. The constructor now returns the adapter; model translation
is contained in host-only files.

Notifications, deliveries and email attachments now use feature-owned payloads.
The host publisher converts envelopes, channels, recipient/template/payload
maps, schedules, retries and raw/base64/path attachment representations. The
host email normalizer retains existing validation/limits. Publication still
uses best-effort notification forwarding and existing event labels.

The service consumes a non-secret `Settings` snapshot instead of host Config.
Host constructors and configuration-based construction remain available through
adapters; a portable constructor accepts only repository, dependency ports and
settings. The service, attachment orchestration, inputs and contracts now build
without host files. HTTP request/response DTOs and mapping stay in the host.

`infra/scripts/test-announcements-core.py` checks a temporary isolated module's
dependency graph and runs read-only build/vet/uncached race checks with module
downloads disabled. Its tests exercise create/publish, delivery metadata,
repository errors and normalization delegation. The PR workflow runs this check.

Announcements remains planned/host-wired. Registry integration for all route
groups, module-disablement checks, independent module packaging and published
version selection remain subsequent phases. No migration or delivery-worker
ownership changed, and the deferred document/storage modules were not modified.

Phase 2 verification passed the isolated portable-core dependency/build/vet/race
checks, targeted Announcements/bootstrap/API race tests, the full backend suite,
vet, existing module checks, catalogue validation and its seven tests, workflow
actionlint, and all 366 current API contracts. Host regression tests exercised
all 20 model round trips with both nullable validity states, unknown enum
strings, repository context/error propagation, nil/empty collection shape,
notification wire/byte payload parity, normalization limits and non-secret
settings conversion. SQL definitions, migrations and generated SQLC files are
unchanged.

### Announcements phase 3 — registry integration and disablement

Announcements is registry-hosted as `announcements`, reporting the host build
version pending independent packaging. All 26 public, protected and admin
routes are composed through the registry. Legacy registrars run only without
a registry, so disabled modules cannot leak routes through the compatibility
path. Existing permissions and write/publication rate limits stay attached to
the same handlers. Enabled initialization requires a handler; disabled
initialization requires no Announcements dependencies.

`DISABLED_MODULES=announcements` removes all announcement endpoints, including
public listings and both attachment download paths, after restart. Records,
audience mappings and stored attachments remain available when re-enabled.
Announcements owns no worker. Shared notification workers and queued deliveries,
core audit recording and migrations retain their host lifecycle. Disablement
does not cancel previously queued notifications.

The runtime catalogue now marks Announcements registry-hosted. Independent
packaging, external-consumer verification, publishing and host pin selection
remain subsequent phases. No SQL definitions, migrations or generated SQLC
files changed, and deferred document/storage features remain outside this phase.

Phase 3 verification passed the portable-core build/vet/race check, targeted
Announcements/bootstrap/API race checks, the full backend suite, vet, all
packaged-module checks, catalogue validation and its seven tests, and all 366
API contracts. The standard production Dockerfile built successfully. Its
disposable real-Keycloak/PostgreSQL fixture passed publication and permission
checks, CSV attachment uploads and both download paths, disablement of all
three route groups, and re-enablement with records and attachment bytes intact.
Existing manifest lifecycle, database/upload backup restore, core audit
recording and scheduler stop/resume checks also passed.

### Announcements phase 4 — standalone package and external consumer

`backend/modules/announcements` contains the unchanged portable lifecycle,
audience, persistence, attachment and notification contracts/core. It depends
only on the standard library and UUID v1.6.0; its module path is
`github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements`.
It exposes dependency version reporting with `dev` for local replacements.

The package has no HTTP/SQLC/Keycloak/cache adapter, schema migration or worker.
Existing host adapters retain authentication/authorization, storage, SQL
transactions, email normalization, notification execution and core audit.
The dashboard retains its current core until a published version is selected;
the standalone verifier rejects drift between the two portable copies.

The package verifier copies the module outside the host tree with workspaces
and network downloads disabled. A private cache contains only the pinned UUID
artifacts checked against reviewed checksums. It checks dependency graphs,
build/vet/uncached race tests and unchanged metadata. An independent consumer
builds and runs the public constructor, models, lifecycle, best-effort
notifications, attachment streams/checksums, context and error propagation.
CI runs the verifier; `modules-check` automatically discovers the new module.

Announcements remains registry-hosted in the runtime catalogue. Publication,
immutable tag verification, host Go pin selection and replacement of retained
host implementation with adapters are the next phase. No host Go pin, runtime
version, route, migration or worker ownership changed.

Phase 4 verification passed isolated build/vet/uncached race checks with the
exact UUID-only dependency graph, checksum verification, immutable dependency
metadata and the independent public-API consumer. Portable tests cover audience
filtering/deduplication, group fallbacks, health-context scope, lifecycle
metadata, notification best-effort behavior, attachment checksums/orphan cleanup,
configured upload limits and missing email-normalizer behavior. The retained
host-core verifier, full backend suite, backend vet, all packaged-module checks,
all 366 API contracts, catalogue validation/seven tests and workflow actionlint
also passed. Publication and dashboard dependency integration were not performed.

### Announcements phase 5 — publication and pinned host integration

Published `backend/modules/announcements/v0.1.0` upstream from the verified
phase-4 commit `e384437e`. The host selects
`github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements v0.1.0`
with Go checksum verification and no local replacement. Runtime inventory
reports the selected dependency version rather than the host build version.

The retained models, contracts, lifecycle, audience and attachment core have
been removed from the host. Compatibility aliases point to the selected
package; a thin composition wrapper embeds its service. Existing startup-only
setters recompose ports before serving requests. SQLC/postgres operations,
mapping, identity/storage adapters, HTTP permissions/rate limits, normalization,
shared delivery workers, core audit recording and migrations remain host-owned.

Portable business regression tests remain in the independent module. Host
tests retain SQLC/model and wire-payload regression coverage and add public
delegation/compatibility composition checks. CI runs the standalone package
verifier and packaged-module checks. The obsolete retained-core verifier was
removed.

`infra/scripts/test-announcements-release.py` downloads v0.1.0 into a fresh
module cache, verifies reviewed host checksums and absence of replacements,
builds/vets/runs a separate consumer with race instrumentation and checks its
reported dependency version. It compares all shipped Go sources to the release.
This explicit publication check requires network access; the package verifier
continues to run with downloads disabled.

Phase 5 verification passed the fresh-cache published consumer and checksum/
source-parity check; the isolated package build/vet/uncached race check; targeted
Announcements/bootstrap/API race tests; the full backend suite and vet; all
packaged-module checks and 366 API contracts; catalogue validation/seven tests;
32 CD tests and workflow actionlint. The standard production Dockerfile built
successfully using the published dependency. Its disposable real-Keycloak/
PostgreSQL fixture verified inventory v0.1.0, permissions, publication, both
stored attachment download paths, all-group disablement and re-enablement with
records/bytes preserved, backup/restore, core audit recording and scheduler
stop/resume. SQL definitions, migrations and generated SQLC files are unchanged.
The release tag is upstream; publication does not deploy the backend.

### Administrative Units phase 1 — portable contracts and visibility ports

Created `feat/extract-admin-units` from the completed Announcements integration.
Administrative Units remains planned/host-wired. Its five models and nine
repository contracts now live in portable files. Existing PostgreSQL queries
remain in a host adapter; the hierarchy builder moved unchanged from the HTTP
handler to the portable core.

A feature-owned visibility port returns only node code and alias namespace/
external ID. Portable resolution retains DWH alias requirements, namespace/
whitespace/deduplication behavior and the exact UG-descendants exception.
Filtering retains ancestor matching for full/facility rows and direct matching
for simple rows. The host adapts validated selected UUID/scope context and shared
health-context calls, classifies forbidden errors and preserves provider causes.
Existing unselected behavior, all nine routes, error envelopes and hierarchy
semantics remain unchanged. Bootstrap composes the ports explicitly and the
previous constructor remains available for compatibility.

The catalogue now identifies the actual warehouse dependency (the existing DWH
connection), alongside health_context. Database/migration ownership and query
contents are unchanged. No registry module or new worker is introduced.

`infra/scripts/test-admin-units-core.py` builds/vets/race-tests the portable
core in a temporary standard-library-only module with workspaces/downloads
disabled and an empty module cache. CI runs the check. Portable tests cover
hierarchy and visibility behavior; host regressions exercise all nine route
families against allowed/denied/missing mapping and context cases.
Registry integration and disablement checks remain the next phase.

Phase 1 verification passed the isolated standard-library-only build/vet/
uncached race check, targeted Administrative Units/bootstrap/API race tests,
the full backend suite, backend vet, existing packaged-module checks, all 366
API contracts, catalogue validation/seven tests and workflow actionlint.
Model fields/JSON tags and the hierarchy algorithm were compared with the
preceding commit and are unchanged. SQL implementation bodies match after
trailing whitespace cleanup in three moved SELECT lines. Registry
integration, standalone publication and deployment were not performed.

### Administrative Units phase 2 — registry integration and disablement

Administrative Units is registry-hosted as `admin-units`, reporting the host
build version until independently packaged/published. Enabled initialization
requires both its handler and the shared health-context handler. Disabled
initialization and registration are skipped.

All nine `/api/v1/visualizer/adminunits/*` routes now enter through the
registry. Legacy registration is gated on the absence of a registry. Shared
host route-group helpers retain the same Data Statistics system requirement,
health-context selection validation and OR read permissions (report browser,
documents, data quality or surveillance). Unrelated Visualizer routes retain
their registration and guards.

`DISABLED_MODULES=admin-units` removes all nine routes after a backend restart.
The module owns no workers. Shared health-context services, warehouse rows and
schema ownership remain unchanged; re-enablement restores existing lookups.

Unit regressions exercise production-router route parity and disabled 404s,
unrelated Visualizer preservation, required enabled dependencies, host version
inventory, anonymous/system/permission denials, all read alternatives, selected
context requirements/errors and scoped rows. The real authenticated fixture
creates a small warehouse dimension only in its disposable database and checks
all nine reads, context selection, disablement and data parity after restart.
The catalogue now marks the feature registry-hosted. A catalogue negative test
now explicitly marks its malformed entry planned instead of depending on the
first feature's current stage. Standalone packaging remains the next phase.

Phase 2 verification passed the portable-core dependency/build/vet/uncached
race check, targeted Administrative Units/health-context/bootstrap/API race
tests, the full backend suite and vet, all existing packaged-module checks and
366 API contracts, catalogue validation/seven tests and workflow actionlint.
The standard production Dockerfile built successfully. Its disposable real
Keycloak/PostgreSQL fixture passed all nine warehouse reads, system and
authentication denial, required/selected national context, inventory,
disablement and re-enablement with identical data. Backup/restore, core audit
recording and scheduler stop/resume checks also passed. No production SQL,
migrations, dependency pins or worker ownership changed.

### Administrative Units phase 3 — standalone package and external consumer

`backend/modules/admin-units` contains the unchanged portable models,
repository/service contracts, hierarchy and visibility core with their
regression tests. The module path is
`github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units`; Go 1.25.4
and the standard library are sufficient. A public `BuildHierarchyTree`
wrapper lets host adapters delegate construction after scoping caller-supplied
rows. Dependency version reporting returns dev for local source replacements.

The package owns no HTTP, database/identity adapters, migrations or workers.
The host retains its core until a verified release is selected. The standalone
verifier rejects drift between matching portable source/test files, copies the
package outside the host tree and checks exact module graphs, build/vet/
uncached race tests and unchanged metadata with an empty cache and workspaces/
downloads disabled. A separate consumer module builds/vets/runs public APIs
with race instrumentation, covering all nine lookups, optional filter values,
hierarchy construction/deduplication/nullability, visibility and error causes.
CI runs this check; modules-check discovers the new nested Go module.

Administrative Units remains registry-hosted with the host build version.
Publication, download/checksum verification, host Go pin selection and replacing
retained implementation with adapters remain the next phase. No host Go pin,
runtime routes, SQL, migration or worker ownership changes in this phase.

Phase 3 verification passed the isolated standard-library-only dependency graph,
build/vet/uncached race tests and separate public-API consumer with race
instrumentation, unchanged dependency metadata and retained-core parity checks.
The full backend suite, backend vet, all packaged-module checks, all 366 API
contracts, catalogue validation/seven tests and workflow actionlint also passed.
No version was published or pinned, and no runtime deployment was performed.

### Administrative Units phase 4 — publication and pinned host integration

Published `backend/modules/admin-units/v0.1.0` upstream from the verified
phase-3 commit `d299ea76`. The host selects
`github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units v0.1.0`
with reviewed Go checksums and no local replacement. Runtime inventory now
reports the selected dependency version instead of the host build version.

Removed the retained host models, repository/service contracts and
implementations, hierarchy algorithm and visibility policy. Compatibility names
refer to published types/functions and the same forbidden sentinel. The HTTP
adapter delegates hierarchy construction after existing row scoping.
Warehouse queries, health-context authorization/selection, HTTP response
mapping, system/read permissions, registration, migrations and lifecycle remain
host-owned and unchanged.

Portable regressions remain in the independent package; host tests retain the
nine-route/adapter coverage and verify type/sentinel identity, public delegation,
request contexts and error causes. The obsolete retained-core verifier and CI
step are removed. The standalone package verifier and modules-check remain in
CI, while CD and real-runtime checks require the v0.1.0 dependency/inventory.

`infra/scripts/test-admin-units-release.py` downloads the published version
into a fresh module cache, validates exact reviewed checksums and the selected
dependency graph, rejects replacements and builds/vets/runs an independent
consumer with race instrumentation and v0.1.0 version reporting. All shipped Go
sources are compared to the release. This explicit release check requires
network access; the isolated package verifier keeps downloads disabled.

Phase 4 verification passed the fresh-cache published consumer and exact
checksum/dependency/source-parity check, isolated package build/vet/uncached
race tests, targeted Administrative Units/health-context/bootstrap/API race
checks, the full backend suite and vet, all packaged-module checks and 366 API
contracts, catalogue validation/seven tests, 32 CD tests and workflow actionlint.
The standard production Dockerfile built successfully with the published pin.
Its disposable real-Keycloak/PostgreSQL fixture verified inventory v0.1.0, all
nine warehouse reads, authentication/system/context guards, disablement and
re-enablement with data parity, backup/restore, core audit recording and
scheduler stop/resume. Warehouse SQL, migrations and worker ownership are
unchanged. The release tag is upstream; publication does not deploy the backend.
