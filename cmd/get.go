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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var getCmd = &cobra.Command{
	Use:   "get USERNAME BOOKNAME EMAIL",
	Short: "lookup address in address book",
	Long: `
Search for address in CardDAV address book BOOKNAME under the user account
USERNAME. Output matching records. Set exit code 0 if at least one record exists.
`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		bookname := args[1]
		email := args[2]
		response, err := MAB.QueryAddress(username, bookname, email)
		cobra.CheckErr(err)
		if !HandleResponse(response, response.Addresses) {
			if !viper.GetBool("quiet") {
				for _, addr := range response.Addresses {
					fmt.Println(addr.Path)
				}
			}
		}
		exitCode := 1
		if len(response.Addresses) > 0 {
			exitCode = 0
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
