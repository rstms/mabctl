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
	"path/filepath"
	"strings"
)

const Version = "1.4.5"

const ProgramName = "mabctl"

var MAB *api.Controller

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mabctl",
	Short: "mabctl address book control tool library",
	Long: `
CLI toolkit for administering a baikal carddav/caldav server.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch cmd.Use {
		case "version", "config":
			return
		}
		var err error
		MAB, err = api.NewAddressBookController()
		cobra.CheckErr(err)
	},
	//PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	//Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "~/.mabctl", "config file (default is ~/.mabctl)")

	optionSwitch("insecure", "", "disable server certificate validation")
	optionSwitch("json", "j", "select JSON output")
	optionSwitch("verbose", "v", "enable diagnostic output")
	optionSwitch("quiet", "q", "suppress output")
	optionString("domain", "d", "", "CardDAV server domain")
	optionString("admin-username", "U", "admin", "baikal admin username")
	optionString("admin-password", "P", "", "baikal admin password")
	optionString("bcc-url", "", "", "bcc API URL")
	optionString("dav-url", "", "", "baikal carddav URL")
	optionString("api-key", "", "", "bcc API key")
	optionString("client-cert", "", "/etc/mabctl/mabctl.pem", "client certificate file")
	optionString("client-key", "", "/etc/mabctl/mabctl.key", "client certificate key file")

}

func viperKey(name string) string {
	return strings.Replace(name, "-", "_", -1)
}

func optionSwitch(name, flag, description string) {
	if flag == "" {
		rootCmd.PersistentFlags().Bool(name, false, description)
	} else {
		rootCmd.PersistentFlags().BoolP(name, flag, false, description)
	}

	viper.BindPFlag(viperKey(name), rootCmd.PersistentFlags().Lookup(name))
}

func optionString(name, flag, value, description string) {
	if flag == "" {
		rootCmd.PersistentFlags().String(name, value, description)
	} else {
		rootCmd.PersistentFlags().StringP(name, flag, value, description)
	}
	viper.BindPFlag(viperKey(name), rootCmd.PersistentFlags().Lookup(name))
}

func pathname(filename string) string {
	if strings.HasPrefix(filename, "~") {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		filename = filepath.Join(home, filename[1:])
	}
	return filename
}

func isFile(filename string) bool {
	_, err := os.Stat(pathname(filename))
	return !os.IsNotExist(err)
}

func initConfig() {
	basename := rootCmd.Name()
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix(basename)
	viper.AutomaticEnv()
	switch {
	case isFile(cfgFile):
		viper.SetConfigFile(pathname(cfgFile))
	case isFile("~/." + basename):
		viper.SetConfigFile(pathname("~/." + basename))
	case isFile("." + basename):
		viper.SetConfigFile("." + basename)
	default:
		// search for 'config' in directories: ~/.name, '.', '/etc/name'
		viper.SetConfigName("config")
		configDirs := []string{"~/.", ".", "/etc/"}
		for _, dir := range configDirs {
			viper.AddConfigPath(pathname(dir + basename))
		}
	}

	err := viper.ReadInConfig()
	cobra.CheckErr(err)
	file := viper.ConfigFileUsed()
	if file != "" && viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Configured from file: %v\n", file)
	}

}

func HandleResponse(response interface{}, data interface{}) bool {
	if viper.GetBool("verbose") {
		viper.Set("json", true)
		PrintResponse(response)
		return true
	}
	if viper.GetBool("json") {
		PrintResponse(data)
		return true
	}
	return false
}

func PrintMessage(response *api.Response) {
	if viper.GetBool("verbose") {
		PrintResponse(response)
	} else {
		PrintResponse(response.Message)
	}
}

func PrintResponse(response interface{}) {

	if viper.GetBool("quiet") {
		return
	}
	if viper.GetBool("json") {
		buf, err := json.MarshalIndent(response, "", "  ")
		cobra.CheckErr(err)
		fmt.Println(string(buf))
		return
	}
	fmt.Println(response)
}
