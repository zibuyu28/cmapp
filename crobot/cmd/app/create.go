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
	"github.com/zibuyu28/cmapp/crobot/internal/cengine"
)

var (
	createUUID    string
	driverName    string
	driverVersion string
	driverID      int
	coreGRPCPort  int
	param string
)

// createCmd create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create chain",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(createUUID) == 0 {
			return errors.New("'uuid' flag is required")
		}
		inf := cengine.InitInfo{
			DriverName:    driverName,
			DriverVersion: driverVersion,
			DriverID:      driverID,
			CoreGRPCPort:  coreGRPCPort,
		}
		err := cengine.CreateChain(context.Background(),inf, createUUID, param)
		if err != nil {
			return errors.Wrap(err, "create chain command")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createUUID, "uuid", "u", "", "the create machine's uuid")
	createCmd.Flags().StringVarP(&param, "param", "p", "", "param to create action, format in json")

	createCmd.Flags().StringVarP(&driverName, "driver-name", "", "", "name of the driver which use to create chain")
	createCmd.Flags().StringVarP(&driverVersion, "driver-version", "", "", "version of the driver which use to create chain")
	createCmd.Flags().IntVarP(&driverID, "driver-id", "", 0, "id of the driver which use to create chain")

	createCmd.Flags().IntVarP(&coreGRPCPort, "core-grpc-port", "", 0, "core grpc port")
}
