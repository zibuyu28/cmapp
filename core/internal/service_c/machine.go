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

package service_c

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	DefaultGrpcPort int = 9009
)

const (
	MachineEngineCoreGRPCPORT = "MACHINE_ENGINE_CORE_GRPC_PORT"
	MachineEngineDriverID     = "MACHINE_ENGINE_DRIVER_ID"
	MachineEngineDriverName   = "MACHINE_ENGINE_DRIVER_NAME"
)

// Create execute driver create command to initialization machine
func Create(ctx context.Context, driverid int) error {
	drv, err := model.GetDriverByID(driverid)
	if err != nil {
		return errors.Wrap(err, "get driver by id")
	}
	err = CreateAction(ctx, "drv.DriverPath", drv.Name, driverid, uuid.New().String())
	if err != nil {
		return errors.Wrap(err, "create aciton")
	}
	return nil
}

// CreateAction driver to create machine
func CreateAction(ctx context.Context, driverRootPath, driverName string, driverId int, args ...string) error {
	abs, _ := filepath.Abs(filepath.Join(driverRootPath, driverName))
	binaryPath := fmt.Sprintf("%s/exec/driver", abs)
	_, err := os.Stat(binaryPath)
	if err != nil {
		if os.ErrNotExist == err {
			return errors.Wrapf(err, "driver executable file [%s]", binaryPath)
		}
		return err
	}
	command := fmt.Sprintf("%s ro create", binaryPath)
	newCmd := cmd.NewDefaultCMD(command, args, cmd.WithEnvs(map[string]string{
		MachineEngineCoreGRPCPORT: strconv.Itoa(DefaultGrpcPort),
		MachineEngineDriverName:   driverName,
		MachineEngineDriverID:     strconv.Itoa(driverId),
	}), cmd.WithTimeout(300))
	out, err := newCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "fail to execute command [%s]", command)
	}
	log.Infof(ctx, "Currently ro create command execute result : %s", out)
	return nil
}

type RMD struct {
	repo sync.Map
}

type clientIns struct {

}

var RMDIns = RMD{repo: sync.Map{}}

func (R *RMD) NewApp(in *ag.NewAppReq) (*ag.App, error) {
	// in.MachineID, save to repo

	panic("implement me")
}

func (R *RMD) StartApp(in *ag.App) error {
	panic("implement me")
}

func (R *RMD) StopApp(in *ag.App) error {
	panic("implement me")
}

func (R *RMD) DestroyApp(in *ag.App) error {
	panic("implement me")
}

func (R *RMD) TagEx(appUUID string, in *ag.Tag) error {
	panic("implement me")
}

func (R *RMD) FileMountEx(appUUID string, in *ag.FileMount) error {
	panic("implement me")
}

func (R *RMD) EnvEx(appUUID string, in *ag.EnvVar) error {
	panic("implement me")
}

func (R *RMD) NetworkEx(appUUID string, in *ag.Network) error {
	panic("implement me")
}

func (R *RMD) FilePremiseEx(appUUID string, in *ag.File) error {
	panic("implement me")
}

func (R *RMD) LimitEx(appUUID string, in *ag.Limit) error {
	panic("implement me")
}

func (R *RMD) HealthEx(appUUID string, in *ag.Health) error {
	panic("implement me")
}

func (R *RMD) LogEx(appUUID string, in *ag.Log) error {
	panic("implement me")
}
