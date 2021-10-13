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
		err := cengine.CreateChain(context.Background(), inf)
		if err != nil {
			return errors.Wrap(err, "create chain command")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.PersistentFlags().StringVarP(&createUUID, "uuid", "u", "", "create action uuid")

	createCmd.PersistentFlags().StringVarP(&driverName, "driver-name", "", "", "name of the driver which use to create chain")
	createCmd.PersistentFlags().StringVarP(&driverVersion, "driver-version", "", "", "version of the driver which use to create chain")
	createCmd.PersistentFlags().IntVarP(&driverID, "driver-id", "", 0, "id of the driver which use to create chain")

	createCmd.PersistentFlags().IntVarP(&coreGRPCPort, "core-grpc-port", "", 0, "core grpc port")
}
