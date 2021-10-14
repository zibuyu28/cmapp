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
	"github.com/zibuyu28/cmapp/core/proto/ch_manager"
)

type CoreChainManager struct {
}

func (c CoreChainManager) ReportNodes(ctx context.Context, nodes *ch_manager.TypedNodes) (*ch_manager.TypedNodes, error) {
	panic("implement me")
}

func (c CoreChainManager) UpdateChain(ctx context.Context, chain *ch_manager.TypedChain) (*ch_manager.TypedChain, error) {
	panic("implement me")
}

func (c CoreChainManager) ReportChain(ctx context.Context, chain *ch_manager.TypedChain) (*ch_manager.TypedChain, error) {
	panic("implement me")
}
