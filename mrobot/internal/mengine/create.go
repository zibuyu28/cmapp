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
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	coreproto "github.com/zibuyu28/cmapp/core/proto"
	"github.com/zibuyu28/cmapp/mrobot/pkg/plugin/localbinary"
	"github.com/zibuyu28/cmapp/mrobot/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"strconv"
	"time"
)

const (
	MachineEngineCoreGRPCPORT  = "MACHINE_ENGINE_CORE_GRPC_PORT"
	MachineEngineDriverName    = "MACHINE_ENGINE_DRIVER_NAME"
	MachineEngineDriverID      = "MACHINE_ENGINE_DRIVER_ID"
	MachineEngineDriverVersion = "MACHINE_ENGINE_DRIVER_VERSION"
)

// CreateMachine create machine
func CreateMachine(ctx context.Context, uuid string) error {
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

	grpcPortStr := os.Getenv(MachineEngineCoreGRPCPORT)
	if len(grpcPortStr) == 0 {
		return errors.Errorf("fail to get core grpc port from env, please check env [%s]", MachineEngineCoreGRPCPORT)
	}
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		return errors.Wrapf(err, "parse grpc port str [%s] to number", grpcPortStr)
	}

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

	machine, err := meIns.InitMachine(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "init machine")
	}
	machine.DriverID = int32(driverID)
	if machine.UUID != uuid {
		return errors.Errorf("machine uuid not correct expect [%s], but got [%s]", uuid, machine.UUID)
	}

	log.Debugf(ctx, "Currently init machine success, info [%+v]", machine)

	// get grpc connect
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	// grpc.WithBlock() : use to make sure the connection is up
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("127.0.0.1:%d", grpcPort), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.Wrap(err, "conn grpc")
	}

	var cli = coreproto.NewMachineManageClient(conn)

	pm := &coreproto.TypedMachine{
		UUID:        uuid,
		State:       machine.State,
		DriverID:    machine.DriverID,
		MachineTags: machine.Tags,
		CustomInfo:  machine.CustomInfo,
	}
	initMachine, err := cli.ReportInitMachine(ctx, pm)
	if err != nil {
		return errors.Wrap(err, "report init machine")
	}

	log.Debugf(ctx, "Currently report init machine id [%d]", initMachine.ID)

	ctx = metadata.AppendToOutgoingContext(ctx, "CoreID", fmt.Sprintf("%d", initMachine.ID))

	_, err = meIns.CreateExec(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "create execute")
	}
	log.Debug(ctx, "Currently execute create machine action success")

	_, err = meIns.InstallMRobot(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "install machine robot")
	}

	log.Debug(ctx, "Currently install machine robot success")

	_, err = meIns.MRoHealthCheck(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "machine robot health check")
	}

	log.Debug(ctx, "Currently install machine robot success")

	ctx.Done()

	// register machine to machine center

	log.Debug(ctx, "Currently register machine to machine center success")
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
			return
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
