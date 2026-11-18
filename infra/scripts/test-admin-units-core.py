#!/usr/bin/env python3
"""Build and race-test portable Administrative Units without host files or downloads."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'backend/internal/features/admin_units'
FILES = ('models.go', 'repository.go', 'service.go', 'hierarchy.go', 'visibility.go', 'core_test.go')


def main():
    with tempfile.TemporaryDirectory(prefix='moh-admin-units-core-') as tmp:
        root = Path(tmp)
        for name in FILES:
            shutil.copyfile(SOURCE / name, root / name)
        module = 'example.com/admin-units-core-check'
        (root / 'go.mod').write_text('module ' + module + '\n\ngo 1.25.4\n')
        env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local',
               'GOPROXY': 'off', 'GOMODCACHE': str(root / 'empty-module-cache')}
        go = os.environ.get('GO_CMD', 'go')

        def run(*args):
            return subprocess.run([go, *args], cwd=root, env=env, text=True,
                                  capture_output=True, check=True).stdout.strip()

        before = (root / 'go.mod').read_bytes()
        modules = set(run('list', '-mod=readonly', '-deps', '-test', '-f',
                          '{{if .Module}}{{.Module.Path}}{{end}}', './...').splitlines()) - {''}
        assert modules == {module}, modules
        run('mod', 'verify')
        run('build', '-mod=readonly', './...')
        run('vet', '-mod=readonly', './...')
        print(run('test', '-race', '-mod=readonly', '-count=1', '-timeout=2m', './...'), flush=True)
        assert (root / 'go.mod').read_bytes() == before
        assert not (root / 'go.sum').exists() and not (root / 'go.work').exists()
        print('PASS: portable Administrative Units models, repository, service, hierarchy and visibility; standard library only', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Administrative Units core verification failed: {error.stdout}\n{error.stderr}') from None
