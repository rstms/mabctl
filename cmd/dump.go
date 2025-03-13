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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dumpUser string

var dumpCmd = &cobra.Command{
	Use:   "dump [USERNAME]",
	Short: "dump CardDAV config",
	Long: `
Output all cardDAV data for USERNAME.  If USERNAME is not specified, output data for all users.
`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		user := ""
		if len(args) > 0 {
		    user = args[0]
		}
		if dumpUser != "" {
		    user = dumpUser
		}
		response, err := MAB.Dump(user)
		cobra.CheckErr(err)

		if !HandleResponse(response, response.Dump) {
			viper.Set("json", true)
			PrintResponse(response.Dump)
		}
	},
}

func init() {
	dumpCmd.Flags().StringVar(&dumpUser, "user", "", "dump username")
	rootCmd.AddCommand(dumpCmd)
}
