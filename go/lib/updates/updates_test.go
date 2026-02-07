package updates

import (
	"os"
	"testing"
)

func TestGetDNFUpdates(t *testing.T) {
	// Skip if dnf is not available
	if _, err := os.Stat("/usr/bin/dnf"); os.IsNotExist(err) {
		t.Skip("DNF not available, skipping")
	}

	secUpdates, bugUpdates, err := GetDNFUpdates()
	if err != nil {
		t.Fatalf("GetDNFUpdates failed: %v", err)
	}

	t.Logf("Found %d security updates, %d bugfix updates", len(secUpdates), len(bugUpdates))

	// Verify update structure if any found
	for _, upd := range secUpdates {
		if upd.Name == "" {
			t.Error("Security update name should not be empty")
		}
		if upd.Version == "" {
			t.Error("Security update version should not be empty")
		}
	}
}

func TestGetZypperUpdates(t *testing.T) {
	// Skip if zypper is not available
	if _, err := os.Stat("/usr/bin/zypper"); os.IsNotExist(err) {
		t.Skip("Zypper not available, skipping")
	}

	secUpdates, bugUpdates, err := GetZypperUpdates()
	if err != nil {
		t.Fatalf("GetZypperUpdates failed: %v", err)
	}

	t.Logf("Found %d security updates, %d bugfix updates", len(secUpdates), len(bugUpdates))
}

func TestGetAPTUpdates(t *testing.T) {
	// Skip if apt is not available
	if _, err := os.Stat("/usr/bin/apt"); os.IsNotExist(err) {
		t.Skip("APT not available, skipping")
	}

	secUpdates, bugUpdates, err := GetAPTUpdates()
	if err != nil {
		t.Fatalf("GetAPTUpdates failed: %v", err)
	}

	t.Logf("Found %d security updates, %d bugfix updates", len(secUpdates), len(bugUpdates))
}
