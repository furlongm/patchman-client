package updates

import (
	"regexp"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
)

// Update represents an available update
type Update struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Repo    string `json:"repo,omitempty"`
}

// GetDNFUpdates returns available security and bugfix updates from DNF
func GetDNFUpdates() (secUpdates []Update, bugUpdates []Update, err error) {
	// Check for available updates via dnf updateinfo
	cmd, err := util.RunCommand("dnf", "updateinfo", "list", "--available")
	if err != nil {
		return nil, nil, err
	}
	if cmd == nil || len(cmd) == 0 {
		return nil, nil, nil
	}

	// Parse output: Advisory-ID  Type  Package-NEVRA
	// Example: RLSA-2025:0377  Moderate/Sec.  NetworkManager-1:1.48.10-5.el9_5.x86_64
	lines := strings.Split(string(cmd), "\n")

	// Regex to parse package NEVRA (Name-Epoch:Version-Release.Arch)
	// Format: package-name-[epoch:]version-release.arch
	nevraRegex := regexp.MustCompile(`^(.+)-([0-9]+:)?([^-]+)-([^.]+)\.(\S+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Last metadata") {
			continue
		}

		// Split by whitespace
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		// advisory := fields[0]  // e.g., RLSA-2025:0377
		updateType := fields[1] // e.g., Moderate/Sec. or bugfix
		pkgNEVRA := fields[2]   // e.g., NetworkManager-1:1.48.10-5.el9_5.x86_64

		// Parse the package NEVRA
		matches := nevraRegex.FindStringSubmatch(pkgNEVRA)
		if len(matches) < 6 {
			continue
		}

		update := Update{
			Name:    matches[1],
			Version: matches[3] + "-" + matches[4], // version-release
			Arch:    matches[5],
		}

		// Classify by type
		if strings.Contains(strings.ToLower(updateType), "sec") {
			secUpdates = append(secUpdates, update)
		} else {
			bugUpdates = append(bugUpdates, update)
		}
	}

	return secUpdates, bugUpdates, nil
}

// GetYUMUpdates is an alias for GetDNFUpdates (same command works for both)
func GetYUMUpdates() (secUpdates []Update, bugUpdates []Update, err error) {
	return GetDNFUpdates()
}
