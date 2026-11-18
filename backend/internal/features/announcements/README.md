# Announcements extraction

Announcements remains a host-wired feature. It has not been registered as a
disableable module, packaged independently, published or pinned yet.

## Phase 1: dependency contracts

`ports.go` defines feature-owned audience models and narrow interfaces:

| Contract | Responsibility |
| --- | --- |
| `AudienceUser` | Identity, enabled status, email and display-name fields needed for recipients |
| `AudienceGroupMember` | Group-member identity and fallback email/name |
| `ClientAccess` | Both internal and public client identifiers used by existing matching |
| `UserDirectory` | User listing/detail, realm-role membership and client-access lookup |
| `GroupDirectory` | Group validation and membership expansion |
| `HealthAudience` | Actor-authorized audience resolution and separate stored-audience expansion |
| `AttachmentStorage` | Upload, download and orphan cleanup through streams/object keys |

These contracts use standard-library types and UUIDs without importing users,
RBAC, Keycloak, health-context or storage implementation packages.

`ServicePorts` composes those dependencies. Bootstrap uses
`NewServiceWithPorts` and the host user/group adapters.
`ports_host.go` preserves the existing `NewService`, `SetRBACRepository` and
`SetHealthContextService` entry points. A nil concrete health-context pointer
remains a nil optional interface.

Host adapters preserve recipient fields, both client identifiers, group lookup
errors and email fallback. Context-aware host calls receive the request context.
Legacy user list/detail methods cannot cancel in-flight operations because
their host signatures lack a context; adapters reject already-canceled calls
before invoking them.

`NotificationPublisher` narrows the notification dependency to `Notify`.
Notification payloads are still host model types in this phase. Their portable
conversion is subsequent work; this feature is not yet independently buildable.

## Preserved behavior

Recipient resolution retains enabled-user filtering, empty-email rejection,
case-insensitive deduplication, realm/client/group audience matching and separate
actor-scoped versus stored health-context resolution. The host remains
authoritative for contextual authorization; extraction must not bypass it.
Map iteration still leaves recipient ordering unspecified.

Attachments retain their host configuration, limits, checksums, object-key
layout, persistence and cleanup. A failed metadata insert removes the uploaded
object. Delivery remains best effort with existing notification envelopes,
channels, scheduling, email/SMS behavior and default admin targeting.

Public, protected and admin routes, permission gates and rate limits have not
changed. SQLC persistence, configuration, email attachment normalization,
notification payload conversion, core audit recording and migrations remain
host-owned at this stage. Delivery workers belong to the existing host services;
no worker ownership is transferred in phase 1.

## Verification and next phases

Regression coverage checks user filtering/deduplication, client identifiers,
group fallback, actor scope, context/error forwarding, host model translation,
optional nil dependencies, attachment checksums/orphan cleanup, best-effort
notifications and permission denial before dependencies are accessed.

```sh
go test -race ./internal/features/announcements ./internal/bootstrap ./internal/api
make contracts-check
```

Next:

1. Separate portable announcement/repository and delivery models, configuration
   values and email normalization from host representations.
2. Integrate all public/protected/admin routes into the registry and verify
   disablement, permissions, attachment access and delivery behavior together.
3. Build and verify the standalone package and a separate external consumer.
4. Publish an immutable version, verify its download/checksums, pin it and
   replace the retained host implementation with adapters.
