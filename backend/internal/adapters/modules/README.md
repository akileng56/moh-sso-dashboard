# Published module host adapters

These directories integrate pinned, independently versioned modules with the
dashboard. Domain implementation lives in `backend/modules/<module-id>`;
there are no corresponding implementations in `backend/internal/features`.

| Adapter | Published module |
| --- | --- |
| `audit` | `backend/modules/audit` |
| `metrics` | `backend/modules/metrics` |
| `report_scheduler` | `backend/modules/report-scheduler` |
| `announcements` | `backend/modules/announcements` |
| `admin_units` | `backend/modules/admin-units` |

Host-only responsibilities include HTTP authorization/envelopes, SQLC and
identity/context mapping, and shared storage/delivery composition. Aliases and
constructors delegate to the reviewed published dependency. Host-owned audit
recording and migrations remain independent of optional reader modules.

The catalogue and architecture checks enforce these ownership locations and
the existing approved dependencies on mandatory host features.


GeoJSON v0.2.0 is self-contained and has no directory here. Its published package
owns its endpoint and embedded maps; the host uses the shared
`internal/modulehttp` registrar. Other listed adapters remain transitional
until their own self-contained package release is integrated.
