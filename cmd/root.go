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
	"os"
	"path/filepath"
	"github.com/rstms/mabctl/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const Version = "1.1.0"

var cfgFile string
var writeConfig bool
var adminClient *api.Client

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mabctl",
	Short: "mail address book control",
	Long: `
Administration tool for a baikal carddav/caldav server.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		if ! writeConfig {
			a, err := api.NewClient(
				viper.GetString("admin_username"),
				viper.GetString("admin_password"),
				viper.GetString("admin_url"),
				viper.GetString("cert"),
				viper.GetString("key"),
				viper.GetString("api_key"),
				viper.GetBool("insecure"),
			)
			cobra.CheckErr(err)
			adminClient = a
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
	    if writeConfig {
		writeConfigFile()
	    }
	},
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
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mabctl)")

	rootCmd.PersistentFlags().BoolVar(&writeConfig, "initconfig", false, "write default config to $HOME/.mabctl")

	rootCmd.PersistentFlags().BoolP("terse", "t", false, "output text instead of JSON)")
	viper.BindPFlag("terse", rootCmd.PersistentFlags().Lookup("terse"))

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "output full response")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	rootCmd.PersistentFlags().String("username", "", "username or email address")
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	rootCmd.PersistentFlags().String("password", "", "carddav account password")
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))

	rootCmd.PersistentFlags().String("cert", filepath.Join(home, "certs", "client.pem"), "client certificate PEM file")
	viper.BindPFlag("cert", rootCmd.PersistentFlags().Lookup("cert"))
	rootCmd.PersistentFlags().String("key", filepath.Join(home, "certs", "client.key"), "client certificate private key PEM file")
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	rootCmd.PersistentFlags().String("url", "", "carddav server URL")
	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))
	rootCmd.PersistentFlags().String("admin-username", "admin", "baikalctl admin username")
	viper.BindPFlag("admin_username", rootCmd.PersistentFlags().Lookup("admin-username"))
	rootCmd.PersistentFlags().String("admin-password", "", "baikalctl admin password")
	viper.BindPFlag("admin_password", rootCmd.PersistentFlags().Lookup("admin-password"))
	rootCmd.PersistentFlags().String("admin-url", "", "baikalctl admin URL")
	viper.BindPFlag("admin_url", rootCmd.PersistentFlags().Lookup("admin-url"))
	rootCmd.PersistentFlags().Bool("insecure", false, "accept host certificate without validation")
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure"))
	rootCmd.PersistentFlags().String("api-key", "", "baikalctl api key")
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
}

func setupConfigFileSearch() {
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		_, err := os.Stat(cfgFile)
		if !os.IsNotExist(err) {
			// --config set, and is a file, so read the file and don't search
			viper.SetConfigFile(cfgFile)
			return
		}

		// --config set, but not a file, so use it as the search name
		viper.SetConfigName(cfgFile)

	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		_, err = os.Stat(filepath.Join(home, ".mabctl"))
		if !os.IsNotExist(err) {
			// --config unset and ~/.mabctl exists, so use it and don't search
			viper.SetConfigFile(filepath.Join(home, ".mabctl"))
			return
		}

		// --config unset, so use the default search name
		viper.SetConfigName("config")
	}

	// setup config search dirs
	user, err := os.UserConfigDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(user, "mabctl"))
	viper.AddConfigPath("/etc/mabctl")
}

func initConfig() {

	setupConfigFileSearch()

	viper.SetEnvPrefix("mabctl")
	viper.AutomaticEnv() // read in environment variables that match

	if !writeConfig {
		err := viper.ReadInConfig()
		cobra.CheckErr(err)

		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Configured from file: %v\n", viper.ConfigFileUsed())
		}
	}
}

func writeConfigFile() {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgFile = filepath.Join(home, ".mabctl")
		_, err = os.Stat(cfgFile)
		if !os.IsNotExist(err) {
			cobra.CheckErr(fmt.Errorf("not overwriting existing config file: %v", cfgFile))
		}
	}
	err := viper.WriteConfigAs(cfgFile)
	cobra.CheckErr(err)
	fmt.Fprintf(os.Stderr, "Default config written to %v\n", cfgFile)
	os.Exit(0)
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

	var out string
	if viper.GetBool("terse") {
		out = fmt.Sprintf("%v", response)
	} else {
		buf, err := json.MarshalIndent(response, "", "  ")
		cobra.CheckErr(err)
		out = string(buf)
	}
	fmt.Println(out)
}
