Name:           patchman-client
Version:        %{?version}%{!?version:%(git describe --tags --always 2>/dev/null | sed 's/^v//' || echo "0.0.0")}
Release:        1%{?dist}
Summary:        Patchman client for reporting system packages
License:        GPLv3
URL:            https://github.com/furlongm/patchman-client
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21
BuildRequires:  python3-devel

# Disable debuginfo - Go binaries don't have traditional debug symbols
%global debug_package %{nil}

%description
Patchman client sends system package information to a patchman server.
This meta-package installs package manager hooks and depends on one of
the client implementations (bash, python, or go).

%package bash
Summary:        Patchman client - bash implementation
Requires:       bash curl coreutils gawk grep sed
Recommends:     jq
Provides:       patchman-client-impl

%description bash
Patchman client sends system package information to a patchman server.
This package provides the bash implementation of the patchman client.
It is the most tested and recommended implementation.
jq is recommended for protocol 2 (REST/JSON) support.

%package python
Summary:        Patchman client - Python implementation
Requires:       python3 python3-requests
Provides:       patchman-client-impl

%description python
Patchman client sends system package information to a patchman server.
This package provides the Python implementation of the patchman client.
It uses the REST API (protocol 2) only.

%package go
Summary:        Patchman client - Go implementation
Provides:       patchman-client-impl

%description go
Patchman client sends system package information to a patchman server.
This package provides the Go implementation of the patchman client.
It is a compiled binary with no runtime dependencies.
It uses the REST API (protocol 2) only.

%prep
%autosetup

%build
cd go
go build -ldflags "-X github.com/furlongm/patchman-client/go/lib/report.Version=%{version}" -o patchman-client-go ./cmd

%install
# Bash client
install -D -m 755 bash/patchman-client %{buildroot}%{_sbindir}/patchman-client-bash
install -D -m 644 bash/patchman-client.conf %{buildroot}%{_sysconfdir}/patchman/patchman-client.conf

# Python client
install -D -m 755 python/patchman_client/client.py %{buildroot}%{_sbindir}/patchman-client-python

# Go client
install -D -m 755 go/patchman-client-go %{buildroot}%{_sbindir}/patchman-client-go

# Hooks (in base package)
install -D -m 644 hooks/dnf/patchman.action %{buildroot}%{_sysconfdir}/dnf/plugins/post-transaction-actions.d/patchman.action
install -D -m 644 hooks/yum/patchman.conf %{buildroot}%{_sysconfdir}/yum/pluginconf.d/patchman.conf
install -D -m 644 hooks/yum/patchman.py %{buildroot}%{_prefix}/lib/yum-plugins/patchman.py

# Systemd files (in base package)
install -D -m 644 etc/systemd/patchman-client.service %{buildroot}%{_unitdir}/patchman-client.service
install -D -m 644 etc/systemd/patchman-client.timer %{buildroot}%{_unitdir}/patchman-client.timer

%post bash
alternatives --install %{_sbindir}/patchman-client patchman-client %{_sbindir}/patchman-client-bash 30

%preun bash
if [ $1 -eq 0 ]; then
    alternatives --remove patchman-client %{_sbindir}/patchman-client-bash
fi

%post python
alternatives --install %{_sbindir}/patchman-client patchman-client %{_sbindir}/patchman-client-python 20

%preun python
if [ $1 -eq 0 ]; then
    alternatives --remove patchman-client %{_sbindir}/patchman-client-python
fi

%post go
alternatives --install %{_sbindir}/patchman-client patchman-client %{_sbindir}/patchman-client-go 10

%preun go
if [ $1 -eq 0 ]; then
    alternatives --remove patchman-client %{_sbindir}/patchman-client-go
fi

%files
%license LICENSE
%doc README.md
%{_sysconfdir}/dnf/plugins/post-transaction-actions.d/patchman.action
%{_sysconfdir}/yum/pluginconf.d/patchman.conf
%{_prefix}/lib/yum-plugins/patchman.py
%{_unitdir}/patchman-client.service
%{_unitdir}/patchman-client.timer

%files bash
%license LICENSE
%{_sbindir}/patchman-client-bash
%config(noreplace) %{_sysconfdir}/patchman/patchman-client.conf

%files python
%license LICENSE
%{_sbindir}/patchman-client-python

%files go
%license LICENSE
%{_sbindir}/patchman-client-go

%changelog
* Sat Feb 07 2026 Marcus Furlong <furlongm@gmail.com> - 4.0.8-1
- Initial release of unified patchman-client package
- Includes bash, python, and go implementations
- Uses alternatives for switching implementations
