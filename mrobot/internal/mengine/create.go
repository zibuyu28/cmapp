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
	"github.com/zibuyu28/cmapp/mrobot/internal/plugin/localbinary"
	machineproto "github.com/zibuyu28/cmapp/mrobot/proto"
	"google.golang.org/grpc"
)

// CreateMachine create machine
func CreateMachine(ctx context.Context, uuid string, corePort int, driverId string) error {

	ctx = context.WithValue(ctx, "UUID", uuid)
	ctx = context.WithValue(ctx, "CoreID", 0)

	var meIns machineproto.MachineDriverClient
	//var meIns ma.MEngine
	meIns, err := getMEngineInstance(ctx, driverId)
	if err != nil {
		log.Errorf(ctx, "Currently fail to new machine engine instance, driverId [%s]", driverId)
		return errors.Wrap(err, "fail to new machine engine instance")
	}

	machine, err := meIns.InitMachine(ctx, &machineproto.Empty{})
	if err != nil {
		return errors.Wrap(err, "init machine")
	}
	if machine.UUID != uuid {
		return errors.Errorf("machine uuid not correct expect [%s], but got [%s]", uuid, machine.UUID)
	}

	log.Debugf(ctx, "Currently init machine success, info [%+v]", machine)

	// get grpc connect
	// grpc.WithBlock() : use to make sure the connection is up
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("127.0.0.1:%d", corePort), grpc.WithInsecure(), grpc.WithBlock())
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

	ctx = context.WithValue(ctx, "CoreID", int(initMachine.ID))

	_, err = meIns.CreateExec(ctx, &machineproto.Empty{})
	if err != nil {
		return errors.Wrap(err, "create execute")
	}
	log.Debug(ctx, "Currently execute create machine action success")

	_, err = meIns.InstallMRobot(ctx, &machineproto.Empty{})
	if err != nil {
		return errors.Wrap(err, "install machine robot")
	}

	log.Debug(ctx, "Currently install machine robot success")

	_, err = meIns.MRoHealthCheck(ctx, &machineproto.Empty{})
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
func getMEngineInstance(ctx context.Context, driverID string) (machineproto.MachineDriverClient, error) {
	// 启动 plugin
	plugin, err := localbinary.NewPlugin(ctx, driverID)
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
	return machineproto.NewMachineDriverClient(conn), nil
}
