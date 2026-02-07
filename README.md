# Patchman Client

Patchman client sends system package information to a [Patchman](https://github.com/furlongm/patchman) server.

## Implementations

Three client implementations are available:

| Implementation | Description | Dependencies |
|---------------|-------------|--------------|
| **Bash** (default) | Most tested, recommended | curl, awk, jq |
| **Python** | REST API only (protocol 2) | Python 3.6+, requests |
| **Go** | Compiled binary | None |

## Installation

### Debian/Ubuntu

```bash
# Install with bash client (default)
apt install patchman-client patchman-client-bash

# Or with python client
apt install patchman-client patchman-client-python

# Or with go client
apt install patchman-client patchman-client-go
```

### RHEL/Fedora/CentOS

```bash
# Install with bash client (default)
dnf install patchman-client patchman-client-bash

# Or with python/go client
dnf install patchman-client patchman-client-python
dnf install patchman-client patchman-client-go
```

### Arch Linux

```bash
# Install from AUR or build from source
makepkg -si
```

### Gentoo

```bash
# Install from overlay or build from source
emerge patchman-client
```

### From Source (Go)

```bash
go install github.com/furlongm/patchman-client/go/cmd@latest
```

### From PyPI (Python)

```bash
pip install patchman-client
```

## Configuration

Edit `/etc/patchman/patchman-client.conf`:

```ini
# Patchman server URL
server=https://patchman.example.com

# Options to curl (bash client only)
curl_options="--insecure --connect-timeout 60 --max-time 300"

# Space delimited tags to send
tags="Server"

# Protocol version (1 = text, 2 = json)
# Protocol 2 requires jq (bash) or is default (python/go)
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

# Send with tags
patchman-client -t web,production
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

- **APT** (Debian/Ubuntu): `/etc/apt/apt.conf.d/05patchman`
- **DNF** (Fedora/RHEL 8+): `/etc/dnf/plugins/post-transaction-actions.d/patchman.action`
- **YUM** (RHEL 7): `/etc/yum/pluginconf.d/patchman.conf`
- **Zypper** (openSUSE/SLES): zypper system plugin
- **Pacman** (Arch): `/etc/pacman.d/hooks/patchman.hook`
- **Portage** (Gentoo): `/etc/portage/bashrc`

## Systemd Timer

A systemd timer is included for periodic reporting (disabled by default):

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
