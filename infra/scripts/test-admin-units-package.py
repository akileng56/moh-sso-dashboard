#!/usr/bin/env python3
"""Verify the Administrative Units module and a separate consumer outside the host tree."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / 'backend/modules/admin-units'
MODULE_PATH = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units'
CONSUMER_PATH = 'example.com/admin-units-consumer'
METADATA = ('go.mod', 'go.sum', 'go.work', 'go.work.sum')


def metadata(root):
    return {name: (root / name).read_bytes() if (root / name).exists() else None
            for name in METADATA}


def main():
    go = os.environ.get('GO_CMD', 'go')
    with tempfile.TemporaryDirectory(prefix='moh-admin-units-package-') as tmp:
        root = Path(tmp)
        standalone = root / 'module'
        # The host retains its core until the release/pinning phase.
        host = ROOT / 'backend/internal/features/admin_units'
        for name in ('models.go', 'repository.go', 'service.go', 'hierarchy.go', 'visibility.go', 'core_test.go'):
            assert (host / name).read_bytes() == (MODULE / name).read_bytes(), 'Retained core drift: ' + name
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
            selected = set(run(cwd, 'list', '-mod=readonly', '-m', '-f', '{{.Path}}', 'all').splitlines())
            assert selected == expected, f'Unexpected selected modules: {selected}'

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
        shutil.copyfile(MODULE / 'testdata/consumer/main.go', consumer / 'main.go')
        consumer_before = metadata(consumer)
        check_graph(consumer, {CONSUMER_PATH, MODULE_PATH})
        run(consumer, 'build', '-mod=readonly', './...')
        run(consumer, 'vet', '-mod=readonly', './...')
        output = run(consumer, 'run', '-race', '-mod=readonly', '.')
        assert output == 'PASS: separate Administrative Units consumer (dev), lookup contracts, hierarchy, visibility, nullable values and errors', output
        print(output, flush=True)
        assert metadata(consumer) == consumer_before, 'Consumer dependency metadata changed'
        assert metadata(standalone) == before, 'Consumer modified module metadata'
        print('PASS: isolated Administrative Units build, dependency boundary, vet, race tests and separate consumer; downloads disabled', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Administrative Units package verification failed: {error.stdout}\n{error.stderr}') from None
