#!/usr/bin/env python3
"""Verify candidate HTTP/provider composition through the independent PostgreSQL harness."""
import importlib.util
from pathlib import Path
import subprocess

spec = importlib.util.spec_from_file_location(
    'announcements_postgres', Path(__file__).with_name('test-announcements-postgres-package.py'))
harness = importlib.util.module_from_spec(spec)
spec.loader.exec_module(harness)

if __name__ == '__main__':
    try:
        harness.main('testdata/http-consumer/main.go')
    except subprocess.CalledProcessError as error:
        raise RuntimeError('Announcements HTTP verification failed:\n' +
                           (error.stdout or '') + '\n' + (error.stderr or '')) from None
