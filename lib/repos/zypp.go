package repos

import (
	"encoding/xml"
	"fmt"
	"os"
	"slices"

	"github.com/furlongm/patchman-client/go/lib/util"
	"github.com/jedib0t/go-pretty/v6/table"
)

type ZyppXMLStream struct {
	XMLName   xml.Name  `xml:"stream"`
	ZyppRepos ZyppRepos `xml:"repo-list"`
}

type ZyppRepos struct {
	XMLName   xml.Name   `xml:"repo-list"`
	ZyppRepos []ZyppRepo `xml:"repo"`
}

type ZyppRepo struct {
	XMLName  xml.Name `xml:"repo"`
	Alias    string   `xml:"alias,attr"`
	Name     string   `xml:"name,attr"`
	Priority int      `xml:"priority,attr"`
	Enabled  bool     `xml:"enabled,attr"`
	Url      string   `xml:"url"`
}

// 'rpm' 'furlongm x86_64' 'furlongm' '99' 'http://download.opensuse.org/repositories/home:/furlongm/15.5/'

func PrintZyppRepo(repo ZyppRepo) {
	fmt.Printf("%s %s %d %s\n", repo.Alias, repo.Name, repo.Priority, repo.Url)
}

func ZyppRepoDisabled(repo ZyppRepo) bool {
	return !repo.Enabled
}

func PrintZyppReposTable(repos []ZyppRepo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Alias", "Name", "Priority", "URL"})
	for _, repo := range repos {
		t.AppendRows([]table.Row{
			{repo.Alias, repo.Name, repo.Priority, repo.Url},
		})
	}
	t.Render()
}

func GetZyppRepos() ([]ZyppRepo, error) {
	cmd, err := util.RunCommand("zypper", "-x", "lr")
	if err != nil {
		return nil, fmt.Errorf("failed to run zypper: %w", err)
	}
	if cmd == nil {
		return nil, nil // zypper not available
	}
	var zypp_xml_stream ZyppXMLStream
	xml.Unmarshal(cmd, &zypp_xml_stream)
	zypp_repos := zypp_xml_stream.ZyppRepos.ZyppRepos
	zypp_repos = slices.DeleteFunc(zypp_repos, ZyppRepoDisabled)
	return zypp_repos, nil
}

func PrintZyppRepos(repos []ZyppRepo) {
	for _, repo := range repos {
		PrintZyppRepo(repo)
	}
}
