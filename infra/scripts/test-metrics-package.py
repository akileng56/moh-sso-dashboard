#!/usr/bin/env python3
"""Verify Metrics outside the dashboard, optionally against disposable PostgreSQL."""
import argparse
import json
import os
from pathlib import Path
import secrets
import shutil
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
MODULE = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics'
CONSUMER = 'example.com/metrics-consumer'
METADATA = ('go.mod', 'go.sum', 'go.work', 'go.work.sum')

def run(*args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, text=True, **kwargs)

def metadata(path):
    return {name: (path / name).read_bytes() if (path / name).exists() else None for name in METADATA}

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--postgres', action='store_true')
    args = parser.parse_args()
    go = os.environ.get('GO_CMD', 'go')
    # Third-party dependencies use the installed, checksum-verified Go cache.
    # No workspace, download, host replacement or host source is available.
    env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local', 'GOPROXY': 'off'}
    env.pop('METRICS_TEST_DSN', None)
    container = None
    try:
        with tempfile.TemporaryDirectory(prefix='moh-metrics-package-') as temporary:
            root = Path(temporary)
            standalone = root / 'module'
            shutil.copytree(ROOT / 'backend/modules/metrics', standalone)
            before = metadata(standalone)
            def go_run(cwd, *commands):
                return run(go, *commands, cwd=cwd, env=env).stdout.strip()
            go_run(standalone, 'mod', 'verify')
            dependencies = set(go_run(standalone, 'list', '-mod=readonly', '-m', '-f', '{{.Path}}', 'all').splitlines())
            assert MODULE in dependencies
            for name in dependencies - {MODULE}:
                assert not name.startswith('github.com/moh'), 'Host or sibling module dependency: ' + name
            for line in go_run(standalone, 'list', '-mod=readonly', '-m', '-f', '{{.Path}} {{if .Replace}}REPLACED{{end}}', 'all').splitlines():
                assert 'REPLACED' not in line, 'Candidate has a replacement: ' + line
            imports = go_run(standalone, 'list', '-mod=readonly', '-deps', '-test', '-f', '{{.ImportPath}}', './...')
            assert 'github.com/moh-sso-dashboard/' not in imports, 'Candidate imports host packages'
            go_run(standalone, 'build', '-mod=readonly', './...')
            go_run(standalone, 'vet', '-mod=readonly', './...')
            print(go_run(standalone, 'test', '-race', '-mod=readonly', '-count=1', '-timeout=2m', './...'), flush=True)
            assert metadata(standalone) == before, 'Candidate dependency metadata changed'
            consumer = root / 'consumer'
            consumer.mkdir()
            manifest = (standalone / 'go.mod').read_text().replace('module ' + MODULE, 'module ' + CONSUMER, 1)
            manifest += '\nrequire ' + MODULE + ' v0.0.0\nreplace ' + MODULE + ' => ../module\n'
            (consumer / 'go.mod').write_text(manifest)
            shutil.copyfile(standalone / 'go.sum', consumer / 'go.sum')
            shutil.copyfile(standalone / 'testdata/consumer/main.go', consumer / 'main.go')
            consumer_before = metadata(consumer)
            graph = set(go_run(consumer, 'list', '-mod=readonly', '-m', '-f', '{{.Path}}', 'all').splitlines())
            assert graph == dependencies | {CONSUMER}, 'External consumer dependency graph drift'
            replacements = [line for line in go_run(consumer, 'list', '-mod=readonly', '-m', '-f',
                            '{{if .Replace}}{{.Path}}{{end}}', 'all').splitlines() if line]
            assert replacements == [MODULE], 'External consumer has a host replacement'
            binary = root / 'metrics-consumer'
            go_run(consumer, 'build', '-race', '-mod=readonly', '-o', str(binary), '.')
            go_run(consumer, 'vet', '-mod=readonly', './...')
            empty = root / 'empty'
            empty.mkdir()
            output = run(str(binary), cwd=empty, env=env).stdout.strip()
            assert output == 'PASS: standalone Metrics HTTP consumer, verified capability wiring and package-owned contract', output
            print(output, flush=True)
            if args.postgres:
                container = 'moh-metrics-test-' + secrets.token_hex(6)
                password = secrets.token_hex(24)
                run('docker', 'run', '--detach', '--rm', '--name', container,
                    '-e', 'POSTGRES_PASSWORD=' + password, '-p', '127.0.0.1::5432',
                    os.environ.get('METRICS_POSTGRES_IMAGE', 'postgres:15-alpine'))
                for _ in range(60):
                    ready = subprocess.run(['docker', 'exec', container, 'pg_isready', '-h', '127.0.0.1', '-U', 'postgres'], capture_output=True)
                    if ready.returncode == 0:
                        break
                    time.sleep(0.5)
                else:
                    raise RuntimeError('Disposable PostgreSQL did not become ready')
                sql = (standalone / 'testdata/consumer/schema.sql').read_text()
                sql += "\nCREATE ROLE metrics_reader LOGIN PASSWORD '" + password + "';\n"
                sql += 'GRANT USAGE ON SCHEMA public TO metrics_reader;\n'
                sql += 'GRANT SELECT ON ALL TABLES IN SCHEMA public TO metrics_reader;\n'
                sql += 'ALTER ROLE metrics_reader SET default_transaction_read_only=on;\n'
                run('docker', 'exec', '-i', container, 'psql', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1', input=sql)
                port = json.loads(run('docker', 'inspect', container).stdout)[0]['NetworkSettings']['Ports']['5432/tcp'][0]['HostPort']
                env['METRICS_TEST_DSN'] = ('host=127.0.0.1 port=' + port +
                    ' user=metrics_reader password=' + password + ' dbname=postgres sslmode=disable')
                output = run(str(binary), cwd=empty, env=env).stdout.strip()
                assert output == 'PASS: Metrics PostgreSQL consumer, 32 repository reads, 19 HTTP routes, NULLs, bound filters, cancellation and caller-owned pool', output
                print(output, flush=True)
            assert metadata(consumer) == consumer_before, 'Consumer dependency metadata changed'
            assert metadata(standalone) == before, 'Consumer modified candidate metadata'
            print('PASS: independent Metrics build, vet, race tests, dependency boundary and external consumer; downloads disabled', flush=True)
    finally:
        if container is not None:
            run('docker', 'rm', '-f', container)

if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        # Never echo command arguments; fixture credentials are temporary.
        raise RuntimeError('Metrics package verification failed:\n' + (error.stdout or '') + '\n' + (error.stderr or '')) from None
