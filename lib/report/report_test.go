package report

import (
	"encoding/json"
	"testing"

	"github.com/furlongm/patchman-client/go/lib/packages"
	"github.com/furlongm/patchman-client/go/lib/repos"
	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/furlongm/patchman-client/go/lib/updates"
)

func TestNewReport(t *testing.T) {
	osInfo := si.OSInfo{
		Hostname:  "testhost",
		MachineID: "test-machine-id",
		OSName:    "Test Linux",
		OSVendor:  "Test",
		OSVersion: "1.0",
		OSArch:    "x86_64",
		Kernel:    "5.10.0-test",
	}

	rpt := NewReport(osInfo)

	if rpt.Protocol != 2 {
		t.Errorf("Expected protocol 2, got %d", rpt.Protocol)
	}

	if rpt.Hostname != "testhost" {
		t.Errorf("Expected hostname 'testhost', got '%s'", rpt.Hostname)
	}

	if rpt.Arch != "x86_64" {
		t.Errorf("Expected arch 'x86_64', got '%s'", rpt.Arch)
	}
}

func TestSetTags(t *testing.T) {
	osInfo := si.OSInfo{Hostname: "test"}
	rpt := NewReport(osInfo)

	rpt.SetTags("web,prod,us-east-1")

	if len(rpt.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(rpt.Tags))
	}

	if rpt.Tags[0] != "web" || rpt.Tags[1] != "prod" || rpt.Tags[2] != "us-east-1" {
		t.Errorf("Tags not parsed correctly: %v", rpt.Tags)
	}
}

func TestAddRPMs(t *testing.T) {
	osInfo := si.OSInfo{Hostname: "test"}
	rpt := NewReport(osInfo)

	rpms := []packages.RPM{
		{Name: "bash", Epoch: "", Version: "5.1", Release: "1.el9", Arch: "x86_64"},
		{Name: "kernel", Epoch: "1", Version: "5.14.0", Release: "100.el9", Arch: "x86_64"},
	}

	rpt.AddRPMs(rpms)

	if len(rpt.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(rpt.Packages))
	}

	if rpt.Packages[0].Type != "rpm" {
		t.Errorf("Expected type 'rpm', got '%s'", rpt.Packages[0].Type)
	}
}

func TestAddYumRepos(t *testing.T) {
	osInfo := si.OSInfo{Hostname: "test"}
	rpt := NewReport(osInfo)

	yumRepos := []repos.YumRepo{
		{ID: "baseos", Name: "BaseOS", BaseURL: "http://mirror/baseos", Priority: 0, Enabled: true},
		{ID: "appstream", Name: "AppStream", BaseURL: "http://mirror/appstream", Priority: 10, Enabled: true},
	}

	rpt.AddYumRepos(yumRepos)

	if len(rpt.Repos) != 2 {
		t.Errorf("Expected 2 repos, got %d", len(rpt.Repos))
	}

	if rpt.Repos[0].Type != "rpm" {
		t.Errorf("Expected type 'rpm', got '%s'", rpt.Repos[0].Type)
	}
}

func TestAddSecurityUpdates(t *testing.T) {
	osInfo := si.OSInfo{Hostname: "test"}
	rpt := NewReport(osInfo)

	secUpdates := []updates.Update{
		{Name: "openssl", Version: "3.0.8-1", Arch: "x86_64"},
		{Name: "curl", Version: "8.0.1-1", Arch: "x86_64"},
	}

	rpt.AddSecurityUpdates(secUpdates)

	if len(rpt.SecUpdates) != 2 {
		t.Errorf("Expected 2 security updates, got %d", len(rpt.SecUpdates))
	}
}

func TestToJSON(t *testing.T) {
	osInfo := si.OSInfo{
		Hostname: "testhost",
		OSName:   "Test Linux",
		OSArch:   "x86_64",
		Kernel:   "5.10.0-test",
	}

	rpt := NewReport(osInfo)
	rpt.SetTags("test")

	jsonData, err := rpt.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check protocol
	if proto, ok := parsed["protocol"].(float64); !ok || proto != 2 {
		t.Errorf("Expected protocol 2, got %v", parsed["protocol"])
	}

	// Check hostname
	if hostname, ok := parsed["hostname"].(string); !ok || hostname != "testhost" {
		t.Errorf("Expected hostname 'testhost', got %v", parsed["hostname"])
	}
}

func TestCheckRebootRequired(t *testing.T) {
	// Just check that it doesn't panic
	result := CheckRebootRequired()
	t.Logf("Reboot required: %v", result)
}
