// Package report provides the report model and submission logic for Patchman client.
package report

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/furlongm/patchman-client/go/lib/packages"
	"github.com/furlongm/patchman-client/go/lib/repos"
	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/furlongm/patchman-client/go/lib/updates"
)

// Version is set at build time via ldflags
var Version = "dev"

// Package represents a package in the report (matches server PackageSerializer)
type Package struct {
	Name     string `json:"name"`
	Epoch    string `json:"epoch,omitempty"`
	Version  string `json:"version"`
	Release  string `json:"release,omitempty"`
	Arch     string `json:"arch"`
	Type     string `json:"type"`               // deb, rpm, arch, gentoo
	Category string `json:"category,omitempty"` // Gentoo only
	Repo     string `json:"repo,omitempty"`     // Gentoo only
}

// Repo represents a repository in the report (matches server RepoSerializer)
type Repo struct {
	Type     string   `json:"type"` // deb, rpm, arch, gentoo
	Name     string   `json:"name"`
	ID       string   `json:"id,omitempty"`
	Priority int      `json:"priority,omitempty"`
	URLs     []string `json:"urls,omitempty"`
}

// Module represents a DNF module in the report (matches server ModuleSerializer)
type Module struct {
	Name     string   `json:"name"`
	Stream   string   `json:"stream"`
	Version  string   `json:"version"`
	Context  string   `json:"context"`
	Arch     string   `json:"arch"`
	Repo     string   `json:"repo,omitempty"`
	Packages []string `json:"packages,omitempty"`
}

// Update represents a security or bugfix update (matches server UpdateSerializer)
type Update struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Repo    string `json:"repo,omitempty"`
}

// Report is the full report payload (matches server ReportUploadSerializer)
type Report struct {
	Protocol       int       `json:"protocol"`
	Hostname       string    `json:"hostname"`
	Arch           string    `json:"arch"`
	Kernel         string    `json:"kernel"`
	OS             string    `json:"os"`
	Tags           []string  `json:"tags,omitempty"`
	RebootRequired bool      `json:"reboot_required"`
	Packages       []Package `json:"packages,omitempty"`
	Repos          []Repo    `json:"repos,omitempty"`
	Modules        []Module  `json:"modules,omitempty"`
	SecUpdates     []Update  `json:"sec_updates,omitempty"`
	BugUpdates     []Update  `json:"bug_updates,omitempty"`
}

// SubmitResponse represents the API response
type SubmitResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// NewReport creates a new report with basic system info
func NewReport(osInfo si.OSInfo) *Report {
	return &Report{
		Protocol:       2,
		Hostname:       osInfo.Hostname,
		Arch:           osInfo.OSArch,
		Kernel:         osInfo.Kernel,
		OS:             osInfo.OSName,
		Tags:           []string{},
		RebootRequired: CheckRebootRequired(),
		Packages:       []Package{},
		Repos:          []Repo{},
		Modules:        []Module{},
		SecUpdates:     []Update{},
		BugUpdates:     []Update{},
	}
}

// SetTags sets the tags from a comma or space delimited string
func (r *Report) SetTags(tags string) {
	parts := strings.Fields(tags)
	if strings.Contains(tags, ",") {
		parts = strings.Split(tags, ",")
	}
	for _, tag := range parts {
		if tag = strings.TrimSpace(tag); tag != "" {
			r.Tags = append(r.Tags, tag)
		}
	}
}

// AddRPMs converts and adds RPM packages to the report
func (r *Report) AddRPMs(rpms []packages.RPM) {
	for _, rpm := range rpms {
		r.Packages = append(r.Packages, Package{
			Name:    rpm.Name,
			Epoch:   rpm.Epoch,
			Version: rpm.Version,
			Release: rpm.Release,
			Arch:    rpm.Arch,
			Type:    "rpm",
		})
	}
}

// AddDEBs converts and adds DEB packages to the report
func (r *Report) AddDEBs(debs []packages.DEB) {
	for _, deb := range debs {
		r.Packages = append(r.Packages, Package{
			Name:    deb.Name,
			Epoch:   deb.Epoch,
			Version: deb.Version,
			Release: deb.Release,
			Arch:    deb.Arch,
			Type:    "deb",
		})
	}
}

// AddYumRepos converts and adds YUM/DNF repositories to the report.
// cpeName is prepended to repo IDs for consistency with bash/python clients.
func (r *Report) AddYumRepos(yumRepos []repos.YumRepo, cpeName string) {
	for _, repo := range yumRepos {
		urls := []string{}
		if repo.BaseURL != "" {
			urls = append(urls, repo.BaseURL)
		}
		repoID := repo.ID
		if cpeName != "" {
			repoID = cpeName + "-" + repoID
		}
		// Normalize repo name: strip duplicate " - arch arch" suffix
		name := repo.Name
		archSuffix := " - " + r.Arch + " " + r.Arch
		if strings.HasSuffix(name, archSuffix) {
			name = name[:len(name)-len(archSuffix)] + " " + r.Arch
		}
		if !strings.Contains(name, r.Arch) {
			name = name + " " + r.Arch
		}
		r.Repos = append(r.Repos, Repo{
			Type:     "rpm",
			Name:     name,
			ID:       repoID,
			Priority: repo.Priority,
			URLs:     urls,
		})
	}
}

// AddAptRepos converts and adds APT repositories to the report
func (r *Report) AddAptRepos(aptRepos []repos.AptRepo) {
	for _, repo := range aptRepos {
		priority := 0
		fmt.Sscanf(repo.Priority, "%d", &priority)
		r.Repos = append(r.Repos, Repo{
			Type:     "deb",
			Name:     repo.Name,
			Priority: priority,
			URLs:     repo.Urls,
		})
	}
}

// AddZyppRepos converts and adds Zypper repositories to the report
func (r *Report) AddZyppRepos(zyppRepos []repos.ZyppRepo) {
	for _, repo := range zyppRepos {
		r.Repos = append(r.Repos, Repo{
			Type:     "rpm",
			Name:     repo.Name,
			ID:       repo.Alias,
			Priority: repo.Priority,
			URLs:     []string{repo.Url},
		})
	}
}

// AddPacmanPkgs converts and adds Arch Linux packages to the report
func (r *Report) AddPacmanPkgs(pkgs []packages.PacmanPkg) {
	for _, pkg := range pkgs {
		r.Packages = append(r.Packages, Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Arch:    pkg.Arch,
			Type:    "arch",
		})
	}
}

// AddPacmanRepos converts and adds Arch Linux repositories to the report
func (r *Report) AddPacmanRepos(pacmanRepos []repos.PacmanRepo) {
	for _, repo := range pacmanRepos {
		urls := repo.Servers
		if len(urls) > 10 {
			urls = urls[:10] // Limit to first 10 mirrors
		}
		r.Repos = append(r.Repos, Repo{
			Type: "arch",
			Name: repo.Name,
			ID:   repo.Name,
			URLs: urls,
		})
	}
}

// AddPortagePkgs converts and adds Gentoo packages to the report
func (r *Report) AddPortagePkgs(pkgs []packages.PortagePkg) {
	// Use the report's arch for Gentoo packages since they don't track arch in package name
	arch := r.Arch
	if arch == "" {
		arch = "x86_64" // fallback
	}
	for _, pkg := range pkgs {
		r.Packages = append(r.Packages, Package{
			Name:     pkg.Name,
			Version:  pkg.Version,
			Type:     "gentoo",
			Category: pkg.Category,
			Repo:     pkg.Repo,
			Arch:     arch,
		})
	}
}

// AddPortageRepos converts and adds Gentoo repositories to the report
func (r *Report) AddPortageRepos(portageRepos []repos.PortageRepo) {
	for _, repo := range portageRepos {
		urls := []string{}
		if repo.SyncURI != "" {
			urls = append(urls, repo.SyncURI)
		}
		r.Repos = append(r.Repos, Repo{
			Type:     "gentoo",
			Name:     repo.Name,
			ID:       repo.Name,
			Priority: repo.Priority,
			URLs:     urls,
		})
	}
}

// AddDNFModules converts and adds DNF modules to the report
func (r *Report) AddDNFModules(dnfModules []packages.DNFModule) {
	for _, mod := range dnfModules {
		r.Modules = append(r.Modules, Module{
			Name:     mod.Name,
			Stream:   mod.Stream,
			Version:  mod.Version,
			Context:  mod.Context,
			Arch:     mod.Arch,
			Repo:     mod.Repo,
			Packages: mod.Packages,
		})
	}
}

// AddSecurityUpdates adds security updates to the report
func (r *Report) AddSecurityUpdates(upds []updates.Update) {
	for _, upd := range upds {
		r.SecUpdates = append(r.SecUpdates, Update{
			Name:    upd.Name,
			Version: upd.Version,
			Arch:    upd.Arch,
			Repo:    upd.Repo,
		})
	}
}

// AddBugfixUpdates adds bugfix updates to the report
func (r *Report) AddBugfixUpdates(upds []updates.Update) {
	for _, upd := range upds {
		r.BugUpdates = append(r.BugUpdates, Update{
			Name:    upd.Name,
			Version: upd.Version,
			Arch:    upd.Arch,
			Repo:    upd.Repo,
		})
	}
}

// ToJSON serializes the report to JSON
func (r *Report) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// SubmitOptions contains options for submitting reports
type SubmitOptions struct {
	ServerURL  string
	APIKey     string
	VerifySSL  bool
	CAFile     string // Path to custom CA certificate file
	Timeout    time.Duration
	MaxRetries int
}

// Submit sends the report to the Patchman server
func (r *Report) Submit(serverURL, apiKey string, verifySSL bool) (*SubmitResponse, error) {
	return r.SubmitWithOptions(SubmitOptions{
		ServerURL:  serverURL,
		APIKey:     apiKey,
		VerifySSL:  verifySSL,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	})
}

// SubmitWithOptions sends the report with advanced options
func (r *Report) SubmitWithOptions(opts SubmitOptions) (*SubmitResponse, error) {
	// Build the endpoint URL
	endpoint := strings.TrimSuffix(opts.ServerURL, "/") + "/api/report/"

	// Serialize report to JSON
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize report: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{}
	if !opts.VerifySSL {
		tlsConfig.InsecureSkipVerify = true
	} else if opts.CAFile != "" {
		// Load custom CA certificate
		caCert, err := os.ReadFile(opts.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Create HTTP client
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Retry logic
	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := r.doSubmit(client, endpoint, jsonData, opts.APIKey)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// Don't retry on client errors (4xx)
		if strings.Contains(err.Error(), "status 4") {
			return resp, err
		}

		// Wait before retry with exponential backoff
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

func (r *Report) doSubmit(client *http.Client, endpoint string, jsonData []byte, apiKey string) (*SubmitResponse, error) {
	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "patchman-client-go/"+Version)
	if apiKey != "" {
		req.Header.Set("Authorization", "Api-Key "+apiKey)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var submitResp SubmitResponse
	if err := json.Unmarshal(body, &submitResp); err != nil {
		// If we can't parse JSON, include the raw response
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		errMsg := submitResp.Error
		if errMsg == "" {
			errMsg = submitResp.Detail
		}
		if errMsg == "" {
			errMsg = string(body)
		}
		return &submitResp, fmt.Errorf("server returned status %d: %s", resp.StatusCode, errMsg)
	}

	return &submitResp, nil
}

// CheckRebootRequired checks if a system reboot is required
func CheckRebootRequired() bool {
	// Debian/Ubuntu: /var/run/reboot-required
	if _, err := os.Stat("/var/run/reboot-required"); err == nil {
		return true
	}

	// Compare running vs installed kernel via /boot/vmlinuz symlink
	runningKernel, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	running := strings.TrimSpace(string(runningKernel))

	if target, err := os.Readlink("/boot/vmlinuz"); err == nil {
		installedKernel := strings.TrimPrefix(target, "vmlinuz-")
		if installedKernel != running {
			return true
		}
	}

	return false
}
