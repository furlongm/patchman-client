package si

import "strings"

// buildOSString constructs the OS name string from os-release fields,
// matching the format used by the bash and python clients.
// Uses PRETTY_NAME (preferred) or NAME, appends version if not present,
// and appends CPE suffix in square brackets.
func buildOSString(osRelease map[string]string, cpeName string) string {
	name := osRelease["PRETTY_NAME"]
	if name == "" {
		name = osRelease["NAME"]
	}
	if name == "" {
		return ""
	}

	// Append version if not already present in the name
	if ver := osRelease["VERSION_ID"]; ver != "" && !strings.Contains(name, ver) {
		name = name + " " + ver
	} else if ver := osRelease["VERSION"]; ver != "" && !strings.Contains(name, ver) {
		name = name + " " + ver
	}

	if cpeName != "" {
		name = name + " [" + cpeName + "]"
	}

	return name
}
