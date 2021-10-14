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
	"github.com/zibuyu28/cmapp/core/proto/ch_manager"
)


// UpdateChain update chain
func UpdateChain(ctx context.Context, chain *ch_manager.TypedChain) error {
	mc := model.Chain{ID: int(chain.ID)}
	var fields []string
	if chain.State != 0 {
		mc.State = int(chain.State)
		fields = append(fields, "state")
	}
	if len(chain.Tags) != 0 {
		mc.Tags = chain.Tags
		fields = append(fields, "tags")
	}
	if len(chain.CustomInfo) != 0 {
		mc.CustomInfo = chain.CustomInfo
		fields = append(fields, "custom_info")
	}
	if len(fields) != 0 {
		err := model.UpdateChain(&mc, mc.ID, fields)
		if err != nil {
			return errors.Wrap(err, "udpate chain")
		}
	}
	return nil
}


// StoreChainRec store chain record
func StoreChainRec(ctx context.Context, chain *ch_manager.TypedChain) error {
	mc := &model.Chain{
		Name:       chain.Name,
		UUID:       chain.UUID,
		Type:       chain.Type,
		Version:    chain.Version,
		State:      int(chain.State),
		DriverID:   int(chain.DriverID),
		Tags:       chain.Tags,
		CustomInfo: chain.CustomInfo,
	}

	err := model.InsertChain(mc)
	if err != nil {
		return errors.Wrap(err, "insert chain")
	}
	chain.ID = int32(mc.ID)
	return nil
}

