#!/usr/bin/env python3
"""
Patchman Python Client - REST API Only (Protocol 2)

A Python implementation of the patchman client that uses the REST API.
Collects system information (packages, repos, modules, updates) and
submits reports to a Patchman server.

Usage:
    patchman-client.py [-v] [-d] [-n] [-u] [-s SERVER] [-c FILE]
                       [-t TAGS] [-h HOSTNAME] [-k API_KEY]

Requirements:
    - Python 3.6+
    - requests library (pip install requests)
"""

import argparse
import atexit
import fcntl
import json
import os
import platform
import re
import shutil
import socket
import subprocess
import sys
from pathlib import Path
from typing import Any

# Optional: try to import requests, fall back to urllib if not available
try:
    import requests
    HAS_REQUESTS = True
except ImportError:
    import ssl
    import urllib.error
    import urllib.request
    HAS_REQUESTS = False


__version__ = '2.1.0'
DEFAULT_CONFIG = '/etc/patchman/patchman-client.conf'
DEFAULT_CERT_PATH = '/etc/patchman/client.crt'
DEFAULT_KEY_PATH = '/etc/patchman/client.key'
LOCK_DIR = '/run/lock/patchman'
LOCK_FILE = os.path.join(LOCK_DIR, 'patchman.lock')

# Try to import cryptography for certificate operations
try:
    from cryptography import x509
    from cryptography.hazmat.primitives import hashes, serialization
    from cryptography.hazmat.primitives.asymmetric import rsa
    from cryptography.x509.oid import NameOID
    HAS_CRYPTOGRAPHY = True
except ImportError:
    HAS_CRYPTOGRAPHY = False


class PatchmanClient:
    """Patchman client for collecting and reporting system information."""

    def __init__(
        self,
        server: str = '',
        api_key: str = '',
        hostname: str = '',
        tags: str = '',
        verbose: bool = False,
        debug: bool = False,
        local_updates: bool = False,
        repo_check: bool = True,
        verify_ssl: bool = True,
        timeout: int = 300,
        protocol: int = 2,
        dry_run: bool = False,
        client_cert: str = '',
        client_key: str = '',
        cert_renew_threshold: int = 30,
    ):
        self.server = server.rstrip('/') if server else ''
        self.api_key = api_key
        self.hostname = hostname or self._get_hostname()
        self.tags = [t.strip() for t in tags.split(',') if t.strip()] if tags else []
        self.verbose = verbose
        self.debug = debug
        self.local_updates = local_updates
        self.repo_check = repo_check
        self.verify_ssl = verify_ssl
        self.timeout = timeout
        self.protocol = protocol
        self.dry_run = dry_run
        self.client_cert = client_cert
        self.client_key = client_key or client_cert
        self.cert_renew_threshold = cert_renew_threshold

        # System info (populated by collect methods)
        self.kernel = ''
        self.arch = ''
        self.os_string = ''
        self.cpe_name = ''  # CPE name for prefixing repos/updates
        self.reboot_required = False

        # Collected data
        self.packages: list[dict[str, Any]] = []
        self.repos: list[dict[str, Any]] = []
        self.modules: list[dict[str, Any]] = []
        self.sec_updates: list[dict[str, Any]] = []
        self.bug_updates: list[dict[str, Any]] = []

    def log(self, message: str, level: str = 'info') -> None:
        """Print log messages based on verbosity level."""
        if level == 'debug' and not self.debug:
            return
        if level == 'info' and not self.verbose:
            return
        print(message, file=sys.stderr)

    def _get_hostname(self) -> str:
        """Get the fully qualified hostname."""
        try:
            return socket.getfqdn()
        except Exception:
            return socket.gethostname()

    def _run_command(
        self, cmd: list[str], check: bool = False
    ) -> tuple[int, str, str]:
        """Run a shell command and return (returncode, stdout, stderr)."""
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=60,
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, '', 'Command timed out'
        except FileNotFoundError:
            return -1, '', f'Command not found: {cmd[0]}'
        except Exception as e:
            return -1, '', str(e)

    def _command_exists(self, cmd: str) -> bool:
        """Check if a command exists in PATH."""
        return shutil.which(cmd) is not None

    # -------------------------------------------------------------------------
    # System Information Collection
    # -------------------------------------------------------------------------

    def collect_host_data(self) -> None:
        """Collect kernel, architecture, and OS information."""
        self.kernel = platform.release()
        self.arch = platform.machine()
        self.os_string = self._detect_os()
        self.log(f'Kernel:   {self.kernel}')
        self.log(f'Arch:     {self.arch}')
        self.log(f'OS:       {self.os_string}')

    def _detect_os(self) -> str:
        """Detect the operating system name and version."""
        os_release = {}

        # Try /etc/os-release first
        if os.path.exists('/etc/os-release'):
            with open('/etc/os-release') as f:
                for line in f:
                    line = line.strip()
                    if '=' in line:
                        key, _, value = line.partition('=')
                        os_release[key] = value.strip('"')

            os_id = os_release.get('ID', '')
            pretty_name = os_release.get('PRETTY_NAME', '')
            name = os_release.get('NAME', '')
            version = os_release.get('VERSION', '')
            version_id = os_release.get('VERSION_ID', '')
            id_like = os_release.get('ID_LIKE', '')
            cpe_name = os_release.get('CPE_NAME', '')

            # Store CPE name for prefixing repos/updates
            self.cpe_name = cpe_name

            if os_id == 'debian':
                deb_version = self._read_file('/etc/debian_version', '').replace('/', '-')
                return f'Debian {deb_version}'
            elif os_id == 'raspbian':
                deb_version = self._read_file('/etc/debian_version', '')
                return f'Raspbian {deb_version}'
            elif os_id == 'ubuntu':
                return pretty_name
            elif os_id == 'centos':
                release = self._read_file('/etc/centos-release', pretty_name)
                if cpe_name:
                    return f'{release} [{cpe_name}]'
                return release
            elif os_id == 'rhel':
                release = self._read_file('/etc/redhat-release', pretty_name)
                if cpe_name:
                    return f'{release} [{cpe_name}]'
                return release
            elif os_id == 'fedora':
                if cpe_name:
                    return f'{pretty_name} [{cpe_name}]'
                return pretty_name
            elif os_id == 'arch':
                return name
            elif os_id == 'gentoo':
                return f'{pretty_name} {version_id}'
            elif 'suse' in id_like:
                if cpe_name:
                    return f'{pretty_name} [{cpe_name}]'
                return pretty_name
            elif os_id == 'astra':
                astra_version = self._read_file('/etc/astra_version', version_id)
                return f'{name} {astra_version}'
            elif os_id in ('rocky', 'almalinux'):
                if cpe_name:
                    return f'{pretty_name} [{cpe_name}]'
                return pretty_name
            else:
                result = f'{name} {version}'
                if cpe_name:
                    result = f'{result} [{cpe_name}]'
                return result

        # Fallback: check various release files
        release_files = [
            ('/etc/SuSE-release', lambda c: next((line for line in c.splitlines() if 'suse' in line.lower()), '')),
            ('/etc/lsb-release', self._parse_lsb_release),
            ('/etc/debian_version', lambda c: f'Debian {c.strip()}'),
            ('/etc/redhat-release', lambda c: c.strip()),
            ('/etc/fedora-release', lambda c: c.strip()),
            ('/etc/centos-release', lambda c: c.strip()),
        ]

        for path, parser in release_files:
            if os.path.exists(path):
                content = self._read_file(path, '')
                if content:
                    result = parser(content)
                    if result:
                        return result

        return 'unknown'

    def _read_file(self, path: str, default: str = '') -> str:
        """Read file contents, return default on error."""
        try:
            with open(path) as f:
                return f.read().strip()
        except Exception:
            return default

    def _parse_lsb_release(self, content: str) -> str:
        """Parse /etc/lsb-release content."""
        for line in content.splitlines():
            if line.startswith('DISTRIB_DESCRIPTION='):
                return line.split('=', 1)[1].strip('"')
            elif line.startswith('DISTRIB_DESC='):
                return line.split('=', 1)[1].strip('"')
        return ''

    def check_reboot_required(self) -> None:
        """Check if a reboot is required."""
        # Debian/Ubuntu: update-notifier-common sets this file
        if os.path.exists('/var/run/reboot-required'):
            self.reboot_required = True
            self.log('Reboot required: Yes (reboot-required file)')
            return

        # Compare running vs installed kernel via /boot/vmlinuz symlink
        try:
            running = Path('/proc/sys/kernel/osrelease').read_text().strip()
            vmlinuz = Path('/boot/vmlinuz')
            if vmlinuz.is_symlink():
                installed = vmlinuz.resolve().name.removeprefix('vmlinuz-')
                if running != installed:
                    self.reboot_required = True
                    self.log('Reboot required: Yes (kernel mismatch)')
                    return
        except OSError:
            pass

        self.reboot_required = False
        self.log('Reboot required: No (server will check)')

    # -------------------------------------------------------------------------
    # Package Collection
    # -------------------------------------------------------------------------

    def collect_packages(self) -> None:
        """Collect installed packages from all package managers."""
        self.packages = []
        self._collect_rpm_packages()
        self._collect_deb_packages()
        self._collect_arch_packages()
        self._collect_gentoo_packages()
        self.log(f'Collected {len(self.packages)} packages')

    def _collect_rpm_packages(self) -> None:
        """Collect packages from RPM database."""
        if not self._command_exists('rpm'):
            return

        self.log('Finding installed rpms...')
        fmt = '%{NAME}\\t%{EPOCH}\\t%{VERSION}\\t%{RELEASE}\\t%{ARCH}\\n'
        rc, stdout, _ = self._run_command(['rpm', '-qa', '--queryformat', fmt])

        if rc != 0:
            return

        for line in stdout.splitlines():
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) < 5:
                continue

            name, epoch, version, release, arch = parts[:5]

            # Skip gpg-pubkey entries
            if name == 'gpg-pubkey':
                continue

            # Clean up epoch
            if epoch == '(none)':
                epoch = ''

            self.packages.append({
                'name': name,
                'epoch': epoch,
                'version': version,
                'release': release,
                'arch': arch,
                'type': 'rpm',
            })

    def _collect_deb_packages(self) -> None:
        """Collect packages from dpkg database."""
        if not self._command_exists('dpkg-query'):
            return

        self.log('Finding installed debs...')
        fmt = '${Status}|${Package}|${Version}|${Architecture}\\n'
        rc, stdout, _ = self._run_command(['dpkg-query', '-W', '--showformat', fmt])

        if rc != 0:
            return

        for line in stdout.splitlines():
            if not line.strip():
                continue

            # Only include installed packages
            if not line.startswith(('install ok installed|', 'hold ok installed|')):
                continue

            parts = line.split('|')
            if len(parts) < 4:
                continue

            _, name, version, arch = parts[:4]

            # Parse epoch from version
            epoch = ''
            if ':' in version:
                epoch, version = version.split(':', 1)

            # Parse release from version
            release = ''
            if '-' in version:
                idx = version.rfind('-')
                release = version[idx + 1:]
                version = version[:idx]

            self.packages.append({
                'name': name,
                'epoch': epoch,
                'version': version,
                'release': release,
                'arch': arch,
                'type': 'deb',
            })

    def _collect_arch_packages(self) -> None:
        """Collect packages from pacman database."""
        if not self._command_exists('pacman'):
            return

        self.log('Finding installed Arch packages...')
        rc, stdout, _ = self._run_command(['pacman', '-Q', '-i'])

        if rc != 0:
            return

        current_pkg: dict[str, str] = {}

        for line in stdout.splitlines():
            line = line.strip()
            if not line:
                # End of package entry
                if current_pkg.get('name'):
                    self.packages.append(current_pkg)
                current_pkg = {}
                continue

            if line.startswith('Name'):
                current_pkg = {'type': 'arch', 'epoch': '', 'release': ''}
                current_pkg['name'] = line.split(':', 1)[1].strip()
            elif line.startswith('Version') and current_pkg:
                version = line.split(':', 1)[1].strip()
                # Parse epoch if present
                if ':' in version:
                    epoch, version = version.split(':', 1)
                    current_pkg['epoch'] = epoch
                # Parse release
                if '-' in version:
                    version, release = version.rsplit('-', 1)
                    current_pkg['release'] = release
                current_pkg['version'] = version
            elif line.startswith('Architecture') and current_pkg:
                current_pkg['arch'] = line.split(':', 1)[1].strip()

        # Don't forget last package
        if current_pkg.get('name'):
            self.packages.append(current_pkg)

    def _collect_gentoo_packages(self) -> None:
        """Collect packages from Gentoo portage."""
        if not self._command_exists('qlist'):
            return

        self.log('Finding installed Gentoo packages...')

        # Get architecture
        arch = ''
        if self._command_exists('qkeyword'):
            rc, stdout, _ = self._run_command(['qkeyword', '-A'])
            if rc == 0:
                arch = stdout.strip()

        # Get packages
        fmt = '%{PN}\\t%{SLOT}\\t%{PV}\\t%{PR}\\t%{CAT}\\t%{REPO}'
        rc, stdout, _ = self._run_command(['qlist', '-Ic', '-F', fmt])

        if rc != 0:
            return

        for line in stdout.splitlines():
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) < 6:
                continue

            name, slot, version, release, category, repo = parts[:6]

            # Clean up release (remove 'r' prefix if it's just 'r0')
            if release == 'r0':
                release = ''
            elif release.startswith('r'):
                release = release[1:]

            self.packages.append({
                'name': name,
                'epoch': slot,  # Gentoo uses slots like epochs
                'version': version,
                'release': release,
                'arch': arch,
                'type': 'gentoo',
                'category': category,
                'repo': repo,
            })

    # -------------------------------------------------------------------------
    # Repository Collection
    # -------------------------------------------------------------------------

    def collect_repos(self) -> None:
        """Collect configured repositories."""
        self.repos = []
        self._collect_yum_repos()
        self._collect_apt_repos()
        self._collect_zypper_repos()
        self._collect_pacman_repos()
        self._collect_gentoo_repos()
        self.log(f'Collected {len(self.repos)} repositories')

        # Auto-enable local_updates if CDN repos detected (like bash client)
        if not self.local_updates:
            for repo in self.repos:
                for url in repo.get('urls', []):
                    if (re.search(r'https://cdn[^.]*\.redhat\.com/', url) or
                            '/XMLRPC' in url or
                            url.startswith('https://updates.suse.com/')):
                        self.local_updates = True
                        self.log('CDN repo detected, enabling local update collection')
                        break
                if self.local_updates:
                    break

    def _collect_yum_repos(self) -> None:
        """Collect YUM/DNF repositories."""
        if not self._command_exists('yum'):
            return

        self.log('Finding yum repos...')

        # Parse priorities from .repo files (yum repolist doesn't show them)
        repo_priorities = self._parse_yum_priorities()

        rc, stdout, _ = self._run_command(
            ['yum', 'repolist', 'enabled', '--verbose'],
        )

        if rc != 0:
            return

        current_repo: dict[str, Any] = {}
        urls: list[str] = []

        for line in stdout.splitlines():
            line = line.strip()

            if line.startswith('Repo-id'):
                # Save previous repo if exists
                if current_repo.get('id'):
                    current_repo['urls'] = urls
                    self.repos.append(current_repo)
                # Start new repo
                repo_id = line.split(':', 1)[1].strip() if ':' in line else ''
                # Remove any trailing /version info
                repo_id = repo_id.split('/')[0]
                # Look up priority from .repo files by name (matched later)
                # Default yum priority is 99
                priority = 99
                # Prefix with CPE name if available
                if self.cpe_name:
                    repo_id = f'{self.cpe_name}-{repo_id}'
                current_repo = {
                    'type': 'rpm',
                    'id': repo_id,
                    'name': '',
                    'priority': priority,
                    'urls': [],
                }
                urls = []

            elif line.startswith('Repo-name') and current_repo:
                name = line.split(':', 1)[1].strip() if ':' in line else ''
                # Remove trailing " - arch" pattern and re-add arch consistently
                if name.endswith(f' - {self.arch}'):
                    name = name[:-len(f' - {self.arch}')]
                if self.arch not in name:
                    name = f'{name} {self.arch}'
                current_repo['name'] = name
                # Match priority by repo name (with arch appended)
                if name in repo_priorities:
                    current_repo['priority'] = repo_priorities[name]

            elif line.startswith('Repo-baseurl') and current_repo:
                url = line.split(':', 1)[1].strip() if ':' in line else ''
                # Handle "Repo-baseurl: url1, url2" format
                for u in url.split(','):
                    u = u.strip()
                    # Strip "(N more)" suffix from URL
                    if ' (' in u:
                        u = u.split(' (')[0]
                    if u and u.startswith(('http://', 'https://')):
                        urls.append(u)

            elif (line.startswith('Repo-metalink') or
                  line.startswith('Repo-mirrors')) and current_repo:
                url = line.split(':', 1)[1].strip() if ':' in line else ''
                if url and url.startswith(('http://', 'https://')):
                    urls.append(url)

            elif line.startswith(':') and current_repo and urls:
                # Continuation line for URLs
                url = line[1:].strip()
                if url.startswith(('http://', 'https://')):
                    urls.append(url.rstrip(','))

        # Don't forget last repo
        if current_repo.get('id'):
            current_repo['urls'] = urls
            self.repos.append(current_repo)

    def _parse_yum_priorities(self) -> dict[str, int]:
        """Parse priorities from /etc/yum.repos.d/*.repo files.

        Returns a dict mapping 'repo name arch' to priority value.
        The bash client does this with sed; we parse the .repo INI files.
        """
        import glob
        priorities: dict[str, int] = {}
        repo_files = glob.glob('/etc/yum.repos.d/*.repo')
        current_name = ''
        for repo_file in repo_files:
            try:
                with open(repo_file) as f:
                    for line in f:
                        line = line.strip()
                        if line.startswith('name'):
                            parts = line.split('=', 1)
                            if len(parts) == 2:
                                name = parts[1].strip()
                                name = name.replace('$releasever', '')
                                if self.arch not in name:
                                    name = f'{name} {self.arch}'
                                current_name = name
                        elif line.startswith('priority') and current_name:
                            parts = line.split('=', 1)
                            if len(parts) == 2:
                                try:
                                    priorities[current_name] = int(parts[1].strip())
                                except ValueError:
                                    pass
            except OSError:
                continue
        return priorities

    def _collect_apt_repos(self) -> None:
        """Collect APT repositories."""
        if not self._command_exists('apt-cache'):
            return

        self.log('Finding apt repos...')
        rc, stdout, _ = self._run_command(['apt-cache', 'policy'])

        if rc != 0:
            return

        # Parse OS info for repo naming
        os_name = self.os_string.split()[0] if self.os_string else 'Unknown'
        version_id = ''
        if os.path.exists('/etc/os-release'):
            with open('/etc/os-release') as f:
                for line in f:
                    if line.startswith('VERSION_ID='):
                        version_id = line.split('=', 1)[1].strip().strip('"')
                        break

        # Parse apt-cache policy output
        # Matches both http(s):// URLs and mirror+file: entries
        repo_pattern = re.compile(
            r'^\s*(\d+)\s+(https?://\S+|mirror\+file:\S+)\s+(\S+)\s+(\S+)',
        )

        mirror_entries = []
        http_entries = []

        for line in stdout.splitlines():
            match = repo_pattern.match(line)
            if not match:
                continue

            # Skip Translation entries
            if 'Translation' in line:
                continue

            priority, base_url, dist, component = match.groups()

            if base_url.startswith('mirror+file:'):
                mirror_entries.append((priority, base_url, dist, component))
            else:
                http_entries.append((priority, base_url, dist, component))

        # Resolve mirror+file: entries by reading the local mirror list file
        for priority, mirror_url, dist, component in mirror_entries:
            mirror_path = mirror_url.replace('mirror+file:', '')
            if os.path.isfile(mirror_path):
                try:
                    with open(mirror_path) as f:
                        for url_line in f:
                            url = url_line.strip()
                            if url and url.startswith(('http://', 'https://')):
                                http_entries.append(
                                    (priority, url, dist, component),
                                )
                except OSError:
                    pass

        # Build repo entries from resolved URLs
        for priority, base_url, dist, component in http_entries:
            if component.endswith('Packages'):
                url = f'{base_url}/{dist}'
            else:
                url = f'{base_url}/dists/{dist}/binary-{self.arch}'

            repo_name = f'{os_name} {version_id} {self.arch} repo at {url}'

            self.repos.append({
                'type': 'deb',
                'name': repo_name,
                'id': '',
                'priority': int(priority),
                'urls': [url],
            })

    def _collect_zypper_repos(self) -> None:
        """Collect Zypper repositories."""
        if not self._command_exists('zypper'):
            return

        self.log('Finding zypper repos...')
        rc, stdout, _ = self._run_command(
            ['zypper', '-q', '--no-refresh', 'lr', '-E', '-u', '--details'],
        )

        if rc != 0:
            return

        # Get PRETTY_NAME for repo naming (matches bash client behavior)
        pretty_name = ''
        if os.path.exists('/etc/os-release'):
            with open('/etc/os-release') as f:
                for line in f:
                    if line.startswith('PRETTY_NAME='):
                        pretty_name = line.split('=', 1)[1].strip().strip('"')
                        break

        # Check if output has "Keep" column (shifts indices)
        # Format with Keep:    # | Alias | Name | Enabled | GPG | Refresh | Keep | Priority | Type | URI | Service
        # Format without Keep: # | Alias | Name | Enabled | GPG | Refresh | Priority | Type | URI
        lines = stdout.splitlines()
        has_keep_column = any('Keep' in line for line in lines[:3])

        if has_keep_column:
            priority_idx = 7
            uri_idx = 9
        else:
            priority_idx = 6
            uri_idx = 8

        # Skip header lines
        for line in lines:
            if not line.strip() or line.startswith('#') or line.startswith('-'):
                continue

            parts = [p.strip() for p in line.split('|')]
            if len(parts) <= uri_idx:
                continue

            # Format: # | Alias | Name | ...
            try:
                alias = parts[1]
                name = parts[2]
                priority_str = parts[priority_idx]
                priority = int(priority_str) if priority_str.isdigit() else 99
                uri = parts[uri_idx]
            except (IndexError, ValueError):
                continue

            self.repos.append({
                'type': 'rpm',
                'name': f'{pretty_name} {name} {self.arch}',
                'id': alias,
                'priority': priority,
                'urls': [uri],
            })

    def _collect_pacman_repos(self) -> None:
        """Collect Pacman repositories."""
        if not self._command_exists('pacman'):
            return

        self.log('Finding pacman repos...')

        pacman_conf = Path('/etc/pacman.conf')
        if not pacman_conf.exists():
            return

        repos: dict[str, list[str]] = {}
        current_repo = ''

        with open(pacman_conf) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue

                # Section header
                if line.startswith('[') and line.endswith(']'):
                    section = line[1:-1]
                    if section != 'options':
                        current_repo = section
                        repos[current_repo] = []
                    continue

                # Include directive
                if current_repo and line.startswith('Include'):
                    include_path = line.split('=', 1)[1].strip()
                    # Handle glob patterns
                    for conf_file in Path('/').glob(include_path.lstrip('/')):
                        if conf_file.is_file():
                            servers = self._parse_pacman_mirrorlist(conf_file)
                            repos[current_repo].extend(servers)

                # Server directive
                elif current_repo and line.startswith('Server'):
                    url = line.split('=', 1)[1].strip()
                    # Expand variables
                    url = url.replace('$repo', current_repo)
                    url = url.replace('$arch', self.arch)
                    repos[current_repo].append(url)

        for repo, urls in repos.items():
            if urls:
                self.repos.append({
                    'type': 'arch',
                    'name': f'Arch Linux {repo} {self.arch}',
                    'id': repo,
                    'priority': 0,
                    'urls': urls,
                })

    def _parse_pacman_mirrorlist(self, path: Path) -> list[str]:
        """Parse a pacman mirrorlist file."""
        servers = []
        try:
            with open(path) as f:
                for line in f:
                    line = line.strip()
                    if line.startswith('Server'):
                        url = line.split('=', 1)[1].strip()
                        servers.append(url)
        except Exception:
            pass
        return servers

    def _collect_gentoo_repos(self) -> None:
        """Collect Gentoo portage repositories."""
        if not self._command_exists('portageq'):
            return

        self.log('Finding portage repos...')
        rc, stdout, _ = self._run_command(['portageq', 'repos_config', '/'])

        if rc != 0:
            return

        current_repo = ''
        priority = 0
        sync_uri = ''

        for line in stdout.splitlines():
            line = line.strip()

            # Section header
            if line.startswith('[') and line.endswith(']'):
                # Save previous repo
                if current_repo and sync_uri:
                    self.repos.append({
                        'type': 'gentoo',
                        'name': f'Gentoo {current_repo}',
                        'id': current_repo,
                        'priority': priority,
                        'urls': [sync_uri],
                    })
                current_repo = line[1:-1]
                priority = 0
                sync_uri = ''
                continue

            if '=' in line:
                key, value = line.split('=', 1)
                key = key.strip()
                value = value.strip()

                if key == 'priority':
                    try:
                        priority = int(value)
                    except ValueError:
                        priority = 0
                elif key == 'sync-uri':
                    sync_uri = value

        # Don't forget last repo
        if current_repo and sync_uri:
            self.repos.append({
                'type': 'gentoo',
                'name': f'Gentoo {current_repo}',
                'id': current_repo,
                'priority': priority,
                'urls': [sync_uri],
            })

    # -------------------------------------------------------------------------
    # Module Collection (RHEL/CentOS 8+ modularity)
    # -------------------------------------------------------------------------

    def collect_modules(self) -> None:
        """Collect enabled DNF modules."""
        self.modules = []

        if not self._command_exists('yum'):
            return

        # Check if modularity is supported
        rc, _, stderr = self._run_command(['yum', 'module', 'list', '--enabled'])
        if rc != 0 or 'No such command' in stderr:
            return

        self.log('Finding enabled yum modules...')
        rc, stdout, _ = self._run_command(['yum', 'module', 'list', '--enabled'])

        if rc != 0:
            return

        # Parse enabled modules
        for line in stdout.splitlines():
            if '[e]' not in line:
                continue
            if line.startswith('Hint'):
                continue

            parts = line.split()
            if not parts:
                continue

            module_name = parts[0]

            # Get detailed module info
            self._collect_module_info(module_name)

        self.log(f'Collected {len(self.modules)} modules')

    def _collect_module_info(self, module_name: str) -> None:
        """Collect detailed information for a module."""
        rc, stdout, _ = self._run_command(['yum', 'module', 'info', module_name])

        if rc != 0:
            return

        current_module: dict[str, Any] = {}
        in_artifacts = False
        packages: list[str] = []

        for line in stdout.splitlines():
            line = line.strip()

            if not line:
                # End of module entry - save if enabled
                if current_module.get('stream') and '[e]' in current_module.get('stream', ''):
                    current_module['stream'] = current_module['stream'].replace(' [e]', '').strip()
                    current_module['packages'] = packages
                    self.modules.append(current_module)
                current_module = {}
                packages = []
                in_artifacts = False
                continue

            if line.startswith('Hint'):
                continue

            if line.startswith('Name'):
                current_module = {'name': line.split(':', 1)[1].strip()}
            elif line.startswith('Stream') and current_module:
                current_module['stream'] = line.split(':', 1)[1].strip()
            elif line.startswith('Version') and current_module:
                current_module['version'] = line.split(':', 1)[1].strip()
            elif line.startswith('Context') and current_module:
                current_module['context'] = line.split(':', 1)[1].strip()
            elif line.startswith('Architecture') and current_module:
                current_module['arch'] = line.split(':', 1)[1].strip()
            elif line.startswith('Repo') and current_module:
                repo = line.split(':', 1)[1].strip()
                # Prefix with CPE name if available
                if self.cpe_name:
                    repo = f'{self.cpe_name}-{repo}'
                current_module['repo'] = repo
            elif line.startswith('Artifacts'):
                in_artifacts = True
            elif in_artifacts and line.startswith(':'):
                pkg = line.lstrip(': ').strip()
                if pkg:
                    packages.append(pkg)
            elif in_artifacts and not line.startswith(':'):
                in_artifacts = False

    # -------------------------------------------------------------------------
    # Updates Collection
    # -------------------------------------------------------------------------

    def collect_updates(self) -> None:
        """Collect available updates."""
        self.sec_updates = []
        self.bug_updates = []

        if self.local_updates:
            self._collect_yum_updates()
            self._collect_zypper_updates()
            self._collect_apt_updates()

        self.log(f'Collected {len(self.sec_updates)} security updates')
        self.log(f'Collected {len(self.bug_updates)} bug fix updates')

    def _collect_yum_updates(self) -> None:
        """Collect updates from YUM/DNF."""
        if not self._command_exists('yum'):
            return

        self.log('Finding yum updates...')

        # Get security updates
        rc, stdout, _ = self._run_command(
            ['yum', '-q', '-C', '--security', 'list', 'updates'],
        )

        if rc == 0:
            self._parse_yum_updates(stdout, is_security=True)

        # Get all updates
        rc, stdout, _ = self._run_command(
            ['yum', '-q', '-C', 'list', 'updates'],
        )

        if rc == 0:
            self._parse_yum_updates(stdout, is_security=False)

    def _parse_yum_updates(self, output: str, is_security: bool) -> None:
        """Parse yum list updates output."""
        skip_patterns = [
            'Available Upgrades',
            'Updated Packages',
            'excluded',
            'Last metadata expiration',
            'needed for security',
            'Loaded plugins',
            'Subscription Management',
            'Failed to set locale',
            'Limiting package lists',
        ]

        target = self.sec_updates if is_security else self.bug_updates

        for line in output.splitlines():
            line = line.strip()
            if not line:
                continue

            # Skip info lines
            if any(p in line for p in skip_patterns):
                continue

            # Parse package line: name.arch version repo
            parts = line.split()
            if len(parts) < 3:
                continue

            name_arch = parts[0]
            version = parts[1]
            repo = parts[2]

            # Prefix repo with CPE name if available
            if self.cpe_name:
                repo = f'{self.cpe_name}-{repo}'

            # Split name.arch
            if '.' in name_arch:
                name, arch = name_arch.rsplit('.', 1)
            else:
                name = name_arch
                arch = self.arch

            target.append({
                'name': name,
                'version': version,
                'arch': arch,
                'repo': repo,
            })

    def _collect_zypper_updates(self) -> None:
        """Collect updates from Zypper."""
        if not self._command_exists('zypper'):
            return

        self.log('Finding zypper updates...')
        rc, stdout, _ = self._run_command(
            ['zypper', '-q', '-n', '-s11', 'lu'],
        )

        if rc != 0:
            return

        for line in stdout.splitlines():
            if not line.startswith('v'):
                continue

            parts = line.split('|')
            if len(parts) < 5:
                continue

            name = parts[1].strip()
            version = parts[3].strip()
            arch = parts[4].strip()
            repo = parts[5].strip() if len(parts) > 5 else ''

            # Zypper doesn't differentiate security/bugfix easily
            self.bug_updates.append({
                'name': name,
                'version': version,
                'arch': arch,
                'repo': repo,
            })

    def _collect_apt_updates(self) -> None:
        """Collect updates from APT (apt list --upgradable)."""
        if not self._command_exists('apt'):
            return

        self.log('Finding apt updates...')
        rc, stdout, _ = self._run_command(['apt', 'list', '--upgradable'])

        if rc != 0:
            return

        # Format: package/suite version arch [upgradable from: old-version]
        # Example: curl/jammy-security 7.81.0-1ubuntu1.15 amd64 [upgradable from: ...]
        update_pattern = re.compile(r'^([^/]+)/(\S+)\s+(\S+)\s+(\S+)')

        for line in stdout.splitlines():
            line = line.strip()
            if not line or line.startswith('Listing'):
                continue

            match = update_pattern.match(line)
            if not match:
                continue

            pkg_name, suite, version, arch = match.groups()

            update = {
                'name': pkg_name,
                'version': version,
                'arch': arch,
            }

            if 'security' in suite.lower():
                self.sec_updates.append(update)
            else:
                self.bug_updates.append(update)

    # -------------------------------------------------------------------------
    # Report Building and Submission
    # -------------------------------------------------------------------------

    def build_report(self) -> dict[str, Any]:
        """Build the JSON report payload."""
        report = {
            'protocol': self.protocol,
            'hostname': self.hostname,
            'arch': self.arch,
            'kernel': self.kernel,
            'os': self.os_string,
            'tags': self.tags,
            'reboot_required': self.reboot_required,
            'packages': self.packages,
            'repos': self.repos,
            'modules': self.modules,
            'sec_updates': self.sec_updates,
            'bug_updates': self.bug_updates,
        }

        # Include compliance data if available
        compliance = self._load_compliance_data()
        if compliance is not None:
            report['compliance'] = compliance

        return report

    def _load_compliance_data(self):
        """Load compliance JSON from disk if it exists."""
        compliance_file = '/var/lib/patchman/compliance/latest.json'
        if not os.path.isfile(compliance_file):
            return None
        try:
            with open(compliance_file) as f:
                data = json.load(f)
            self.log(f'Loaded compliance data from {compliance_file}')
            return data
        except (json.JSONDecodeError, OSError) as e:
            self.log(f'Warning: could not load compliance data: {e}',
                     level='error')
            return None

    def submit_report(self) -> bool:
        """Submit the report to the Patchman server."""
        if not self.server:
            self.log('Error: No server configured', level='error')
            return False

        report = self.build_report()
        url = f'{self.server}/api/report/'

        self.log(f'Submitting report to {url}')
        self.log(f'Report size: {len(json.dumps(report))} bytes', level='debug')

        if self.debug:
            self.log(f'Report JSON:\n{json.dumps(report, indent=2)}', level='debug')

        headers = {
            'Content-Type': 'application/json',
        }
        if self.api_key:
            headers['Authorization'] = f'Api-Key {self.api_key}'

        # Prepare client certificate if configured
        cert = None
        if self.client_cert and os.path.exists(self.client_cert):
            if self.client_key and self.client_key != self.client_cert:
                cert = (self.client_cert, self.client_key)
            else:
                cert = self.client_cert
            self.log(f'Using client certificate: {self.client_cert}', level='debug')

        try:
            if HAS_REQUESTS:
                response = requests.post(
                    url,
                    json=report,
                    headers=headers,
                    verify=self.verify_ssl,
                    timeout=self.timeout,
                    cert=cert,
                )
                status_code = response.status_code
                response_text = response.text
            else:
                # Fallback to urllib
                data = json.dumps(report).encode('utf-8')
                req = urllib.request.Request(url, data=data, headers=headers)

                ctx = ssl.create_default_context()
                if not self.verify_ssl:
                    ctx.check_hostname = False
                    ctx.verify_mode = ssl.CERT_NONE
                if cert:
                    if isinstance(cert, tuple):
                        ctx.load_cert_chain(cert[0], cert[1])
                    else:
                        ctx.load_cert_chain(cert)

                try:
                    with urllib.request.urlopen(req, timeout=self.timeout, context=ctx) as resp:
                        status_code = resp.status
                        response_text = resp.read().decode('utf-8')
                except urllib.error.HTTPError as e:
                    status_code = e.code
                    response_text = e.read().decode('utf-8')

            if status_code in (200, 201, 202):
                try:
                    result = json.loads(response_text)
                    report_id = result.get('report_id', 'unknown')
                    self.log(f'Report accepted (ID: {report_id})')
                except json.JSONDecodeError:
                    self.log('Report accepted')
                return True
            else:
                self.log(f'Server returned status {status_code}: {response_text}', level='error')
                return False

        except Exception as e:
            self.log(f'Error submitting report: {e}', level='error')
            return False

    # -------------------------------------------------------------------------
    # Certificate Management
    # -------------------------------------------------------------------------

    def _build_cert_tuple(self) -> 'str | tuple[str, str] | None':
        """Build cert parameter for requests/urllib from configured paths."""
        if not self.client_cert or not os.path.exists(self.client_cert):
            return None
        if self.client_key and self.client_key != self.client_cert:
            return (self.client_cert, self.client_key)
        return self.client_cert

    def _post_json(self, url: str, payload: dict, cert: 'str | tuple[str, str] | None' = None) -> tuple[int, str]:
        """POST JSON to a URL, returning (status_code, response_text)."""
        headers = {'Content-Type': 'application/json'}

        if HAS_REQUESTS:
            response = requests.post(
                url,
                json=payload,
                headers=headers,
                cert=cert,
                verify=self.verify_ssl,
                timeout=self.timeout,
            )
            return response.status_code, response.text
        else:
            data = json.dumps(payload).encode('utf-8')
            req = urllib.request.Request(url, data=data, headers=headers)
            ctx = ssl.create_default_context()
            if not self.verify_ssl:
                ctx.check_hostname = False
                ctx.verify_mode = ssl.CERT_NONE
            if cert:
                if isinstance(cert, tuple):
                    ctx.load_cert_chain(cert[0], cert[1])
                else:
                    ctx.load_cert_chain(cert)
            try:
                with urllib.request.urlopen(req, timeout=self.timeout, context=ctx) as resp:
                    return resp.status, resp.read().decode('utf-8')
            except urllib.error.HTTPError as e:
                return e.code, e.read().decode('utf-8')

    def check_cert_expiry(self) -> tuple[bool, int]:
        """Check if client certificate needs renewal.

        Returns:
            Tuple of (needs_renewal, days_until_expiry)
        """
        if not self.client_cert or not os.path.exists(self.client_cert):
            return False, -1

        if not HAS_CRYPTOGRAPHY:
            self.log('cryptography library not available, skipping cert check', level='debug')
            return False, -1

        try:
            with open(self.client_cert, 'rb') as f:
                cert_data = f.read()
            cert = x509.load_pem_x509_certificate(cert_data)

            from datetime import datetime, timezone
            now = datetime.now(timezone.utc)
            expires = (
                cert.not_valid_after_utc
                if hasattr(cert, 'not_valid_after_utc')
                else cert.not_valid_after.replace(tzinfo=timezone.utc)
            )
            days_until_expiry = (expires - now).days

            self.log(f'Certificate expires in {days_until_expiry} days')

            if days_until_expiry <= self.cert_renew_threshold:
                return True, days_until_expiry

            return False, days_until_expiry

        except Exception as e:
            self.log(f'Error checking certificate expiry: {e}', level='error')
            return False, -1

    def renew_certificate(self) -> bool:
        """Renew the client certificate using the existing cert for authentication."""
        if not HAS_CRYPTOGRAPHY:
            self.log('cryptography library required for certificate renewal', level='error')
            return False

        if not self.client_cert:
            self.log('No client certificate configured', level='error')
            return False

        self.log('Renewing client certificate...')

        try:
            key = rsa.generate_private_key(public_exponent=65537, key_size=2048)

            csr = x509.CertificateSigningRequestBuilder().subject_name(
                x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, self.hostname)])
            ).sign(key, hashes.SHA256())

            csr_pem = csr.public_bytes(serialization.Encoding.PEM).decode()

            url = f'{self.server}/api/cert/renew/'
            cert = self._build_cert_tuple()

            status_code, response_text = self._post_json(
                url, {'csr': csr_pem}, cert=cert,
            )

            if status_code != 200:
                self.log(f'Certificate renewal failed: {response_text}', level='error')
                return False

            result = json.loads(response_text)

            with open(self.client_cert, 'w') as f:
                f.write(result['certificate'])
            os.chmod(self.client_cert, 0o644)

            key_path = self.client_key or self.client_cert
            key_pem = key.private_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PrivateFormat.TraditionalOpenSSL,
                encryption_algorithm=serialization.NoEncryption(),
            )

            if key_path == self.client_cert:
                with open(key_path, 'a') as f:
                    f.write(key_pem.decode())
            else:
                with open(key_path, 'wb') as f:
                    f.write(key_pem)
                os.chmod(key_path, 0o600)

            self.log(f'Certificate renewed, expires in {result["expires_in_days"]} days')
            return True

        except Exception as e:
            self.log(f'Certificate renewal error: {e}', level='error')
            return False

    def enroll(self, token: str, cert_path: str = '', key_path: str = '') -> bool:
        """Enroll for a new client certificate using an enrollment token.

        Args:
            token: Enrollment token from server
            cert_path: Where to save the certificate
            key_path: Where to save the key
        """
        if not HAS_CRYPTOGRAPHY:
            self.log('Enrollment requires cryptography library', level='error')
            return False

        cert_path = cert_path or DEFAULT_CERT_PATH
        key_path = key_path or DEFAULT_KEY_PATH

        self.log(f'Enrolling {self.hostname} for client certificate...')

        try:
            key = rsa.generate_private_key(public_exponent=65537, key_size=2048)

            csr = x509.CertificateSigningRequestBuilder().subject_name(
                x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, self.hostname)])
            ).sign(key, hashes.SHA256())

            csr_pem = csr.public_bytes(serialization.Encoding.PEM).decode()

            url = f'{self.server}/api/cert/enroll/'
            payload = {
                'token': token,
                'csr': csr_pem,
                'hostname': self.hostname,
            }

            status_code, response_text = self._post_json(url, payload)

            if status_code != 200:
                self.log(f'Enrollment failed: {response_text}', level='error')
                return False

            result = json.loads(response_text)

            cert_dir = os.path.dirname(cert_path)
            if cert_dir and not os.path.exists(cert_dir):
                os.makedirs(cert_dir, mode=0o755)

            with open(cert_path, 'w') as f:
                f.write(result['certificate'])
            os.chmod(cert_path, 0o644)

            with open(key_path, 'wb') as f:
                f.write(key.private_bytes(
                    encoding=serialization.Encoding.PEM,
                    format=serialization.PrivateFormat.TraditionalOpenSSL,
                    encryption_algorithm=serialization.NoEncryption(),
                ))
            os.chmod(key_path, 0o600)

            self.log('Enrolled successfully!')
            self.log(f'  Certificate: {cert_path}')
            self.log(f'  Key: {key_path}')
            self.log(f'  Expires in: {result["expires_in_days"]} days')
            return True

        except Exception as e:
            self.log(f'Enrollment error: {e}', level='error')
            return False

    # -------------------------------------------------------------------------
    # Main Entry Point
    # -------------------------------------------------------------------------

    def print_summary(self) -> None:
        """Print a dry-run summary of collected data."""
        report = self.build_report()
        print()
        print('=== Dry Run Summary ===')
        print(f'Hostname: {self.hostname}')
        print(f'OS: {self.os_string}')
        print(f'Arch: {self.arch}')
        print(f'Kernel: {self.kernel}')
        print(f'Packages: {len(self.packages)}')
        print(f'Repos: {len(self.repos)}')
        print(f'Modules: {len(self.modules)}')
        print(f'Security updates: {len(self.sec_updates)}')
        print(f'Bugfix updates: {len(self.bug_updates)}')
        print(f'Reboot required: {report["reboot_required"]}')
        if self.tags:
            print(f'Tags: {",".join(self.tags)}')

        if self.verbose:
            if self.packages:
                print('\n=== Packages ===')
                for p in sorted(self.packages, key=lambda x: x['name']):
                    print(f'  {p["name"]}-{p["version"]} ({p["type"]})')
            if self.repos:
                print('\n=== Repos ===')
                for r in self.repos:
                    print(f'  [{r["type"]}] {r["name"]} {" ".join(r.get("urls", []))}')
            if self.sec_updates:
                print('\n=== Security Updates ===')
                for u in self.sec_updates:
                    print(f'  {u["name"]}-{u["version"]}')
            if self.bug_updates:
                print('\n=== Bugfix Updates ===')
                for u in self.bug_updates:
                    print(f'  {u["name"]}-{u["version"]}')
            if self.modules:
                print('\n=== Modules ===')
                for m in self.modules:
                    print(f'  {m["name"]}:{m["stream"]}')

        if self.debug:
            import json
            print('\n=== Full JSON Report ===')
            print(json.dumps(report, indent=2))

    def run(self) -> bool:
        """Run the client: collect data and submit report."""
        self.log(f'Patchman Python Client v{__version__}')
        self.log(f'Hostname: {self.hostname}')

        # Check certificate status if configured
        if self.client_cert:
            if not os.path.exists(self.client_cert):
                self.log(f'Client certificate not found: {self.client_cert}', level='error')
                self.log('Use --enroll to obtain a certificate first', level='error')
                return False

            needs_renewal, days_left = self.check_cert_expiry()
            if needs_renewal:
                self.log(f'Certificate expires in {days_left} days, renewing...')
                if not self.renew_certificate():
                    self.log('Certificate renewal failed, aborting', level='error')
                    return False

        # Collect system information
        self.collect_host_data()
        self.check_reboot_required()

        # Collect packages
        self.collect_packages()

        # Collect repos (if enabled)
        if self.repo_check:
            self.collect_repos()

        # Collect modules
        self.collect_modules()

        # Collect updates (if local updates enabled)
        self.collect_updates()

        # Dry run: print summary and exit
        if self.dry_run:
            self.print_summary()
            return True

        # Submit report
        return self.submit_report()


def load_config(config_file: str) -> dict[str, Any]:
    """Load configuration from file."""
    config: dict[str, Any] = {
        'server': '',
        'api_key': '',
        'tags': '',
        'verbose': False,
        'debug': False,
        'local_updates': False,
        'repo_check': True,
        'verify_ssl': True,
        'timeout': 300,
        'client_cert': '',
        'client_key': '',
        'cert_renew_threshold': 30,
    }

    if not os.path.exists(config_file):
        return config

    load_config_file(config, config_file)

    # Load conf.d drop-in files (like bash client)
    conf_dir = os.path.join(os.path.dirname(config_file), 'conf.d')
    if os.path.isdir(conf_dir):
        for f in sorted(os.listdir(conf_dir)):
            path = os.path.join(conf_dir, f)
            if os.path.isfile(path):
                load_config_file(config, path)

    return config


def load_config_file(config: dict[str, Any], path: str) -> None:
    """Parse a single shell-style config file into config dict."""
    try:
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                if '=' in line:
                    key, value = line.split('=', 1)
                    key = key.strip()
                    value = value.strip().strip('"').strip("'")

                    if key == 'server':
                        config['server'] = value
                    elif key == 'api_key':
                        config['api_key'] = value
                    elif key == 'tags':
                        config['tags'] = value
                    elif key == 'verbose':
                        config['verbose'] = value.lower() in ('true', '1', 'yes')
                    elif key == 'debug':
                        config['debug'] = value.lower() in ('true', '1', 'yes')
                    elif key == 'local_updates':
                        config['local_updates'] = value.lower() in ('true', '1', 'yes')
                    elif key == 'repo_check':
                        config['repo_check'] = value.lower() in ('true', '1', 'yes')
                    elif key == 'verify_ssl':
                        config['verify_ssl'] = value.lower() in ('true', '1', 'yes')
                    elif key == 'timeout':
                        try:
                            config['timeout'] = int(value)
                        except ValueError:
                            pass
                    elif key == 'client_cert':
                        config['client_cert'] = value
                    elif key == 'client_key':
                        config['client_key'] = value
                    elif key == 'cert_renew_threshold':
                        try:
                            config['cert_renew_threshold'] = int(value)
                        except ValueError:
                            pass
    except OSError:
        pass


def main() -> int:
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description='Patchman Python Client - REST API',
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument('-v', '--verbose', action='store_true',
                        help='Verbose output')
    parser.add_argument('-d', '--debug', action='store_true',
                        help='Debug output (implies verbose)')
    parser.add_argument('-n', '--no-repo-check', action='store_true',
                        help='Skip repository collection')
    parser.add_argument('-u', '--local-updates', action='store_true',
                        help='Find updates locally (for RHEL/SUSE)')
    parser.add_argument('-y', '--dry-run', action='store_true',
                        help='Collect data but do not submit')
    parser.add_argument('-s', '--server', metavar='SERVER',
                        help='Patchman server URL')
    parser.add_argument('-c', '--config', metavar='FILE',
                        default=DEFAULT_CONFIG,
                        help=f'Config file (default: {DEFAULT_CONFIG})')
    parser.add_argument('-t', '--tags', metavar='TAGS',
                        help='Comma-separated list of tags')
    parser.add_argument('-H', '--hostname', metavar='HOSTNAME',
                        help='Override hostname')
    parser.add_argument('-k', '--api-key', metavar='KEY',
                        help='API key for authentication')
    parser.add_argument('-p', '--protocol', metavar='PROTOCOL', type=int,
                        default=2, help='Protocol version (default: 2)')
    parser.add_argument('--no-verify-ssl', action='store_true',
                        help='Disable SSL certificate verification')
    parser.add_argument('--client-cert', metavar='FILE',
                        help='Client certificate file for mTLS')
    parser.add_argument('--client-key', metavar='FILE',
                        help='Client key file (if separate from cert)')
    parser.add_argument('--enroll', action='store_true',
                        help='Enroll for a new client certificate')
    parser.add_argument('--token', metavar='TOKEN',
                        help='Enrollment token (required with --enroll)')
    parser.add_argument('--no-renew', action='store_true',
                        help='Skip automatic certificate renewal')
    parser.add_argument('--regenerate-compliance', action='store_true',
                        help='Run patchman-openscap before submitting report')
    parser.add_argument('--version', action='version',
                        version=f'patchman-client.py {__version__}')

    args = parser.parse_args()

    # Load config file
    config = load_config(args.config)

    # Command line overrides config file
    server = args.server or config['server']
    api_key = args.api_key or config['api_key']
    tags = args.tags or config['tags']
    verbose = args.verbose or args.debug or config['verbose']
    debug = args.debug or config['debug']
    local_updates = args.local_updates or config['local_updates']
    repo_check = not args.no_repo_check and config['repo_check']
    verify_ssl = not args.no_verify_ssl and config['verify_ssl']
    timeout = config['timeout']
    client_cert = args.client_cert or config['client_cert']
    client_key = args.client_key or config['client_key']
    cert_renew_threshold = config['cert_renew_threshold']

    # Validate protocol
    supported_protocols = [2]
    if args.protocol not in supported_protocols:
        print(f'Error: unsupported protocol {args.protocol}. '
              f'Supported: {", ".join(str(p) for p in supported_protocols)}',
              file=sys.stderr)
        return 1

    if not args.dry_run:
        if not server:
            print('patchman-client: server not configured.', file=sys.stderr)
            print(f'  Edit {args.config} and set server= to your patchman server URL.',
                  file=sys.stderr)
            return 1

        if 'patchman.example.com' in server:
            print('patchman-client: server not configured.', file=sys.stderr)
            print(f'  Edit {args.config} and set server= to your patchman server URL.',
                  file=sys.stderr)
            return 1

    # Acquire exclusive lock to prevent concurrent runs
    lock_fd = None
    try:
        os.makedirs(LOCK_DIR, exist_ok=True)
        lock_fd = open(LOCK_FILE, 'w')
        fcntl.flock(lock_fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except OSError:
        if lock_fd:
            lock_fd.close()
        print('Error: another patchman-client instance is running.', file=sys.stderr)
        return 1

    def _release_lock():
        try:
            fcntl.flock(lock_fd, fcntl.LOCK_UN)
            lock_fd.close()
            os.unlink(LOCK_FILE)
        except OSError:
            pass

    atexit.register(_release_lock)

    # Create client
    client = PatchmanClient(
        server=server,
        api_key=api_key,
        hostname=args.hostname or '',
        tags=tags,
        verbose=verbose,
        debug=debug,
        local_updates=local_updates,
        repo_check=repo_check,
        verify_ssl=verify_ssl,
        timeout=timeout,
        protocol=args.protocol,
        dry_run=args.dry_run,
        client_cert=client_cert,
        client_key=client_key,
        cert_renew_threshold=cert_renew_threshold if not args.no_renew else 0,
    )

    # Handle enrollment
    if args.enroll:
        if not args.token:
            print('Error: --token is required with --enroll', file=sys.stderr)
            return 1
        success = client.enroll(
            token=args.token,
            cert_path=client_cert or DEFAULT_CERT_PATH,
            key_path=client_key or DEFAULT_KEY_PATH,
        )
        return 0 if success else 1

    # Regenerate compliance data if requested
    if args.regenerate_compliance:
        client.log('Running patchman-openscap to regenerate compliance data...')
        try:
            result = subprocess.run(
                ['patchman-openscap', '--force'],
                capture_output=True, text=True, timeout=3600,
            )
            if result.returncode != 0:
                client.log(f'Warning: patchman-openscap failed (exit {result.returncode})',
                           level='error')
                if result.stderr:
                    client.log(result.stderr.strip(), level='error')
        except FileNotFoundError:
            client.log('Warning: patchman-openscap not found, skipping compliance scan',
                       level='error')
        except subprocess.TimeoutExpired:
            client.log('Warning: patchman-openscap timed out', level='error')

    # Normal operation: collect and submit report
    success = client.run()
    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())
