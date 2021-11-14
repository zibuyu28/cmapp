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
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/internal/service_g"
	"github.com/zibuyu28/cmapp/core/proto/ch_manager"
)

type CoreChainManager struct {
}

// ReportNodes report nodes
func (c CoreChainManager) ReportNodes(ctx context.Context, nodes *ch_manager.TypedNodes) (*ch_manager.TypedNodes, error) {
	defer func() {
		if e := recover(); e != nil {
			for _, node := range nodes.Nodes {
				if node.ID != 0 {
					_ = service_g.RemoveNodeRec(ctx, int(node.ID))
				}
			}
		}
	}()
	for i, node := range nodes.Nodes {
		err := service_g.StoreNodeRec(ctx, node)
		if err != nil {
			return nil, errors.Wrap(err, "store node")
		}
		nodes.Nodes[i].ID = node.ID
	}
	return nodes, nil
}

func (c CoreChainManager) UpdateChain(ctx context.Context, chain *ch_manager.TypedChain) (*ch_manager.TypedChain, error) {
	if chain.ID != 0 {
		return nil, errors.New("chain id is nil")
	}
	err := service_g.UpdateChain(ctx, chain)
	if err != nil {
		return nil, errors.Wrap(err, "update chain")
	}
	return chain, nil
}

func (c CoreChainManager) ReportChain(ctx context.Context, chain *ch_manager.TypedChain) (*ch_manager.TypedChain, error) {
	err := service_g.StoreChainRec(ctx, chain)
	if err != nil {
		return nil, errors.Wrap(err, "store chain")
	}
	return chain, nil
}
