package repos

import (
	"os"
	"testing"
)

func TestGetYumRepos(t *testing.T) {
	// Skip if not on RPM-based system
	if _, err := os.Stat("/etc/yum.repos.d"); os.IsNotExist(err) {
		t.Skip("Not an RPM-based system, skipping")
	}

	repos, err := GetYumRepos()
	if err != nil {
		t.Fatalf("GetYumRepos failed: %v", err)
	}

	// Should find at least one repo on RPM-based systems
	if len(repos) == 0 {
		t.Log("Warning: no YUM repos found")
	}

	// Verify repo structure
	for _, repo := range repos {
		if repo.ID == "" {
			t.Error("Repo ID should not be empty")
		}
		if repo.Name == "" {
			t.Error("Repo Name should not be empty")
		}
	}
}

func TestParseYumRepoFile(t *testing.T) {
	// Create a temporary repo file for testing
	content := `[test-repo]
name=Test Repository
baseurl=http://example.com/repo/$releasever/$basearch
enabled=1
gpgcheck=0
priority=10

[disabled-repo]
name=Disabled Repository
baseurl=http://example.com/disabled
enabled=0
`
	tmpFile, err := os.CreateTemp("", "test-*.repo")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	repos, err := parseRepoFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseRepoFile failed: %v", err)
	}

	// parseRepoFile returns all repos (enabled and disabled)
	if len(repos) != 2 {
		t.Errorf("Expected 2 repos (all), got %d", len(repos))
	}

	// Filter to enabled only (like GetYumRepos does)
	var enabledRepos []YumRepo
	for _, repo := range repos {
		if repo.Enabled {
			enabledRepos = append(enabledRepos, repo)
		}
	}

	if len(enabledRepos) != 1 {
		t.Errorf("Expected 1 enabled repo, got %d", len(enabledRepos))
	}

	if enabledRepos[0].ID != "test-repo" {
		t.Errorf("Expected ID 'test-repo', got '%s'", enabledRepos[0].ID)
	}

	if enabledRepos[0].Name != "Test Repository" {
		t.Errorf("Expected Name 'Test Repository', got '%s'", enabledRepos[0].Name)
	}

	if enabledRepos[0].Priority != 10 {
		t.Errorf("Expected Priority 10, got %d", enabledRepos[0].Priority)
	}
}

func TestGetPacmanRepos(t *testing.T) {
	// Skip if not on Arch system
	if _, err := os.Stat("/etc/pacman.conf"); os.IsNotExist(err) {
		t.Skip("Not an Arch Linux system, skipping")
	}

	repos, err := GetPacmanRepos()
	if err != nil {
		t.Fatalf("GetPacmanRepos failed: %v", err)
	}

	// Should find at least one repo on Arch systems
	if len(repos) == 0 {
		t.Error("Expected at least one Pacman repo")
	}
}

func TestGetPortageRepos(t *testing.T) {
	// Skip if not on Gentoo system
	if _, err := os.Stat("/etc/portage"); os.IsNotExist(err) {
		t.Skip("Not a Gentoo system, skipping")
	}

	repos, err := GetPortageRepos()
	if err != nil {
		t.Fatalf("GetPortageRepos failed: %v", err)
	}

	// Log results (may be empty on non-Gentoo)
	t.Logf("Found %d Portage repos", len(repos))
}
