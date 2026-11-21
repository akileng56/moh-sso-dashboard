#!/usr/bin/env python3
"""Verify the GeoJSON module and a separate consumer outside the host tree."""

import os
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / 'backend/modules/geojson'
MODULE_PATH = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson'
CONSUMER_PATH = 'example.com/geojson-consumer'
METADATA = ('go.mod', 'go.sum', 'go.work', 'go.work.sum')
CONSUMER_SOURCE = r'''package main

import (
 "bytes"
 "context"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"

 geojson "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson"
)

type stream struct { *bytes.Reader }
func (*stream) Close() error { return nil }
type reader struct {}
func (reader) Open(_ context.Context, filename string) (geojson.Asset, error) {
 if filename != "uga_water.geojson" { panic("unexpected filename: " + filename) }
 return geojson.Asset{Content: &stream{bytes.NewReader([]byte("{}"))}}, nil
}

func main() {
 layer, err := geojson.NewService(reader{}).Open(context.Background(), "geo_water")
 if err != nil { panic(err) }
 body, err := io.ReadAll(layer.Content)
 if err != nil { panic(err) }
 if err := layer.Content.Close(); err != nil { panic(err) }
 if layer.Filename != "uga_water.geojson" || string(body) != "{}" ||
  geojson.ModulePath != "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson" ||
  geojson.Version() != "dev" { panic(fmt.Sprintf("public consumer contract changed: filename=%q body=%q module=%q version=%q", layer.Filename, body, geojson.ModulePath, geojson.Version())) }
 // Exercise the complete package, including embedded assets, through only its
 // public API. The executable is run from a separate empty working directory.
 mux := http.NewServeMux()
 _, err = geojson.Register(mux, geojson.Dependencies{
  BasePath: "/api/v1/geojson",
  AuthorizeAny: func(request *http.Request, permissions ...string) error {
   if strings.Join(permissions, ",") != "surveillance:read,data_quality:read,report_browser:read" { panic("permission contract changed") }
   if request.Header.Get("X-Fixture-Identity") != "allowed" { return geojson.ErrUnauthorized }
   return nil
  },
 })
 if err != nil { panic(err) }
 serve := func(authorized bool, headers map[string]string) *httptest.ResponseRecorder {
  request := httptest.NewRequest("GET", "/api/v1/geojson/geo_water", nil)
  if authorized { request.Header.Set("X-Fixture-Identity", "allowed") }
  for key,value := range headers { request.Header.Set(key,value) }
  response := httptest.NewRecorder()
  mux.ServeHTTP(response, request)
  return response
 }
 if response:=serve(false,nil); response.Code!=401 { panic("unauthenticated module request was accepted") }
 response := serve(true,nil)
 if response.Code!=200 || !strings.Contains(response.Body.String(),"FeatureCollection") ||
  response.Header().Get("Content-Type")!="application/geo+json" || response.Header().Get("ETag")=="" { panic("self-contained HTTP assets failed") }
 etag:=response.Header().Get("ETag")
 if response:=serve(true,map[string]string{"If-None-Match":etag});response.Code!=304 || response.Body.Len()!=0 { panic("conditional request failed") }
 if response:=serve(true,map[string]string{"Range":"bytes=0-9"});response.Code!=206 || response.Body.Len()!=10 { panic("range request failed") }
 if response:=serve(false,map[string]string{"If-None-Match":etag});response.Code!=401 { panic("conditional request bypassed authorization") }
 fmt.Println("PASS: separate self-contained GeoJSON consumer (dev)")
}
'''


def metadata(root):
    return {name: (root / name).read_bytes() if (root / name).exists() else None
            for name in METADATA}


def main():
    go = os.environ.get('GO_CMD', 'go')
    with tempfile.TemporaryDirectory(prefix='moh-geojson-package-') as tmp:
        root = Path(tmp)
        standalone = root / 'module'
        shutil.copytree(MODULE, standalone)
        env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local',
               'GOPROXY': 'off', 'GOMODCACHE': str(root / 'empty-module-cache')}

        def run(cwd, *args):
            result = subprocess.run([go, *args], cwd=cwd, env=env, text=True,
                                    capture_output=True, check=True)
            return result.stdout.strip()

        def check_graph(cwd, expected):
            modules = set(run(cwd, 'list', '-mod=readonly', '-deps', '-test',
                              '-f', '{{if .Module}}{{.Module.Path}}{{end}}', './...').splitlines()) - {''}
            assert modules == expected, f'Unexpected dependency modules: {modules}'

        snapshots = json.loads((standalone / 'maps/sha256.json').read_text())
        assert set(snapshots) == {'uga_water.geojson', 'uganda_districts.geojson', 'uganda_subcounties.geojson'}
        for name, expected in snapshots.items():
            content = (standalone / 'maps' / name).read_bytes()
            assert hashlib.sha256(content).hexdigest() == expected, 'Embedded map checksum drift: ' + name
            assert not (ROOT / 'backend/assets/geojson' / name).exists(), 'Self-contained module retains a duplicate host asset: ' + name

        before = metadata(standalone)
        run(standalone, 'mod', 'verify')
        check_graph(standalone, {MODULE_PATH})
        run(standalone, 'build', '-mod=readonly', './...')
        run(standalone, 'vet', '-mod=readonly', './...')
        print(run(standalone, 'test', '-race', '-mod=readonly', '-count=1',
                  '-timeout=2m', './...'), flush=True)
        assert metadata(standalone) == before, 'Module dependency metadata changed'

        consumer = root / 'consumer'
        consumer.mkdir()
        (consumer / 'go.mod').write_text(
            'module ' + CONSUMER_PATH + '\n\ngo 1.25.4\n\nrequire ' +
            MODULE_PATH + ' v0.0.0\n\nreplace ' + MODULE_PATH + ' => ../module\n')
        (consumer / 'main.go').write_text(CONSUMER_SOURCE)
        consumer_before = metadata(consumer)
        check_graph(consumer, {CONSUMER_PATH, MODULE_PATH})
        run(consumer, 'build', '-mod=readonly', './...')
        run(consumer, 'vet', '-mod=readonly', './...')
        binary = root / 'geojson-consumer'
        run(consumer, 'build', '-mod=readonly', '-o', str(binary), '.')
        empty = root / 'empty-working-directory'
        empty.mkdir()
        output = subprocess.run([str(binary)], cwd=empty, env=env, text=True,
                                capture_output=True, check=True).stdout.strip()
        assert output == 'PASS: separate self-contained GeoJSON consumer (dev)', output
        print(output, flush=True)
        assert metadata(consumer) == consumer_before, 'Consumer dependency metadata changed'
        assert metadata(standalone) == before, 'Consumer modified module metadata'
        print('PASS: isolated self-contained GeoJSON build, asset checksums, dependency boundary, vet, race tests and separate HTTP consumer from empty directory; downloads disabled', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'GeoJSON package verification failed: {error.stdout}\n{error.stderr}') from None
