package packages

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
	"github.com/knqyf263/go-deb-version"
)

type DEB struct {
	Name    string `json:"name,omitempty"`
	Epoch   string `json:"epoch,omitempty"`
	Version string `json:"version,omitempty"`
	Release string `json:"release,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func PrintDEB(deb DEB) {
	output := deb.Name
	if deb.Epoch != "0" {
		output += "-" + deb.Epoch
	}
	output += "-" + deb.Version
	if deb.Release != "" {
		output += "-" + deb.Release
	}
	output += "-" + deb.Arch
	fmt.Println(output)
}

func GetInstalledDEBs() ([]DEB, error) {
	cmd, err := util.RunCommand("dpkg-query", "-W", "--showformat", "${Status}|${Package}|${Version}|${Architecture}\n")
	if err != nil {
		return nil, fmt.Errorf("failed to query dpkg database: %w", err)
	}
	if cmd == nil {
		return nil, nil // dpkg-query not available
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
			Epoch:   fmt.Sprintf("%d", d_epoch),
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
