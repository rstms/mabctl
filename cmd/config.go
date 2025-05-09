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
	"os"

	"github.com/rstms/mabctl/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "output configuration",
	Long: `
Output configuration in YAML format."
`,
	Run: func(cmd *cobra.Command, args []string) {
		tempfile, err := os.CreateTemp("", "temp-config-*")
		cobra.CheckErr(err)
		defer os.Remove(tempfile.Name())
		err = api.SetDefaults()
		cobra.CheckErr(err)
		err = viper.WriteConfigAs(tempfile.Name())
		cobra.CheckErr(err)
		data, err := os.ReadFile(tempfile.Name())
		cobra.CheckErr(err)
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
