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
)

var booksCmd = &cobra.Command{
	Use:   "books USERNAME",
	Short: "list address books",
	Long: `
List address boooks for a user account.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		if viper.GetBool("admin") {
			response, err := MAB.GetBooksAdmin(username)
			cobra.CheckErr(err)
			if !HandleResponse(response, response.Books) {
				for _, book := range response.Books {
					fmt.Println(book.Token)
				}
			}
		} else {
			response, err := MAB.GetBooks(username)
			cobra.CheckErr(err)
			if !HandleResponse(response, response.Books) {
				names, err := response.Names()
				cobra.CheckErr(err)
				for _, name := range names {
					fmt.Println(name)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(booksCmd)
}
