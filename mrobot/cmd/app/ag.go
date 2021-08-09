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
	"github.com/zibuyu28/cmapp/mrobot/drivers"
	agfw "github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/worker"
	"os"
)

const (
	AgentPluginName    = "AGENT_PLUGIN_DRIVER_NAME"
	AgentPluginBuildIn = "AGENT_PLUGIN_BUILD_IN"
)

// agCmd represents the ag command
var agCmd = &cobra.Command{
	Use:   "ag",
	Short: "ag machine agent",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		pbi := os.Getenv(AgentPluginBuildIn)
		if len(pbi) != 0 && pbi == "true" {
			return RunAsAgentPlugin(args)
		}

		fmt.Println("this is ag server")
		panic("implement me")

		return nil
	},
}

// RunAsAgentPlugin run as agent plugin
func RunAsAgentPlugin(args []string) error {
	pluginName := os.Getenv(AgentPluginName)
	if len(pluginName) == 0 {
		return errors.New("set to remote plugin mode, plugin name not found")
	}
	parsePlugin, err := drivers.ParsePlugin(pluginName)
	if err != nil {
		return errors.Wrap(err, "parse build in plugin")
	}
	agfw.RegisterWorker(parsePlugin.GrpcPluginServer)
	agfw.Start(context.Background())
	return nil
}

func init() {
	rootCmd.AddCommand(agCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// maCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// maCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
