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
	"github.com/rstms/mabctl/card"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cardsCmd represents the cards command
var cardsCmd = &cobra.Command{
	Use:   "cards [USERNAME]",
	Short: "list carddav cards",
	Long: `
List data from the carddav server.
`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var username string
		if len(args) > 0 {
			username = args[0]
		} else {
			username = viper.GetString("username")
		}
		password := viper.GetString("password")
		url := viper.GetString("url")
		if viper.GetBool("discover") {
			url = ""
		}
		cardClient, err := card.NewClient(username, password, url)
		cobra.CheckErr(err)
		ret, err := cardClient.List()
		cobra.CheckErr(err)
		fmt.Println(ret)
	},
}

func init() {
	rootCmd.AddCommand(cardsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cardsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cardsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
