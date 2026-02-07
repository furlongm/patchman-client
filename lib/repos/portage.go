package repos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// PortageRepo represents a Gentoo repository (overlay)
type PortageRepo struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	SyncType string `json:"sync_type,omitempty"`
	SyncURI  string `json:"sync_uri,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// GetPortageRepos returns configured Gentoo repositories
func GetPortageRepos() ([]PortageRepo, error) {
	// Check if Portage is available
	if _, err := os.Stat("/etc/portage"); os.IsNotExist(err) {
		return nil, nil
	}

	repos := []PortageRepo{}

	// Read repos from /etc/portage/repos.conf (file or directory)
	reposConf := "/etc/portage/repos.conf"
	stat, err := os.Stat(reposConf)
	if err == nil {
		if stat.IsDir() {
			// Directory of .conf files
			files, _ := filepath.Glob(filepath.Join(reposConf, "*.conf"))
			for _, f := range files {
				parsed, _ := parseReposConfFile(f)
				repos = append(repos, parsed...)
			}
			// Also check for files without extension
			entries, _ := os.ReadDir(reposConf)
			for _, e := range entries {
				if !e.IsDir() && !strings.HasSuffix(e.Name(), ".conf") {
					parsed, _ := parseReposConfFile(filepath.Join(reposConf, e.Name()))
					repos = append(repos, parsed...)
				}
			}
		} else {
			// Single file
			parsed, _ := parseReposConfFile(reposConf)
			repos = append(repos, parsed...)
		}
	}

	// Also check legacy /var/db/repos for overlays
	varDbRepos := "/var/db/repos"
	if entries, err := os.ReadDir(varDbRepos); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				repoPath := filepath.Join(varDbRepos, e.Name())
				// Check if it's a valid repo (has profiles/repo_name)
				repoNameFile := filepath.Join(repoPath, "profiles", "repo_name")
				if _, err := os.Stat(repoNameFile); err == nil {
					// Check if we already have this repo
					found := false
					for _, r := range repos {
						if r.Location == repoPath || r.Name == e.Name() {
							found = true
							break
						}
					}
					if !found {
						repos = append(repos, PortageRepo{
							Name:     e.Name(),
							Location: repoPath,
						})
					}
				}
			}
		}
	}

	return repos, nil
}

// parseReposConfFile parses a repos.conf file in INI format
func parseReposConfFile(path string) ([]PortageRepo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	repos := []PortageRepo{}
	var currentRepo *PortageRepo

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header [repo-name]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous repo
			if currentRepo != nil && currentRepo.Name != "DEFAULT" {
				repos = append(repos, *currentRepo)
			}
			sectionName := line[1 : len(line)-1]
			currentRepo = &PortageRepo{Name: sectionName}
			continue
		}

		if currentRepo == nil {
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "location":
			currentRepo.Location = value
		case "sync-type":
			currentRepo.SyncType = value
		case "sync-uri":
			currentRepo.SyncURI = value
		case "priority":
			fmt.Sscanf(value, "%d", &currentRepo.Priority)
		}
	}

	// Don't forget the last repo
	if currentRepo != nil && currentRepo.Name != "DEFAULT" {
		repos = append(repos, *currentRepo)
	}

	return repos, nil
}

// PrintPortageReposTable prints repos in a table format
func PrintPortageReposTable(repos []PortageRepo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Location", "Sync URI"})
	for _, repo := range repos {
		t.AppendRow(table.Row{repo.Name, repo.Location, repo.SyncURI})
	}
	t.Render()
}

// PrintPortageRepo prints a single repo
func PrintPortageRepo(repo PortageRepo) {
	fmt.Printf("%s: %s (%s)\n", repo.Name, repo.Location, repo.SyncURI)
}
