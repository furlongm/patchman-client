package repos

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/furlongm/patchman-client/go/lib/si"
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

func TestParseRepoFileMetalinkOnly(t *testing.T) {
	// Metalink-only repos (like EPEL) have no baseurl - just a metalink URL.
	// The client must capture the metalink URL so the server knows about the repo.
	content := `[epel]
name=Extra Packages for Enterprise Linux 10 - x86_64
metalink=https://mirrors.fedoraproject.org/metalink?repo=epel-10&arch=x86_64
enabled=1
gpgcheck=1

[baseos]
name=Rocky Linux 10 - BaseOS
baseurl=https://example.com/rocky/10/BaseOS/x86_64/os/
enabled=1
priority=10
`
	tmpFile, err := os.CreateTemp("", "test-metalink-*.repo")
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

	if len(repos) != 2 {
		t.Fatalf("Expected 2 repos, got %d", len(repos))
	}

	// EPEL repo - metalink only, no baseurl
	epel := repos[0]
	if epel.ID != "epel" {
		t.Errorf("Expected ID 'epel', got '%s'", epel.ID)
	}
	if epel.BaseURL == "" {
		t.Error("Metalink-only repo must have BaseURL populated from metalink")
	}
	if !strings.Contains(epel.BaseURL, "metalink") {
		t.Errorf("Expected metalink URL, got '%s'", epel.BaseURL)
	}

	// BaseOS repo - has baseurl
	baseos := repos[1]
	if baseos.ID != "baseos" {
		t.Errorf("Expected ID 'baseos', got '%s'", baseos.ID)
	}
	if baseos.BaseURL == "" {
		t.Error("BaseOS repo must have BaseURL")
	}
	if baseos.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", baseos.Priority)
	}
}

func TestParseRepoFileMirrorlistOnly(t *testing.T) {
	// Mirrorlist-only repos should also have their URL captured.
	content := `[extras]
name=Extras
mirrorlist=https://mirrors.example.com/mirrorlist?repo=extras
enabled=1
`
	tmpFile, err := os.CreateTemp("", "test-mirrorlist-*.repo")
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

	if len(repos) != 1 {
		t.Fatalf("Expected 1 repo, got %d", len(repos))
	}

	if repos[0].BaseURL == "" {
		t.Error("Mirrorlist-only repo must have BaseURL populated from mirrorlist")
	}
	if !strings.Contains(repos[0].BaseURL, "mirrorlist") {
		t.Errorf("Expected mirrorlist URL, got '%s'", repos[0].BaseURL)
	}
}

func TestParseRepoFileBaseURLNotOverwrittenByMetalink(t *testing.T) {
	// When a repo has both baseurl and metalink, baseurl should be kept.
	content := `[epel]
name=EPEL
baseurl=https://example.com/epel/10/
metalink=https://mirrors.fedoraproject.org/metalink?repo=epel-10
enabled=1
`
	tmpFile, err := os.CreateTemp("", "test-both-*.repo")
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

	if len(repos) != 1 {
		t.Fatalf("Expected 1 repo, got %d", len(repos))
	}

	// baseurl should be preserved, not overwritten by metalink
	if !strings.Contains(repos[0].BaseURL, "example.com") {
		t.Errorf("BaseURL should be the baseurl value, got '%s'", repos[0].BaseURL)
	}
}

func TestConstructAptRepoURL(t *testing.T) {
	tests := []struct {
		url        string
		components []string
		want       string
	}{
		{"http://deb.debian.org/debian", []string{"bookworm/main", "amd64", "Packages"},
			"http://deb.debian.org/debian/dists/bookworm/main/binary-amd64"},
		{"http://deb.debian.org/debian/", []string{"bookworm/main", "amd64", "Packages"},
			"http://deb.debian.org/debian/dists/bookworm/main/binary-amd64"},
		{"https://repo.example.com/flat", []string{"./", "Packages"},
			"https://repo.example.com/flat/./"},
	}
	for _, tt := range tests {
		if got := ConstructAptRepoURL(tt.url, tt.components); got != tt.want {
			t.Errorf("ConstructAptRepoURL(%q, %q) = %q, want %q", tt.url, tt.components, got, tt.want)
		}
	}
}

func TestParseAptPolicy(t *testing.T) {
	output := `Package files:
 100 /var/lib/dpkg/status
     release a=now
 500 http://deb.debian.org/debian bookworm/main amd64 Packages
     release v=12,o=Debian,a=stable,n=bookworm,l=Debian,c=main,b=amd64
     origin deb.debian.org
 500 http://deb.debian.org/debian bookworm/main Translation-en
   1 http://deb.debian.org/debian experimental/main amd64 Packages
1001 http://apt.example.com/pinned stable/main amd64 Packages
 -10 http://apt.example.com/never stable/main amd64 Packages
 500 https://repo.example.com/flat ./ Packages
Pinned packages:
`
	want := []AptRepo{
		{Priority: "500", Urls: []string{"http://deb.debian.org/debian/dists/bookworm/main/binary-amd64"}},
		{Priority: "1", Urls: []string{"http://deb.debian.org/debian/dists/experimental/main/binary-amd64"}},
		{Priority: "1001", Urls: []string{"http://apt.example.com/pinned/dists/stable/main/binary-amd64"}},
		{Priority: "-10", Urls: []string{"http://apt.example.com/never/dists/stable/main/binary-amd64"}},
		{Priority: "500", Urls: []string{"https://repo.example.com/flat/./"}},
	}
	got := parseAptPolicy(output, si.OSInfo{})
	if len(got) != len(want) {
		t.Fatalf("got %d repos, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Priority != want[i].Priority || !reflect.DeepEqual(got[i].Urls, want[i].Urls) {
			t.Errorf("repo %d: got %s %q, want %s %q", i, got[i].Priority, got[i].Urls, want[i].Priority, want[i].Urls)
		}
	}
}

func TestParseAptPolicyMirrorFile(t *testing.T) {
	dir := t.TempDir()
	mirrors := dir + "/mirrors.txt"
	if err := os.WriteFile(mirrors, []byte("http://a.example.com/debian/\nhttps://b.example.com/debian\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := dir + "/empty.txt"
	if err := os.WriteFile(empty, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := " 500 mirror+file:" + mirrors + " bookworm/main amd64 Packages\n" +
		" 500 mirror+file:" + empty + " bookworm/main amd64 Packages\n"
	got := parseAptPolicy(output, si.OSInfo{})
	want := []string{
		"http://a.example.com/debian/dists/bookworm/main/binary-amd64",
		"https://b.example.com/debian/dists/bookworm/main/binary-amd64",
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0].Urls, want) {
		t.Errorf("got %+v, want one repo with %q", got, want)
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
