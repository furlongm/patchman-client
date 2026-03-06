package packages

import (
	"regexp"
	"strings"

	"github.com/furlongm/patchman-client/go/lib/util"
)

// DNFModule represents an installed DNF module
type DNFModule struct {
	Name     string   `json:"name"`
	Stream   string   `json:"stream"`
	Version  string   `json:"version"`
	Context  string   `json:"context"`
	Arch     string   `json:"arch"`
	Repo     string   `json:"repo"`
	Packages []string `json:"packages,omitempty"`
}

// GetInstalledDNFModules returns installed DNF modules
func GetInstalledDNFModules() ([]DNFModule, error) {
	// Check if dnf is available
	cmd, err := util.RunCommand("dnf", "module", "list", "--installed")
	if err != nil {
		return nil, err
	}
	if len(cmd) == 0 {
		return nil, nil
	}

	modules := []DNFModule{}

	// Parse the tabular output
	// Format: Name Stream Profiles Summary
	// Lines start with module name (not indented)
	lines := strings.Split(string(cmd), "\n")

	// Find the data section (after header line)
	inData := false
	moduleRegex := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and hints
		if line == "" || strings.HasPrefix(line, "Hint:") || strings.HasPrefix(line, "Last metadata") {
			continue
		}

		// Detect header by looking for "Name" column
		if strings.HasPrefix(line, "Name") && strings.Contains(line, "Stream") {
			inData = true
			continue
		}

		if !inData {
			continue
		}

		// Parse module line
		matches := moduleRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			name := matches[1]
			stream := matches[2]

			// Get detailed info for this module
			module, err := getModuleInfo(name, stream)
			if err == nil && module != nil {
				modules = append(modules, *module)
			} else {
				// Basic info only
				modules = append(modules, DNFModule{
					Name:   name,
					Stream: stream,
				})
			}
		}
	}

	return modules, nil
}

// getModuleInfo gets detailed info for a specific module
func getModuleInfo(name, stream string) (*DNFModule, error) {
	cmd, err := util.RunCommand("dnf", "module", "info", name+":"+stream)
	if err != nil {
		return nil, err
	}
	if cmd == nil {
		return nil, nil
	}

	module := &DNFModule{
		Name:   name,
		Stream: stream,
	}

	lines := strings.Split(string(cmd), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Version") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				module.Version = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Context") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				module.Context = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Architecture") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				module.Arch = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Repo") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				module.Repo = strings.TrimSpace(parts[1])
			}
		}
	}

	return module, nil
}
