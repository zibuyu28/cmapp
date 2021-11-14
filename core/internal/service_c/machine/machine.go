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

package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	DefaultGrpcPort   int    = 9009
	DefaultDriverPath string = "drv.DriverPath"
)

const (
	MachineEngineCoreGRPCPORT  = "MACHINE_ENGINE_CORE_GRPC_PORT"
	MachineEngineDriverID      = "MACHINE_ENGINE_DRIVER_ID"
	MachineEngineDriverName    = "MACHINE_ENGINE_DRIVER_NAME"
	MachineEngineDriverVersion = "MACHINE_ENGINE_DRIVER_VERSION"
)

// Create execute driver create command to initialization machine
func Create(ctx context.Context, driverid int, param interface{}) error {
	drv, err := model.GetDriverByID(driverid)
	if err != nil {
		return errors.Wrap(err, "get driver by id")
	}
	if drv.Type != model.MachineDriver {
		return errors.Errorf("driver id [%d]'s type not machine driver", drv.ID)
	}
	marshal, err := json.Marshal(param)
	if err != nil {
		return errors.Wrap(err, "marshal param")
	}
	err = CreateAction(ctx, DefaultDriverPath, drv, machineUuid(drv.Name), string(marshal))
	if err != nil {
		return errors.Wrap(err, "create action")
	}
	return nil
}

func machineUuid(driverName string) string {
	return fmt.Sprintf("%s-%s", driverName, md5.MD5(fmt.Sprintf("%d", time.Now().UnixNano()))[:8])
}

// CreateAction driver to create machine
func CreateAction(ctx context.Context, driverRootPath string, drv *model.Driver, uuid, param string) error {
	abs, _ := filepath.Abs(filepath.Join(driverRootPath, drv.Name, drv.Version))
	binaryPath := fmt.Sprintf("%s/driver", abs)
	_, err := os.Stat(binaryPath)
	if err != nil {
		if os.ErrNotExist == err {
			return errors.Wrapf(err, "driver executable file [%s]", binaryPath)
		}
		return err
	}
	outCh := make(chan string, 10)
	defer close(outCh)
	timeout, cancelFunc := context.WithTimeout(ctx, 600*time.Second)
	defer cancelFunc()

	command := fmt.Sprintf("%s ma create -u %s -p '%s'", binaryPath, uuid, param)
	newCmd := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(map[string]string{
		MachineEngineCoreGRPCPORT:  strconv.Itoa(DefaultGrpcPort),
		MachineEngineDriverName:    drv.Name,
		MachineEngineDriverVersion: drv.Version,
		MachineEngineDriverID:      strconv.Itoa(drv.ID),
		"BASE_CORE_ADDR":           "192.168.31.63:9009",
		"BASE_IMAGE_REPOSITORY":    "harbor.hyeprchain.cn",
		"BASE_IMAGE_STORE_PATH":    "platform",
	}), cmd.WithTimeout(600), cmd.WithStream(outCh))
	go driverOutput(timeout, outCh)
	_, err = newCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "fail to execute command [%s]", command)
	}
	//log.Infof(ctx, "Currently ro create command execute result : %s", out)
	return nil
}

func driverOutput(ctx context.Context, outCh chan string) {
	for {
		select {
		case l := <-outCh:
			log.Info(ctx, l)
		case <-ctx.Done():
			return
		}
	}
}
