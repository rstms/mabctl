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
	"github.com/spf13/viper"
	"os"
)

var restoreUser string

var restoreCmd = &cobra.Command{
	Use:   "restore [FILENAME]",
	Short: "DESTRUCTIVE restore from dump file",
	Long: `
Restore the CardDAV server config from a JSON dump file.  If FILENAME is 
provided read from the file.  If FILENAME is absent or '-' read from STDIN
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		var file *os.File
		var err error
		if filename == "" || filename == "-" {
			// read from STDIN
			file = os.Stdin
		} else {
			// read from file
			file, err = os.Open(filename)
			cobra.CheckErr(err)
			defer file.Close()
		}
		var dump api.ConfigDump
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&dump)
		cobra.CheckErr(err)

		if viper.GetBool("force") {
		    if restoreUser != "" {
			_, err := MAB.DeleteUser(restoreUser)
			cobra.CheckErr(err)
		    } else {
			_, err := MAB.Clear()
			cobra.CheckErr(err)
		    }
		}

		response, err := MAB.Restore(&dump, restoreUser)
		cobra.CheckErr(err)
		if !HandleResponse(response, response) {
			fmt.Println(response.Message)
		}
	},
}

func init() {
	restoreCmd.Flags().StringVar(&restoreUser, "user", "", "restore username")
	rootCmd.AddCommand(restoreCmd)
}
