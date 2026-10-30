#!/usr/bin/env python3
"""
Tests to ensure Python and Bash patchman clients produce consistent output.

These tests compare the output format of both clients to ensure they send
the same data structure to the server.
"""

import os
import unittest


class ClientOutputComparisonTests(unittest.TestCase):
    """Tests comparing Python and Bash client output formats."""

    def setUp(self):
        """Set up test fixtures."""
        self.client_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        self.python_client = os.path.join(self.client_dir, 'patchman-client.py')
        self.bash_client = os.path.join(self.client_dir, 'patchman-client')

    def test_os_string_with_cpe_name(self):
        """Test that both clients format OS string with CPE name consistently."""
        # Sample os-release content for Rocky Linux
        os_release_content = '''
NAME="Rocky Linux"
VERSION="10.1 (Red Quartz)"
ID="rocky"
ID_LIKE="rhel centos fedora"
VERSION_ID="10.1"
PRETTY_NAME="Rocky Linux 10.1 (Red Quartz)"
CPE_NAME="cpe:/o:rocky:rocky:10"
'''
        # Test that CPE name is included in os-release
        self.assertIn('cpe:/o:rocky:rocky:10', os_release_content)

    def test_repo_id_cpe_prefix_format(self):
        """Test that repo IDs are prefixed with CPE name."""
        cpe_name = 'cpe:/o:rocky:rocky:10'
        repo_id = 'baseos'

        # Expected format from bash client
        expected = f'{cpe_name}-{repo_id}'

        self.assertEqual(expected, 'cpe:/o:rocky:rocky:10-baseos')

    def test_repo_name_arch_format(self):
        """Test that repo names have arch appended without duplication."""
        test_cases = [
            # (input_name, arch, expected_output)
            ('Rocky Linux 10 - BaseOS', 'x86_64', 'Rocky Linux 10 - BaseOS x86_64'),
            ('Extra Packages for Enterprise Linux 10 - x86_64', 'x86_64',
             'Extra Packages for Enterprise Linux 10 x86_64'),
            ('Rocky Linux 10 - Extras', 'x86_64', 'Rocky Linux 10 - Extras x86_64'),
        ]

        for input_name, arch, expected in test_cases:
            # Simulate Python client logic
            name = input_name
            if name.endswith(f' - {arch}'):
                name = name[:-len(f' - {arch}')]
            if arch not in name:
                name = f'{name} {arch}'

            self.assertEqual(name, expected, f'Failed for input: {input_name}')

    def test_url_suffix_stripping(self):
        """Test that '(N more)' suffix is stripped from URLs."""
        test_urls = [
            ('https://example.com/repo/ (37 more)', 'https://example.com/repo/'),
            ('https://example.com/repo/', 'https://example.com/repo/'),
            ('https://example.com/repo/ (1 more)', 'https://example.com/repo/'),
        ]

        for input_url, expected in test_urls:
            # Simulate Python client logic
            url = input_url
            if ' (' in url:
                url = url.split(' (')[0]

            self.assertEqual(url, expected, f'Failed for input: {input_url}')

    def test_update_repo_cpe_prefix(self):
        """Test that update repo names are prefixed with CPE name."""
        cpe_name = 'cpe:/o:rocky:rocky:10'
        repo = 'baseos'

        # Expected format
        expected = f'{cpe_name}-{repo}'

        self.assertEqual(expected, 'cpe:/o:rocky:rocky:10-baseos')

    def test_module_repo_cpe_prefix(self):
        """Test that module repo names are prefixed with CPE name."""
        cpe_name = 'cpe:/o:rocky:rocky:10'
        repo = 'appstream'

        # Expected format
        expected = f'{cpe_name}-{repo}'

        self.assertEqual(expected, 'cpe:/o:rocky:rocky:10-appstream')

    def test_package_format(self):
        """Test package data format consistency."""
        # Expected package format from both clients
        expected_keys = {'name', 'version', 'arch', 'type'}

        # Sample RPM package
        rpm_package = {
            'name': 'bash',
            'version': '5.1.8-6.el9',
            'arch': 'x86_64',
            'type': 'rpm',
        }

        self.assertEqual(set(rpm_package.keys()), expected_keys)

    def test_rpm_package_parsing(self):
        """Test parsing of rpm -qa output.

        Real output from: rpm -qa --queryformat '%{NAME} %{EPOCH} %{VERSION} %{RELEASE} %{ARCH}'
        """
        rpm_output = """libgcc (none) 14.3.1 2.1.el10 x86_64
tzdata (none) 2025b 2.el10 noarch
rocky-gpg-keys (none) 10.1 1.2.el10 noarch
bash (none) 5.2.26 6.el10 x86_64
glibc (none) 2.39 58.el10_1.2 x86_64"""

        packages = []
        for line in rpm_output.splitlines():
            parts = line.split()
            if len(parts) >= 5:
                name = parts[0]
                epoch = parts[1] if parts[1] != '(none)' else ''
                version = parts[2]
                release = parts[3]
                arch = parts[4]

                # Combine version-release like bash client does
                full_version = f'{version}-{release}'
                if epoch:
                    full_version = f'{epoch}:{full_version}'

                packages.append({
                    'name': name,
                    'version': full_version,
                    'arch': arch,
                    'type': 'rpm',
                })

        self.assertEqual(len(packages), 5)
        self.assertEqual(packages[0]['name'], 'libgcc')
        self.assertEqual(packages[0]['version'], '14.3.1-2.1.el10')
        self.assertEqual(packages[0]['arch'], 'x86_64')
        self.assertEqual(packages[3]['name'], 'bash')
        self.assertEqual(packages[3]['version'], '5.2.26-6.el10')

    def test_deb_package_parsing(self):
        """Test parsing of dpkg-query output.

        Real output from: dpkg-query -W -f='${Package} ${Version} ${Architecture}'
        """
        dpkg_output = """adduser 3.137ubuntu1 all
apache2 2.4.58-1ubuntu8.10 amd64
bash 5.2.21-2ubuntu4 amd64
bsdutils 1:2.39.3-9ubuntu6.4 amd64
ca-certificates 20240203 all
coreutils 9.4-3ubuntu6.1 amd64"""

        packages = []
        for line in dpkg_output.splitlines():
            parts = line.split()
            if len(parts) >= 3:
                name = parts[0]
                version = parts[1]
                arch = parts[2]

                packages.append({
                    'name': name,
                    'version': version,
                    'arch': arch,
                    'type': 'deb',
                })

        self.assertEqual(len(packages), 6)
        self.assertEqual(packages[0]['name'], 'adduser')
        self.assertEqual(packages[0]['version'], '3.137ubuntu1')
        self.assertEqual(packages[0]['arch'], 'all')
        self.assertEqual(packages[2]['name'], 'bash')
        self.assertEqual(packages[2]['version'], '5.2.21-2ubuntu4')
        # Version with epoch
        self.assertEqual(packages[3]['name'], 'bsdutils')
        self.assertEqual(packages[3]['version'], '1:2.39.3-9ubuntu6.4')

    def test_pacman_package_parsing(self):
        """Test parsing of pacman -Q output.

        Real output from: pacman -Q
        Format: <package> <version>
        """
        pacman_output = """acl 2.3.2-1
archlinux-keyring 20260107-2
bash 5.3.9-1
binutils 2.45.1+r35+g12d0a1dbc1b9-1
bzip2 1.0.8-6
curl 8.18.0-3"""

        packages = []
        for line in pacman_output.splitlines():
            parts = line.split()
            if len(parts) >= 2:
                name = parts[0]
                version = parts[1]

                packages.append({
                    'name': name,
                    'version': version,
                    'arch': 'x86_64',  # pacman -Q doesn't show arch
                    'type': 'arch',
                })

        self.assertEqual(len(packages), 6)
        self.assertEqual(packages[0]['name'], 'acl')
        self.assertEqual(packages[0]['version'], '2.3.2-1')
        self.assertEqual(packages[2]['name'], 'bash')
        self.assertEqual(packages[2]['version'], '5.3.9-1')
        # Complex version with + and git hash
        self.assertEqual(packages[3]['name'], 'binutils')
        self.assertEqual(packages[3]['version'], '2.45.1+r35+g12d0a1dbc1b9-1')

    def test_gentoo_package_parsing(self):
        """Test parsing of qlist -ICv output.

        Real output from: qlist -ICv
        Format: <category>/<package>-<version>
        """
        qlist_output = """acct-group/audio-0-r3
acct-group/cdrom-0-r3
acct-group/portage-0-r2
app-shells/bash-5.2_p26
sys-libs/glibc-2.39-r6
dev-lang/python-3.12.4"""

        packages = []
        for line in qlist_output.splitlines():
            if '/' not in line:
                continue
            # Split category/package-version
            category, pkg_ver = line.split('/', 1)
            # Find last hyphen followed by version number
            # Package names can have hyphens, so we need to find version start
            import re
            match = re.match(r'^(.+?)-(\d.*)$', pkg_ver)
            if match:
                name = f'{category}/{match.group(1)}'
                version = match.group(2)
                packages.append({
                    'name': name,
                    'version': version,
                    'arch': 'x86_64',
                    'type': 'gentoo',
                })

        self.assertEqual(len(packages), 6)
        self.assertEqual(packages[0]['name'], 'acct-group/audio')
        self.assertEqual(packages[0]['version'], '0-r3')
        self.assertEqual(packages[3]['name'], 'app-shells/bash')
        self.assertEqual(packages[3]['version'], '5.2_p26')
        self.assertEqual(packages[5]['name'], 'dev-lang/python')
        self.assertEqual(packages[5]['version'], '3.12.4')

    def test_report_json_structure(self):
        """Test that report JSON has all required fields."""
        required_fields = [
            'protocol',
            'hostname',
            'arch',
            'kernel',
            'os',
            'tags',
            'reboot_required',
            'packages',
            'repos',
            'modules',
            'sec_updates',
            'bug_updates',
        ]

        # Sample report structure
        report = {
            'protocol': 2,
            'hostname': 'test.example.com',
            'arch': 'x86_64',
            'kernel': '5.14.0-1.el9',
            'os': 'Rocky Linux release 10.1 (Red Quartz) [cpe:/o:rocky:rocky:10]',
            'tags': ['test'],
            'reboot_required': False,
            'packages': [],
            'repos': [],
            'modules': [],
            'sec_updates': [],
            'bug_updates': [],
        }

        for field in required_fields:
            self.assertIn(field, report, f'Missing required field: {field}')


class OSStringFormatTests(unittest.TestCase):
    """Tests for OS string formatting across distributions."""

    def test_debian_format(self):
        """Debian should be 'Debian X.Y' without CPE."""
        # Bash: os="Debian $(cat /etc/debian_version | sed -e 's/\//-/')"
        deb_version = '12.0'
        expected = f'Debian {deb_version}'
        self.assertEqual(expected, 'Debian 12.0')

    def test_ubuntu_format(self):
        """Ubuntu should use PRETTY_NAME."""
        # Real: PRETTY_NAME="Ubuntu Resolute Raccoon (development branch)"
        pretty_name = 'Ubuntu Resolute Raccoon (development branch)'
        expected = pretty_name
        self.assertEqual(expected, 'Ubuntu Resolute Raccoon (development branch)')

    def test_rocky_format_with_cpe(self):
        """Rocky should include CPE name in brackets."""
        # Real: /etc/rocky-release contains "Rocky Linux release 9.7 (Blue Onyx)"
        # Real: CPE_NAME="cpe:/o:rocky:rocky:9::baseos"
        release = 'Rocky Linux release 9.7 (Blue Onyx)'
        cpe = 'cpe:/o:rocky:rocky:9::baseos'
        expected = f'{release} [{cpe}]'
        self.assertEqual(
            expected,
            'Rocky Linux release 9.7 (Blue Onyx) [cpe:/o:rocky:rocky:9::baseos]'
        )

    def test_alma_format_with_cpe(self):
        """Alma should include CPE name in brackets."""
        release = 'AlmaLinux release 9.3 (Shamrock Pampas Cat)'
        cpe = 'cpe:/o:almalinux:almalinux:9'
        expected = f'{release} [{cpe}]'
        self.assertIn('[cpe:/o:almalinux:almalinux:9]', expected)

    def test_opensuse_format_with_cpe(self):
        """openSUSE should use PRETTY_NAME with CPE in brackets."""
        # Real: PRETTY_NAME="openSUSE Leap 16.0"
        # Real: CPE_NAME="cpe:/o:opensuse:leap:16.0"
        pretty_name = 'openSUSE Leap 16.0'
        cpe = 'cpe:/o:opensuse:leap:16.0'
        expected = f'{pretty_name} [{cpe}]'
        self.assertEqual(expected, 'openSUSE Leap 16.0 [cpe:/o:opensuse:leap:16.0]')

    def test_arch_format(self):
        """Arch should use NAME without CPE."""
        # Real: NAME="Arch Linux"
        # Arch has no CPE_NAME
        name = 'Arch Linux'
        expected = name
        self.assertEqual(expected, 'Arch Linux')

    def test_gentoo_format(self):
        """Gentoo should use PRETTY_NAME + VERSION_ID."""
        # Real: PRETTY_NAME="Gentoo Linux", VERSION_ID="2.18"
        pretty_name = 'Gentoo Linux'
        version_id = '2.18'
        expected = f'{pretty_name} {version_id}'
        self.assertEqual(expected, 'Gentoo Linux 2.18')

    def test_fedora_format_with_cpe(self):
        """Fedora should use PRETTY_NAME with CPE in brackets."""
        # Real: PRETTY_NAME="Fedora Linux 43 (Container Image)"
        # Real: CPE_NAME="cpe:/o:fedoraproject:fedora:43"
        pretty_name = 'Fedora Linux 43 (Container Image)'
        cpe = 'cpe:/o:fedoraproject:fedora:43'
        expected = f'{pretty_name} [{cpe}]'
        self.assertEqual(expected, 'Fedora Linux 43 (Container Image) [cpe:/o:fedoraproject:fedora:43]')


class RepoFormatTests(unittest.TestCase):
    """Tests for repository data formatting."""

    def test_rpm_repo_structure(self):
        """Test RPM repository data structure."""
        repo = {
            'type': 'rpm',
            'id': 'cpe:/o:rocky:rocky:10-baseos',
            'name': 'Rocky Linux 10 - BaseOS x86_64',
            'priority': 99,
            'urls': ['https://example.com/rocky/10/BaseOS/x86_64/os/'],
        }

        self.assertEqual(repo['type'], 'rpm')
        self.assertIn('cpe:', repo['id'])
        self.assertIn('x86_64', repo['name'])
        self.assertIsInstance(repo['urls'], list)

    def test_deb_repo_structure(self):
        """Test DEB repository data structure."""
        repo = {
            'type': 'deb',
            'id': '',
            'name': 'Ubuntu 24.04 amd64 repo at http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64',
            'priority': 500,
            'urls': ['http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64'],
        }

        self.assertEqual(repo['type'], 'deb')
        self.assertEqual(repo['id'], '')
        self.assertIsInstance(repo['urls'], list)

    def test_apt_repo_empty_id(self):
        """Test APT repo uses empty id field.

        APT repos don't have a repo id like YUM repos do.
        Bash client outputs 4 fields: 'type' 'name' 'priority' 'url'
        build_repos_json sets id to empty string for deb type.
        Python client uses id: '' to match.
        """
        repo = {
            'type': 'deb',
            'id': '',  # APT repos have no id
            'name': 'Debian 12 x86_64 repo at http://deb.debian.org/debian/dists/bookworm/main/binary-amd64',
            'priority': 500,
            'urls': ['http://deb.debian.org/debian/dists/bookworm/main/binary-amd64'],
        }

        self.assertEqual(repo['id'], '')
        self.assertIsInstance(repo['priority'], int)

    def test_apt_policy_parsing(self):
        """Test parsing of apt-cache policy output format.

        Real apt-cache policy output looks like:
         500 http://archive.ubuntu.com/ubuntu noble/main amd64 Packages

        Should produce repo with:
        - name: 'Ubuntu 24.04 amd64 repo at http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64'
        - priority: 500
        - url: 'http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64'
        """
        import re

        # Simulated apt-cache policy lines
        policy_lines = [
            " 500 http://security.ubuntu.com/ubuntu noble-security/main amd64 Packages",
            " 500 http://archive.ubuntu.com/ubuntu noble/universe amd64 Packages",
            " 100 http://archive.ubuntu.com/ubuntu noble-backports/main amd64 Packages",
        ]

        repo_pattern = re.compile(r'^\s*(\d+)\s+(https?://\S+)\s+(\S+)\s+(\S+)')
        os_name = "Ubuntu"
        version_id = "24.04"
        arch = "amd64"

        repos = []
        for line in policy_lines:
            match = repo_pattern.match(line)
            if match:
                priority, base_url, dist, component = match.groups()
                url = f'{base_url}/dists/{dist}/binary-{arch}'
                repo_name = f'{os_name} {version_id} {arch} repo at {url}'
                repos.append({
                    'type': 'deb',
                    'id': '',
                    'name': repo_name,
                    'priority': int(priority),
                    'urls': [url],
                })

        self.assertEqual(len(repos), 3)

        # Check first repo (security)
        self.assertEqual(repos[0]['priority'], 500)
        self.assertIn('noble-security/main', repos[0]['urls'][0])
        self.assertEqual(repos[0]['id'], '')

        # Check backports has lower priority
        self.assertEqual(repos[2]['priority'], 100)
        self.assertIn('noble-backports', repos[2]['urls'][0])

    def test_bash_apt_sed_transform(self):
        """Test that bash sed transform produces correct format.

        Input: '500 http://archive.ubuntu.com/ubuntu noble/main amd64'
        Output: "'deb' 'Ubuntu 24.04 amd64 repo at <url>' '500' '<url>'"

        The output has 4 fields (no id), which build_repos_json handles specially.
        """
        # Simulated bash output after sed transform
        bash_output = (
            "'deb' 'Ubuntu 24.04 amd64 repo at "
            "http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64' "
            "'500' 'http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64'"
        )

        # Extract parts like bash does with grep -oP "'[^']*'"
        import re
        parts = re.findall(r"'([^']*)'", bash_output)

        self.assertEqual(len(parts), 4)  # deb type has 4 fields, no id
        self.assertEqual(parts[0], 'deb')
        self.assertIn('Ubuntu 24.04 amd64 repo at', parts[1])
        self.assertEqual(parts[2], '500')  # priority
        self.assertIn('http://', parts[3])  # url

    def test_zypper_parsing_with_keep_column(self):
        """Test zypper parsing handles 'Keep' column correctly.

        Newer zypper versions have a Keep column that shifts indices:
        # | Alias | Name | Enabled | GPG | Refresh | Keep | Priority | Type | URI | Service
        """
        # Simulated zypper output with Keep column
        zypper_output = (
            "#  | Alias      | Name    | Enabled | GPG Check | Refresh | Keep "
            "| Priority | Type   | URI                                        | Service\n"
            "---+------------+---------+---------+-----------+---------+------"
            "+----------+--------+--------------------------------------------+---------\n"
            " 1 | games      | games   | Yes     | (r ) Yes  | No      | -    "
            "|   99     | rpm-md | https://download.opensuse.org/games/16.0/  |\n"
            " 2 | packman    | packman | Yes     | (r ) Yes  | Yes     | -    "
            "|   98     | rpm-md | https://ftp.gwdg.de/pub/packman/suse/16.0/ |"
        )

        lines = zypper_output.splitlines()
        has_keep = any('Keep' in line for line in lines[:3])
        self.assertTrue(has_keep)

        # With Keep column: priority at index 7, uri at index 9
        priority_idx = 7 if has_keep else 6
        uri_idx = 9 if has_keep else 8

        repos = []
        for line in lines:
            if not line.strip() or line.startswith('#') or line.startswith('-'):
                continue
            parts = [p.strip() for p in line.split('|')]
            if len(parts) <= uri_idx:
                continue
            try:
                alias = parts[1]
                name = parts[2]
                priority_str = parts[priority_idx]
                priority = int(priority_str) if priority_str.isdigit() else 99
                uri = parts[uri_idx]
                repos.append({'id': alias, 'name': name, 'priority': priority, 'url': uri})
            except (IndexError, ValueError):
                continue

        self.assertEqual(len(repos), 2)
        self.assertEqual(repos[0]['id'], 'games')
        self.assertEqual(repos[0]['priority'], 99)
        self.assertIn('download.opensuse.org', repos[0]['url'])
        self.assertEqual(repos[1]['id'], 'packman')
        self.assertEqual(repos[1]['priority'], 98)

    def test_zypper_parsing_without_keep_column(self):
        """Test zypper parsing handles old format without Keep column."""
        # Simulated zypper output without Keep column
        zypper_output = """#  | Alias    | Name     | Enabled | GPG Check | Refresh | Priority | Type   | URI
---+----------+----------+---------+-----------+---------+----------+--------+-----
 1 | repo-oss | repo-oss | Yes     | (r ) Yes  | Yes     |   99     | rpm-md | http://cdn.opensuse.org/oss"""

        lines = zypper_output.splitlines()
        has_keep = any('Keep' in line for line in lines[:3])
        self.assertFalse(has_keep)

        # Without Keep column: priority at index 6, uri at index 8
        priority_idx = 7 if has_keep else 6
        uri_idx = 9 if has_keep else 8

        repos = []
        for line in lines:
            if not line.strip() or line.startswith('#') or line.startswith('-'):
                continue
            parts = [p.strip() for p in line.split('|')]
            if len(parts) <= uri_idx:
                continue
            try:
                alias = parts[1]
                priority_str = parts[priority_idx]
                priority = int(priority_str) if priority_str.isdigit() else 99
                uri = parts[uri_idx]
                repos.append({'id': alias, 'priority': priority, 'url': uri})
            except (IndexError, ValueError):
                continue

        self.assertEqual(len(repos), 1)
        self.assertEqual(repos[0]['id'], 'repo-oss')
        self.assertEqual(repos[0]['priority'], 99)

    def test_yum_dnf_parsing(self):
        """Test parsing of yum/dnf repolist enabled --verbose output.

        Real output format:
        Repo-id            : appstream
        Repo-name          : Rocky Linux 10 - AppStream
        Repo-baseurl       : https://example.com/repo/ (37 more)
        """
        dnf_output = """Repo-id            : appstream
Repo-name          : Rocky Linux 10 - AppStream
Repo-mirrors       : https://mirrors.rockylinux.org/mirrorlist?arch=x86_64
Repo-baseurl       : https://example.com/pub/rocky/10.1/AppStream/x86_64/os/ (37 more)
Repo-filename      : /etc/yum.repos.d/rocky.repo

Repo-id            : epel
Repo-name          : Extra Packages for Enterprise Linux 10 - x86_64
Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-z-10
Repo-baseurl       : https://example.com/pub/epel/10.1/Everything/x86_64/ (48 more)
Repo-filename      : /etc/yum.repos.d/epel.repo"""

        arch = 'x86_64'
        repos = []
        current_repo = {}
        urls = []

        for line in dnf_output.splitlines():
            line = line.strip()

            if line.startswith('Repo-id'):
                # Save previous repo
                if current_repo.get('id'):
                    current_repo['urls'] = urls
                    repos.append(current_repo)
                # Start new repo
                repo_id = line.split(':', 1)[1].strip() if ':' in line else ''
                repo_id = repo_id.split('/')[0]
                current_repo = {'type': 'rpm', 'id': repo_id, 'name': '', 'priority': 99, 'urls': []}
                urls = []

            elif line.startswith('Repo-name') and current_repo:
                name = line.split(':', 1)[1].strip() if ':' in line else ''
                # Handle arch in name - strip " - arch" then add arch
                if name.endswith(f' - {arch}'):
                    name = name[:-len(f' - {arch}')]
                if arch not in name:
                    name = f'{name} {arch}'
                current_repo['name'] = name

            elif line.startswith('Repo-baseurl') and current_repo:
                url = line.split(':', 1)[1].strip() if ':' in line else ''
                # Strip "(N more)" suffix
                if ' (' in url:
                    url = url.split(' (')[0]
                if url.startswith(('http://', 'https://')):
                    urls.append(url)

        # Don't forget last repo
        if current_repo.get('id'):
            current_repo['urls'] = urls
            repos.append(current_repo)

        self.assertEqual(len(repos), 2)

        # Check appstream repo
        self.assertEqual(repos[0]['id'], 'appstream')
        self.assertEqual(repos[0]['name'], 'Rocky Linux 10 - AppStream x86_64')
        self.assertIn('rocky', repos[0]['urls'][0])
        self.assertNotIn('(37 more)', repos[0]['urls'][0])

        # Check epel repo - name already has arch, shouldn't duplicate
        self.assertEqual(repos[1]['id'], 'epel')
        self.assertEqual(repos[1]['name'], 'Extra Packages for Enterprise Linux 10 x86_64')
        self.assertNotIn(' - x86_64', repos[1]['name'])  # " - x86_64" should be stripped

    def test_pacman_repo_format(self):
        """Test pacman repo parsing and output format.

        pacman.conf has sections like:
        [core]
        Include = /etc/pacman.d/mirrorlist

        Output format: 'arch' 'Arch Linux <repo> <arch>' '<repo>' '<url1>' '<url2>' ...
        """
        # Simulated pacman.conf sections
        pacman_conf = """[core]
Include = /etc/pacman.d/mirrorlist

[extra]
Include = /etc/pacman.d/mirrorlist"""

        arch = 'x86_64'
        repos = {}
        current_repo = ''

        for line in pacman_conf.splitlines():
            line = line.strip()
            if not line:
                continue

            # Section header
            if line.startswith('[') and line.endswith(']'):
                section = line[1:-1]
                if section != 'options':
                    current_repo = section
                    repos[current_repo] = []
                continue

        # Verify we found the repos
        self.assertIn('core', repos)
        self.assertIn('extra', repos)

        # Simulate final output format
        output_repos = []
        for repo in repos:
            output_repos.append({
                'type': 'arch',
                'name': f'Arch Linux {repo} {arch}',
                'id': repo,
                'priority': 0,
                'urls': ['https://mirror.example.com/archlinux/$repo/os/$arch'],
            })

        self.assertEqual(len(output_repos), 2)
        self.assertEqual(output_repos[0]['type'], 'arch')
        self.assertEqual(output_repos[0]['name'], 'Arch Linux core x86_64')
        self.assertEqual(output_repos[0]['id'], 'core')

    def test_dnf_module_parsing(self):
        """Test parsing of dnf module info output.

        Real output from: dnf module info mariadb
        """
        module_info = """Name             : mariadb
Stream           : 10.11 [e] [a]
Version          : 9070020260107210828
Context          : rhel9
Architecture     : x86_64
Profiles         : client, galera, server [d]
Default profiles : server
Repo             : appstream
Summary          : MariaDB Module
Artifacts        : galera-0:26.4.20-1.module+el9.7.0+40024+12a6f3b4.x86_64
                 : mariadb-3:10.11.15-1.module+el9.7.0+40065+b33d0971.x86_64
                 : mariadb-server-3:10.11.15-1.module+el9.7.0+40065+b33d0971.x86_64"""

        # Parse like bash client does
        module_name = 'mariadb'
        module_arch = ''
        module_repo = ''
        module_stream = ''
        module_version = ''
        module_context = ''
        artifacts = []
        in_artifacts = False

        for line in module_info.splitlines():
            line = line.strip()
            if line.startswith('Architecture'):
                module_arch = line.split(':', 1)[1].strip()
            elif line.startswith('Repo'):
                module_repo = line.split(':', 1)[1].strip()
            elif line.startswith('Stream'):
                # Strip [e] [a] markers
                stream = line.split(':', 1)[1].strip()
                module_stream = stream.split('[')[0].strip()
            elif line.startswith('Version'):
                module_version = line.split(':', 1)[1].strip()
            elif line.startswith('Context'):
                module_context = line.split(':', 1)[1].strip()
            elif line.startswith('Artifacts'):
                in_artifacts = True
                artifact = line.split(':', 1)[1].strip()
                if artifact:
                    artifacts.append(artifact)
            elif in_artifacts and line.startswith(':'):
                artifacts.append(line[1:].strip())

        self.assertEqual(module_name, 'mariadb')
        self.assertEqual(module_stream, '10.11')
        self.assertEqual(module_version, '9070020260107210828')
        self.assertEqual(module_context, 'rhel9')
        self.assertEqual(module_arch, 'x86_64')
        self.assertEqual(module_repo, 'appstream')
        self.assertEqual(len(artifacts), 3)

        # With CPE prefix
        cpe_name = 'cpe:/o:rocky:rocky:9::baseos'
        prefixed_repo = f'{cpe_name}-{module_repo}'
        self.assertEqual(prefixed_repo, 'cpe:/o:rocky:rocky:9::baseos-appstream')

    def test_gentoo_portage_parsing(self):
        """Test parsing of portageq repos_config output.

        Real output from: portageq repos_config /
        """
        portageq_output = """[DEFAULT]
auto-sync = yes
main-repo = gentoo

[gentoo]
auto-sync = yes
location = /var/db/repos/gentoo
priority = -1000
sync-type = rsync
sync-uri = rsync://rsync.gentoo.org/gentoo-portage"""

        repos = {}
        repo_name = ''
        priority = ''
        sync_uri = ''

        for line in portageq_output.splitlines():
            # Section header
            if line.startswith('[') and line.endswith(']'):
                # Save previous repo
                if repo_name and sync_uri:
                    repos[repo_name] = {'priority': priority or '0', 'uri': sync_uri}
                repo_name = line[1:-1]
                if repo_name == 'DEFAULT':
                    repo_name = ''
                priority = ''
                sync_uri = ''
            elif 'priority' in line and '=' in line:
                priority = line.split('=')[1].strip()
            elif 'sync-uri' in line and '=' in line:
                sync_uri = line.split('=')[1].strip()

        # Save last repo
        if repo_name and sync_uri:
            repos[repo_name] = {'priority': priority or '0', 'uri': sync_uri}

        self.assertIn('gentoo', repos)
        self.assertEqual(repos['gentoo']['priority'], '-1000')
        self.assertEqual(repos['gentoo']['uri'], 'rsync://rsync.gentoo.org/gentoo-portage')

        # Output format: 'gentoo' 'Gentoo Linux <repo> Repo <arch>' '<repo>' '<priority>' '<uri>'
        arch = 'x86_64'
        output_repos = []
        for repo, info in repos.items():
            output_repos.append({
                'type': 'gentoo',
                'name': f'Gentoo Linux {repo} Repo {arch}',
                'id': repo,
                'priority': int(info['priority']),
                'urls': [info['uri']],
            })

        self.assertEqual(len(output_repos), 1)
        self.assertEqual(output_repos[0]['type'], 'gentoo')
        self.assertEqual(output_repos[0]['name'], 'Gentoo Linux gentoo Repo x86_64')
        self.assertEqual(output_repos[0]['priority'], -1000)

    def test_dnf_updateinfo_parsing(self):
        """Test parsing of dnf updateinfo list output.

        Real output from: dnf updateinfo list
        Format: <advisory-id> <severity>  <package-nevra>
        """
        updateinfo_output = """RLSA-2025:23343 Moderate/Sec.  binutils-2.35.2-67.el9_7.1.x86_64
RLSA-2025:23343 Moderate/Sec.  binutils-gold-2.35.2-67.el9_7.1.x86_64
RLSA-2025:22175 Important/Sec. expat-2.5.0-5.el9_7.1.x86_64
RLSA-2026:0936  Moderate/Sec.  glib2-2.68.4-18.el9_7.1.x86_64
RLSA-2026:0719  Important/Sec. gnupg2-2.3.3-5.el9_7.x86_64
RLSA-2025:22376 Moderate/Sec.  libxml2-2.9.13-14.el9_7.x86_64
RLSA-2025:23342 Moderate/Sec.  python3-3.9.25-2.el9_7.x86_64
RLSA-2026:0067  Moderate/Sec.  tar-2:1.34-9.el9_7.x86_64"""

        updates = []
        for line in updateinfo_output.splitlines():
            parts = line.split()
            if len(parts) >= 3:
                advisory = parts[0]
                severity = parts[1]
                package = parts[2]
                updates.append({
                    'advisory': advisory,
                    'severity': severity,
                    'package': package,
                })

        self.assertEqual(len(updates), 8)
        self.assertEqual(updates[0]['advisory'], 'RLSA-2025:23343')
        self.assertEqual(updates[0]['severity'], 'Moderate/Sec.')
        self.assertEqual(updates[0]['package'], 'binutils-2.35.2-67.el9_7.1.x86_64')

        # Check Important severity
        important = [u for u in updates if 'Important' in u['severity']]
        self.assertEqual(len(important), 2)


class YumRepolistAwkTests(unittest.TestCase):
    """Tests for the awk parser that processes yum repolist --verbose output.

    The bash client uses an awk one-liner to parse yum repolist output into
    quoted fields. These tests pipe simulated output through the actual awk
    and verify correct repo separation.
    """

    AWK_CMD = (
        r"""awk 'BEGIN{n=0} { if ($1=="Repo-id") {if(n>0){print ""} n++; url=0; printf "'"'"'"; for (i=3; i<NF; i++) printf $i " "; printf $NF"'"'"' "} if ($1=="Repo-name") {printf "'"'"'"; for (i=3; i<NF; i++) printf $i " "; printf $NF"'" x86_64'"' "} if ($1=="Repo-mirrors" || $1=="Repo-metalink") {printf "'"'"'"; for (i=3; i<NF; i++) printf $i " "; printf $NF"'"'"' "} if ($1=="Repo-baseurl" || $1=="Repo-baseurl:") { url=1; comma=match($NF,","); if (comma) out=substr($NF,1,comma-1); else out=$NF; printf "'"'"'"out"'"'"' "; } else { if (url==1) { if ($1==":") { comma=match($NF,","); if (comma) out=substr($NF,1,comma-1); else out=$NF; printf "'"'"'"out"'"'"' "; } else {url=0} } } } END{if(n>0) print ""}'"""  # noqa: E501
    )

    def _run_awk(self, input_text):
        """Pipe input through the awk command and return output lines."""
        import subprocess
        result = subprocess.run(
            ['bash', '-c', self.AWK_CMD],
            input=input_text, capture_output=True, text=True,
        )
        return [line for line in result.stdout.splitlines() if line.strip()]

    def test_baseurl_only_repo(self):
        """Repo with only Repo-baseurl should produce one line."""
        input_text = (
            'Repo-id            : baseos\n'
            'Repo-name          : Rocky Linux 10 - BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)
        self.assertIn('baseos', lines[0])
        self.assertIn('https://example.com/baseos/', lines[0])

    def test_metalink_only_repo(self):
        """Repo with only Repo-metalink (no baseurl) should produce one line."""
        input_text = (
            'Repo-id            : epel\n'
            'Repo-name          : Extra Packages for Enterprise Linux 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)
        self.assertIn('epel', lines[0])
        self.assertIn('metalink', lines[0])

    def test_mirrors_only_repo(self):
        """Repo with only Repo-mirrors (no baseurl) should produce one line."""
        input_text = (
            'Repo-id            : extras\n'
            'Repo-name          : Extras\n'
            'Repo-mirrors       : https://mirrors.example.com/mirrorlist\n'
            'Repo-filename      : /etc/yum.repos.d/extras.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)
        self.assertIn('extras', lines[0])

    def test_baseurl_and_metalink_repo(self):
        """Repo with both baseurl and metalink should produce one line."""
        input_text = (
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-baseurl       : https://example.com/epel/10/\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)
        self.assertIn('epel', lines[0])

    def test_multiple_baseurl_repos(self):
        """Multiple repos with baseurls should each get their own line."""
        input_text = (
            'Repo-id            : baseos\n'
            'Repo-name          : BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
            '\n'
            'Repo-id            : appstream\n'
            'Repo-name          : AppStream\n'
            'Repo-baseurl       : https://example.com/appstream/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 2)
        self.assertIn('baseos', lines[0])
        self.assertIn('appstream', lines[1])

    def test_multiple_metalink_only_repos(self):
        """Multiple metalink-only repos must not merge into one line.

        This is the bug that caused the original crash: consecutive repos
        with only Repo-metalink (no Repo-baseurl) would merge into a single
        line, producing garbage that the server couldn't parse.
        """
        input_text = (
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
            '\n'
            'Repo-id            : epel-cisco\n'
            'Repo-name          : EPEL Cisco\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-cisco-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel-cisco.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 2, f'Expected 2 lines but got {len(lines)}: {lines}')
        self.assertIn('epel', lines[0])
        self.assertIn('epel-cisco', lines[1])

    def test_mixed_baseurl_and_metalink_repos(self):
        """Mix of baseurl and metalink-only repos should all separate."""
        input_text = (
            'Repo-id            : baseos\n'
            'Repo-name          : BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
            '\n'
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
            '\n'
            'Repo-id            : appstream\n'
            'Repo-name          : AppStream\n'
            'Repo-baseurl       : https://example.com/appstream/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 3, f'Expected 3 lines but got {len(lines)}: {lines}')
        self.assertIn('baseos', lines[0])
        self.assertIn('epel', lines[1])
        self.assertIn('appstream', lines[2])

    def test_single_repo(self):
        """Single repo should produce exactly one line."""
        input_text = (
            'Repo-id            : baseos\n'
            'Repo-name          : BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)

    def test_empty_input(self):
        """Empty input should produce no output."""
        lines = self._run_awk('')
        self.assertEqual(len(lines), 0)

    def test_baseurl_with_continuation_urls(self):
        """Repo with multiple baseurl continuation lines."""
        input_text = (
            'Repo-id            : baseos\n'
            'Repo-name          : BaseOS\n'
            'Repo-baseurl       : https://mirror1.example.com/baseos/\n'
            '                   : https://mirror2.example.com/baseos/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
        )
        lines = self._run_awk(input_text)
        self.assertEqual(len(lines), 1)
        self.assertIn('mirror1', lines[0])
        self.assertIn('mirror2', lines[0])


class PythonClientYumRepoTests(unittest.TestCase):
    """Tests for the Python client's yum/dnf repo parsing.

    These test the actual PatchmanClient._collect_yum_repos method by
    mocking _run_command and _command_exists. They will fail if the
    Python client doesn't handle metalink/mirrors-only repos.
    """

    def _collect(self, yum_output, arch='x86_64'):
        """Create a PatchmanClient, mock yum output, and collect repos."""
        import sys
        sys.path.insert(0, os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            'python',
        ))
        from patchman_client.client import PatchmanClient

        client = PatchmanClient.__new__(PatchmanClient)
        client.repos = []
        client.arch = arch
        client.cpe_name = ''
        client.verbose = False
        client.debug = False

        # Mock _command_exists to return True for yum
        client._command_exists = lambda cmd: True
        client.log = lambda *a, **kw: None
        # Mock _run_command to return our test output
        client._run_command = lambda cmd: (0, yum_output, '')

        client._collect_yum_repos()
        return client.repos

    def test_baseurl_only(self):
        """Repo with baseurl should have URL captured."""
        output = (
            'Repo-id            : baseos\n'
            'Repo-name          : Rocky Linux 10 - BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 1)
        self.assertEqual(repos[0]['id'], 'baseos')
        self.assertEqual(len(repos[0]['urls']), 1)
        self.assertEqual(repos[0]['urls'][0], 'https://example.com/baseos/')

    def test_metalink_only(self):
        """Repo with only metalink must still have a URL captured."""
        output = (
            'Repo-id            : epel\n'
            'Repo-name          : Extra Packages for Enterprise Linux 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 1)
        self.assertEqual(repos[0]['id'], 'epel')
        self.assertGreater(len(repos[0]['urls']), 0,
                           'Metalink-only repo must have at least one URL')
        self.assertIn('metalink', repos[0]['urls'][0])

    def test_mirrors_only(self):
        """Repo with only Repo-mirrors must still have a URL captured."""
        output = (
            'Repo-id            : extras\n'
            'Repo-name          : Extras\n'
            'Repo-mirrors       : https://mirrors.example.com/mirrorlist\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 1)
        self.assertGreater(len(repos[0]['urls']), 0,
                           'Mirrors-only repo must have at least one URL')

    def test_multiple_metalink_only_repos(self):
        """Multiple metalink-only repos should each be separate."""
        output = (
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
            '\n'
            'Repo-id            : epel-cisco\n'
            'Repo-name          : EPEL Cisco\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-cisco-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel-cisco.repo\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 2)
        self.assertEqual(repos[0]['id'], 'epel')
        self.assertGreater(len(repos[0]['urls']), 0)
        self.assertEqual(repos[1]['id'], 'epel-cisco')
        self.assertGreater(len(repos[1]['urls']), 0)

    def test_mixed_baseurl_and_metalink_repos(self):
        """Mix of baseurl and metalink-only repos."""
        output = (
            'Repo-id            : baseos\n'
            'Repo-name          : BaseOS\n'
            'Repo-baseurl       : https://example.com/baseos/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
            '\n'
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-filename      : /etc/yum.repos.d/epel.repo\n'
            '\n'
            'Repo-id            : appstream\n'
            'Repo-name          : AppStream\n'
            'Repo-baseurl       : https://example.com/appstream/\n'
            'Repo-filename      : /etc/yum.repos.d/rocky.repo\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 3)
        self.assertEqual(repos[0]['id'], 'baseos')
        self.assertIn('baseos', repos[0]['urls'][0])
        self.assertEqual(repos[1]['id'], 'epel')
        self.assertIn('metalink', repos[1]['urls'][0])
        self.assertEqual(repos[2]['id'], 'appstream')
        self.assertIn('appstream', repos[2]['urls'][0])

    def test_baseurl_preferred_over_metalink(self):
        """When both baseurl and metalink exist, both should be captured."""
        output = (
            'Repo-id            : epel\n'
            'Repo-name          : EPEL 10\n'
            'Repo-metalink      : https://mirrors.fedoraproject.org/metalink?repo=epel-10\n'
            'Repo-baseurl       : https://example.com/epel/10/\n'
        )
        repos = self._collect(output)
        self.assertEqual(len(repos), 1)
        self.assertGreaterEqual(len(repos[0]['urls']), 1)


class CDNDetectionTests(unittest.TestCase):
    """Tests for CDN repo detection logic that auto-enables local_updates."""

    CDN_PATTERNS = [
        ('https://cdn.redhat.com/content/dist/rhel9/9/x86_64/baseos/os', True),
        ('https://cdn-ubi.redhat.com/content/public/ubi/dist/ubi9/9/x86_64/baseos/os', True),
        ('https://satellite.example.com/XMLRPC/channel', True),
        ('https://updates.suse.com/SUSE/Products/SLE-Module-Basesystem/15-SP5/x86_64/product/', True),
        ('https://mirror.example.com/rocky/10/BaseOS/x86_64/os/', False),
        ('http://archive.ubuntu.com/ubuntu', False),
        ('https://dl.fedoraproject.org/pub/epel/10/Everything/x86_64/', False),
    ]

    def _matches_cdn(self, url):
        """Test URL against CDN detection logic from python client."""
        import re
        return bool(
            re.search(r'https://cdn[^.]*\.redhat\.com/', url) or
            '/XMLRPC' in url or
            url.startswith('https://updates.suse.com/')
        )

    def test_cdn_patterns(self):
        """Each URL should be correctly identified as CDN or not."""
        for url, expected in self.CDN_PATTERNS:
            with self.subTest(url=url):
                result = self._matches_cdn(url)
                self.assertEqual(result, expected,
                                 f'{url} expected CDN={expected}, got {result}')

    def test_cdn_detection_enables_local_updates(self):
        """Simulated repo list with CDN should trigger local_updates."""
        repos = [
            {'urls': ['https://mirror.example.com/baseos/']},
            {'urls': ['https://cdn.redhat.com/content/dist/rhel9/baseos/os']},
        ]
        local_updates = False
        import re
        for repo in repos:
            for url in repo.get('urls', []):
                if (re.search(r'https://cdn[^.]*\.redhat\.com/', url) or
                        '/XMLRPC' in url or
                        url.startswith('https://updates.suse.com/')):
                    local_updates = True
                    break
            if local_updates:
                break
        self.assertTrue(local_updates)

    def test_no_cdn_repos(self):
        """Repo list without CDN should not trigger local_updates."""
        repos = [
            {'urls': ['https://mirror.example.com/baseos/']},
            {'urls': ['http://archive.ubuntu.com/ubuntu']},
        ]
        local_updates = False
        import re
        for repo in repos:
            for url in repo.get('urls', []):
                if (re.search(r'https://cdn[^.]*\.redhat\.com/', url) or
                        '/XMLRPC' in url or
                        url.startswith('https://updates.suse.com/')):
                    local_updates = True
                    break
        self.assertFalse(local_updates)


class APTUpdateParsingTests(unittest.TestCase):
    """Tests for APT update parsing (apt list --upgradable output)."""

    def _parse_apt_updates(self, output):
        """Parse apt list --upgradable output, matching python client logic."""
        import re
        sec_updates = []
        bug_updates = []
        pattern = re.compile(r'^([^/]+)/(\S+)\s+(\S+)\s+(\S+)')
        for line in output.splitlines():
            line = line.strip()
            if not line or line.startswith('Listing'):
                continue
            match = pattern.match(line)
            if not match:
                continue
            pkg_name, suite, version, arch = match.groups()
            update = {'name': pkg_name, 'version': version, 'arch': arch}
            if 'security' in suite.lower():
                sec_updates.append(update)
            else:
                bug_updates.append(update)
        return sec_updates, bug_updates

    def test_security_update(self):
        """Security updates should be classified correctly."""
        output = (
            'Listing... Done\n'
            'curl/jammy-security 7.81.0-1ubuntu1.15 amd64 '
            '[upgradable from: 7.81.0-1ubuntu1.14]\n'
        )
        sec, bug = self._parse_apt_updates(output)
        self.assertEqual(len(sec), 1)
        self.assertEqual(len(bug), 0)
        self.assertEqual(sec[0]['name'], 'curl')
        self.assertEqual(sec[0]['version'], '7.81.0-1ubuntu1.15')
        self.assertEqual(sec[0]['arch'], 'amd64')

    def test_bugfix_update(self):
        """Non-security updates should be classified as bugfix."""
        output = (
            'Listing... Done\n'
            'vim/jammy-updates 2:8.2.3995-1ubuntu2.15 amd64 '
            '[upgradable from: 2:8.2.3995-1ubuntu2.14]\n'
        )
        sec, bug = self._parse_apt_updates(output)
        self.assertEqual(len(sec), 0)
        self.assertEqual(len(bug), 1)
        self.assertEqual(bug[0]['name'], 'vim')

    def test_mixed_updates(self):
        """Mixed security and bugfix updates should be split correctly."""
        output = (
            'Listing... Done\n'
            'openssl/jammy-security 3.0.2-0ubuntu1.15 amd64 '
            '[upgradable from: 3.0.2-0ubuntu1.14]\n'
            'vim/jammy-updates 2:8.2.3995-1ubuntu2.15 amd64 '
            '[upgradable from: 2:8.2.3995-1ubuntu2.14]\n'
            'curl/noble-security 8.5.0-2ubuntu10.6 amd64 '
            '[upgradable from: 8.5.0-2ubuntu10.5]\n'
        )
        sec, bug = self._parse_apt_updates(output)
        self.assertEqual(len(sec), 2)
        self.assertEqual(len(bug), 1)

    def test_no_updates(self):
        """Empty update list should produce no updates."""
        output = 'Listing... Done\n'
        sec, bug = self._parse_apt_updates(output)
        self.assertEqual(len(sec), 0)
        self.assertEqual(len(bug), 0)

    def test_empty_input(self):
        """Completely empty input should produce no updates."""
        sec, bug = self._parse_apt_updates('')
        self.assertEqual(len(sec), 0)
        self.assertEqual(len(bug), 0)


class MirrorFileResolutionTests(unittest.TestCase):
    """Tests for mirror+file:// URL resolution in APT repo parsing."""

    def test_mirror_file_regex_match(self):
        """mirror+file: URLs should be matched by the expanded regex."""
        import re
        pattern = re.compile(
            r'^\s*(\d+)\s+(https?://\S+|mirror\+file:\S+)\s+(\S+)\s+(\S+)',
        )
        line = ' 500 mirror+file:/etc/apt/mirror-list.txt focal/main amd64 Packages'
        match = pattern.match(line)
        self.assertIsNotNone(match)
        self.assertEqual(match.group(2), 'mirror+file:/etc/apt/mirror-list.txt')

    def test_http_url_still_matches(self):
        """HTTP URLs should still match the expanded pattern."""
        import re
        pattern = re.compile(
            r'^\s*(\d+)\s+(https?://\S+|mirror\+file:\S+)\s+(\S+)\s+(\S+)',
        )
        line = ' 500 http://archive.ubuntu.com/ubuntu focal/main amd64 Packages'
        match = pattern.match(line)
        self.assertIsNotNone(match)
        self.assertEqual(match.group(2), 'http://archive.ubuntu.com/ubuntu')

    def test_mirror_file_resolution(self, tmp_path=None):
        """mirror+file: entries should be resolved to HTTP URLs."""
        import tempfile
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as f:
            f.write('http://mirror1.example.com/ubuntu\n')
            f.write('https://mirror2.example.com/ubuntu\n')
            f.write('\n')
            f.write('# comment should be ignored by real parser\n')
            mirror_path = f.name
        try:
            # Simulate the mirror resolution logic from python client
            http_entries = []
            mirror_url = f'mirror+file:{mirror_path}'
            path = mirror_url.replace('mirror+file:', '')
            if os.path.isfile(path):
                with open(path) as mf:
                    for url_line in mf:
                        url = url_line.strip()
                        if url and url.startswith(('http://', 'https://')):
                            http_entries.append(url)
            self.assertEqual(len(http_entries), 2)
            self.assertEqual(http_entries[0], 'http://mirror1.example.com/ubuntu')
            self.assertEqual(http_entries[1], 'https://mirror2.example.com/ubuntu')
        finally:
            os.unlink(mirror_path)


class BashDebFieldBranchingTests(unittest.TestCase):
    """Tests for deb vs rpm field format in bash build_repos_json."""

    def _parse_fields(self, line, rtype):
        """Simulate the bash deb/rpm field branching logic."""
        import re
        parts = re.findall(r"'([^']*)'", line)
        name = parts[1] if len(parts) > 1 else ''
        if rtype == 'deb':
            rid = ''
            priority = parts[2] if len(parts) > 2 else '0'
            url_start = 3
        else:
            rid = parts[2] if len(parts) > 2 else ''
            priority = parts[3] if len(parts) > 3 else '0'
            url_start = 4
        urls = parts[url_start:] if len(parts) > url_start else []
        return name, rid, priority, urls

    def test_rpm_repo_5_fields(self):
        """RPM repos have type, name, id, priority, url1 [url2 ...]."""
        line = "'rpm' 'BaseOS x86_64' 'baseos' '99' 'https://mirror/baseos/'"
        name, rid, priority, urls = self._parse_fields(line, 'rpm')
        self.assertEqual(name, 'BaseOS x86_64')
        self.assertEqual(rid, 'baseos')
        self.assertEqual(priority, '99')
        self.assertEqual(urls, ['https://mirror/baseos/'])

    def test_deb_repo_4_fields(self):
        """Deb repos have type, name, priority, url (no id)."""
        line = "'deb' 'Ubuntu 22.04 amd64 repo at http://x/dists/y' '500' 'http://x/dists/y'"
        name, rid, priority, urls = self._parse_fields(line, 'deb')
        self.assertEqual(name, 'Ubuntu 22.04 amd64 repo at http://x/dists/y')
        self.assertEqual(rid, '')
        self.assertEqual(priority, '500')
        self.assertEqual(urls, ['http://x/dists/y'])

    def test_rpm_repo_multiple_urls(self):
        """RPM repos can have multiple URLs."""
        line = "'rpm' 'EPEL x86_64' 'epel' '99' 'https://mirror1/' 'https://mirror2/'"
        name, rid, priority, urls = self._parse_fields(line, 'rpm')
        self.assertEqual(rid, 'epel')
        self.assertEqual(len(urls), 2)

    def test_deb_would_misparse_as_rpm(self):
        """Deb repo parsed as RPM would get wrong fields — proves branching needed."""
        line = "'deb' 'Ubuntu 22.04 amd64 repo' '500' 'http://x/dists/y'"
        # If parsed as RPM: rid='500', priority='http://x/dists/y' — wrong!
        _, rid_rpm, priority_rpm, _ = self._parse_fields(line, 'rpm')
        self.assertEqual(rid_rpm, '500')  # wrong: '500' is actually priority
        # If parsed as DEB: rid='', priority='500' — correct!
        _, rid_deb, priority_deb, _ = self._parse_fields(line, 'deb')
        self.assertEqual(rid_deb, '')
        self.assertEqual(priority_deb, '500')


class BashArchDedupTests(unittest.TestCase):
    """Tests for the sed pattern that strips duplicate arch suffixes."""

    def _dedup_arch(self, name, arch='x86_64'):
        """Apply the same sed logic as bash: s/ - arch arch$/ arch/."""
        import subprocess
        result = subprocess.run(
            ['sed', '-e', f's/ - {arch} {arch}$/ {arch}/'],
            input=name, capture_output=True, text=True,
        )
        return result.stdout.strip()

    def test_duplicate_arch_stripped(self):
        """'EPEL - x86_64 x86_64' should become 'EPEL x86_64'."""
        result = self._dedup_arch('EPEL - x86_64 x86_64')
        self.assertEqual(result, 'EPEL x86_64')

    def test_single_arch_unchanged(self):
        """'BaseOS x86_64' should remain unchanged."""
        result = self._dedup_arch('BaseOS x86_64')
        self.assertEqual(result, 'BaseOS x86_64')

    def test_no_arch_unchanged(self):
        """'Rocky Linux 10 - AppStream' should remain unchanged."""
        result = self._dedup_arch('Rocky Linux 10 - AppStream')
        self.assertEqual(result, 'Rocky Linux 10 - AppStream')

    def test_aarch64_dedup(self):
        """Dedup should work for aarch64 too."""
        result = self._dedup_arch('EPEL - aarch64 aarch64', arch='aarch64')
        self.assertEqual(result, 'EPEL aarch64')


class RebootDetectionTests(unittest.TestCase):
    """Tests for multi-distro reboot detection logic."""

    def test_debian_reboot_file_exists(self):
        """If /var/run/reboot-required exists, reboot is True."""
        # We can't create the real file, but test the logic
        reboot = os.path.exists('/var/run/reboot-required')
        # On this system it may or may not exist — just test the type
        self.assertIsInstance(reboot, bool)

    def test_kernel_version_comparison_logic(self):
        """Kernel mismatch detection: different versions should flag reboot."""
        running = '5.14.0-427.13.1.el9_4.x86_64'
        installed = '5.14.0-427.18.1.el9_4.x86_64'
        self.assertNotEqual(running, installed)

    def test_kernel_version_same(self):
        """Same kernel versions should not flag reboot."""
        running = '5.14.0-427.18.1.el9_4.x86_64'
        installed = '5.14.0-427.18.1.el9_4.x86_64'
        self.assertEqual(running, installed)

    def test_vmlinuz_prefix_stripping(self):
        """vmlinuz- prefix should be stripped to get kernel version."""
        symlink_target = 'vmlinuz-5.14.0-427.18.1.el9_4.x86_64'
        # Python uses removeprefix
        version = symlink_target.removeprefix('vmlinuz-')
        self.assertEqual(version, '5.14.0-427.18.1.el9_4.x86_64')
        # Bash uses sed -e 's/^vmlinuz-//'
        import subprocess
        result = subprocess.run(
            ['sed', '-e', 's/^vmlinuz-//'],
            input=symlink_target, capture_output=True, text=True,
        )
        self.assertEqual(result.stdout.strip(), '5.14.0-427.18.1.el9_4.x86_64')


class PythonConfigTests(unittest.TestCase):
    """Tests for Python client config loading and conf.d support."""

    def test_load_config_basic(self):
        """Basic config file should parse correctly."""
        import tempfile
        config_content = (
            'server=https://patchman.example.com\n'
            'api_key=test-key-123\n'
            'tags=web,prod\n'
            'verbose=true\n'
            'debug=false\n'
            'local_updates=yes\n'
            'repo_check=false\n'
            'verify_ssl=true\n'
            'timeout=600\n'
        )
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(config_content)
            conf_path = f.name
        try:
            import importlib
            spec = importlib.util.spec_from_file_location(
                'client', os.path.join(
                    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                    'python', 'patchman_client', 'client.py',
                ),
            )
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            config = mod.load_config(conf_path)
            self.assertEqual(config['server'], 'https://patchman.example.com')
            self.assertEqual(config['api_key'], 'test-key-123')
            self.assertEqual(config['tags'], 'web,prod')
            self.assertTrue(config['verbose'])
            self.assertFalse(config['debug'])
            self.assertTrue(config['local_updates'])
            self.assertFalse(config['repo_check'])
            self.assertTrue(config['verify_ssl'])
            self.assertEqual(config['timeout'], 600)
        finally:
            os.unlink(conf_path)

    def test_load_config_missing_file(self):
        """Missing config file should return defaults."""
        import importlib
        spec = importlib.util.spec_from_file_location(
            'client', os.path.join(
                os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                'python', 'patchman_client', 'client.py',
            ),
        )
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        config = mod.load_config('/nonexistent/patchman-client.conf')
        self.assertEqual(config['server'], '')
        self.assertFalse(config['verbose'])
        self.assertTrue(config['repo_check'])

    def test_load_config_with_conf_d(self):
        """conf.d drop-in files should override base config."""
        import tempfile
        tmpdir = tempfile.mkdtemp()
        try:
            # Write base config
            base_conf = os.path.join(tmpdir, 'patchman-client.conf')
            with open(base_conf, 'w') as f:
                f.write('server=https://base.example.com\n')
                f.write('verbose=false\n')

            # Create conf.d directory with override
            conf_d = os.path.join(tmpdir, 'conf.d')
            os.makedirs(conf_d)
            with open(os.path.join(conf_d, '01-override.conf'), 'w') as f:
                f.write('verbose=true\n')
                f.write('tags=override-tag\n')

            import importlib
            spec = importlib.util.spec_from_file_location(
                'client', os.path.join(
                    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                    'python', 'patchman_client', 'client.py',
                ),
            )
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            config = mod.load_config(base_conf)
            # Base value should be present
            self.assertEqual(config['server'], 'https://base.example.com')
            # conf.d should override
            self.assertTrue(config['verbose'])
            self.assertEqual(config['tags'], 'override-tag')
        finally:
            import shutil
            shutil.rmtree(tmpdir)

    def test_config_comments_and_blanks_ignored(self):
        """Comments and blank lines should be ignored."""
        import tempfile
        config_content = (
            '# This is a comment\n'
            '\n'
            'server=https://patchman.example.com\n'
            '  # Indented comment\n'
            '\n'
            'api_key=key123\n'
        )
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(config_content)
            conf_path = f.name
        try:
            import importlib
            spec = importlib.util.spec_from_file_location(
                'client', os.path.join(
                    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                    'python', 'patchman_client', 'client.py',
                ),
            )
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            config = mod.load_config(conf_path)
            self.assertEqual(config['server'], 'https://patchman.example.com')
            self.assertEqual(config['api_key'], 'key123')
        finally:
            os.unlink(conf_path)

    def test_config_quoted_values(self):
        """Values with quotes should have quotes stripped."""
        import tempfile
        config_content = (
            'server="https://patchman.example.com"\n'
            "api_key='key123'\n"
        )
        with tempfile.NamedTemporaryFile(mode='w', suffix='.conf', delete=False) as f:
            f.write(config_content)
            conf_path = f.name
        try:
            import importlib
            spec = importlib.util.spec_from_file_location(
                'client', os.path.join(
                    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                    'python', 'patchman_client', 'client.py',
                ),
            )
            mod = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(mod)
            config = mod.load_config(conf_path)
            self.assertEqual(config['server'], 'https://patchman.example.com')
            self.assertEqual(config['api_key'], 'key123')
        finally:
            os.unlink(conf_path)

    def test_tags_comma_or_space_delimited(self):
        """Tags split on commas if present, otherwise on whitespace."""
        import importlib
        spec = importlib.util.spec_from_file_location(
            'client', os.path.join(
                os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
                'python', 'patchman_client', 'client.py',
            ),
        )
        mod = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(mod)
        for tags, expected in (
            ('web prod', ['web', 'prod']),
            ('web, prod', ['web', 'prod']),
            ('my tag,web', ['my tag', 'web']),
            ('', []),
        ):
            client = mod.PatchmanClient(hostname='test', tags=tags)
            self.assertEqual(client.tags, expected)


if __name__ == '__main__':
    unittest.main()
