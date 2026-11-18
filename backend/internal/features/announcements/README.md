# Announcements host adapters

The host selects `github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements v0.1.0`
with reviewed checksums and no local replacement. Runtime inventory reports the
selected dependency version. The release tag is
`backend/modules/announcements/v0.1.0`.

## Boundaries

| Source | Host responsibility |
| --- | --- |
| `contracts.go` | Compatibility aliases to published types, ports, inputs, enums and notification constants |
| `service_host.go` | Compose and embed the selected core; no lifecycle or audience implementation remains |
| `ports_host.go` | Map users/groups, snapshot non-secret settings, retain configuration constructors and startup setters |
| `repository_host.go`, `repository_postgres.go` | SQLC persistence contract, existing SQL queries and transactions |
| `repository_adapter.go`, `mapper_postgres.go` | Translate SQLC records into selected package models and back |
| `delivery_host.go` | Map notification payloads and delegate existing host email normalization |
| `handler.go`, `dto.go`, `mapper.go`, `routes.go` | HTTP DTOs/responses, authorization, route contracts and rate limits |

`Service` embeds the published service. Compatibility setters rebuild its
composition from the same repository, settings and selected ports. They are
startup-only and run before requests/workers begin; they are not runtime
configuration or concurrent mutation APIs. Prefer supplying `ServicePorts`
when dependencies are already available.

All 26 public/protected/admin routes remain registered as one module.
`DISABLED_MODULES=announcements` removes all these routes after restart,
including both attachment download paths. Re-enabling restores access to
existing records and objects. Shared notification workers, queued deliveries,
audit recording and migrations retain their host lifecycle.

## Verification

```sh
python3 infra/scripts/test-announcements-package.py
python3 infra/scripts/test-announcements-release.py
go -C backend test -race ./internal/features/announcements ./internal/bootstrap ./internal/api
make -C backend modules-check contracts-check
```

The package verifier uses an isolated UUID-only cache with downloads disabled.
The release verifier intentionally downloads v0.1.0 into a fresh cache, validates
the host checksums, rejects replacements and builds/runs an independent public
consumer that must report v0.1.0. It also compares the shipped Go sources.

Host regressions cover model round trips, SQLC mapping and context/errors,
notification wire/byte payloads, normalization limits, settings/secret exclusion,
compatibility composition, permissions and full route disablement. Portable
lifecycle, audience and attachment tests now live in the standalone module.
The authenticated distribution fixture checks real Keycloak/PostgreSQL,
publication, stored attachment downloads, version inventory and re-enablement.
