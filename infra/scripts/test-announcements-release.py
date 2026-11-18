#!/usr/bin/env python3
"""Verify the pinned published Announcements release with a fresh module cache."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
MODULE = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements'
VERSION = 'v0.1.0'


def main():
    go = os.environ.get('GO_CMD', 'go')
    with tempfile.TemporaryDirectory(prefix='moh-announcements-release-') as tmp:
        root = Path(tmp)
        consumer = root / 'consumer'
        consumer.mkdir()
        env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local',
               'GOMODCACHE': str(root / 'fresh-module-cache')}
        # This is a release verification, so normal checksum verification and
        # registry downloads remain enabled. There is no local replacement.
        (consumer / 'go.mod').write_text(
            'module example.com/announcements-release-consumer\n\ngo 1.25.4\n\n'
            'require ' + MODULE + ' ' + VERSION + '\n'
            'require github.com/google/uuid v1.6.0\n')
        sums = [line for line in (ROOT / 'backend/go.sum').read_text().splitlines()
                if line.startswith((MODULE + ' ' + VERSION, 'github.com/google/uuid v1.6.0'))]
        assert len(sums) == 4, 'Reviewed release and UUID checksums are required'
        (consumer / 'go.sum').write_text('\n'.join(sums) + '\n')
        source = (ROOT / 'backend/modules/announcements/testdata/consumer/main.go').read_text()
        source = source.replace('announcements.Version() != "dev"', 'announcements.Version() != "' + VERSION + '"')
        source = source.replace('consumer (dev)', 'consumer (' + VERSION + ')')
        (consumer / 'main.go').write_text(source)
        before = {name: (consumer / name).read_bytes() for name in ('go.mod', 'go.sum')}

        def run(*args):
            result = subprocess.run([go, *args], cwd=consumer, env=env,
                                    text=True, capture_output=True, check=True)
            return result.stdout.strip()

        downloaded = json.loads(run('mod', 'download', '-json', MODULE + '@' + VERSION))
        assert downloaded['Sum'] in (consumer / 'go.sum').read_text()
        assert downloaded['GoModSum'] in (consumer / 'go.sum').read_text()
        selected = json.loads(run('list', '-mod=readonly', '-m', '-json', MODULE))
        assert selected['Version'] == VERSION and not selected.get('Replace'), selected
        run('mod', 'download', 'github.com/google/uuid')
        run('mod', 'verify')
        run('build', '-mod=readonly', './...')
        run('vet', '-mod=readonly', './...')
        output = run('run', '-race', '-mod=readonly', '.')
        assert 'separate Announcements consumer (' + VERSION + ')' in output, output
        print(output, flush=True)
        assert all((consumer / name).read_bytes() == content for name, content in before.items())
        # Compare all shipped Go files to the selected immutable release.
        released = Path(downloaded['Dir'])
        for path in (ROOT / 'backend/modules/announcements').rglob('*.go'):
            relative = path.relative_to(ROOT / 'backend/modules/announcements')
            assert path.read_bytes() == (released / relative).read_bytes(), 'Release source drift: ' + str(relative)
        print('PASS: published v0.1.0 download/checksums, fresh-cache consumer, selected version and source parity', flush=True)


if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        raise RuntimeError(f'Published Announcements verification failed: {error.stdout}\n{error.stderr}') from None
