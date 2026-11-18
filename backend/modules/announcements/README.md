# Announcements Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements`.
Requires Go 1.25.4 and `github.com/google/uuid v1.6.0`.

Version `v0.1.0` is published under `backend/modules/announcements/v0.1.0`.
The dashboard selects this immutable Go dependency with reviewed checksums.
Hosts supply persistence, identity, storage, normalization and notification
adapters. Local/unversioned consumers report `dev`.

## Public API

| API | Responsibility |
| --- | --- |
| `NewServiceWithSettings`, `Service` | Announcement creation, updates, lifecycle actions, listing and audience/delivery orchestration |
| `Repository` | All 53 persistence operations using package-owned models and inputs |
| `UserDirectory`, `GroupDirectory`, `HealthAudience` | User/group lookup and actor-scoped or stored health audience expansion |
| `AttachmentStorage` | Upload/download streams and deletion; resource ownership stays with the host |
| `NotificationPublisher` | Forward package-owned notification envelopes and delivery requests |
| `EmailAttachmentNormalizer` | Apply the host's attachment validation and normalization policy |
| `ServicePorts`, `Settings` | Explicit dependencies and non-secret branding, URL, SMS and attachment settings |
| `ModulePath`, `Version` | Build dependency identity; local replacements report `dev` |

The repository returns package-owned announcements, attachments, enums and
statistics. Standard-library nullable SQL values and UUID nullable values are
data values, preserving null versus zero semantics without a database driver.
No SQLC aliases, Gin, Keycloak, Redis, SMTP or dashboard imports are included.

Compose the service using a host implementation of `Repository` and the ports
needed by the chosen operations:

```go
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
including read/write/publish and attachment access. It implements repository
queries/transactions and supplies context-aware adapters. Health audience
resolution keeps actor-scoped selection distinct from stored expansion.

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

HTTP routes/DTOs, permissions, rate limits, SQLC/PostgreSQL adapters, identity
lookups, blob storage, configuration/secrets, notification delivery workers,
core audit recording and migrations remain in the dashboard. The module
neither starts infrastructure nor applies migrations. Disabling announcement
routes does not cancel previously queued shared notifications.

## Verification

From the repository root:

```sh
python3 infra/scripts/test-announcements-package.py
make -C backend modules-check
```

The verifier copies this package outside the source tree, disables workspaces
and network module downloads, and seeds a private cache with only UUID's
reviewed pinned artifacts. It verifies checksums, the exact dependency graph,
build/vet/uncached race checks and unchanged dependency metadata. It then
builds, vets and runs `testdata/consumer/main.go` in a separate module using
only public APIs. The consumer exercises lifecycle, notifications, nullable
values, attachment upload/download and repository error propagation.

The dashboard integration uses aliases and adapters; no retained host core
remains. To verify the published dependency with a fresh module cache and no
replacement, run `python3 infra/scripts/test-announcements-release.py`.
This release check requires network access and verifies downloaded checksums,
selected version, source parity and an independent consumer.
