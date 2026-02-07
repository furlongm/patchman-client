"""Patchman Python Client - REST API (Protocol 2)."""

import subprocess


def _get_version():
    """Get version from git tags or fallback."""
    try:
        result = subprocess.run(
            ["git", "describe", "--tags", "--always"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode == 0:
            return result.stdout.strip().lstrip("v")
    except Exception:
        pass
    return "0.0.0"


__version__ = _get_version()
