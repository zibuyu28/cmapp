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

package mengine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/plugin/localbinary"
	coreproto "github.com/zibuyu28/cmapp/core/proto/ma_manager"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	MachineEngineCoreGRPCPORT  = "MACHINE_ENGINE_CORE_GRPC_PORT"
	MachineEngineCoreGRPCAddr  = "MACHINE_ENGINE_CORE_GRPC_ADDR"
	MachineEngineCoreHttpAddr = "MACHINE_ENGINE_CORE_HTTP_ADDR"
	MachineEngineDriverName    = "MACHINE_ENGINE_DRIVER_NAME"
	MachineEngineDriverID      = "MACHINE_ENGINE_DRIVER_ID"
	MachineEngineDriverVersion = "MACHINE_ENGINE_DRIVER_VERSION"
)

const (
	BaseCoreHTTPAddr = "CoreHTTPAddr"
	BaseCoreGRPCAddr = "CoreGRPCAddr"
	BaseCoreAddr = "CoreAddr"
	BaseRepository = "Repository"
	BaseStorePath = "StorePath"
)

// CreateMachine create machine
func CreateMachine(ctx context.Context, uuid, param string) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	log.Debugf(ctx, "Currently create machine logic, uuid [%v]", uuid)

	driverName := os.Getenv(MachineEngineDriverName)
	if len(driverName) == 0 {
		return errors.Errorf("fail to get driver name from env, please check env [%s]", MachineEngineDriverName)
	}

	driverVersion := os.Getenv(MachineEngineDriverVersion)
	if len(driverVersion) == 0 {
		return errors.Errorf("fail to get driver version from env, please check env [%s]", MachineEngineDriverVersion)
	}

	driverIDStr := os.Getenv(MachineEngineDriverID)
	if len(driverIDStr) == 0 {
		return errors.Errorf("fail to get driver id from env, please check env [%s]", MachineEngineDriverID)
	}

	driverID, err := strconv.Atoi(driverIDStr)
	if err != nil {
		return errors.Errorf("fail to parse driver id by driverStr [%s], please check env [%s]", driverIDStr, MachineEngineDriverID)
	}

	grpcAddr := os.Getenv(MachineEngineCoreGRPCAddr)
	if len(grpcAddr) == 0 {
		return errors.Errorf("fail to get core grpc addr from env, please check env [%s]", MachineEngineCoreGRPCAddr)
	}
	log.Debugf(ctx, "get core grpc addr [%s]", grpcAddr)

	httpAddr := os.Getenv(MachineEngineCoreHttpAddr)
	if len(httpAddr) == 0 {
		return errors.Errorf("fail to get core http addr from env, please check env [%s]", MachineEngineCoreHttpAddr)
	}
	log.Debugf(ctx, "get core http addr [%s]", httpAddr)

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"UUID": uuid,
	}))

	var meIns driver.MachineDriverClient
	//var meIns ma.MEngine
	meIns, err = getMEnginePluginInstance(ctx, driverID, driverName, driverVersion)
	if err != nil {
		log.Errorf(ctx, "Currently fail to new machine engine instance, driverName [%s]", driverName)
		return errors.Wrap(err, "fail to new machine engine instance")
	}
	flags, err := meIns.GetCreateFlags(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "get create flags")
	}
	var p = make(map[string]string)
	err = json.Unmarshal([]byte(param), &p)
	if err != nil {
		return errors.Wrap(err, "param not in json format")
	}
	p[BaseCoreHTTPAddr] = httpAddr
	p[BaseCoreGRPCAddr] = grpcAddr
	p[BaseCoreAddr] = httpAddr
	p[BaseRepository] = "" // TODO: check need
	p[BaseStorePath] = "" // TODO: check need

	for i, flag := range flags.Flags {
		if v, ok := p[flag.Name]; ok {
			flags.Flags[i].Value = []string{v}
		}
	}
	_, err = meIns.SetConfigFromFlags(ctx, flags)
	if err != nil {
		return errors.Wrap(err, "set config flags")
	}

	machine, err := meIns.InitMachine(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "init machine")
	}
	machine.DriverID = int32(driverID)
	if machine.UUID != uuid {
		return errors.Errorf("machine uuid not correct expect [%s], but got [%s]", uuid, machine.UUID)
	}

	log.Debugf(ctx, "Currently init machine success, info [%+v]", machine)

	// get grpc connectUUID
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(600))
	defer cancel()
	// grpc.WithBlock() : use to make sure the connection is up
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.Wrap(err, "conn core grpc")
	}

	var cli = coreproto.NewMachineManageClient(conn)

	pm := &coreproto.TypedMachine{
		UUID:        uuid,
		State:       machine.State,
		DriverID:    machine.DriverID,
		MachineTags: machine.Tags,
		CustomInfo:  machine.CustomInfo,
		AGGRPCAddr:  "",
	}
	initMachine, err := cli.ReportInitMachine(ctx, pm)
	if err != nil {
		return errors.Wrap(err, "report init machine")
	}

	log.Debugf(ctx, "Currently report init machine id [%d]", initMachine.ID)

	ctx = metadata.AppendToOutgoingContext(ctx, "CoreMachineID", fmt.Sprintf("%d", initMachine.ID))

	ma, err := meIns.CreateExec(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "create execute")
	}
	log.Debug(ctx, "Currently execute create machine action success")
	err = MachineUpdate(ctx, cli, ma, initMachine)
	if err != nil {
		return errors.Wrap(err, "update machine info")
	}

	ma, err = meIns.InstallMRobot(ctx, ma)
	if err != nil {
		return errors.Wrap(err, "install machine robot")
	}

	log.Debug(ctx, "Currently install machine robot success")
	err = MachineUpdate(ctx, cli, ma, initMachine)
	if err != nil {
		return errors.Wrap(err, "update machine info")
	}

	for range [5][0]int{} {
		_, err = meIns.MRoHealthCheck(ctx, ma)
		if err != nil {
			log.Errorf(ctx, "check health [%v]", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	ma, err = meIns.MRoHealthCheck(ctx, ma)
	if err != nil {
		return errors.Wrap(err, "machine robot health check")
	}

	log.Debug(ctx, "Currently install machine robot success")
	err = MachineUpdate(ctx, cli, ma, initMachine)
	if err != nil {
		return errors.Wrap(err, "update machine info")
	}

	// TODO: need stop plugin server

	// register machine to machine center

	log.Debug(ctx, "Currently register machine to machine center success")
	return nil
}

// MachineUpdate machine info update
func MachineUpdate(ctx context.Context, cli coreproto.MachineManageClient, ma *driver.Machine, tma *coreproto.TypedMachine) error {
	if ma == nil {
		log.Debug(ctx, "machine obj is nil")
		return nil
	}
	var nu bool
	if ma.State != 0 {
		tma.State = ma.State
		nu = true
	}
	if len(ma.Tags) != 0 {
		ma.Tags = append(ma.Tags, tma.MachineTags...)
		var m = make(map[string]struct{})
		var tags []string
		for _, tag := range ma.Tags {
			if _, ok := m[tag]; ok {
				continue
			}
			tags = append(tags, tag)
			m[tag] = struct{}{}
		}
		tma.MachineTags = tags
		nu = true
	}
	if len(ma.CustomInfo) != 0 {
		for k, v := range ma.CustomInfo {
			if ev, ok := tma.CustomInfo[k]; ok && ev == v {
				continue
			}
			tma.CustomInfo[k] = v
			nu = true
		}
	}
	if len(ma.AGGRPCAddr) != 0 && ma.AGGRPCAddr != tma.AGGRPCAddr {
		tma.AGGRPCAddr = ma.AGGRPCAddr
		nu = true
	}
	if nu {
		if tma.ID == 0 {
			return errors.New("typed machine id is nil")
		}
		_, err := cli.UpdateMachine(ctx, tma)
		if err != nil {
			return errors.Wrap(err, "update machine")
		}
	}
	return nil
}

// TODODone: 这个是一个很大的问题, 该怎么嵌入驱动 ----> 使用grpc嵌入
func getMEnginePluginInstance(ctx context.Context, driverID int, driverName, driverVersion string) (driver.MachineDriverClient, error) {
	// 启动 plugin
	plugin, err := localbinary.NewPlugin(ctx, driverID, driverName, driverVersion)
	if err != nil {
		return nil, errors.Wrap(err, "new plugin")
	}
	go func() {
		if err = plugin.Serve(); err != nil {
			// TODO: Is this best approach?
			log.Warn(ctx, err.Error())
			//return nil, errors.Wrap(err, "plugin serve")
		}
	}()

	address, err := plugin.Address()
	if err != nil {
		return nil, errors.Wrap(err, "get plugin serve address")
	}

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, "create grpc connection")
	}
	return driver.NewMachineDriverClient(conn), nil
}
