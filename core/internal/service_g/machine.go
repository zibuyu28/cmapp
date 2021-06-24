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
	"github.com/zibuyu28/cmapp/core/internal/model"
	"github.com/zibuyu28/cmapp/core/proto"
)

// StoreMachineRec store machine record
func StoreMachineRec(ctx context.Context, machine *proto.Machine) error {
	m := &model.Machine{
		UUID:       machine.UUID,
		State:      int(machine.State),
		DriverID:   int(machine.DriverID),
		Tags:       machine.MachineTags,
		CustomInfo: machine.CustomInfo,
	}
	err := model.InsertMachine(m)
	if err != nil {
		return errors.Wrap(err, "insert machine")
	}
	machine.ID = int32(m.ID)
	return nil
}