package repos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// PacmanRepo represents an Arch Linux repository
type PacmanRepo struct {
	Name     string   `json:"name"`
	Servers  []string `json:"servers"`
	SigLevel string   `json:"siglevel,omitempty"`
}

// GetPacmanRepos parses /etc/pacman.conf for configured repositories
func GetPacmanRepos() ([]PacmanRepo, error) {
	confPath := "/etc/pacman.conf"
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		return nil, nil // pacman not available
	}

	file, err := os.Open(confPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	repos := []PacmanRepo{}
	var currentRepo *PacmanRepo
	serverRegex := regexp.MustCompile(`^Server\s*=\s*(.+)`)
	includeRegex := regexp.MustCompile(`^Include\s*=\s*(.+)`)
	sigLevelRegex := regexp.MustCompile(`^SigLevel\s*=\s*(.+)`)
	repoRegex := regexp.MustCompile(`^\[([^\]]+)\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for repository section
		if matches := repoRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous repo
			if currentRepo != nil && currentRepo.Name != "options" {
				repos = append(repos, *currentRepo)
			}
			currentRepo = &PacmanRepo{
				Name:    matches[1],
				Servers: []string{},
			}
			continue
		}

		if currentRepo == nil || currentRepo.Name == "options" {
			continue
		}

		// Parse Server directive
		if matches := serverRegex.FindStringSubmatch(line); len(matches) > 1 {
			server := strings.TrimSpace(matches[1])
			currentRepo.Servers = append(currentRepo.Servers, server)
			continue
		}

		// Parse Include directive (mirrorlist)
		if matches := includeRegex.FindStringSubmatch(line); len(matches) > 1 {
			mirrorPath := strings.TrimSpace(matches[1])
			mirrors, _ := parseMirrorlist(mirrorPath)
			currentRepo.Servers = append(currentRepo.Servers, mirrors...)
			continue
		}

		// Parse SigLevel
		if matches := sigLevelRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentRepo.SigLevel = strings.TrimSpace(matches[1])
		}
	}

	// Don't forget the last repo
	if currentRepo != nil && currentRepo.Name != "options" {
		repos = append(repos, *currentRepo)
	}

	return repos, nil
}

// parseMirrorlist reads a mirrorlist file and returns uncommented Server lines
func parseMirrorlist(path string) ([]string, error) {
	// Handle glob patterns like /etc/pacman.d/mirrorlist
	matches, err := filepath.Glob(path)
	if err != nil || len(matches) == 0 {
		return nil, err
	}

	servers := []string{}
	for _, filePath := range matches {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		serverRegex := regexp.MustCompile(`^Server\s*=\s*(.+)`)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			if matches := serverRegex.FindStringSubmatch(line); len(matches) > 1 {
				servers = append(servers, strings.TrimSpace(matches[1]))
			}
		}
		file.Close()
	}

	return servers, nil
}

// PrintPacmanReposTable prints repos in a table format
func PrintPacmanReposTable(repos []PacmanRepo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Servers"})
	for _, repo := range repos {
		serverStr := ""
		if len(repo.Servers) > 0 {
			serverStr = repo.Servers[0]
			if len(repo.Servers) > 1 {
				serverStr += fmt.Sprintf(" (+%d more)", len(repo.Servers)-1)
			}
		}
		t.AppendRow(table.Row{repo.Name, serverStr})
	}
	t.Render()
}

// PrintPacmanRepo prints a single repo
func PrintPacmanRepo(repo PacmanRepo) {
	fmt.Printf("%s: %v\n", repo.Name, repo.Servers)
}
