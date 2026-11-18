#!/usr/bin/env python3
"""Compile and race-test portable Announcements contracts without host files."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'backend/internal/features/announcements'
FILES = ('models.go', 'repository.go', 'notification_port.go', 'settings.go',
         'ports.go', 'service.go', 'attachment_service.go', 'inputs.go', 'values.go',
         'core_portable_test.go')


def main():
    with tempfile.TemporaryDirectory(prefix='moh-announcements-core-') as tmp:
        root = Path(tmp)
        for name in FILES:
            shutil.copyfile(SOURCE / name, root / name)
        (root / 'go.mod').write_text('module example.com/announcements-core-check\n\ngo 1.25.4\n\nrequire github.com/google/uuid v1.6.0\n')
        sums = [line for line in (ROOT / 'backend/go.sum').read_text().splitlines()
                if line.startswith('github.com/google/uuid v1.6.0')]
        assert len(sums) == 2, 'Reviewed UUID checksums are required'
        (root / 'go.sum').write_text('\n'.join(sums) + '\n')
        before = {name: (root / name).read_bytes() for name in ('go.mod', 'go.sum')}
        env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local', 'GOPROXY': 'off'}
        go = os.environ.get('GO_CMD', 'go')

        def run(*args):
            return subprocess.run([go, *args], cwd=root, env=env, text=True,
                                  capture_output=True, check=True).stdout.strip()

        modules = set(run('list', '-mod=readonly', '-deps', '-test', '-f',
                          '{{if .Module}}{{.Module.Path}}{{end}}', './...').splitlines()) - {''}
        assert modules == {'example.com/announcements-core-check', 'github.com/google/uuid'}, modules
        run('mod', 'verify')
        run('build', '-mod=readonly', './...')
        run('vet', '-mod=readonly', './...')
        print(run('test', '-race', '-mod=readonly', '-count=1', '-timeout=2m', './...'), flush=True)
        assert all((root / name).read_bytes() == value for name, value in before.items())
        print('PASS: portable Announcements core builds, vets and race-tests without host files or downloads', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Portable Announcements verification failed: {error.stdout}\n{error.stderr}') from None
