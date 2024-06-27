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
	Name     string `json:"name,omitempty"`
	Priority string `json:"priority,omitempty"`
	Url      string `json:"url,omitempty"`
}

// 'deb' 'Debian 12.5 x86_64 repo at https://deb.debian.org/debian/dists/bookworm/main/binary-amd64' '500' 'https://deb.debian.org/debian/dists/bookworm/main/binary-amd64'

func PrintAptRepo(repo AptRepo) {
	fmt.Println("%s %s %s", repo.Name, repo.Priority, repo.Url)
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
		match, _ := regexp.MatchString("^ *[0-9]{1,5} mirror+file|http(s)?:", line)
		if !match {
			continue
		}
		v := strings.Split(line, " ")
		priority := v[1]
		//path := v[2:]
		mirrors := []string{}
		if strings.HasPrefix(v[2], "mirror+file://") {
			mirror_file := strings.Replace(v[2], "mirror+file://", "", 1)
			content, err := os.ReadFile(mirror_file)
			util.CheckError(err)
			for _, mirror := range strings.Split(string(content), "\n") {
				mirrors = append(mirrors, mirror)
			}
		}
		if strings.HasSuffix(line, "Packages") {
			dist_repo := true
		} else {
			dist_repo := false
		}
		url := v[3]
		name := fmt.Sprintf("%s %s %s repo at %s ", os_info.OSVendor, os_info.OSRelease, os_info.OSArch, url)
		apt_repos = append(apt_repos, AptRepo{
			Name:     name,
			Priority: priority,
			Url:      url,
		})
	}

	return apt_repos, nil
}

func PrintAptRepos(repos []AptRepo) {
	for _, repo := range repos {
		PrintAptRepo(repo)
	}
}
