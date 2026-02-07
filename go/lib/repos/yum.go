package repos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

type YumRepo struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	BaseURL  string `json:"baseurl,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

func GetYumRepos() ([]YumRepo, error) {
	repoFiles, err := filepath.Glob("/etc/yum.repos.d/*.repo")
	if err != nil {
		return nil, err
	}

	var repos []YumRepo
	for _, repoFile := range repoFiles {
		fileRepos, err := parseRepoFile(repoFile)
		if err != nil {
			continue
		}
		repos = append(repos, fileRepos...)
	}

	// Filter to enabled repos only
	var enabledRepos []YumRepo
	for _, repo := range repos {
		if repo.Enabled {
			enabledRepos = append(enabledRepos, repo)
		}
	}

	return enabledRepos, nil
}

func parseRepoFile(filename string) ([]YumRepo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var repos []YumRepo
	var currentRepo *YumRepo

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// New section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if currentRepo != nil && currentRepo.ID != "" {
				repos = append(repos, *currentRepo)
			}
			repoID := strings.Trim(line, "[]")
			currentRepo = &YumRepo{
				ID:      repoID,
				Enabled: true, // Default to enabled
			}
			continue
		}

		if currentRepo == nil {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			currentRepo.Name = value
		case "baseurl":
			currentRepo.BaseURL = value
		case "metalink", "mirrorlist":
			if currentRepo.BaseURL == "" {
				currentRepo.BaseURL = value
			}
		case "enabled":
			currentRepo.Enabled = value == "1" || strings.ToLower(value) == "true"
		case "priority":
			fmt.Sscanf(value, "%d", &currentRepo.Priority)
		}
	}

	// Don't forget the last repo
	if currentRepo != nil && currentRepo.ID != "" {
		repos = append(repos, *currentRepo)
	}

	return repos, scanner.Err()
}

func PrintYumRepo(repo YumRepo) {
	fmt.Printf("%s %s %s\n", repo.ID, repo.Name, repo.BaseURL)
}

func PrintYumReposTable(repos []YumRepo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Name", "Priority", "URL"})
	for _, repo := range repos {
		t.AppendRows([]table.Row{
			{repo.ID, repo.Name, repo.Priority, repo.BaseURL},
		})
	}
	t.Render()
}

func PrintYumRepos(repos []YumRepo) {
	for _, repo := range repos {
		PrintYumRepo(repo)
	}
}
