package si

import (
	"testing"
)

func TestParseOSRelease(t *testing.T) {
	result := parseOSRelease()

	// Should have at least some keys from /etc/os-release
	if len(result) == 0 {
		t.Skip("No /etc/os-release found, skipping test")
	}

	// NAME should be present
	if _, ok := result["NAME"]; !ok {
		t.Error("Expected NAME key in os-release")
	}

	// ID should be present
	if _, ok := result["ID"]; !ok {
		t.Error("Expected ID key in os-release")
	}
}

func TestGetOSInfo(t *testing.T) {
	osInfo := GetOSInfo()

	// Hostname should not be empty
	if osInfo.Hostname == "" {
		t.Error("Expected non-empty hostname")
	}

	// OSName should not be empty
	if osInfo.OSName == "" {
		t.Error("Expected non-empty OS name")
	}

	// OSArch should not be empty
	if osInfo.OSArch == "" {
		t.Error("Expected non-empty OS arch")
	}

	// Kernel should not be empty
	if osInfo.Kernel == "" {
		t.Error("Expected non-empty kernel")
	}
}

func TestGetOSInfoCPE(t *testing.T) {
	osInfo := GetOSInfo()

	// CPE may or may not be present depending on the distro
	// Just check that we don't crash
	t.Logf("CPE: %s", osInfo.CPE)
	t.Logf("Codename: %s", osInfo.Codename)
}
