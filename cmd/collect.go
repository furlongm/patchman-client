/*
Copyright © 2023 Marcus Furlong <furlongm@gmail.com>

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
	"github.com/furlongm/patchman-client-go/lib/os"
	"github.com/furlongm/patchman-client-go/lib/packages"
	"github.com/furlongm/patchman-client-go/lib/repos"
	"github.com/spf13/cobra"
)

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Inside subCmd Run with args: %v\n", args)

		var os_info = os.GetOSInfo()
		os.PrintOSInfo(os_info)

		var err error

		var rpms []packages.RPM
		rpms, err = packages.GetInstalledRPMs()
		if err == nil {
			packages.PrintRPMs(rpms)
		}

		var debs []packages.DEB
		debs, err = packages.GetInstalledDEBs()
		if err == nil {
			packages.PrintDEBs(debs)
		}

		var zypp_repos []repos.ZyppRepo
		zypp_repos, err = repos.GetZyppRepos()
		if err == nil {
			repos.PrintZyppRepos(zypp_repos)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// collectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// collectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
