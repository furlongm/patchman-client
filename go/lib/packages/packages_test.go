package packages

import (
	"os"
	"testing"
)

func TestGetInstalledRPMs(t *testing.T) {
	// Skip if not an RPM-based system (check for /var/lib/rpm database)
	if _, err := os.Stat("/var/lib/rpm"); os.IsNotExist(err) {
		t.Skip("Not an RPM-based system, skipping")
	}

	rpms, err := GetInstalledRPMs()
	if err != nil {
		t.Fatalf("GetInstalledRPMs failed: %v", err)
	}

	t.Logf("Found %d RPM packages", len(rpms))

	// Verify package structure if we have packages
	for _, rpm := range rpms[:min(5, len(rpms))] {
		if rpm.Name == "" {
			t.Error("RPM name should not be empty")
		}
		if rpm.Version == "" {
			t.Error("RPM version should not be empty")
		}
		if rpm.Arch == "" {
			t.Error("RPM arch should not be empty")
		}
	}
}

func TestGetInstalledDEBs(t *testing.T) {
	// Skip if dpkg-query is not available
	if _, err := os.Stat("/usr/bin/dpkg-query"); os.IsNotExist(err) {
		t.Skip("dpkg-query not available, skipping")
	}

	debs, err := GetInstalledDEBs()
	if err != nil {
		t.Fatalf("GetInstalledDEBs failed: %v", err)
	}

	t.Logf("Found %d DEB packages", len(debs))
}

func TestGetInstalledPacmanPkgs(t *testing.T) {
	// Skip if pacman is not available
	if _, err := os.Stat("/usr/bin/pacman"); os.IsNotExist(err) {
		t.Skip("Pacman not available, skipping")
	}

	pkgs, err := GetInstalledPacmanPkgs()
	if err != nil {
		t.Fatalf("GetInstalledPacmanPkgs failed: %v", err)
	}

	t.Logf("Found %d Pacman packages", len(pkgs))
}

func TestGetInstalledPortagePkgs(t *testing.T) {
	// Skip if Portage is not available
	if _, err := os.Stat("/var/db/pkg"); os.IsNotExist(err) {
		t.Skip("Portage not available, skipping")
	}

	pkgs, err := GetInstalledPortagePkgs()
	if err != nil {
		t.Fatalf("GetInstalledPortagePkgs failed: %v", err)
	}

	t.Logf("Found %d Portage packages", len(pkgs))
}

func TestGetInstalledDNFModules(t *testing.T) {
	// Skip if dnf is not available
	if _, err := os.Stat("/usr/bin/dnf"); os.IsNotExist(err) {
		t.Skip("DNF not available, skipping")
	}

	modules, err := GetInstalledDNFModules()
	if err != nil {
		t.Fatalf("GetInstalledDNFModules failed: %v", err)
	}

	t.Logf("Found %d DNF modules", len(modules))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
