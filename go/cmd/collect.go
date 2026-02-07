/*
Copyright © 2024 Marcus Furlong <furlongm@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"

	"github.com/furlongm/patchman-client/go/lib/packages"
	"github.com/furlongm/patchman-client/go/lib/report"
	"github.com/furlongm/patchman-client/go/lib/repos"
	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect system information, packages, and repositories",
	Long:  `Collect system information, installed packages, and configured repositories from the local system.`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose := viper.GetBool("verbose")

		osInfo := si.GetOSInfo()
		si.PrintOSInfo(osInfo)
		fmt.Println()

		// Check reboot status
		if report.CheckRebootRequired() {
			fmt.Println("Reboot required: YES")
		} else {
			fmt.Println("Reboot required: NO")
		}
		fmt.Println()

		// Collect packages (RPM, DEB, Pacman, or Portage)
		rpms, err := packages.GetInstalledRPMs()
		if err == nil && len(rpms) > 0 {
			fmt.Printf("Found %d RPM packages\n", len(rpms))
			if verbose {
				packages.PrintRPMs(rpms)
			}
		}

		debs, err := packages.GetInstalledDEBs()
		if err == nil && len(debs) > 0 {
			fmt.Printf("Found %d DEB packages\n", len(debs))
			if verbose {
				packages.PrintDEBs(debs)
			}
		}

		pacmanPkgs, err := packages.GetInstalledPacmanPkgs()
		if err == nil && len(pacmanPkgs) > 0 {
			fmt.Printf("Found %d Pacman packages\n", len(pacmanPkgs))
			if verbose {
				packages.PrintPacmanPkgs(pacmanPkgs)
			}
		}

		portagePkgs, err := packages.GetInstalledPortagePkgs()
		if err == nil && len(portagePkgs) > 0 {
			fmt.Printf("Found %d Portage packages\n", len(portagePkgs))
			if verbose {
				packages.PrintPortagePkgs(portagePkgs)
			}
		}

		// Collect repositories
		fmt.Println()

		yumRepos, err := repos.GetYumRepos()
		if err == nil && len(yumRepos) > 0 {
			fmt.Printf("Found %d YUM/DNF repositories\n", len(yumRepos))
			repos.PrintYumReposTable(yumRepos)
		}

		aptRepos, err := repos.GetAptRepos(osInfo)
		if err == nil && len(aptRepos) > 0 {
			fmt.Printf("Found %d APT repositories\n", len(aptRepos))
			repos.PrintAptReposTable(aptRepos)
		}

		zyppRepos, err := repos.GetZyppRepos()
		if err == nil && len(zyppRepos) > 0 {
			fmt.Printf("Found %d Zypper repositories\n", len(zyppRepos))
			repos.PrintZyppReposTable(zyppRepos)
		}

		pacmanRepos, err := repos.GetPacmanRepos()
		if err == nil && len(pacmanRepos) > 0 {
			fmt.Printf("Found %d Pacman repositories\n", len(pacmanRepos))
			repos.PrintPacmanReposTable(pacmanRepos)
		}

		portageRepos, err := repos.GetPortageRepos()
		if err == nil && len(portageRepos) > 0 {
			fmt.Printf("Found %d Portage repositories\n", len(portageRepos))
			repos.PrintPortageReposTable(portageRepos)
		}

		// Collect DNF modules
		dnfModules, err := packages.GetInstalledDNFModules()
		if err == nil && len(dnfModules) > 0 {
			fmt.Printf("\nFound %d DNF modules\n", len(dnfModules))
			for _, mod := range dnfModules {
				fmt.Printf("  %s:%s (%s)\n", mod.Name, mod.Stream, mod.Version)
			}
		}
	},
}

func init() {
}
