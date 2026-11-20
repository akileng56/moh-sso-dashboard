#!/usr/bin/env python3
"""Build and test only the Audit module outside the dashboard source tree."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / 'backend/modules/audit'


def main():
    env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local'}
    go = os.environ.get('GO_CMD', 'go')
    with tempfile.TemporaryDirectory(prefix='moh-audit-package-') as tmp:
        standalone = Path(tmp) / 'module'
        shutil.copytree(MODULE, standalone)

        def run(*args):
            return subprocess.run([go, *args], cwd=standalone, env=env, text=True,
                                  check=True, capture_output=True).stdout.strip()

        before = {name: (standalone / name).read_bytes() for name in ('go.mod', 'go.sum')}
        run('mod', 'verify')
        graph = run('list', '-mod=readonly', '-deps', '-test', '-f', '{{.ImportPath}}', './...')
        forbidden = [path for path in graph.splitlines() if path.startswith((
            'github.com/moh-sso-dashboard/', 'github.com/gin-gonic/',
            'github.com/redis/', 'github.com/sqlc-dev/'))]
        assert not forbidden, f'Host/framework dependency leaked into Audit module: {forbidden}'
        run('build', '-mod=readonly', './...')
        run('vet', '-mod=readonly', './...')
        print(run('test', '-race', '-mod=readonly', '-count=1', '-timeout=2m', './...'), flush=True)
        assert all((standalone / name).read_bytes() == contents for name, contents in before.items()), 'Dependency metadata changed during verification'
        print('PASS: standalone Audit build, dependency boundary, vet, race tests and external consumer', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Audit package verification failed: {error.stdout}\n{error.stderr}') from None
