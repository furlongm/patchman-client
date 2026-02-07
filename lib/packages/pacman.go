package packages

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
)

// PacmanPkg represents an Arch Linux package
type PacmanPkg struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
}

// PrintPacmanPkg prints a single pacman package
func PrintPacmanPkg(pkg PacmanPkg) {
	fmt.Printf("%s-%s-%s\n", pkg.Name, pkg.Version, pkg.Arch)
}

// GetInstalledPacmanPkgs returns all installed Arch Linux packages
func GetInstalledPacmanPkgs() ([]PacmanPkg, error) {
	// pacman -Q outputs: package-name version
	cmd, err := util.RunCommand("pacman", "-Q")
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		return nil, nil // pacman not available
	}

	pkgs := []PacmanPkg{}

	// Get architecture for all packages via pacman -Qi
	// For efficiency, we'll query each package's arch
	// Or use: pacman -Qqi | grep -E "^(Name|Version|Architecture)"

	// Simpler: use pacman -Qi to get detailed info
	detailCmd, err := util.RunCommand("pacman", "-Qi")
	if err != nil {
		// Fall back to basic query without arch
		for _, line := range strings.Split(string(cmd), "\n") {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkgs = append(pkgs, PacmanPkg{
					Name:    parts[0],
					Version: parts[1],
					Arch:    "x86_64", // default
				})
			}
		}
		return pkgs, nil
	}

	// Parse pacman -Qi output
	var currentPkg PacmanPkg
	for _, line := range strings.Split(string(detailCmd), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name") {
			// Start of new package
			if currentPkg.Name != "" {
				pkgs = append(pkgs, currentPkg)
			}
			currentPkg = PacmanPkg{}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentPkg.Name = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Version") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentPkg.Version = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Architecture") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentPkg.Arch = strings.TrimSpace(parts[1])
			}
		}
	}
	// Don't forget the last package
	if currentPkg.Name != "" {
		pkgs = append(pkgs, currentPkg)
	}

	return pkgs, nil
}

// PrintPacmanPkgs prints all pacman packages
func PrintPacmanPkgs(pkgs []PacmanPkg) {
	for _, pkg := range pkgs {
		PrintPacmanPkg(pkg)
	}
}

// GetPacmanUpdates returns available security and bugfix updates
// Arch doesn't distinguish between security/bugfix, so all updates go to bugfix
func GetPacmanUpdates() (secUpdates []PacmanPkg, bugUpdates []PacmanPkg, err error) {
	// checkupdates is from pacman-contrib, outputs: package oldver -> newver
	cmd, err := util.RunCommand("checkupdates")
	if err != nil {
		return nil, nil, err
	}
	if cmd == nil {
		return nil, nil, nil
	}

	// Parse output: package-name oldver -> newver
	updateRegex := regexp.MustCompile(`^(\S+)\s+\S+\s+->\s+(\S+)`)
	for _, line := range strings.Split(string(cmd), "\n") {
		if line == "" {
			continue
		}
		matches := updateRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			bugUpdates = append(bugUpdates, PacmanPkg{
				Name:    matches[1],
				Version: matches[2],
				Arch:    "x86_64", // default
			})
		}
	}

	// Arch Security Tracker doesn't have a simple CLI tool
	// All updates treated as bugfix unless we integrate with arch-audit
	return secUpdates, bugUpdates, nil
}
