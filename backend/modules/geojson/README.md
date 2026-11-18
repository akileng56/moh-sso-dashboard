# GeoJSON Go module

Module path: `github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson`.
Requires Go 1.25.4 and has no third-party dependencies.

Version `v0.1.0` is the first standalone GeoJSON release. Its immutable
nested-module release tag is `backend/modules/geojson/v0.1.0`. Hosts select the
verified dependency version and supply storage, asset packaging and HTTP policy.

## Public API

| API | Responsibility |
| --- | --- |
| `ResolveLayer` | Immutable public-name to canonical-filename mapping |
| `AssetReader`, `Asset` | Caller-supplied seekable asset stream and modification time |
| `Service`, `NewService`, `Service.Open` | Resolution, context propagation, read orchestration and failure cleanup |
| `Layer` | Resolved filename and successful stream transferred to the caller |
| `ErrUnknownLayer` | Rejected public layer name, before any storage access |
| `ErrLoadFailed` | Read/configuration/cancellation failure with stable public text |
| `ModulePath`, `Version` | Selected consumer dependency version; local/unversioned builds report `dev` |

The allowlist is fixed and case-sensitive:

| Public name | Canonical filename |
| --- | --- |
| geo_water | uga_water.geojson |
| geo_districts | uganda_districts.geojson |
| geo_subcounties | uganda_subcounties.geojson |

## Ownership and errors

Authenticate and authorize access before calling `Service.Open`; the library
does not implement host permission policy. Implement `AssetReader` to open the
canonical filename in host-controlled storage, using the supplied context.
Readers return a seekable, closeable stream and its modification time.

After a successful service call, the caller closes `Layer.Content` after use.
On a failed or canceled read, the service closes any returned stream and returns
an empty layer. Underlying read, cancellation and close-error causes remain
available through `errors.Is/As`, while the load error's text remains stable.
The reader port transfers ownership of every non-nil returned stream, including
one returned alongside an error.

Only unknown public names map to `ErrUnknownLayer` during resolution. A missing
file for a known name is a load failure. HTTP adapters should prioritize
`ErrLoadFailed` if an underlying reader error happens to match both sentinels.
The dashboard retains its existing 404/500 envelopes and `http.ServeContent`
handling for raw GeoJSON, cache/nosniff headers, conditional requests and ranges.

## Host integration

The package contains no map files, filesystem adapter, HTTP framework, permission
resolver, database driver, migrations or background workers. Hosts provide
storage and ship the map assets; packaging this library does not package or
deploy those assets.

Dashboard consumers select a reviewed Go dependency version and report the
selected module version through their compiled inventory. Local workspace or
directory replacements report `dev`; they must not be used as a production pin.
The host retains only its storage/HTTP adapters after release integration.

## Verification

From the repository root:

```sh
python3 infra/scripts/test-geojson-package.py
make -C backend modules-check
```

The verifier copies only this module outside the dashboard, uses an empty module
cache with workspaces and network module downloads disabled, checks dependency
boundaries and runs read-only build/vet/race checks. It also compiles and runs a
separate consumer module using a temporary local replacement. This replacement
exists only in the disposable verification fixture; the host dependency files
are unchanged.

A `go.sum` is not needed because this module uses only the standard library.
Release preparation must verify a clean published download and checksums before
the host selects the immutable dependency version.
