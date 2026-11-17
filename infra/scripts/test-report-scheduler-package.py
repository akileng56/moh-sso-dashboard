#!/usr/bin/env python3
"""Verify the scheduler outside the host tree using disposable PostgreSQL only."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time
import uuid

ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / 'backend/modules/report-scheduler'


def run(*args, cwd=ROOT, env=None, sql=None):
    return subprocess.run(args, cwd=cwd, env=env, input=sql, text=True,
                          check=True, capture_output=True).stdout.strip()


def main():
    name = 'moh-scheduler-package-' + uuid.uuid4().hex[:10]
    env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local'}
    # Never inherit an operator's database connection for these tests.
    for key in ('TEST_REPORT_SCHEDULER_DB_URL', 'TEST_REPORT_SCHEDULER_SCHEMA_DB_URL'):
        env.pop(key, None)
    go = os.environ.get('GO_CMD', 'go')
    contract = json.loads((MODULE / 'schema-contract.json').read_text())
    assert contract['owner'] == 'dashboard-host'
    migrations = contract['requiredMigrations']
    assert migrations and max(int(value[:6]) for value in migrations) == contract['minimumHostMigration']
    with tempfile.TemporaryDirectory(prefix='moh-scheduler-standalone-') as tmp:
        standalone = Path(tmp) / 'module'
        shutil.copytree(MODULE, standalone)
        run(go, 'mod', 'verify', cwd=standalone, env=env)
        graph = run(go, 'list', '-mod=readonly', '-deps', '-f', '{{.ImportPath}}', './...', cwd=standalone, env=env)
        forbidden = [value for value in graph.splitlines() if value.startswith('github.com/moh-sso-dashboard/')]
        assert not forbidden, f'Host dependency leaked into standalone module: {forbidden}'
        run(go, 'build', '-mod=readonly', './...', cwd=standalone, env=env)
        run(go, 'vet', '-mod=readonly', './...', cwd=standalone, env=env)
        print('Standalone build, vet and dependency boundary passed', flush=True)
        try:
            run('docker', 'run', '-d', '--name', name, '--label', 'moh.fixture=scheduler-package',
                '--tmpfs', '/var/lib/postgresql/data', '-p', '127.0.0.1::5432',
                '-e', 'POSTGRES_PASSWORD=fixture-only', '-e', 'POSTGRES_DB=scheduler_fixture', 'postgres:15-alpine')
            deadline = time.monotonic() + 60
            while subprocess.run(['docker', 'exec', name, 'pg_isready', '-h', '127.0.0.1', '-U', 'postgres', '-d', 'scheduler_fixture'], capture_output=True).returncode:
                if time.monotonic() > deadline:
                    raise RuntimeError('Disposable PostgreSQL did not become ready')
                time.sleep(.5)
            for migration in migrations:
                assert Path(migration).name == migration
                sql = (ROOT / 'backend/internal/db/migrations' / migration).read_text()
                run('docker', 'exec', '-i', name, 'psql', '-X', '-v', 'ON_ERROR_STOP=1',
                    '-U', 'postgres', '-d', 'scheduler_fixture', sql=sql)
            print('Host-owned scheduler migrations applied: 40 through 47', flush=True)
            port = run('docker', 'port', name, '5432/tcp').rsplit(':', 1)[1]
            dsn = f'postgres://postgres:fixture-only@127.0.0.1:{port}/scheduler_fixture?sslmode=disable'
            env.update(TEST_REPORT_SCHEDULER_DB_URL=dsn, TEST_REPORT_SCHEDULER_SCHEMA_DB_URL=dsn)
            print(run(go, 'test', '-race', '-mod=readonly', '-count=1', '-timeout=3m', './...', cwd=standalone, env=env), flush=True)
            print('PASS: isolated package, PostgreSQL schema, concurrent claims, retries, stale-claim recovery and delivery idempotence', flush=True)
        finally:
            subprocess.run(['docker', 'rm', '-f', name], capture_output=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        # Fixture credentials only; command output contains no operator settings.
        raise RuntimeError(f'Package verification failed: {error.stdout}\n{error.stderr}') from None
