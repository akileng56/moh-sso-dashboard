# GeoJSON host adapters

The dashboard selects
`github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson v0.1.0`
without local replacements. The immutable module tag is
`backend/modules/geojson/v0.1.0`. Compiled inventory reports the selected
dependency version. `DISABLED_MODULES=geojson` removes the routes on restart;
the module owns no workers or migrations.

## Responsibilities

`package.go` aliases the published asset/reader, layer and service types,
exposes the error sentinels, and delegates construction and layer resolution.
The retained host asset contracts and portable service implementation have
been removed. Portable source and external-consumer tests live in the
[standalone module](../../../modules/geojson/README.md).

`assets_directory.go` adapts host-controlled files to the reader contract.
Bootstrap injects it through `NewHandlerWithReader`; `NewHandler(basePath)`
remains compatible. The adapter accepts only regular files with canonical
single-component filenames and checks request cancellation.

`handler.go` and `routes.go` retain HTTP serving and authorization. Authentication
and one of `surveillance:read`, `data_quality:read` or `report_browser:read`
are required before storage access. Unknown public layer names return the
existing `NOT_FOUND` response; load failures return `GEOJSON_LOAD_FAILED`
without exposing paths. A storage failure cannot turn a known layer into a 404.

The package owns the fixed mapping below; assets and Docker packaging remain
host-owned:

| Public name | Host asset |
| --- | --- |
| geo_water | uga_water.geojson |
| geo_districts | uganda_districts.geojson |
| geo_subcounties | uganda_subcounties.geojson |

## Stream and HTTP contracts

A successful service call transfers stream ownership to the HTTP adapter,
which closes it after serving. The package closes any stream returned on a
failed or canceled read. Original and close-error causes remain available
through `errors.Is/As`, while public load-error text stays stable.

Responses remain raw `application/geo+json` with
`Cache-Control: public, max-age=86400` and `X-Content-Type-Options: nosniff`.
`http.ServeContent` preserves conditional requests and byte ranges.

## Verification

Host tests exercise the pinned dependency through real bundled files, all three
permission alternatives, denial before storage access, safe failures, stream
closure, cancellation, ranges and conditional reads. Registry tests cover
enabled/disabled routes and dependency-version reporting.

```sh
go test -race ./internal/features/geojson ./internal/api ./internal/bootstrap
make contracts-check
```

From the repository root, the standalone verifier builds/vets/race-tests only the
module and a separate consumer with downloads disabled:

```sh
python3 infra/scripts/test-geojson-package.py
```

The authenticated distribution fixture additionally checks the selected version,
all three packaged layers, headers, ranges, modification-time validation,
real Keycloak permissions and registry disablement.
