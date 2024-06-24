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
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Post struct {
	Userid string `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type Response struct {
	Count    int         `json:"count,omitempty"`
	Next     interface{} `json:"next,omitempty"`
	Previous interface{} `json:"previous,omitempty"`
	Results  []Host      `json:"results,omitempty"`
}

type Host struct {
	ID                  int           `json:"id,omitempty"`
	Hostname            string        `json:"hostname,omitempty"`
	Ipaddress           string        `json:"ipaddress,omitempty"`
	Reversedns          interface{}   `json:"reversedns,omitempty"`
	CheckDNS            bool          `json:"check_dns,omitempty"`
	Os                  string        `json:"os,omitempty"`
	Kernel              string        `json:"kernel,omitempty"`
	Arch                string        `json:"arch,omitempty"`
	Domain              string        `json:"domain,omitempty"`
	Lastreport          string        `json:"lastreport,omitempty"`
	Repos               []string      `json:"repos,omitempty"`
	Updates             []interface{} `json:"updates,omitempty"`
	RebootRequired      bool          `json:"reboot_required,omitempty"`
	HostReposOnly       bool          `json:"host_repos_only,omitempty"`
	Tags                string        `json:"tags,omitempty"`
	UpdatedAt           string        `json:"updated_at,omitempty"`
	BugfixUpdateCount   int           `json:"bugfix_update_count,omitempty"`
	SecurityUpdateCount int           `json:"security_update_count,omitempty"`
}

var BaseURL = "http://35.237.21.194/api"

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		HostURL := BaseURL + "/host/" + "?hostname=4301e54f4349"
		response, err := http.Get(HostURL)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(responseData))
		var responseObject Response
		json.Unmarshal(responseData, &responseObject)

		if responseObject.Count == 1 {
			HostObject := responseObject.Results[0]
			fmt.Println(string(HostObject.Hostname))

		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// submitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// submitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
