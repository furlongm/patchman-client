package repos

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/furlongm/patchman-client/go/lib/util"
	"github.com/jedib0t/go-pretty/v6/table"
)

type AptRepo struct {
	Name     string   `json:"name,omitempty"`
	Priority string   `json:"priority,omitempty"`
	Urls     []string `json:"url,omitempty"`
}

// 'deb' 'Debian 12.5 x86_64 repo at https://deb.debian.org/debian/dists/bookworm/main/binary-amd64' '500' 'https://deb.debian.org/debian/dists/bookworm/main/binary-amd64'

func PrintAptRepo(repo AptRepo) {
	fmt.Println(repo.Name, repo.Priority, repo.Urls)
}

func PrintAptReposTable(repos []AptRepo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Priority", "URL"})
	for _, repo := range repos {
		t.AppendRows([]table.Row{
			{repo.Name, repo.Priority, repo.Urls},
		})
	}
	t.Render()
}

func ConstructAptRepoURL(url string, components []string) string {
	repo_url := strings.TrimRight(url, "/")
	if len(components) < 2 || components[1] == "Packages" {
		// flat repo, e.g. "./ Packages"
		if len(components) > 0 {
			repo_url += "/" + components[0]
		}
		return repo_url
	}
	return repo_url + "/dists/" + components[0] + "/binary-" + components[1]
}

var aptRepoLineRe = regexp.MustCompile(`^\s*-?[0-9]+\s+(mirror\+file:|https?:)`)

func GetAptRepos(os_info si.OSInfo) ([]AptRepo, error) {
	cmd, err := util.RunCommand("apt-cache", "policy")
	if err != nil {
		return nil, fmt.Errorf("failed to run apt-cache policy: %w", err)
	}
	if cmd == nil {
		return nil, nil // apt-cache not available
	}
	return parseAptPolicy(string(cmd), os_info), nil
}

func parseAptPolicy(output string, os_info si.OSInfo) []AptRepo {
	apt_repos := []AptRepo{}
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Translation") {
			continue
		}
		if !aptRepoLineRe.MatchString(line) {
			continue
		}
		// priority is right-aligned, so the number of leading spaces varies
		v := strings.Fields(line)
		if len(v) < 4 {
			continue
		}
		priority := v[0]
		repo_urls := []string{}
		if strings.HasPrefix(v[1], "mirror+file:/") {
			mirror_file := strings.Replace(v[1], "mirror+file:", "", 1)
			content, err := os.ReadFile(mirror_file)
			if err != nil {
				slog.Warn("failed to read mirror file", "file", mirror_file, "error", err)
				continue
			}
			for _, mirror := range strings.Split(string(content), "\n") {
				if mirror != "" {
					repo_url := ConstructAptRepoURL(mirror, v[2:])
					repo_urls = append(repo_urls, repo_url)
				}
			}
			if len(repo_urls) == 0 {
				continue
			}
		} else {
			repo_url := ConstructAptRepoURL(v[1], v[2:])
			repo_urls = append(repo_urls, repo_url)
		}
		name := fmt.Sprintf("%s %s %s repo at %s", os_info.OSVendor, os_info.OSRelease, os_info.OSArch, repo_urls[0])
		apt_repos = append(apt_repos, AptRepo{
			Name:     name,
			Priority: priority,
			Urls:     repo_urls,
		})
	}
	return apt_repos
}

func PrintAptRepos(repos []AptRepo) {
	for _, repo := range repos {
		PrintAptRepo(repo)
	}
}
