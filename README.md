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

### Debian/Ubuntu

Add the repository and install:

```bash
# Add repository (TODO: update with actual repo URL)
echo "deb https://packagecloud.io/furlongm/patchman/any/ any main" | sudo tee /etc/apt/sources.list.d/patchman.list
curl -fsSL https://packagecloud.io/furlongm/patchman/gpgkey | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/patchman.gpg
sudo apt update

# Install (pulls in bash client by default)
sudo apt install patchman-client

# Or explicitly choose an implementation
sudo apt install patchman-client patchman-client-bash
sudo apt install patchman-client patchman-client-python
sudo apt install patchman-client patchman-client-go
```

### RHEL/Fedora/CentOS

Add the repository and install:

```bash
# Add repository (TODO: update with actual repo URL)
sudo tee /etc/yum.repos.d/patchman.repo << 'EOF'
[patchman]
name=Patchman
baseurl=https://packagecloud.io/furlongm/patchman/el/$releasever/$basearch
gpgcheck=1
gpgkey=https://packagecloud.io/furlongm/patchman/gpgkey
enabled=1
EOF

# Install (pulls in bash client by default)
sudo dnf install patchman-client

# Or explicitly choose an implementation
sudo dnf install patchman-client patchman-client-bash
sudo dnf install patchman-client patchman-client-python
sudo dnf install patchman-client patchman-client-go
```

### Arch Linux

```bash
# Install from AUR
yay -S patchman-client
```

### Gentoo

```bash
# Install from overlay
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

# Protocol version (1 = form data, 2 = REST/JSON)
# Protocol 2 requires jq for bash client
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

- **apt** (Debian/Ubuntu): `/etc/apt/apt.conf.d/05patchman`
- **dnf** (Fedora/RHEL 8+): `/etc/dnf/plugins/post-transaction-actions.d/patchman.action`
- **yum** (RHEL 7): `/etc/yum/pluginconf.d/patchman.conf`
- **zypper** (openSUSE/SLES): zypper system plugin
- **pacman** (Arch): `/etc/pacman.d/hooks/patchman.hook`
- **portage** (Gentoo): `/etc/portage/bashrc`

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
