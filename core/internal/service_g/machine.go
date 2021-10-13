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

package service_g

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"github.com/zibuyu28/cmapp/core/proto/ma_manager"
)

// UpdateMachineRec update machine
func UpdateMachineRec(ctx context.Context, machine *ma_manager.TypedMachine) error {
	var updateFileds []string
	m := &model.Machine{ID: int(machine.ID)}
	if machine.State != 0 {
		m.State = int(machine.State)
		updateFileds = append(updateFileds, "state")
	}
	if len(machine.MachineTags) != 0 {
		m.Tags = machine.MachineTags
		updateFileds = append(updateFileds, "tags")
	}
	if len(machine.CustomInfo) != 0 {
		m.CustomInfo = machine.CustomInfo
		updateFileds = append(updateFileds, "custom_info")
	}
	if len(machine.AGGRPCAddr) != 0 {
		m.AGGRPCAddr = machine.AGGRPCAddr
		updateFileds = append(updateFileds, "ag_grpc_addr")
	}

	err := model.UpdateMachine(m, m.ID, updateFileds)
	if err != nil {
		return errors.Wrap(err, "update machine")
	}
	return nil
}

// StoreMachineRec store machine record
func StoreMachineRec(ctx context.Context, machine *ma_manager.TypedMachine) error {
	m := &model.Machine{
		UUID:       machine.UUID,
		State:      int(machine.State),
		DriverID:   int(machine.DriverID),
		Tags:       machine.MachineTags,
		CustomInfo: machine.CustomInfo,
		AGGRPCAddr: machine.AGGRPCAddr,
	}
	err := model.InsertMachine(m)
	if err != nil {
		return errors.Wrap(err, "insert machine")
	}
	machine.ID = int32(m.ID)
	return nil
}

// RegisterMachine TODO: implement register logic
func RegisterMachine(ctx context.Context, machine *ma_manager.TypedMachine) error {
	log.Warn(ctx, "mock register, please implement me")
	return nil
}