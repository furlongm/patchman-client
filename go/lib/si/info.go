package si

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/zcalusic/sysinfo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	CPE       string `json:"cpe,omitempty"`
	Codename  string `json:"codename,omitempty"`
}

func GetOSInfo() OSInfo {
	if runtime.GOOS != "linux" {
		log.Fatal("Error: only linux is supported")
	}

	var si sysinfo.SysInfo
	si.GetSysInfo()

	caser := cases.Title(language.English)

	var osi = OSInfo{}
	osi.Hostname = si.Node.Hostname
	osi.MachineID = si.Node.MachineID
	osi.OSName = si.OS.Name
	osi.OSVendor = caser.String(si.OS.Vendor)
	osi.OSVersion = si.OS.Version
	osi.OSRelease = si.OS.Release
	osi.OSArch = si.Kernel.Architecture
	osi.Kernel = si.Kernel.Release

	// Parse additional fields from /etc/os-release
	osRelease := parseOSRelease()
	if cpe, ok := osRelease["CPE_NAME"]; ok {
		osi.CPE = cpe
	}

	// Build OS name from os-release for consistency with bash/python clients
	if osName := buildOSString(osRelease, osi.CPE); osName != "" {
		osi.OSName = osName
	}

	// Get codename from VERSION or VERSION_CODENAME
	if codename, ok := osRelease["VERSION_CODENAME"]; ok {
		osi.Codename = codename
	} else if version, ok := osRelease["VERSION"]; ok {
		// Extract codename from VERSION like "9.5 (Blue Onyx)" or "22.04.3 LTS (Jammy Jellyfish)"
		if start := strings.Index(version, "("); start != -1 {
			if end := strings.Index(version, ")"); end > start {
				osi.Codename = version[start+1 : end]
			}
		}
	}

	return osi
}

// parseOSRelease reads /etc/os-release and returns key-value pairs
func parseOSRelease() map[string]string {
	result := make(map[string]string)

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "="); idx != -1 {
			key := line[:idx]
			value := line[idx+1:]
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			result[key] = value
		}
	}

	return result
}

func PrintOSInfo(osi OSInfo) {
	fmt.Println("hostname   :", osi.Hostname)
	fmt.Println("machine id :", osi.MachineID)
	fmt.Println("OS name    :", osi.OSName)
	fmt.Println("OS vendor  :", osi.OSVendor)
	fmt.Println("OS version :", osi.OSVersion)
	fmt.Println("OS release :", osi.OSRelease)
	fmt.Println("OS arch    :", osi.OSArch)
	fmt.Println("Kernel     :", osi.Kernel)
	if osi.CPE != "" {
		fmt.Println("CPE        :", osi.CPE)
	}
	if osi.Codename != "" {
		fmt.Println("Codename   :", osi.Codename)
	}
}
