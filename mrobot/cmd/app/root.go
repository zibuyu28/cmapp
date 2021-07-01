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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers"
	"github.com/zibuyu28/cmapp/mrobot/internal/plugin"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

const (
	PluginEnvDriverName = "MACHINE_PLUGIN_DRIVER_NAME"
	PluginBuildIn       = "MACHINE_PLUGIN_BUILD_IN"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mrobot",
	Short: "machine robot",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		pbi := os.Getenv(PluginBuildIn)
		if len(pbi) != 0 && pbi == "true" {
			return RunAsPlugin(args)
		}
		return errors.New("not implement")
	},
}

// RunAsPlugin run as plugin
func RunAsPlugin(args []string) error {
	driverName := os.Getenv(PluginEnvDriverName)
	if len(driverName) == 0 {
		return errors.New("set to plugin mode, driver name not found")
	}
	driver, err := drivers.ParseDriver(driverName)
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/roc.yaml)")

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

		// Search config in home directory with name "roc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("roc")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof(context.Background(),"Using config file:%s", viper.ConfigFileUsed())
	}
}
