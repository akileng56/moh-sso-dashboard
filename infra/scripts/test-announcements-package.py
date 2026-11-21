#!/usr/bin/env python3
"""Verify the Announcements module and a separate consumer outside the host tree."""

import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
MODULE = ROOT / 'backend/modules/announcements'
MODULE_PATH = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements'
CONSUMER_PATH = 'example.com/announcements-consumer'
METADATA = ('go.mod', 'go.sum', 'go.work', 'go.work.sum')


def metadata(root):
    return {name: (root / name).read_bytes() if (root / name).exists() else None
            for name in METADATA}


def main():
    go = os.environ.get('GO_CMD', 'go')
    with tempfile.TemporaryDirectory(prefix='moh-announcements-package-') as tmp:
        root = Path(tmp)
        standalone = root / 'module'
        shutil.copytree(MODULE, standalone)
        env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local',
               'GOPROXY': 'off', 'GOMODCACHE': str(root / 'private-module-cache')}

        # Only reviewed UUID and PostgreSQL driver artifacts enter the private cache.
        cached = Path(subprocess.run([go, 'env', 'GOMODCACHE'], env={**os.environ, 'GOTOOLCHAIN': 'local'},
                                     text=True, capture_output=True, check=True).stdout.strip())
        checksums = (standalone / 'go.sum').read_text().splitlines()
        reviewed = []
        for dependency, version in [('github.com/google/uuid', 'v1.6.0'),
                                    ('github.com/lib/pq', 'v1.10.9')]:
            sums = [line for line in (ROOT / 'backend/go.sum').read_text().splitlines()
                    if line.startswith(dependency + ' ' + version)]
            assert len(sums) == 2, 'Reviewed dependency checksums missing'
            reviewed.extend(sums)
            relative = Path('cache/download') / dependency / '@v'
            pinned = Path(env['GOMODCACHE']) / relative
            pinned.mkdir(parents=True)
            for suffix in ('mod', 'info', 'zip'):
                shutil.copyfile(cached / relative / (version + '.' + suffix),
                                pinned / (version + '.' + suffix))
            zip_hash = next(line.split()[2] for line in sums if line.split()[1] == version)
            (pinned / (version + '.ziphash')).write_text(zip_hash)
        assert sorted(checksums) == sorted(reviewed), 'Dependency pins/checksums drifted'

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
        check_graph(standalone, {MODULE_PATH, 'github.com/google/uuid', 'github.com/lib/pq'})
        run(standalone, 'build', '-mod=readonly', './...')
        run(standalone, 'vet', '-mod=readonly', './...')
        print(run(standalone, 'test', '-race', '-mod=readonly', '-count=1',
                  '-timeout=2m', './...'), flush=True)
        assert metadata(standalone) == before, 'Module dependency metadata changed'

        consumer = root / 'consumer'
        consumer.mkdir()
        (consumer / 'go.mod').write_text(
            'module ' + CONSUMER_PATH + '\n\ngo 1.25.4\n\nrequire ' +
            MODULE_PATH + ' v0.0.0\nrequire github.com/google/uuid v1.6.0\nrequire github.com/lib/pq v1.10.9\n\nreplace ' + MODULE_PATH + ' => ../module\n')
        shutil.copyfile(MODULE / 'testdata/consumer/main.go', consumer / 'main.go')
        shutil.copyfile(standalone / 'go.sum', consumer / 'go.sum')
        consumer_before = metadata(consumer)
        check_graph(consumer, {CONSUMER_PATH, MODULE_PATH, 'github.com/google/uuid', 'github.com/lib/pq'})
        run(consumer, 'build', '-mod=readonly', './...')
        run(consumer, 'vet', '-mod=readonly', './...')
        output = run(consumer, 'run', '-mod=readonly', '.')
        assert output == 'PASS: separate Announcements consumer (dev), lifecycle, notifications, nullable models, attachments and errors', output
        print(output, flush=True)
        assert metadata(consumer) == consumer_before, 'Consumer dependency metadata changed'
        assert metadata(standalone) == before, 'Consumer modified module metadata'
        print('PASS: isolated Announcements build, dependency boundary, vet, race tests and separate consumer; downloads disabled', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Announcements package verification failed: {error.stdout}\n{error.stderr}') from None
