# Patchman Client

Patchman client sends system package information to a [Patchman](https://github.com/furlongm/patchman) server.

Patchman clients send a list of installed packages and enabled repositories to
the Patchman server. The Patchman server updates its package list for each
repository and determines which hosts require updates, and whether those updates
are normal or security updates.

## Implementations

Three client implementations are available:

| Implementation | Description | Protocol | Dependencies |
|---------------|-------------|----------|--------------|
| **bash** (default) | Most tested, recommended | 1 (form) or 2 (REST) | bash, curl, awk, grep, sed |
| **python** | REST API only | 2 (REST) | Python 3.6+, requests |
| **go** | Compiled binary | 2 (REST) | None |

## Installation

### Ubuntu 24.04 (noble)

```bash
curl -sS https://repo.openbytes.ie/openbytes-1.gpg | tee /usr/share/keyrings/openbytes.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/openbytes.gpg] https://repo.openbytes.ie/patchman/ubuntu noble-backports main" > /etc/apt/sources.list.d/patchman.list
apt update
apt -y install  patchman-client

# or explicitly choose an implementation
apt -y install patchman-client-bash
apt -y install patchman-client-python
apt -y install patchman-client-go
```

### Debian 13 (trixie)
```bash
curl -sS https://repo.openbytes.ie/openbytes-1.gpg | tee /usr/share/keyrings/openbytes.gpg > /dev/null
echo "deb [signed-by=/usr/share/keyrings/openbytes.gpg] https://repo.openbytes.ie/patchman/debian trixie main" > /etc/apt/sources.list.d/patchman.list
apt update
apt -y install  patchman-client

# or explicitly choose an implementation
apt -y install patchman-client-bash
apt -y install patchman-client-python
apt -y install patchman-client-go
```

### RHEL/Rocky/Alma/Suse

```bash
curl -sS https://repo.openbytes.ie/openbytes-2.gpg > /etc/pki/rpm-gpg/RPM-GPG-KEY-openbytes
cat <<EOF >> /etc/yum.repos.d/openbytes.repo
[patchman]
name=Patchman
baseurl=https://repo.openbytes.ie/patchman/el10
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-openbytes
enabled=1
EOF
dnf -y install patchman-client

# or explicitly choose an implementation
dnf -y install patchman-client-bash
dnf -y install patchman-client-python
dnf -y install patchman-client-go
```

### Source (Go)

```bash
cd go && go build -o patchman-client ./cmd
```

### From GitHub Releases

Download pre-built Go binaries from [GitHub Releases](https://github.com/furlongm/patchman-client/releases):

```bash
# Download and install (amd64)
curl -LO https://github.com/furlongm/patchman-client/releases/latest/download/patchman-client-go-linux-amd64
install -m 755 patchman-client-go-linux-amd64 /usr/sbin/patchman-client
```

## Configuration

Edit `/etc/patchman/patchman-client.conf`:

```ini
# patchman server URL
server=https://patchman.example.com

# options to curl (bash client only)
curl_options="--insecure --connect-timeout 60 --max-time 300"

# space delimited tags to send
tags="server"

# protocol version (1 = text/form data, 2 = REST API)
protocol=1

# API key for protocol 2 authentication
#api_key=pm_your_api_key_here
```

## Usage

```bash
# Send report to server
patchman-client

# Verbose output
patchman-client -v

# Debug output
patchman-client -d

# Skip repo check (used by package manager hooks)
patchman-client -n

# Override server
patchman-client -s https://patchman.example.com
```

## Switching Implementations

Use `update-alternatives` to switch between implementations:

```bash
# Show current configuration
update-alternatives --display patchman-client

# Interactively choose implementation
update-alternatives --config patchman-client

# Set specific implementation
update-alternatives --set patchman-client /usr/sbin/patchman-client-bash
update-alternatives --set patchman-client /usr/sbin/patchman-client-python
update-alternatives --set patchman-client /usr/sbin/patchman-client-go
```

## Package Manager Hooks

The `patchman-client` meta-package installs hooks for automatic reporting after package changes:

- **apt** (Debian/Ubuntu): `/etc/apt/apt.conf.d/05patchman`
- **dnf** (Fedora/RHEL 8+): `/etc/dnf/plugins/post-transaction-actions.d/patchman.action`
- **yum** (RHEL 7): `/etc/yum/pluginconf.d/patchman.conf`
- **zypper** (openSUSE/SLES): `/usr/lib/zypp/plugins/commit/patchman.py`
- **pacman** (Arch): `/etc/pacman.d/hooks/patchman.hook`
- **portage** (Gentoo): `/etc/portage/bashrc`

## Systemd Timer

A systemd timer is included for periodic reporting (disabled by default, hooks enabled by default):

```bash
# Enable daily reports
systemctl enable --now patchman-client.timer

# Check timer status
systemctl list-timers patchman-client.timer
```

## License

GPL-3.0 - see [LICENSE](LICENSE)

## Author

Marcus Furlong <furlongm@gmail.com>
