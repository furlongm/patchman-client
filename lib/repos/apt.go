package repos

import (
	"fmt"
	"github.com/furlongm/patchman-client-go/lib/os"
	"github.com/furlongm/patchman-client-go/lib/util"
	"log"
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
	fmt.Printf("%s %s %s %s %s \n", repo.Name, repo.Priority, repo.Url)
}

func GetAptRepos(os_info os.OSInfo) ([]AptRepo, error) {
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
		if len(v) <= 4 {
			continue
		}
		name := fmt.Sprintf("%s %s %s repo at %s ", os_info.OSVendor, os_info.OSRelease, os_info.OSArch)
		name += v[3]
		apt_repos = append(apt_repos, AptRepo{
			Name:     name,
			Priority: v[1],
			Url:      v[3],
		})
	}

	return apt_repos, nil
}

func PrintAptRepos(repos []AptRepo) {
	for _, repo := range repos {
		PrintAptRepo(repo)
	}
}
