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

package api_g

import (
	"context"
	"github.com/zibuyu28/cmapp/core/internal/service_g"
	"github.com/zibuyu28/cmapp/core/proto"
)

type CoreMachineManager struct {
}

// RegisterMachine register machine to center
func (m *CoreMachineManager) RegisterMachine(ctx context.Context, machine *proto.TypedMachine) (*proto.RegisterMachineRes, error) {
	err := service_g.RegisterMachine(ctx, machine)
	if err != nil {
		return nil, err
	}
	return &proto.RegisterMachineRes{Res: true}, nil
}

// ReportInitMachine report init machine
func (m *CoreMachineManager) ReportInitMachine(ctx context.Context, machine *proto.TypedMachine) (*proto.TypedMachine, error) {
	err := service_g.StoreMachineRec(ctx, machine)
	if err != nil {
		return nil, err
	}
	return machine, nil
}
