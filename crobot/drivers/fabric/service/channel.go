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

package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/fabtool"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
)

type ChannelWorker struct {
	BaseWorker
}

// NewChannelWorker new channel worker
func NewChannelWorker(workDir string) *ChannelWorker {
	return &ChannelWorker{
		BaseWorker: BaseWorker{
			workDir: workDir,
		},
	}
}

func (c *ChannelWorker) CreateInitChannel(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "peer", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init peer tool")
	}
	var targetPeer *model.Peer
	for _, peer := range chain.Peers {
		if peer.AnchorPeer {
			targetPeer = &peer
			break
		}
	}
	if targetPeer == nil {
		return errors.New("no anchor peer")
	}
	for _, channel := range chain.Channels {
		err = itool.(fabtool.PeerTool).CreateNewChannel(chain, channel, targetPeer, c.workDir)
		if err != nil {
			return errors.Wrap(err, "generate genesis block")
		}
	}
	return nil
}

func (c *ChannelWorker) JoinInitChannel(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "peer", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init peer tool")
	}
	for _, peer := range chain.Peers {
		for _, channel := range chain.Channels {
			err = itool.(fabtool.PeerTool).JoinChannel(chain, &peer, fmt.Sprintf("%s/%s.block", c.workDir, channel.UUID), c.workDir)
			if err != nil {
				return errors.Wrap(err, "generate genesis block")
			}
		}
	}
	return nil
}

func (c *ChannelWorker) UpdateAnchorPeers(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "peer", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init peer tool")
	}
	for _, peer := range chain.Peers {
		if peer.AnchorPeer {
			for _, channel := range chain.Channels {
				err = itool.(fabtool.PeerTool).UpdateAnchorPeer(chain, &peer, fmt.Sprintf("%s/%s-%sMSPAnchors.tx", c.workDir, channel.UUID, peer.Organization.UUID), c.workDir)
				if err != nil {
					return errors.Wrap(err, "generate genesis block")
				}
			}
		}
	}
	return nil
}
