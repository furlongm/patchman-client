package repos

import (
	"fmt"
	"github.com/furlongm/patchman-client-go/lib/si"
	"github.com/furlongm/patchman-client-go/lib/util"
	"log"
	"os"
	"regexp"
	"strings"
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

func ConstructAptRepoURL(url string, components []string) string {
	repo_url := url
	if components[0] != "" {
		repo_url += "/dists/" + components[0]
	}
	if components[1] != "Packages" {
		repo_url += "/binary-" + components[1]
	}
	fmt.Println(repo_url)
	return repo_url
}

func GetAptRepos(os_info si.OSInfo) ([]AptRepo, error) {
	cmd, err := util.RunCommand("apt-cache", "policy")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	apt_repos := []AptRepo{}
	for _, line := range strings.Split(string(cmd), "\n") {
		if strings.Contains(line, "Translation") {
			continue
		}
		match, _ := regexp.MatchString("^ *[0-9]{1,5} mirror\\+file|http(s)?:", line)
		if !match {
			continue
		}
		v := strings.Split(line, " ")
		priority := v[1]
		repo_urls := []string{}
		if strings.HasPrefix(v[2], "mirror+file:/") {
			mirror_file := strings.Replace(v[2], "mirror+file:", "", 1)
			content, err := os.ReadFile(mirror_file)
			util.CheckError(err)
			for _, mirror := range strings.Split(string(content), "\n") {
				if mirror != "" {
					repo_url := ConstructAptRepoURL(mirror, v[3:])
					repo_urls = append(repo_urls, repo_url)
				}
			}
		} else {
			repo_url := ConstructAptRepoURL(v[2], v[3:])
			repo_urls = append(repo_urls, repo_url)
		}
		name := fmt.Sprintf("%s %s %s repo at %s", os_info.OSVendor, os_info.OSRelease, os_info.OSArch, repo_urls[0])
		apt_repos = append(apt_repos, AptRepo{
			Name:     name,
			Priority: priority,
			Urls:     repo_urls,
		})
	}

	return apt_repos, nil
}

func PrintAptRepos(repos []AptRepo) {
	for _, repo := range repos {
		PrintAptRepo(repo)
	}
}
