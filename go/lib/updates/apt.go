package updates

import (
	"regexp"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
)

// GetAPTUpdates returns available security and bugfix updates from APT
func GetAPTUpdates() (secUpdates []Update, bugUpdates []Update, err error) {
	// First, check for updates via apt list --upgradable
	cmd, err := util.RunCommand("apt", "list", "--upgradable")
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) == 0 {
		return nil, nil, nil
	}

	// Get security updates from apt-get upgrade simulation with security only
	secCmd, _ := util.RunCommand("apt-get", "-s", "-o", "Dir::Etc::SourceList=/etc/apt/sources.list.d/", "-o", "Dir::Etc::SourceParts=/dev/null", "upgrade")

	// Parse security package names
	secPackages := make(map[string]bool)
	if secCmd != nil {
		// Parse lines like: Inst package-name [current] (new-version repo)
		instRegex := regexp.MustCompile(`^Inst\s+(\S+)`)
		for _, line := range strings.Split(string(secCmd), "\n") {
			if matches := instRegex.FindStringSubmatch(line); len(matches) >= 2 {
				secPackages[matches[1]] = true
			}
		}
	}

	// Alternative: check for -security in origin
	// Parse apt list --upgradable output
	// Format: package/suite version arch [upgradable from: old-version]
	// Example: curl/jammy-security 7.81.0-1ubuntu1.15 amd64 [upgradable from: 7.81.0-1ubuntu1.14]
	upgradeRegex := regexp.MustCompile(`^([^/]+)/(\S+)\s+(\S+)\s+(\S+)`)

	for _, line := range strings.Split(string(cmd), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Listing") {
			continue
		}

		matches := upgradeRegex.FindStringSubmatch(line)
		if len(matches) < 5 {
			continue
		}

		pkgName := matches[1]
		suite := matches[2]
		version := matches[3]
		arch := matches[4]

		update := Update{
			Name:    pkgName,
			Version: version,
			Arch:    arch,
		}

		// Check if it's a security update (suite contains -security)
		if strings.Contains(suite, "-security") || strings.Contains(suite, "security") || secPackages[pkgName] {
			secUpdates = append(secUpdates, update)
		} else {
			bugUpdates = append(bugUpdates, update)
		}
	}

	return secUpdates, bugUpdates, nil
}
