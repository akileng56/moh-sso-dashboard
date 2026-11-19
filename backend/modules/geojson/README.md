# GeoJSON Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson`.
Requires Go 1.25.4 and has no third-party dependencies.

Version `v0.2.0` provides the self-contained API below. Its release tag is
`backend/modules/geojson/v0.2.0`. Hosts select its reviewed version/checksums,
supply shared security/trace callbacks, and register the module directly.
`OpenAPI()` supplies its operation contract for generic host documentation.

## Self-contained entry point

The package owns its model/service, map snapshots, storage readers, HTTP
handling, permission requirements, safe errors and registration. GeoJSON owns
no database migrations, background workers or external provider connections.

```go
mux := http.NewServeMux()
_, err := geojson.Register(mux, geojson.Dependencies{
    BasePath:     "/api/v1/geojson",
    AuthorizeAny: platform.AuthorizeAny,
    RequestID:    platform.RequestID,
})
```

`platform` represents the consuming platform's shared security/trace service.
The functions are shared platform capabilities, not GeoJSON-specific adapters.
`Registrar` is a structural `Handle(string, http.Handler)` interface satisfied
by `http.ServeMux`. A framework host can provide one registrar for all modules.

`NewModule` returns the same capability as a standard `http.Handler` for hosts
that mount handlers directly. Its `Pattern()` describes the registered GET
pattern. Go's standard multiplexer also supports HEAD for this GET pattern.
BasePath defaults to `/geojson`; configured prefixes must be clean absolute
paths containing letters, numbers, slash, dot, underscore or hyphen.

| Dependency | Meaning |
| --- | --- |
| `AuthorizeAny` | Required: authenticate the caller and grant at least one supplied permission |
| `RequestID` | Optional: provide the trusted trace ID for JSON error metadata |
| `Assets` | Optional: a caller-selected seekable asset reader |
| `Directory` | Optional: a trusted operator-configured filesystem override |
| Neither asset override | Serve the module's embedded, versioned map files |

Assets and Directory are mutually exclusive. Construction rejects a missing
authorizer before registering any route. No default anonymous or permissive
authorization is provided.

The authorizer must check any of these exact permission keys:
`surveillance:read`, `data_quality:read`, `report_browser:read`.
Each call receives a fresh permission slice. The platform must verify identity
before granting permission; the module does not trust request headers as claims.
Authorization runs before storage access, ranges and conditional response
handling. Errors implementing `StatusCode() int` map 401/403 to the existing safe
unauthorized/forbidden envelopes. Other failures return 503 without exposing
provider diagnostics. Independent callers can use ErrUnauthorized/ErrForbidden.

## Map ownership and HTTP behavior

| Public name | Embedded filename |
| --- | --- |
| geo_water | uga_water.geojson |
| geo_districts | uganda_districts.geojson |
| geo_subcounties | uganda_subcounties.geojson |

The fixed, case-sensitive allowlist never accepts caller-supplied filenames.
The three snapshots are embedded at compile time. They are usable from an empty
working directory without dashboard files, runtime downloads or storage adapters.
See `maps/README.md` for provenance and `maps/sha256.json` for checksums.
The verifier checks every embedded snapshot against its published checksum.

Successful reads return raw GeoJSON with the existing content type,
`public, max-age=86400` and `nosniff` headers. The module handles ranges and
conditional reads with `http.ServeContent` and closes streams on every outcome.
Embedded files have no filesystem timestamp; stable content ETags support 304
responses. Filesystem/custom-reader overrides retain their modification times.
Unknown names return the existing 404 envelope; known-layer load failures remain
500, with no private path/cause disclosure. Request IDs come only from the
trusted trace callback.

## Lower-level compatibility API

`ResolveLayer`, `AssetReader`, `Asset`, `Layer`, `NewService` and
`Service.Open` remain available. Those lower-level calls do not authorize:
their caller must do so before opening an asset. Successful streams transfer to
the caller; failed/cancelled reads close any returned stream. Load failures keep
their stable public message and retain causes through `errors.Is/As`.

`BundledAssets` and `NewDirectoryReader` provide package-owned implementations
of AssetReader. Directory is trusted deployment configuration. The selected
directory and any symlinks it contains are operator-controlled.

`ModulePath` and `Version()` report the consuming build's selected dependency.
Local/unversioned builds report `dev`; production must use a reviewed release.

## Verification

```sh
python3 infra/scripts/test-geojson-package.py
make -C backend modules-check contracts-check
```

The verifier copies the complete module outside the host, uses an empty module
cache with downloads/workspaces disabled, and builds/vets/race-tests it. It
checks embedded snapshot identities and compiles a separate public consumer.
That binary runs from an empty directory and verifies package-owned HTTP
registration/assets, access denial, byte ranges and ETag conditional responses.
Release verification checks the downloaded artifact separately from the local source tree.
