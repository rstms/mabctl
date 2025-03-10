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
)

var mkuserCmd = &cobra.Command{
	Use:   "mkuser EMAIL [DISPLAY_NAME] [PASSWORD]",
	Short: "add a user account",
	Long: `
Add a user account to the caldav/carddav server.  Usernames are email
addresses by convention.  The display name will be visible to the user.
`,
	Args: cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		display := ""
		if len(args) > 1 {
			display = args[1]
		}
		password := ""
		if len(args) > 2 {
			password = args[2]
		}
		response, err := MAB.AddUser(email, display, password)
		cobra.CheckErr(err)
		if !HandleResponse(response, response.User) {
			fmt.Printf("created: %s\n", response.User.UserName)
		}
	},
}

func init() {
	rootCmd.AddCommand(mkuserCmd)
}
