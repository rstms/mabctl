/*
Copyright © 2024 Matt Krueger <mkrueger@rstms.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/rstms/mabctl/api"
	"github.com/spf13/cobra"
	"os"
)

var resetFile string

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "get user accounts",
	Long: `
Write the set of CardDAV usernames and passwords to stdout
`,
	Run: func(cmd *cobra.Command, args []string) {

		var response *api.UserAccountsResponse
		if resetFile != "" {
			// reset accounts from JSON dict
			var file *os.File
			var err error
			if resetFile == "-" {
				// read from STDIN
				file = os.Stdin
			} else {
				// read from file
				file, err = os.Open(resetFile)
				cobra.CheckErr(err)
				defer file.Close()
			}
			var accounts map[string]string
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&accounts)
			cobra.CheckErr(err)
			request := api.UserAccountsRequest{Accounts: accounts}
			response, err = MAB.SetAccounts(&request)
			cobra.CheckErr(err)
		} else {
			// don't set accounts, just get them
			var err error
			response, err = MAB.GetAccounts()
			cobra.CheckErr(err)
		}
		if !HandleResponse(response, response.Accounts) {
			for username, password := range response.Accounts {
				fmt.Printf("%s\t%s\n", username, password)
			}
		}
	},
}

func init() {
	accountsCmd.Flags().StringVarP(&resetFile, "reset", "r", "", "reset accounts from JSON dict file (- reads from stdin)")
	rootCmd.AddCommand(accountsCmd)
}
