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
	"cmapp/internal/proto"
	"cmapp/pkg/ma"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)



// CreateMachine create machine
func CreateMachine(uuid string) error {
	var meIns ma.MEngine
	meIns = getMEngineInstance()

	engineCreateContext := ma.MEngineContext{
		Context: context.Background(),
		UUID:    uuid,
		CoreID:  0,
	}

	machine, err := meIns.InitMachine(engineCreateContext)
	if err != nil {
		return errors.Wrap(err, "init machine")
	}

	var cli = proto.NewMachineManageClient(&grpc.ClientConn{})

	pm := &proto.Machine{
		UUID:        uuid,
		State:       int32(machine.State),
		DriverID:    int32(machine.DriverID),
		MachineTags: machine.Tags,
		CustomInfo:  machine.CustomInfo,
	}
	initMachine, err := cli.ReportInitMachine(context.Background(), pm)
	if err != nil {
		return errors.Wrap(err, "report init machine")
	}

	fmt.Printf("report init machine id [%d]\n", initMachine.ID)

	engineCreateContext.CoreID = int(initMachine.ID)

	err = meIns.CreateExec(engineCreateContext)
	if err != nil {
		return errors.Wrap(err, "create execute")
	}

	err = meIns.InstallMRobot(engineCreateContext)
	if err != nil {
		return errors.Wrap(err, "install machine robot")
	}

	err = meIns.MRoHealthCheck(engineCreateContext, 10)
	if err != nil {
		return errors.Wrap(err, "machine robot health check")
	}

	engineCreateContext.Done()

	return nil
}

// TODO: 这个是一个很大的问题, 该怎么嵌入驱动
func getMEngineInstance() ma.MEngine {
	return nil
}
