# Operating compiled modules

Use [the coverage guide](module-coverage.md) to distinguish packaged modules,
registry-hosted features, planned extraction work and mandatory host capabilities.
Use `GET /api/v1/admin/modules` with an authenticated administrator session to
inspect the modules actually compiled into the running backend.

| Runtime ID | Scope | Version reported | Effect of disablement |
| --- | --- | --- | --- |
| metrics | Dashboard metrics APIs | Selected Go dependency version | Metrics routes are omitted |
| audit | Audit read, metrics and export APIs | Selected Go dependency version | Audit read routes are omitted; core event recording continues |
| geojson | Permission-guarded static map assets | Selected Go dependency version | GeoJSON routes are omitted; map consumers need availability review |
| report-scheduler | Report schedules, execution and delivery | Selected Go dependency version | All scheduler routes, including public downloads, are omitted; worker startup is skipped |
| announcements | Public, audience and administrative announcements, including attachments | Selected Go dependency version | All public, protected and admin announcement routes are omitted; shared delivery workers continue |
| admin-units | Organizational units, facilities and hierarchy lookup | Host build version (not separately published yet) | All nine Administrative Units routes are omitted; Visualizer routes, health-context services and warehouse data remain |

The default empty `DISABLED_MODULES` enables all compiled definitions. Set a
comma-separated list, such as `audit,geojson`, in the backend environment and
restart/recreate the backend through the normal deployment workflow. Whitespace
is trimmed. Unknown IDs fail startup. The registry inventory still lists disabled
definitions with `enabled=false`. Their initialization and registration callbacks
are skipped. Existing core authentication, authorization, session and event
recording services retain their lifecycle.

Review affected application registrations and navigation when changing API
availability. Keep Keycloak roles and application lifecycle approval aligned with
the intended rollout; module configuration does not automatically edit them.
Announcements, the audit read module and map asset module have no background workers to stop.
Announcements disablement takes effect after a backend restart. It preserves
announcement records, audience selections and attachment objects. It does not
cancel notifications already queued in the shared notification outbox; that
worker serves other modules and retains its host lifecycle. Re-enable the module
and restart to restore access to existing records and attachment downloads.
Report Scheduler is pinned to its published Go package and owns its worker through the registry. Its initial persistence
heartbeat must succeed within ten seconds before jobs can be claimed. Startup
fails if scheduler persistence is unavailable. Cleanup cancels and joins the worker
within the host shutdown deadline before database resources close. Disablement
retains schedules, history and pending jobs; re-enabling resumes existing claim,
retry and stale-claim recovery behavior. Previously issued public download links
are unavailable while disabled. Other worker-bearing legacy features remain
tracked as planned.

For version changes, select verified Go/npm releases, review dependency files and
build new images. The frontend pin selection is independent of this backend
availability setting. Database data and migration ownership remain explicit in
the catalogue; disabling an API does not remove its schema or historical data.

For rollout, backup and recovery, use `infra/distribution.md` and
`infra/server/README.md`. Promote the same verified image-pair bundle from staging
to production. Compare application versions, compiled inventory, frontend module
versions, readiness and affected permissions after deployment. Use the documented
backup/compatibility procedure when an image rollback cannot use an upgraded schema.

For local verification, the authenticated distribution fixture exercises real
Keycloak roles, module inventory, API disablement and continued core audit
recording, as well as manifests, backup/restore and restart recovery. It uses
disposable projects and does not contact an existing production identity provider.


Report Scheduler currently selects `v0.1.0` from
`github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler`.
Its immutable tag is `backend/modules/report-scheduler/v0.1.0`. A version change
requires updating the host Go pin/checksums and rebuilding the backend image.
The package release does not deploy a running installation. Verify the compiled
inventory reports the selected version, and review scheduler schema compatibility
before promoting a newly built image pair.


Audit read APIs select `v0.1.0` from
`github.com/mohuganda/moh-sso-dashboard/backend/modules/audit`, published under `backend/modules/audit/v0.1.0`. Its compiled inventory
version comes from the selected dependency. SQLC queries/mapping, optional Redis
cache, HTTP authorization and export limits remain host adapters. Core event
recording and all audit migrations retain host ownership. Publishing a library
tag does not deploy the running dashboard; rebuild and promote the verified
backend image to consume an updated pin.


GeoJSON selects `v0.1.0` from
`github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson`, published
under `backend/modules/geojson/v0.1.0`. The compiled inventory reports that
selected dependency version. Host storage and HTTP adapters retain authorization,
safe errors, headers, byte ranges and conditional reads. The package does not
contain map files: ship the three existing assets in the backend image and
verify actual map responses before rollout. Publishing the library does not
deploy the running dashboard.

Announcements selects `v0.1.0` from
`github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements`, published
under `backend/modules/announcements/v0.1.0`. Its inventory version comes from
the selected dependency. The host retains SQLC/persistence mapping, identity and
storage adapters, HTTP authorization/rate limits, email normalization, shared
notification workers and audit recording. A version change requires reviewing
the Go pin/checksums and rebuilding the backend image. Publishing the module
does not deploy a running installation.


Administrative Units is registry-hosted under `admin-units`. Set
`DISABLED_MODULES=admin-units` and restart/recreate the backend to omit all nine
`/api/v1/visualizer/adminunits/*` endpoints. It owns no workers and does not
modify warehouse data or schemas. Re-enable and restart to restore existing
lookups. Enabled registration retains the same system access, selected health
context and read-permission alternatives as the legacy route path. Other
Visualizer routes retain their registration and guards.

Administrative Units now has a verified standalone candidate in
`backend/modules/admin-units`. The dashboard still uses its retained core and
reports the host build version. Packaging alone does not publish a version,
select a Go dependency or change runtime availability. Publication, verified
download/checksums, host pinning and replacement with adapters are next.
