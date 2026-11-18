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
The audit read module and map asset module have no background workers to stop.
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
