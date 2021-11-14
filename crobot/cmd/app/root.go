/*
 * Copyright © 2021 zibuyu28
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/crobot/drivers"
	"github.com/zibuyu28/cmapp/crobot/pkg/plugin"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string


const ChainPluginBuildIn string = ""

const (
	PluginDriverName = "PLUGIN_DRIVER_NAME"
	PluginBuildIn       = "PLUGIN_BUILD_IN"
)

var (
	appNameSplit = strings.Split(filepath.Base(os.Args[0]), "/")
	appName      = appNameSplit[len(appNameSplit)-1]
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		pbi := os.Getenv(PluginBuildIn)
		if len(pbi) != 0 && pbi == "true" {
			err :=RunAsChainPlugin(args)
			if err != nil {
				log.Fatal(context.Background(), err.Error())
			}
		}
	},
}

func RunAsChainPlugin(args []string) error {
	dName := os.Getenv(PluginDriverName)
	if len(dName) == 0 {
		return errors.New("set to plugin mode, driver name not found")
	}
	driver, err := drivers.ParseDriver(dName)
	if err != nil {
		return errors.Wrap(err, "parse build driver")
	}
	plugin.RegisterDriver(*driver)
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cr.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cr" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cr")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
