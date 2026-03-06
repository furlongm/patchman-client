package si

import (
	"testing"
)

func TestBuildOSStringRocky(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "rocky",
		"PRETTY_NAME": "Rocky Linux 9.5 (Blue Onyx)",
		"CPE_NAME":    "cpe:/o:rocky:rocky:9",
	}
	result := buildOSString(osRelease, osRelease["CPE_NAME"])
	expected := "Rocky Linux 9.5 (Blue Onyx) [cpe:/o:rocky:rocky:9]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringAlmaLinux(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "almalinux",
		"PRETTY_NAME": "AlmaLinux 9.5 (Teal Serval)",
		"CPE_NAME":    "cpe:/o:almalinux:almalinux:9",
	}
	result := buildOSString(osRelease, osRelease["CPE_NAME"])
	expected := "AlmaLinux 9.5 (Teal Serval) [cpe:/o:almalinux:almalinux:9]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringUbuntu(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "ubuntu",
		"PRETTY_NAME": "Ubuntu 22.04.3 LTS",
	}
	result := buildOSString(osRelease, "")
	expected := "Ubuntu 22.04.3 LTS"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringFedora(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "fedora",
		"PRETTY_NAME": "Fedora Linux 42 (Server Edition)",
		"CPE_NAME":    "cpe:/o:fedoraproject:fedora:42",
	}
	result := buildOSString(osRelease, osRelease["CPE_NAME"])
	expected := "Fedora Linux 42 (Server Edition) [cpe:/o:fedoraproject:fedora:42]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringArch(t *testing.T) {
	osRelease := map[string]string{
		"ID":   "arch",
		"NAME": "Arch Linux",
	}
	result := buildOSString(osRelease, "")
	expected := "Arch Linux"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringGentoo(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "gentoo",
		"PRETTY_NAME": "Gentoo Linux",
		"VERSION_ID":  "2.15",
	}
	result := buildOSString(osRelease, "")
	expected := "Gentoo Linux 2.15"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringSUSE(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "opensuse-leap",
		"ID_LIKE":     "suse opensuse",
		"PRETTY_NAME": "openSUSE Leap 15.6",
		"CPE_NAME":    "cpe:/o:opensuse:leap:15.6",
	}
	result := buildOSString(osRelease, osRelease["CPE_NAME"])
	expected := "openSUSE Leap 15.6 [cpe:/o:opensuse:leap:15.6]"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringNoCPE(t *testing.T) {
	osRelease := map[string]string{
		"ID":          "rocky",
		"PRETTY_NAME": "Rocky Linux 9.5 (Blue Onyx)",
	}
	result := buildOSString(osRelease, "")
	expected := "Rocky Linux 9.5 (Blue Onyx)"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringGenericDistro(t *testing.T) {
	osRelease := map[string]string{
		"ID":      "nixos",
		"NAME":    "NixOS",
		"VERSION": "24.11",
	}
	result := buildOSString(osRelease, "")
	expected := "NixOS 24.11"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestBuildOSStringEmpty(t *testing.T) {
	result := buildOSString(map[string]string{}, "")
	if result != "" {
		t.Errorf("Expected empty string for empty os-release, got '%s'", result)
	}
}
