# Announcements extraction

Announcements remains host-wired. It has not been registered as a disableable
module or published as a separately versioned package yet.

## Portable contracts and service

Phase 2 removes host model, SQLC, configuration, email and utility imports from
the announcement service and attachment orchestration.

| Source | Responsibility |
| --- | --- |
| `models.go` | Feature-owned announcements, attachments, enums, statistics and persistence inputs |
| `inputs.go`, `values.go` | Service inputs, recipient/context models and input value normalization |
| `repository.go` | All 53 persistence operations using feature-owned values |
| `ports.go` | User/group/context audience contracts and stream-based attachment storage |
| `notification_port.go` | Feature-owned notifications, delivery requests, attachments, channels and publisher/normalizer contracts |
| `settings.go` | Only non-secret values used for branding, URLs, SMS and attachment decisions |
| `service.go`, `attachment_service.go` | Portable business, audience, attachment and delivery orchestration |

Standard-library `sql.NullString`/`sql.NullTime` and UUID nullable values remain
value types in the contract to preserve null versus zero semantics. No generated
SQLC type alias, PostgreSQL driver, host model or host configuration is exposed.
Unknown enum strings and invalid-but-nonzero nullable values are preserved by
persistence translation rather than silently normalized.

`NewServiceWithSettings` composes the portable service.
`NotificationPublisher.Publish` accepts only the feature's own payload.
`EmailAttachmentNormalizer` delegates attachment validation and normalization;
when normalization is required, a missing adapter fails closed.

## Host adapters

`repository_host.go` retains the SQLC-side contract.
`repository_adapter.go` and `mapper_postgres.go` translate records, pagination,
lifecycle inputs, status values and statistics. The existing SQL queries,
transactions and underlying repository method bodies remain unchanged;
`NewAnnouncementRepository` wraps that implementation in the adapter.

`delivery_host.go` translates notification envelopes, all delivery channels,
templates, recipients, schedules, retries and attachment representations to the
existing host service. It delegates attachment normalization to the existing
email policy, preserving validation and size/type limits. Notification errors
remain best effort in the announcement service.

`ports_host.go` retains `NewService`, `NewServiceWithPorts` and the existing
RBAC/health-context setters. It builds the non-secret settings snapshot and
adapts identity dependencies. Bootstrap explicitly wraps the host notifier.
Context-aware host calls receive request contexts; legacy user list/detail
operations can only reject cancellation before they start because their host
signatures lack contexts.

The host retains HTTP DTOs/mapping, all public/protected/admin routes,
permissions, rate limits, contextual authorization, storage providers, SQLC
persistence, migrations, core audit recording and delivery workers.
Storage Locations, Document Templates and Documents remain deferred.

## Verification and next phases

The isolated verifier copies only portable source files outside the dashboard,
checks that the dependency graph contains only its temporary module and UUID,
and runs build/vet/uncached race checks with downloads disabled.
It tests create/publish flows, delivery metadata, repository errors and the
email normalization boundary without host files. The PR workflow runs it.

Host tests additionally cover all 20 model translations in both nullable
states, future enum values, collection nil/empty distinctions, context/error
forwarding, notification wire and byte payload parity, schedules/channels,
attachment normalization, secret exclusion and existing audience/storage behavior.

```sh
python3 infra/scripts/test-announcements-core.py
go -C backend test -race ./internal/features/announcements ./internal/bootstrap ./internal/api
make -C backend contracts-check
```

Next:

1. Register all public/protected/admin routes as one module and verify
   disablement, permissions, attachment access and delivery behavior.
2. Build and verify the standalone module and an external consumer.
3. Publish a version, verify its download/checksums, pin it and replace the
   retained host implementation with adapters.
