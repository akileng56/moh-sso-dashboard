# GeoJSON extraction

The existing `geojson` registry module serves permission-guarded map assets.
It remains hosted in the dashboard and reports the host version.
`DISABLED_MODULES=geojson` removes its routes on restart. It owns no workers or
database migrations.

## Phase 1: asset and authorization boundaries

`AssetReader.Open(context.Context, filename)` returns an `Asset` containing a
seekable, closeable stream and its modification time. These contracts use only
the standard library. The caller closes every non-nil returned stream, including
one returned alongside an error.

`NewDirectoryReader` adapts host-owned files, checks request cancellation and
accepts only single-component filenames for regular files. Bootstrap explicitly
injects it through `NewHandlerWithReader`. `NewHandler(basePath)` remains a
compatible convenience constructor.

The HTTP handler retains the exact public-name allowlist:

| Public name | Host asset |
| --- | --- |
| geo_water | uga_water.geojson |
| geo_districts | uganda_districts.geojson |
| geo_subcounties | uganda_subcounties.geojson |

Unrecognized names return the existing `NOT_FOUND` response without touching
storage. Missing assets and reader failures return `GEOJSON_LOAD_FAILED`
without exposing internal paths. Streams close after successful reads,
conditional requests, ranges and error responses.

Authentication and the existing any-permission gate remain in the host route
registrar: `surveillance:read`, `data_quality:read` or `report_browser:read`.
These checks execute before the reader is invoked. The package extraction must
not replace authoritative host permission resolution.

The response remains raw `application/geo+json`, with
`Cache-Control: public, max-age=86400` and `X-Content-Type-Options: nosniff`.
`http.ServeContent` retains modification-time validation and byte-range behavior.
Map assets and their Docker packaging remain host-owned.

## Verification and next phases

Tests cover real bundled layers, permission denial before asset access,
allowlist rejection, safe storage errors, stream closure, range/conditional
responses, context cancellation and directory-adapter boundaries. Registry
tests retain enabled/disabled route coverage.

```sh
go test -race ./internal/features/geojson ./internal/api ./internal/bootstrap
make contracts-check
```

The next phases are:

1. Move layer resolution and portable read/error handling behind the reader contract.
2. Build and verify the standalone module and an external consumer.
3. Publish a version, pin its verified checksums and replace the retained
   host implementation with adapters.
