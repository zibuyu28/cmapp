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

// StoreNodeRec store node record
func StoreNodeRec(ctx context.Context, node *ch_manager.TypedNode) error {
	mn := &model.Node{
		Name:       node.Name,
		UUID:       node.UUID,
		Type:       node.Type,
		State:      int(node.State),
		ChainID:    int(node.ChainID),
		MachineID:  int(node.MachineID),
		Tags:       node.Tags,
		CustomInfo: node.CustomInfo,
	}

	err := model.InsertNode(mn)
	if err != nil {
		return errors.Wrap(err, "insert node")
	}
	node.ID = int32(mn.ID)
	return nil
}


// RemoveNodeRec remove node
func RemoveNodeRec(ctx context.Context,id int) error {
	node := model.Node{ID: id}
	err := model.DeleteNode(&node)
	if err != nil {
		return errors.Wrap(err, "insert node")
	}
	return nil
}