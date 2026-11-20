# Announcements Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements`.
Requires Go 1.25.4, `github.com/google/uuid v1.6.0` and `github.com/lib/pq v1.10.9`.

Version `v0.1.0` is published under `backend/modules/announcements/v0.1.0`.
The dashboard selects this immutable Go dependency with reviewed checksums.
The current source adds standalone PostgreSQL persistence as the first
self-contained-module phase; it is not yet published or selected by the host.
Local/unversioned consumers report `dev`.

## Public API

| API | Responsibility |
| --- | --- |
| `NewServiceWithSettings`, `Service` | Announcement creation, updates, lifecycle actions, listing and audience/delivery orchestration |
| `Repository`, `NewRepository(*sql.DB)` | All 53 persistence operations, package-owned SQL and audience transactions; caller-owned pool |
| `UserDirectory`, `GroupDirectory`, `HealthAudience` | User/group lookup and actor-scoped or stored health audience expansion |
| `AttachmentStorage` | Upload/download streams and deletion; resource ownership stays with the host |
| `NotificationPublisher` | Forward package-owned notification envelopes and delivery requests |
| `EmailAttachmentNormalizer` | Apply the host's attachment validation and normalization policy |
| `ServicePorts`, `Settings` | Explicit dependencies and non-secret branding, URL, SMS and attachment settings |
| `ModulePath`, `Version` | Build dependency identity; local replacements report `dev` |

The repository returns package-owned announcements, attachments, enums and
statistics. Standard-library nullable SQL values and UUID nullable values are
data values, preserving null versus zero semantics without a database driver.
Private SQL records and queries are internal implementation details. No public
SQLC aliases, Gin, Keycloak, Redis, SMTP or dashboard imports are included.

Compose the service using package persistence and the capabilities needed by
the chosen operations (a custom `Repository` remains supported):

```go
repo, err := announcements.NewRepository(primaryPool)
if err != nil {
    return err
}
core := announcements.NewServiceWithSettings(repo, announcements.ServicePorts{
    Users:            users,
    Groups:           groups,
    HealthContexts:   healthAudience,
    Storage:          storage,
    Notifications:    publisher,
    EmailAttachments: normalizer,
}, settings)
```

The host authenticates and authorizes each request before calling the service,
including read/write/publish and attachment access. The package implements
repository queries/transactions. Health audience resolution keeps actor-scoped selection distinct from stored expansion.

Notification forwarding remains best effort; a downstream publisher failure
does not turn successful persistence/publication into a failure. Existing
notification labels and queued delivery payloads are preserved. This package
does not own a worker or provide atomic/exactly-once persistence and delivery.
Scheduled delivery execution belongs to the host's notification infrastructure.
Hosts must retain their email normalization policy; nonempty email attachments
that require normalization fail closed when the normalizer is absent.

Attachment uploads retain configured count/type/byte limits, checksums, actor
metadata and orphan cleanup on persistence failure. Download streams must be
closed by the caller. Hosts supply settings appropriate to their deployment;
nil settings use the existing defaults rather than host configuration.

## Host ownership

HTTP routes/DTOs, permissions and rate limits are scheduled for the next
self-contained-module phase. Shared identity lookups, blob storage,
configuration/secrets, notification delivery workers, core audit recording and migrations remain in the dashboard. The module
neither starts infrastructure nor applies migrations. Disabling announcement
routes does not cancel previously queued shared notifications.

`NewRepository` rejects a nil pool, honors request cancellation and preserves
wrapped SQL errors. It does not ping or close the pool. The caller must apply
the existing Announcements schema before using it, including audience and
attachment tables. SMS recipient lookup reads shared notification preferences;
group and health-context mappings reference platform-owned identity/context
rows. These dependencies are explicit schema prerequisites, not owned
infrastructure. All five audience replacements are transactional. Notification
markers and append-only audience snapshots retain existing behavior; they are
not an atomic delivery outbox.

## Verification

From the repository root:

```sh
python3 infra/scripts/test-announcements-package.py
python3 infra/scripts/test-announcements-postgres-package.py
make -C backend modules-check
```

The verifier copies this package outside the source tree, disables workspaces
and network module downloads, and seeds a private cache with only reviewed
pinned artifacts for UUID and lib/pq. It verifies checksums, the exact dependency graph,
build/vet/uncached race checks and unchanged dependency metadata. It then
builds, vets and runs `testdata/consumer/main.go` in a separate module using
only public APIs. The consumer exercises lifecycle, notifications, nullable
values, attachment upload/download and repository error propagation.

The PostgreSQL verifier creates a disposable database from a test-only snapshot
of the existing schema. It builds a race-enabled public-API consumer outside
the repository with workspaces/downloads disabled and a private dependency
cache. It exercises all 53 operations, nullable values, audience isolation,
attachment scoping/deletion, append-only snapshots, cancellation and pool
ownership. All five audience replacement transactions are tested for rollback
after partial writes; clients/users use explicit test-only failure triggers.
The module never applies this fixture schema.

The dashboard still selects v0.1.0 and retains its HTTP/persistence adapters.
They will be removed only after package HTTP/provider wiring and independent
verification, followed by a new published version and reviewed host pin.
To verify the published dependency with a fresh module cache and no replacement, run `python3 infra/scripts/test-announcements-release.py`.
This release check requires network access and verifies downloaded checksums,
selected version, source parity and an independent consumer.
