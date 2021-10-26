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

var hmd = ag.HMD{Version: base.V1}

// CreateChainProcess create chain
func (c *CreateChainWorker) CreateChainProcess(chain *fabric.Fabric) error {
	log.Debug(c.ctx, "create chain process")

	err := newApp(chain)
	if err != nil {
		return errors.Wrap(err, "new app for chain")
	}
	// 主机开端口，主要是为了将端口外部路由给记录下
	// 处理主机的 HostName 字段
	// 生成证书
	basePath, _ := filepath.Abs(fmt.Sprintf("chain_certs/%s_%s", chain.Name, chain.UUID))
	_ = os.MkdirAll(basePath, os.ModePerm)
	certWorker := service.NewCertWorker(c.driveruuid, basePath)
	err = certWorker.InitCert(c.ctx, chain)
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

func constructPeer(ctx context.Context, chain *fabric.Fabric) error {
	for _, peer := range chain.Peers {
		if peer.APP == nil {
			return errors.New("app is nil")
		}
		err := hmd.TagEx(peer.APP.UUID, &ag.Tag{
			Key:   "tag",
			Value: peer.Tag,
		})
		if err != nil {
			return errors.Wrap(err, "set tag")
		}

	}
	panic("implement me")
}

func constructOrder(ctx context.Context, chain *fabric.Fabric) error {
	for _, order := range chain.Orderers {
		if order.APP == nil {
			return errors.New("app is nil")
		}
		err := hmd.TagEx(order.APP.UUID, &ag.Tag{
			Key:   "tag",
			Value: order.Tag,
		})
		if err != nil {
			return errors.Wrap(err, "set tag")
		}
		log.Debugf(ctx, "cert addr [%s]", order.RemoteCert)
		err = hmd.FilePremiseEx(order.APP.UUID, &ag.File{
			Name:        "msp.tar.gz",
			AcquireAddr: order.RemoteCert,
			Shell:       "mkdir cert && tar -zxvf msp.tar.gz --strip-components 1",
		})
		if err != nil {
			return errors.Wrap(err, "set cert file premise")
		}
		log.Debugf(ctx, "genesis block addr [%s]", chain.RemoteGenesisBlock)
		err = hmd.FilePremiseEx(order.APP.UUID, &ag.File{
			Name:        "orderer.genesis.block",
			AcquireAddr: chain.RemoteGenesisBlock,
			Shell:       "",
		})
		if err != nil {
			return errors.Wrap(err, "set genesis block file premise")
		}

		err = hmd.LimitEx(order.APP.UUID, &ag.Limit{
			CPU:    1000,
			Memory: 1024,
		})
		if err != nil {
			return errors.Wrap(err, "set app limit")
		}
		err = hmd.HealthEx(order.APP.UUID, &ag.Health{
			Liveness: ag.Basic{
				MethodType: ag.GET,
				Path:       "/healthz",
				Port:       order.HealthPort,
			},
			Readness: ag.Basic{
				MethodType: ag.GET,
				Path:       "/healthz",
				Port:       order.HealthPort,
			},
		})
		if err != nil {
			return errors.Wrap(err, "set app health")
		}
		var envs = make(map[string]string)
		envs["FABRIC_LOGGING_SPEC"] = order.LogLevel
		envs["ORDERER_GENERAL_LISTENADDRESS"] = "0.0.0.0"
		envs["ORDERER_OPERATIONS_LISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", order.HealthPort)
		envs["ORDERER_GENERAL_GENESISMETHOD"] = "file"
		envs["ORDERER_GENERAL_GENESISFILE"] = fmt.Sprintf("%s/orderer.genesis.block", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_LOCALMSPID"] = "OrdererMSP"
		envs["ORDERER_GENERAL_LOCALMSPDIR"] = fmt.Sprintf("%s/config/msp", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_TLS_ENABLED"] = fmt.Sprintf("%t", chain.TLSEnable)
		envs["ORDERER_GENERAL_TLS_PRIVATEKEY"] = fmt.Sprintf("%s/config/tls/keystore/key.pem", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_TLS_CERTIFICATE"] = fmt.Sprintf("%s/config/tls/signcerts/cert.pem", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_TLS_ROOTCAS"] = fmt.Sprintf("[%s/config/tls/tlscacerts/tls.pem]", order.APP.Workspace.Workspace)
		envs["ORDERER_KAFKA_VERBOSE"] = "true"
		envs["ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY"] = fmt.Sprintf("%s/config/tls/keystore/key.pem", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE"] = fmt.Sprintf("%s/config/tls/signcerts/cert.pem", order.APP.Workspace.Workspace)
		envs["ORDERER_GENERAL_CLUSTER_ROOTCAS"] = fmt.Sprintf("[%s/config/tls/tlscacerts/tls.pem]", order.APP.Workspace.Workspace)
		envs["GODEBUG"] = "netdns=go"
		for k, v := range envs {
			log.Debugf(ctx, "set env key [%s], val [%s]", k, v)
			err = hmd.EnvEx(order.APP.UUID, &ag.EnvVar{Key: k, Value: v})
			if err != nil {
				return errors.Wrapf(err, "set env, key [%s], value [%s]", k, v)
			}
		}
	}
	return nil
}

func newApp(chain *fabric.Fabric) error {

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
				Port:         orderer.GRPCPort,
				Name:         "grpc",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, grpc)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		health := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         orderer.HealthPort,
				Name:         "health",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, health)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		app.Networks = []ag.Network{*grpc, *health}
		for _, s := range grpc.RouteInfo {
			if s.RouteType == ag.OUT {
				chain.Orderers[i].NodeHostName = s.Router
			}
		}
	}
	for i, peer := range chain.Peers {
		app, err := hmd.NewApp(&ag.NewAppReq{
			MachineID: peer.MachineID,
			Name:      "peer",
			Version:   "1.4.3",
		})
		if err != nil {
			return errors.Wrap(err, "new app")
		}
		chain.Peers[i].APP = app
		grpc := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         peer.GRPCPort,
				Name:         "grpc",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, grpc)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		health := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         peer.HealthPort,
				Name:         "health",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, health)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		ccport := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         peer.ChainCodeListenPort,
				Name:         "chaincode",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, ccport)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		event := &ag.Network{
			PortInfo: struct {
				Port         int         `json:"port"`
				Name         string      `json:"name"`
				ProtocolType ag.Protocol `json:"protocol_type"`
			}{
				Port:         peer.EventPort,
				Name:         "event",
				ProtocolType: ag.TCP,
			},
		}
		err = hmd.NetworkEx(app.UUID, event)
		if err != nil {
			return errors.Wrap(err, "net work set ex")
		}
		app.Networks = []ag.Network{*grpc, *health, *ccport, *event}
		for _, s := range grpc.RouteInfo {
			if s.RouteType == ag.OUT {
				chain.Peers[i].NodeHostName = s.Router
			}
		}
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
