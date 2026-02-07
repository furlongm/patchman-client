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
	"os"

	"github.com/furlongm/patchman-client/go/lib/packages"
	"github.com/furlongm/patchman-client/go/lib/report"
	"github.com/furlongm/patchman-client/go/lib/repos"
	"github.com/furlongm/patchman-client/go/lib/si"
	"github.com/furlongm/patchman-client/go/lib/updates"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Collect and submit report to Patchman server",
	Long: `Collect system information, installed packages, and repositories,
then submit as a JSON report to the configured Patchman server.

The server URL and API key can be provided via:
  - Command line flags: --server, --api-key
  - Config file: /etc/patchman/patchman-client.yaml
  - Environment variables: PATCHMAN_SERVER, PATCHMAN_API_KEY`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose := viper.GetBool("verbose")
		debug := viper.GetBool("debug")
		dryRun := viper.GetBool("dry-run")
		serverURL := viper.GetString("server")
		apiKey := viper.GetString("api_key")
		tags := viper.GetString("tags")
		noVerifySSL := viper.GetBool("no_verify_ssl")
		caFile := viper.GetString("ca_file")

		// Validate required config (unless dry-run)
		if serverURL == "" && !dryRun {
			fmt.Fprintln(os.Stderr, "Error: server URL is required (--server or config file)")
			os.Exit(1)
		}

		// Collect system info
		if verbose {
			fmt.Println("Collecting system information...")
		}
		osInfo := si.GetOSInfo()
		if verbose {
			si.PrintOSInfo(osInfo)
			fmt.Println()
		}

		// Create report
		rpt := report.NewReport(osInfo)
		rpt.SetTags(tags)

		// Collect packages
		if verbose {
			fmt.Println("Collecting packages...")
		}

		// RPM packages
		rpms, err := packages.GetInstalledRPMs()
		if err == nil && len(rpms) > 0 {
			if verbose {
				fmt.Printf("  Found %d RPM packages\n", len(rpms))
			}
			rpt.AddRPMs(rpms)
		}

		// DEB packages
		debs, err := packages.GetInstalledDEBs()
		if err == nil && len(debs) > 0 {
			if verbose {
				fmt.Printf("  Found %d DEB packages\n", len(debs))
			}
			rpt.AddDEBs(debs)
		}

		// Arch Linux packages
		pacmanPkgs, err := packages.GetInstalledPacmanPkgs()
		if err == nil && len(pacmanPkgs) > 0 {
			if verbose {
				fmt.Printf("  Found %d Pacman packages\n", len(pacmanPkgs))
			}
			rpt.AddPacmanPkgs(pacmanPkgs)
		}

		// Gentoo packages
		portagePkgs, err := packages.GetInstalledPortagePkgs()
		if err == nil && len(portagePkgs) > 0 {
			if verbose {
				fmt.Printf("  Found %d Portage packages\n", len(portagePkgs))
			}
			rpt.AddPortagePkgs(portagePkgs)
		}

		// Collect repositories
		if verbose {
			fmt.Println("Collecting repositories...")
		}

		// YUM/DNF repos
		yumRepos, err := repos.GetYumRepos()
		if err == nil && len(yumRepos) > 0 {
			if verbose {
				fmt.Printf("  Found %d YUM/DNF repositories\n", len(yumRepos))
			}
			rpt.AddYumRepos(yumRepos)
		}

		// APT repos
		aptRepos, err := repos.GetAptRepos(osInfo)
		if err == nil && len(aptRepos) > 0 {
			if verbose {
				fmt.Printf("  Found %d APT repositories\n", len(aptRepos))
			}
			rpt.AddAptRepos(aptRepos)
		}

		// Zypper repos
		zyppRepos, err := repos.GetZyppRepos()
		if err == nil && len(zyppRepos) > 0 {
			if verbose {
				fmt.Printf("  Found %d Zypper repositories\n", len(zyppRepos))
			}
			rpt.AddZyppRepos(zyppRepos)
		}

		// Pacman repos
		pacmanRepos, err := repos.GetPacmanRepos()
		if err == nil && len(pacmanRepos) > 0 {
			if verbose {
				fmt.Printf("  Found %d Pacman repositories\n", len(pacmanRepos))
			}
			rpt.AddPacmanRepos(pacmanRepos)
		}

		// Portage repos
		portageRepos, err := repos.GetPortageRepos()
		if err == nil && len(portageRepos) > 0 {
			if verbose {
				fmt.Printf("  Found %d Portage repositories\n", len(portageRepos))
			}
			rpt.AddPortageRepos(portageRepos)
		}

		// DNF modules
		dnfModules, err := packages.GetInstalledDNFModules()
		if err == nil && len(dnfModules) > 0 {
			if verbose {
				fmt.Printf("  Found %d DNF modules\n", len(dnfModules))
			}
			rpt.AddDNFModules(dnfModules)
		}

		// Collect available updates
		if verbose {
			fmt.Println("Collecting available updates...")
		}

		// Try DNF/YUM updates first
		secUpdates, bugUpdates, err := updates.GetDNFUpdates()
		if err == nil && (len(secUpdates) > 0 || len(bugUpdates) > 0) {
			if len(secUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d security updates (DNF)\n", len(secUpdates))
				}
				rpt.AddSecurityUpdates(secUpdates)
			}
			if len(bugUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d bugfix updates (DNF)\n", len(bugUpdates))
				}
				rpt.AddBugfixUpdates(bugUpdates)
			}
		}

		// Try Zypper updates
		secUpdates, bugUpdates, err = updates.GetZypperUpdates()
		if err == nil && (len(secUpdates) > 0 || len(bugUpdates) > 0) {
			if len(secUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d security updates (Zypper)\n", len(secUpdates))
				}
				rpt.AddSecurityUpdates(secUpdates)
			}
			if len(bugUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d bugfix updates (Zypper)\n", len(bugUpdates))
				}
				rpt.AddBugfixUpdates(bugUpdates)
			}
		}

		// Try APT updates
		secUpdates, bugUpdates, err = updates.GetAPTUpdates()
		if err == nil && (len(secUpdates) > 0 || len(bugUpdates) > 0) {
			if len(secUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d security updates (APT)\n", len(secUpdates))
				}
				rpt.AddSecurityUpdates(secUpdates)
			}
			if len(bugUpdates) > 0 {
				if verbose {
					fmt.Printf("  Found %d bugfix updates (APT)\n", len(bugUpdates))
				}
				rpt.AddBugfixUpdates(bugUpdates)
			}
		}

		// Debug: print JSON
		if debug {
			jsonData, _ := rpt.ToJSON()
			fmt.Printf("\nReport JSON:\n%s\n\n", string(jsonData))
		}

		// Dry run: just show summary
		if dryRun {
			fmt.Println("\n=== Dry Run Summary ===")
			fmt.Printf("Hostname: %s\n", rpt.Hostname)
			fmt.Printf("OS: %s\n", rpt.OS)
			fmt.Printf("Arch: %s\n", rpt.Arch)
			fmt.Printf("Kernel: %s\n", rpt.Kernel)
			fmt.Printf("Packages: %d\n", len(rpt.Packages))
			fmt.Printf("Repos: %d\n", len(rpt.Repos))
			fmt.Printf("Modules: %d\n", len(rpt.Modules))
			fmt.Printf("Security updates: %d\n", len(rpt.SecUpdates))
			fmt.Printf("Bugfix updates: %d\n", len(rpt.BugUpdates))
			fmt.Printf("Reboot required: %v\n", rpt.RebootRequired)
			if tags != "" {
				fmt.Printf("Tags: %v\n", rpt.Tags)
			}
			if debug {
				jsonData, _ := rpt.ToJSON()
				fmt.Printf("\nFull JSON:\n%s\n", string(jsonData))
			}
			return
		}

		// Submit report
		if verbose {
			fmt.Printf("Submitting report to %s...\n", serverURL)
			fmt.Printf("  Packages: %d\n", len(rpt.Packages))
			fmt.Printf("  Repos: %d\n", len(rpt.Repos))
			fmt.Printf("  Reboot required: %v\n", rpt.RebootRequired)
			if apiKey != "" {
				fmt.Printf("  API Key: %s\n", MaskAPIKey(apiKey))
			}
		}

		resp, err := rpt.SubmitWithOptions(report.SubmitOptions{
			ServerURL:  serverURL,
			APIKey:     apiKey,
			VerifySSL:  !noVerifySSL,
			CAFile:     caFile,
			MaxRetries: 3,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error submitting report: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Success: %s\n", resp.Message)
		} else {
			fmt.Println("Report submitted successfully")
		}
	},
}

func init() {
	// Bind flags to viper for config file support
	submitCmd.Flags().Bool("dry-run", false, "collect and display report without submitting")
	viper.BindPFlag("dry-run", submitCmd.Flags().Lookup("dry-run"))
}
