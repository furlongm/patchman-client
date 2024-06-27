package os

import (
	"fmt"
	"github.com/zcalusic/sysinfo"
	"log"
	"runtime"
	"strings"
)

type OSInfo struct {
	Hostname  string `json:"hostname,omitempty"`
	MachineID string `json:"machine_id,omitempty"`
	OSName    string `json:"os_name,omitempty"`
	OSVendor  string `json:"os_vendor,omitempty"`
	OSVersion string `json:"os_version,omitempty"`
	OSRelease string `json:"os_release,omitempty"`
	OSArch    string `json:"os_arch,omitempty"`
	Kernel    string `json:"kernel,omitempty"`
}

func GetOSInfo() OSInfo {
	if runtime.GOOS != "linux" {
		log.Fatal("Error: only linux is supported")
	}

	var si sysinfo.SysInfo
	si.GetSysInfo()

	//	fmt.Println(si)

	var osi = OSInfo{}
	osi.Hostname = si.Node.Hostname
	osi.MachineID = si.Node.MachineID
	osi.OSName = si.OS.Name
	osi.OSVendor = strings.Title(si.OS.Vendor)
	osi.OSVersion = si.OS.Version
	osi.OSRelease = si.OS.Release
	osi.OSArch = si.Kernel.Architecture
	osi.Kernel = si.Kernel.Release

	return osi
}

func PrintOSInfo(osi OSInfo) {
	fmt.Println("hostname   :", string(osi.Hostname))
	fmt.Println("machine id :", string(osi.MachineID))
	fmt.Println("OS name    :", string(osi.OSName))
	fmt.Println("OS vendor  :", string(osi.OSVendor))
	fmt.Println("OS version :", string(osi.OSVersion))
	fmt.Println("OS release :", string(osi.OSRelease))
	fmt.Println("OS arch    :", string(osi.OSArch))
	fmt.Println("Kernel     :", string(osi.Kernel))
}
