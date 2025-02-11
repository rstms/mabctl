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
	"os"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan USERNAME EMAIL_ADDRESS",
	Short: "report address books containing address",
	Long: `
Scan for EMAIL_ADDRESS in all CardDAV address books under the user account
USERNAME.  Output name of each book containing EMAIL_ADDRESS.  Set exit
code 0 if at least one book contains the address.
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		email := args[1]
		exitCode := 1
		response, err := MAB.ScanAddress(username, email)
		cobra.CheckErr(err)
		if len(response.Books) > 0 {
		    exitCode = 0
		}
		if ! HandleResponse(response, response.Books) {
		    if ! viper.GetBool("quiet") {
			names, err := response.Names()
			cobra.CheckErr(err)
			for _, name := range names {
			    fmt.Println(name)
			}
		    }
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
