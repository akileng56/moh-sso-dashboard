#!/usr/bin/env python3
"""Exercise Announcements persistence from an independent consumer against PostgreSQL."""
import json
import os
from pathlib import Path
import secrets
import shutil
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
MODULE_PATH = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements'

def run(*args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, text=True, **kwargs)

def main():
    name = 'moh-announcements-test-' + secrets.token_hex(6)
    password = secrets.token_hex(24)
    image = os.environ.get('ANNOUNCEMENTS_POSTGRES_IMAGE', 'postgres:15-alpine')
    go = os.environ.get('GO_CMD', 'go')
    created = False
    try:
        run('docker', 'run', '--detach', '--rm', '--name', name,
            '-e', 'POSTGRES_PASSWORD=' + password,
            '-p', '127.0.0.1::5432', image)
        created = True
        for attempt in range(60):
            ready = subprocess.run(['docker', 'exec', name, 'pg_isready', '-h', '127.0.0.1', '-U', 'postgres'],
                                   capture_output=True)
            if ready.returncode == 0:
                break
            time.sleep(0.5)
        else:
            raise RuntimeError('Disposable PostgreSQL did not become ready')
        sql = (ROOT / 'backend/modules/announcements/testdata/postgres-consumer/schema.sql').read_text()
        run('docker', 'exec', '-i', name, 'psql', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1', input=sql)
        bindings = json.loads(run('docker', 'inspect', name).stdout)[0]['NetworkSettings']['Ports']['5432/tcp']
        port = bindings[0]['HostPort']
        with tempfile.TemporaryDirectory(prefix='moh-announcements-postgres-') as tmp:
            root = Path(tmp)
            shutil.copytree(ROOT / 'backend/modules/announcements', root / 'module')
            consumer = root / 'consumer'
            consumer.mkdir()
            (consumer / 'go.mod').write_text(
                'module example.com/announcements-postgres-consumer\n\ngo 1.25.4\n\nrequire (\n' +
                MODULE_PATH + ' v0.0.0\n github.com/google/uuid v1.6.0\n github.com/lib/pq v1.10.9\n)\n\nreplace ' +
                MODULE_PATH + ' => ../module\n')
            sums = (root / 'module/go.sum').read_text()
            (consumer / 'go.sum').write_text(sums)
            shutil.copyfile(root / 'module/testdata/postgres-consumer/main.go', consumer / 'main.go')
            env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local', 'GOPROXY': 'off'}
            cached = Path(run(go, 'env', 'GOMODCACHE').stdout.strip())
            private = root / 'private-module-cache'
            reviewed_sums = (ROOT / 'backend/go.sum').read_text().splitlines()
            for dependency, version in [('github.com/google/uuid', 'v1.6.0'),
                                        ('github.com/lib/pq', 'v1.10.9')]:
                relative = Path('cache/download') / dependency / '@v'
                pinned = private / relative
                pinned.mkdir(parents=True)
                expected = [line for line in reviewed_sums if line.startswith(dependency + ' ' + version)]
                assert len(expected) == 2 and all(line in sums.splitlines() for line in expected)
                for suffix in ('mod', 'info', 'zip'):
                    shutil.copyfile(cached / relative / (version + '.' + suffix),
                                    pinned / (version + '.' + suffix))
                digest = next(line.split()[2] for line in expected if line.split()[1] == version)
                (pinned / (version + '.ziphash')).write_text(digest)
            env['GOMODCACHE'] = str(private)
            before = {p: p.read_bytes() for p in root.rglob('go.*') if p.is_file()}
            env['ANNOUNCEMENTS_TEST_DSN'] = (
                'host=127.0.0.1 port=' + port + ' user=postgres password=' +
                password + ' dbname=postgres sslmode=disable')
            binary = root / 'consumer-bin'
            run(go, 'build', '-race', '-mod=readonly', '-o', str(binary), '.', cwd=consumer, env=env)
            run(go, 'vet', '-mod=readonly', './...', cwd=consumer, env=env)
            empty = root / 'empty'
            empty.mkdir()
            result = run(str(binary), cwd=empty, env=env)
            print(result.stdout.strip(), flush=True)
            assert all(p.read_bytes() == value for p, value in before.items()), 'Dependency metadata changed'
    finally:
        if created:
            run('docker', 'rm', '-f', name)

if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        # Never echo command arguments: the fixture DSN contains its password.
        raise RuntimeError('PostgreSQL package verification failed:\n' +
                           (error.stdout or '') + '\n' + (error.stderr or '')) from None
