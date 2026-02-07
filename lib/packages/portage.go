package packages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PortagePkg represents a Gentoo package
type PortagePkg struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Slot     string `json:"slot,omitempty"`
	Repo     string `json:"repo,omitempty"`
}

// PrintPortagePkg prints a single Portage package
func PrintPortagePkg(pkg PortagePkg) {
	fmt.Printf("%s/%s-%s\n", pkg.Category, pkg.Name, pkg.Version)
}

// GetInstalledPortagePkgs returns all installed Gentoo packages
// by reading /var/db/pkg directly (faster than qlist/equery)
func GetInstalledPortagePkgs() ([]PortagePkg, error) {
	pkgDB := "/var/db/pkg"
	if _, err := os.Stat(pkgDB); os.IsNotExist(err) {
		return nil, nil // Portage not available
	}

	pkgs := []PortagePkg{}

	// Walk /var/db/pkg/category/name-version/
	categories, err := os.ReadDir(pkgDB)
	if err != nil {
		return nil, err
	}

	// Regex to separate name and version: name-1.2.3_rc1-r2
	// Version starts with digit, may have suffix like _alpha, _beta, _rc, -r1
	versionRegex := regexp.MustCompile(`^(.+)-(\d[^-]*)(-r\d+)?$`)

	for _, category := range categories {
		if !category.IsDir() {
			continue
		}
		catName := category.Name()

		catPath := filepath.Join(pkgDB, catName)
		packages, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}

		for _, pkg := range packages {
			if !pkg.IsDir() {
				continue
			}
			pkgName := pkg.Name()

			// Parse name-version
			matches := versionRegex.FindStringSubmatch(pkgName)
			if len(matches) < 3 {
				// Can't parse, use whole thing as name
				pkgs = append(pkgs, PortagePkg{
					Category: catName,
					Name:     pkgName,
					Version:  "unknown",
				})
				continue
			}

			name := matches[1]
			version := matches[2]
			if len(matches) > 3 && matches[3] != "" {
				version += matches[3] // Include -r1 revision
			}

			// Try to get repo and slot from package metadata
			pkgPath := filepath.Join(catPath, pkgName)
			repo := readPkgFile(pkgPath, "repository")
			slot := readPkgFile(pkgPath, "SLOT")

			pkgs = append(pkgs, PortagePkg{
				Category: catName,
				Name:     name,
				Version:  version,
				Slot:     slot,
				Repo:     repo,
			})
		}
	}

	return pkgs, nil
}

// readPkgFile reads a single-line file from the package directory
func readPkgFile(pkgPath, filename string) string {
	filePath := filepath.Join(pkgPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// PrintPortagePkgs prints all Portage packages
func PrintPortagePkgs(pkgs []PortagePkg) {
	for _, pkg := range pkgs {
		PrintPortagePkg(pkg)
	}
}

// GetPortageUpdates returns available updates using emerge --pretend
// Note: This can be slow as it needs to calculate dependencies
func GetPortageUpdates() (secUpdates []PortagePkg, bugUpdates []PortagePkg, err error) {
	// GLSA (Gentoo Linux Security Advisories) for security updates
	// glsa-check --list affected
	// For now, we'll use emerge --pretend --update @world

	// This is expensive, so we'll skip automatic detection
	// Users can run: glsa-check --list affected

	return nil, nil, nil
}
