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

package process

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"github.com/zibuyu28/cmapp/core/pkg/ag/base"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/service"
	"os"
	"path/filepath"
)

type CreateChainWorker struct {
	ctx        context.Context
	driveruuid string
	baseDir    string
}

func NewCreateChainWorker(ctx context.Context, driveruuid, baseDir string) *CreateChainWorker {
	return &CreateChainWorker{
		ctx:        ctx,
		driveruuid: driveruuid,
		baseDir:    baseDir,
	}
}

// CreateChainProcess create chain
func (c *CreateChainWorker) CreateChainProcess(chain *fabric.Fabric) error {
	log.Debug(c.ctx, "create chain process")
	// 主机开端口，主要是为了将端口外部路由给记录下
	// 处理主机的 HostName 字段
	// 生成证书
	basePath, _ := filepath.Abs(fmt.Sprintf("chain_certs/%s_%s", chain.Name, chain.UUID))
	_ = os.MkdirAll(basePath, os.ModePerm)
	certWorker := service.NewCertWorker(c.driveruuid, basePath)
	err := certWorker.InitCert(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init cert")
	}
	log.Debugf(c.ctx, "init cert success")
	// 生成configtx.yaml
	configWorker := service.NewConfigWorker(c.driveruuid, basePath)
	err = configWorker.InitTxFile(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init config tx file")
	}
	log.Debugf(c.ctx, "init config tx file success")
	// 根据 configtx.yaml 生成 genesisblock, channel.tx, anchorPeerArtifacts.block
	err = configWorker.GenesisBlock(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init genesis block file")
	}
	log.Debugf(c.ctx, "init genesis block file success")

	err = configWorker.AllChannelTxFile(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init channel tx file")
	}
	log.Debugf(c.ctx, "init channel tx file success")

	err = configWorker.AnchorPeerArtifacts(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init anchor peer artifacts file")
	}
	log.Debugf(c.ctx, "init anchor peer artifacts file success")

	certMap, err := certWorker.GetNodeCertMap()
	if err != nil {
		return errors.Wrap(err, "get node cert map")
	}
	// 上传证书，创世区块
	var capi ag.CoreAPI
	err = uploadNodeCert(capi, chain, certMap)
	if err != nil {
		return errors.Wrap(err, "upload node cert")
	}
	genesisBlock := fmt.Sprintf("%s/orderer.genesis.block", basePath)
	gb, err := capi.UploadFile(genesisBlock)
	if err != nil {
		return errors.Wrap(err, "upload genesis block file")
	}
	chain.RemoteGenesisBlock = gb
	// 构建几个节点的配置，并启动

	// 创建初始通道
	// 节点加入通道
	// 更新锚节点
	// 完成过程

	return nil
}

func newApp(chain *fabric.Fabric) error {
	hmd := ag.HMD{Version: base.V1}
	for i, orderer := range chain.Orderers {
		app, err := hmd.NewApp(&ag.NewAppReq{
			MachineID: orderer.MachineID,
			Name:      "orderer",
			Version:   "1.4.3",
		})
		if err != nil {
			return errors.Wrap(err, "new app")
		}
		chain.Orderers[i].APP = app
		grpc := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         7050,
				Name:         "grpc",
				ProtocolType: ag.TCP,
			},
		}
		health := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         8443,
				Name:         "health",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, health)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		app.Networks = []ag.Network{*grpc, *health}
	}
	return nil
}

func uploadNodeCert(capi ag.CoreAPI, chain *fabric.Fabric, certMap map[string]string) error {
	for i, orderer := range chain.Orderers {
		if s, ok := certMap[orderer.UUID]; ok {
			file, err := capi.UploadFile(s)
			if err != nil {
				return errors.Wrapf(err, "upload file [%s]", s)
			}
			chain.Orderers[i].RemoteCert = file
		}
	}
	for i, peer := range chain.Peers {
		if s, ok := certMap[peer.UUID]; ok {
			file, err := capi.UploadFile(s)
			if err != nil {
				return errors.Wrapf(err, "upload file [%s]", s)
			}
			chain.Peers[i].RemoteCert = file
		}
	}
	return nil
}
