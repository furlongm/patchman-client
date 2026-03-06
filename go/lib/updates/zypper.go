package updates

import (
	"encoding/xml"

	"github.com/furlongm/patchman-client/go/lib/util"
)

// ZypperPatchesXML represents the XML output of zypper list-patches
type ZypperPatchesXML struct {
	XMLName    xml.Name         `xml:"stream"`
	UpdateList ZypperUpdateList `xml:"update-status>update-list"`
}

type ZypperUpdateList struct {
	Updates []ZypperUpdate `xml:"update"`
}

type ZypperUpdate struct {
	Name        string `xml:"name,attr"`
	Edition     string `xml:"edition,attr"`
	Arch        string `xml:"arch,attr"`
	Category    string `xml:"category,attr"`
	Severity    string `xml:"severity,attr"`
	PkgManager  string `xml:"pkgmanager,attr"`
	Restart     string `xml:"restart,attr"`
	Interactive string `xml:"interactive,attr"`
	Status      string `xml:"status,attr"`
}

// GetZypperUpdates returns available security and bugfix updates from Zypper using XML output
func GetZypperUpdates() (secUpdates []Update, bugUpdates []Update, err error) {
	// Use XML output for robust parsing
	cmd, err := util.RunCommand("zypper", "--xmlout", "list-patches")
	if err != nil {
		return nil, nil, err
	}
	if len(cmd) == 0 {
		return nil, nil, nil
	}

	var patchesXML ZypperPatchesXML
	if err := xml.Unmarshal(cmd, &patchesXML); err != nil {
		return nil, nil, err
	}

	for _, patch := range patchesXML.UpdateList.Updates {
		// Skip already installed patches
		if patch.Status == "installed" || patch.Status == "not-needed" {
			continue
		}

		update := Update{
			Name:    patch.Name,
			Version: patch.Edition,
			Arch:    patch.Arch,
		}

		// Classify by category
		switch patch.Category {
		case "security":
			secUpdates = append(secUpdates, update)
		default:
			// bugfix, recommended, optional, feature, document
			bugUpdates = append(bugUpdates, update)
		}
	}

	return secUpdates, bugUpdates, nil
}
