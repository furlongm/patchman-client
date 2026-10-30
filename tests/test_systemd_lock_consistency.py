"""Guard against #866: hardened unit must carve out a writable lock dir and
all clients must agree on it."""
import os
import re

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SERVICE = os.path.join(REPO, 'contrib', 'systemd', 'patchman-client.service')
BASH = os.path.join(REPO, 'bash', 'patchman-client')
PY = os.path.join(REPO, 'python', 'patchman_client', 'client.py')
GO = os.path.join(REPO, 'go', 'cmd', 'submit.go')


def _read(path):
    with open(path, encoding='utf-8') as handle:
        return handle.read()


def _service_values(key):
    values = []
    for line in _read(SERVICE).splitlines():
        line = line.strip()
        if line.startswith('#') or '=' not in line:
            continue
        name, _, value = line.partition('=')
        if name.strip() == key:
            values.append(value.strip())
    return values


def _runtime_lock_dir():
    runtime = _service_values('RuntimeDirectory')
    assert runtime, 'patchman-client.service must set RuntimeDirectory'
    return '/run/' + runtime[0]


def test_hardened_unit_carves_out_a_writable_runtime_dir():
    # if the unit locks the filesystem down, it must provide a writable /run dir
    protect = _service_values('ProtectSystem')
    if protect and protect[0] == 'strict':
        assert _service_values('RuntimeDirectory'), \
            'ProtectSystem=strict needs RuntimeDirectory= or /run is read-only (#866)'
    # a blanket ReadOnlyPaths=/ re-mounts the runtime dir read-only; forbid it
    assert '/' not in _service_values('ReadOnlyPaths'), \
        'ReadOnlyPaths=/ makes the lock dir read-only even with RuntimeDirectory (#866)'


def test_bash_client_lock_dir_matches_runtime_dir():
    match = re.search(r'^lock_dir=(\S+)', _read(BASH), re.MULTILINE)
    assert match, 'could not find lock_dir in bash client'
    assert match.group(1) == _runtime_lock_dir()


def test_python_client_lock_dir_matches_runtime_dir():
    match = re.search(r"^LOCK_DIR\s*=\s*'([^']+)'", _read(PY), re.MULTILINE)
    assert match, 'could not find LOCK_DIR in python client'
    assert match.group(1) == _runtime_lock_dir()


def test_go_client_lock_dir_matches_runtime_dir():
    match = re.search(r'lockDir\s*:=\s*"([^"]+)"', _read(GO))
    assert match, 'could not find lockDir in go client'
    assert match.group(1) == _runtime_lock_dir()
