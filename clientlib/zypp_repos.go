package clientlib

import (
	"encoding/xml"
	"fmt"
	"log"
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
	fmt.Printf("%s %s %s %s %s \n", repo.Alias, repo.Name, repo.Enabled, repo.Priority, repo.Url)
}

func GetZyppRepos() ([]ZyppRepo, error) {
	cmd, err := RunCommand("zypper", "-x", "lr")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var zypp_xml_stream ZyppXMLStream
	xml.Unmarshal(cmd, &zypp_xml_stream)
	return zypp_xml_stream.ZyppRepos.ZyppRepos, nil
}

func PrintZyppRepos(repos []ZyppRepo) {
	for _, repo := range repos {
		PrintZyppRepo(repo)
	}
}
