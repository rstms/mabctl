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

var addrCmd = &cobra.Command{
	Use:   "addr USERNAME BOOK_NAME EMAIL_ADDRESS",
	Short: "lookup an email address",
	Long: `
Lookup EMAIL_ADDRESS in address book BOOK_NAME of user USERNAME.
The command will fail with an error if BOOK_NAME does not existe.
The exit code is 0 if the address exists in the address book.
`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		bookname := args[1]
		email := args[2]
		response, err := MAB.QueryAddress(username, bookname, email)
		cobra.CheckErr(err)
		exitCode := 1
		if response.Address == nil {
		    exitCode = 1
		}
		if !HandleResponse(response, response.Address) {
			if !viper.GetBool("quiet") {
			    if response.Address != nil {
			    email, err := MAB.EmailAddress(*response.Address)
			    cobra.CheckErr(err)
			    fmt.Println(email)
			    }
			}
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(addrCmd)
}
