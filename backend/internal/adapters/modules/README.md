# Published module host adapters

These directories integrate pinned, independently versioned modules with the
dashboard. Domain implementation lives in `backend/modules/<module-id>`;
there are no corresponding implementations in `backend/internal/features`.

| Adapter | Published module |
| --- | --- |
| `audit` | `backend/modules/audit` |
| `metrics` | `backend/modules/metrics` |
| `geojson` | `backend/modules/geojson` |
| `report_scheduler` | `backend/modules/report-scheduler` |
| `announcements` | `backend/modules/announcements` |
| `admin_units` | `backend/modules/admin-units` |

Host-only responsibilities include HTTP authorization/envelopes, SQLC and
identity/context mapping, and shared storage/delivery composition. Aliases and
constructors delegate to the reviewed published dependency. Host-owned audit
recording and migrations remain independent of optional reader modules.

The catalogue and architecture checks enforce these ownership locations and
the existing approved dependencies on mandatory host features.
