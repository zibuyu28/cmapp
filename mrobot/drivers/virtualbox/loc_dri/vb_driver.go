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

package loc_dri

import (
	"context"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/pkg"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc/metadata"
	"os"
	"strconv"
)

const (
	PluginEnvDriverName    = "MACHINE_PLUGIN_DRIVER_NAME"
	PluginEnvDriverVersion = "MACHINE_PLUGIN_DRIVER_VERSION"
	PluginEnvDriverID      = "MACHINE_PLUGIN_DRIVER_ID"
)

type DriverVB struct {
	pkg.BaseDriver
	ServerHost string `validate:"required"`
	ServerPort string `validate:"required"`
}

func NewDriverVB() *DriverVB {
	return &DriverVB{}
}

func (d *DriverVB) GetCreateFlags(ctx context.Context, empty *driver.Empty) (*driver.Flags, error) {
	baseFlags := &driver.Flags{Flags: d.GetFlags()}
	flags := []*driver.Flag{
		{
			Name:   "VBServerHost",
			Usage:  "virtualbox server host",
			EnvVar: "VB_SERVERHOST",
			Value:  nil,
		},
		{
			Name:   "VBServerPort",
			Usage:  "virtualbox server port",
			EnvVar: "VB_SERVERPORT",
			Value:  nil,
		},
	}
	baseFlags.Flags = append(baseFlags.Flags, flags...)
	return baseFlags, nil
}

func (d *DriverVB) SetConfigFromFlags(ctx context.Context, flags *driver.Flags) (*driver.Empty, error) {
	m := d.ConvertFlags(flags)
	d.CoreAddr = m["CoreAddr"]
	d.ImageRepository.Repository = m["Repository"]
	d.ImageRepository.StorePath = m["StorePath"]

	d.ServerHost = m["VBServerHost"]
	d.ServerPort = m["ServerPort"]

	validate := v.New()
	err := validate.Struct(d)
	if err != nil {
		return nil, errors.Wrap(err, "validate param")
	}

	driverName := os.Getenv(PluginEnvDriverName)
	if len(driverName) == 0 {
		return nil, errors.Errorf("fail to get driver name from env, please check env [%s]", PluginEnvDriverName)
	}

	driverVersion := os.Getenv(PluginEnvDriverVersion)
	if len(driverVersion) == 0 {
		return nil, errors.Errorf("fail to get driver version from env, please check env [%s]", PluginEnvDriverVersion)
	}

	driverIDStr := os.Getenv(PluginEnvDriverID)
	if len(driverIDStr) == 0 {
		return nil, errors.Errorf("fail to get driver id from env, please check env [%s]", PluginEnvDriverID)
	}

	driverID, err := strconv.Atoi(driverIDStr)
	if err != nil {
		return nil, errors.Errorf("fail to parse driver id by driverStr [%s], please check env [%s]", driverIDStr, PluginEnvDriverID)
	}

	d.DriverName = driverName
	d.DriverVersion = driverVersion
	d.DriverID = driverID
	return nil, nil
}

func (d *DriverVB) InitMachine(ctx context.Context, empty *driver.Empty) (*driver.Machine, error) {
	log.Debug(ctx, "Currently virtualbox machine plugin start to init machine")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("UUID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find uuid from metadata")
	}
	var customInfo = map[string]string{
		"server_host": d.ServerHost,
		"server_port": d.ServerPort,
	}
	return &driver.Machine{
		UUID:       datas[0],
		State:      1,
		Tags:       []string{"virtualbox"},
		CustomInfo: customInfo,
	}, nil
}

func (d *DriverVB) CreateExec(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	// 1. 使用sdk请求远程的vbox webserver 创建一个主机
	// 2. 使用远程ssh的方式，使用shell创建主机
	panic("implement me")
}

func (d *DriverVB) InstallMRobot(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	// 1. 请求远程vb webserver 安装 ha
	// TODO: 确认是否可以安装ha
	// 2. 远程shell的方式可以直接创建
	panic("implement me")
}

func (d *DriverVB) MRoHealthCheck(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	panic("implement me")
}

func (d *DriverVB) Exit(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	panic("implement me")
}
