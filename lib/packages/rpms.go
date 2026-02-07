package packages

import (
	"fmt"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
)

type RPM struct {
	Name    string `json:"name,omitempty"`
	Epoch   string `json:"epoch,omitempty"`
	Version string `json:"version,omitempty"`
	Release string `json:"release,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func PrintRPM(rpm RPM) {
	if rpm.Epoch == "" {
		fmt.Printf("%s-%s-%s.%s\n", rpm.Name, rpm.Version, rpm.Release, rpm.Arch)
	} else {
		fmt.Printf("%s-%s:%s-%s.%s\n", rpm.Name, rpm.Epoch, rpm.Version, rpm.Release, rpm.Arch)
	}
}

func GetInstalledRPMs() ([]RPM, error) {
	cmd, err := util.RunCommand("rpm", "-qa", "--queryformat", "%{NAME} %{EPOCH} %{VERSION} %{RELEASE} %{ARCH}\n")
	if err != nil {
		return nil, fmt.Errorf("failed to query RPM database: %w", err)
	}
	if cmd == nil {
		return nil, nil // rpm not available
	}
	rpms := []RPM{}
	for _, line := range strings.Split(string(cmd), "\n") {
		v := strings.Split(line, " ")
		if len(v) <= 1 {
			continue
		}
		if v[1] == "(none)" {
			v[1] = ""
		}

		rpms = append(rpms, RPM{
			Name:    v[0],
			Epoch:   v[1],
			Version: v[2],
			Release: v[3],
			Arch:    v[4],
		})
	}

	return rpms, nil
}

func PrintRPMs(rpms []RPM) {
	for _, rpm := range rpms {
		PrintRPM(rpm)
	}
}
