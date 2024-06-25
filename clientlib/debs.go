package clientlib

import (
	"fmt"
	"github.com/knqyf263/go-deb-version"
	"log"
	"log/slog"
	"strings"
)

type DEB struct {
	Name    string `json:"name,omitempty"`
	Epoch   string `json:"epoch,omitempty"`
	Version string `json:"version,omitempty"`
	Release string `json:"release,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func PrintDEB(deb DEB) {
	if deb.Epoch == "" {
		fmt.Printf("%s-%s-%s.%s.deb\n", deb.Name, deb.Version, deb.Release, deb.Arch)
	} else {
		fmt.Printf("%s-%s:%s-%s.%s.deb\n", deb.Name, deb.Epoch, deb.Version, deb.Release, deb.Arch)
	}
}

func GetInstalledDEBs() ([]DEB, error) {
	cmd, err := RunCommand("dpkg-query", "-W", "--showformat", "${Status}|${Package}|${Version}|${Architecture}\n")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	debs := []DEB{}
	for _, line := range strings.Split(string(cmd), "\n") {
		v := strings.Split(line, "|")
		if len(v) <= 1 {
			continue
		}
		if v[0] != "install ok installed" {
			slog.Warn(fmt.Sprintf("skipping package: %s", v))
			continue
		}

		full_version, err := version.NewVersion(v[2])
		if err != nil {
			slog.Warn(fmt.Sprintf("couldn't parse package version: %s", v))
			continue
		}
		d_epoch := full_version.Epoch()
		d_version := full_version.Version()
		d_release := full_version.Revision()

		debs = append(debs, DEB{
			Name:    v[1],
			Epoch:   fmt.Sprintf("%s", d_epoch),
			Version: d_version,
			Release: d_release,
			Arch:    v[3],
		})
	}
	return debs, nil
}

func PrintDEBs(debs []DEB) {
	for _, deb := range debs {
		PrintDEB(deb)
	}
}
