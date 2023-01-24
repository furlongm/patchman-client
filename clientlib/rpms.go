package clientlib

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type RPM struct {
	Name    string `json:"name,omitempty"`
	Epoch   string `json:"epoch,omitempty"`
	Version string `json:"version,omitempty"`
	Release string `json:"release,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func GetInstalledRPMs() ([]RPM, error) {
    var debug = false
	cmd, err := exec.Command("rpm", "-qa", "--queryformat", "%{NAME} %{EPOCH} %{version} %{RELEASE} %{ARCH}\n").Output()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

    if debug {
    	fmt.Println("Running command: ", string(cmd))
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
