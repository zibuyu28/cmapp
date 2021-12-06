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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/service"
	"os"
	"path/filepath"
	"strings"
)

type CreateChainWorker struct {
	ctx     context.Context
	baseDir string
}

func NewCreateChainWorker(ctx context.Context, coreHttpAddr, baseDir string) *CreateChainWorker {
	cw := &CreateChainWorker{
		ctx:     ctx,
		baseDir: baseDir,
	}
	hmd = &ag.HMD{
		V:            ag.V1,
		CoreHttpAddr: coreHttpAddr,
	}
	ag.NewCore(ag.V1, coreHttpAddr)
	return cw
}

var hmd *ag.HMD

// CreateChainProcess create chain
func (c *CreateChainWorker) CreateChainProcess(chain *model.Fabric) error {
	log.Debug(c.ctx, "create chain process")

	err := newApp(chain)
	if err != nil {
		return errors.Wrap(err, "new app for chain")
	}
	// 主机开端口，主要是为了将端口外部路由给记录下
	// 处理主机的 HostName 字段
	// 生成证书
	basePath, _ := filepath.Abs(filepath.Join(c.baseDir, "chain_certs",fmt.Sprintf("%s_%s", chain.Name, chain.UUID)))
	_ = os.RemoveAll(basePath)
	_ = os.MkdirAll(basePath, os.ModePerm)
	certWorker := service.NewCertWorker(basePath)
	err = certWorker.InitCert(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "init cert")
	}
	log.Debugf(c.ctx, "init cert success")
	// 生成configtx.yaml
	configWorker := service.NewConfigWorker(basePath)
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
	err = uploadNodeCert(ag.CoreIns, chain, certMap)
	if err != nil {
		return errors.Wrap(err, "upload node cert")
	}
	log.Debugf(c.ctx, "upload node cert success")
	genesisBlock := fmt.Sprintf("%s/orderer.genesis.block", basePath)
	gb, err := ag.CoreIns.UploadFile(genesisBlock)
	if err != nil {
		return errors.Wrap(err, "upload genesis block file")
	}
	chain.RemoteGenesisBlock = gb
	log.Debugf(c.ctx, "upload genesis block success")

	// 构建几个节点的配置，并启动
	err = constructOrder(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "construct orderers")
	}
	log.Debugf(c.ctx, "construct order app success")

	err = constructPeer(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "construct peers")
	}
	log.Debugf(c.ctx, "construct peer app success")

	// 创建初始通道
	channelWorker := service.NewChannelWorker(basePath)
	err = channelWorker.CreateInitChannel(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "create init channel")
	}
	log.Debugf(c.ctx, "create init channel success")
	// 节点加入通道
	err = channelWorker.JoinInitChannel(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "join init channel")
	}
	log.Debugf(c.ctx, "join init channel success")
	// 更新锚节点
	err = channelWorker.UpdateAnchorPeers(c.ctx, chain)
	if err != nil {
		return errors.Wrap(err, "join init channel")
	}
	log.Debugf(c.ctx, "update anchor peer success")
	// 完成过程
	return nil
}

func constructPeer(ctx context.Context, chain *model.Fabric) error {
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
		log.Debugf(ctx, "cert addr [%s]", peer.RemoteCert)
		err = hmd.FilePremiseEx(peer.APP.UUID, &ag.File{
			Name:        "msp.tar.gz",
			AcquireAddr: peer.RemoteCert,
			Shell:       "mkdir config && tar -zxvf msp.tar.gz -C config/ --strip-components 1",
		})
		if err != nil {
			return errors.Wrap(err, "set cert file premise")
		}
		err = hmd.LimitEx(peer.APP.UUID, &ag.Limit{
			CPU:    1000,
			Memory: 1024,
		})
		if err != nil {
			return errors.Wrap(err, "set app limit")
		}
		err = hmd.HealthEx(peer.APP.UUID, &ag.Health{
			Liveness: ag.Basic{
				MethodType: ag.GET,
				Path:       "/healthz",
				Port:       peer.HealthPort,
			},
			Readness: ag.Basic{
				MethodType: ag.GET,
				Path:       "/healthz",
				Port:       peer.HealthPort,
			},
		})
		if err != nil {
			return errors.Wrap(err, "set app health")
		}
		var envs = make(map[string]string)
		//envs["FABRIC_CFG_PATH"] = peer.APP.Workspace.Workspace
		envs["CORE_PEER_FILESYSTEMPATH"] = "pro-data"
		envs["CORE_LEDGER_STATE_STATEDATABASE"] = "goleveldb"
		//envs["CORE_LEDGER_STATE_STATEDATABASE"] = "CouchDB"
		//for i, network := range peer.CouchDB.Networks {
		//	if network.PortInfo.Port != 5984 {
		//		continue
		//	}
		//	for _, s := range peer.CouchDB.Networks[i].RouteInfo {
		//		if s.RouteType == ag.OUT {
		//			envs["CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS"] = s.Router
		//		}
		//	}
		//}
		//envs["CORE_LEDGER_STATE_COUCHDBCONFIG_USERNAME"] = ""
		//envs["CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD"] = ""
		//envs["CORE_CHAINCODE_BUILDER"] = ""
		//envs["CORE_CHAINCODE_GOLANG_RUNTIME"] = ""
		envs["CORE_VM_ENDPOINT"] = peer.RMTDocker
		envs["FABRIC_LOGGING_SPEC"] = peer.LogLevel
		envs["CORE_PEER_TLS_ENABLED"] = fmt.Sprintf("%t", chain.TLSEnable)
		envs["CORE_PEER_LISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", peer.GRPCPort)
		envs["CORE_OPERATIONS_LISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", peer.HealthPort)
		envs["CORE_PEER_GOSSIP_USELEADERELECTION"] = "true"
		envs["CORE_PEER_GOSSIP_ORGLEADER"] = "false"
		envs["CORE_PEER_PROFILE_ENABLED"] = "true"
		envs["CORE_PEER_TLS_CERT_FILE"] = "config/tls/server.crt"
		envs["CORE_PEER_TLS_KEY_FILE"] = "config/tls/server.key"
		envs["CORE_PEER_TLS_ROOTCERT_FILE"] = "config/tls/ca.crt"
		envs["CORE_PEER_ADDRESSAUTODETECT"] = "true"
		envs["CORE_PEER_CHAINCODELISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", peer.ChainCodeListenPort)// 链码监听地址
		for i, network := range peer.APP.Networks {
			if network.PortInfo.Port == peer.ChainCodeListenPort {
				for _, s := range peer.APP.Networks[i].RouteInfo {
					if s.RouteType == ag.OUT {
						envs["CORE_PEER_CHAINCODEADDRESS"] = s.Router // 链码回调时访问的地址
					}
				}
			} else if network.PortInfo.Port == peer.GRPCPort {
				for _, s := range peer.APP.Networks[i].RouteInfo {
					if s.RouteType == ag.OUT {
						envs["CORE_PEER_GOSSIP_BOOTSTRAP"] = s.Router
						envs["CORE_PEER_ADDRESS"] = s.Router
						envs["CORE_PEER_GOSSIP_EXTERNALENDPOINT"] = s.Router
					}
				}
			}
		}
		envs["CORE_PEER_ID"] = peer.UUID
		envs["CORE_PEER_LOCALMSPID"] = fmt.Sprintf("%sMSP", peer.Organization.UUID)
		envs["CORE_PEER_MSPCONFIGPATH"] = "config/msp"

		for k, v := range envs {
			log.Debugf(ctx, "set env key [%s], val [%s]", k, v)
			err = hmd.EnvEx(peer.APP.UUID, &ag.EnvVar{Key: k, Value: v})
			if err != nil {
				return errors.Wrapf(err, "set env, key [%s], value [%s]", k, v)
			}
		}

		err = hmd.StartApp(peer.APP.UUID, peer.APP)
		if err != nil {
			return errors.Wrapf(err, "star peer app [%s]", peer.APP.UUID)
		}
	}
	return nil
}

func constructOrder(ctx context.Context, chain *model.Fabric) error {
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
			Shell:       "mkdir config && tar -zxvf msp.tar.gz -C config/ --strip-components 1",
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
		//envs["FABRIC_CFG_PATH"] = order.APP.Workspace.Workspace
		envs["ORDERER_FILELEDGER_LOCATION"] = "pro-data"
		envs["ORDERER_GENERAL_LISTENADDRESS"] = "0.0.0.0"
		envs["ORDERER_GENERAL_LISTENPORT"] = fmt.Sprintf("%d", order.GRPCPort)
		envs["ORDERER_OPERATIONS_LISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", order.HealthPort)
		envs["ORDERER_GENERAL_GENESISMETHOD"] = "file"
		envs["ORDERER_GENERAL_GENESISFILE"] = "orderer.genesis.block"
		envs["ORDERER_GENERAL_LOCALMSPID"] = "OrdererMSP"
		envs["ORDERER_GENERAL_LOCALMSPDIR"] = "config/msp"
		envs["ORDERER_GENERAL_TLS_ENABLED"] = fmt.Sprintf("%t", chain.TLSEnable)
		envs["ORDERER_GENERAL_TLS_PRIVATEKEY"] = "config/tls/server.key"
		envs["ORDERER_GENERAL_TLS_CERTIFICATE"] = "config/tls/server.crt"
		envs["ORDERER_GENERAL_TLS_ROOTCAS"] = "config/tls/ca.crt"
		envs["ORDERER_KAFKA_VERBOSE"] = "true"
		envs["ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY"] = "config/tls/server.key"
		envs["ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE"] = "config/tls/server.crt"
		envs["ORDERER_GENERAL_CLUSTER_ROOTCAS"] = "config/tls/ca.crt"
		envs["GODEBUG"] = "netdns=go"
		for k, v := range envs {
			log.Debugf(ctx, "set env key [%s], val [%s]", k, v)
			err = hmd.EnvEx(order.APP.UUID, &ag.EnvVar{Key: k, Value: v})
			if err != nil {
				return errors.Wrapf(err, "set env, key [%s], value [%s]", k, v)
			}
		}
		err = hmd.StartApp(order.APP.UUID, order.APP)
		if err != nil {
			return errors.Wrapf(err, "star order app [%s]", order.APP.UUID)
		}
	}
	return nil
}

func newApp(chain *model.Fabric) error {
	for i := range chain.Orderers {
		err := newOrdererApp(&chain.Orderers[i])
		if err != nil {
			return errors.Wrap(err, "new orderer app")
		}
	}
	for i := range chain.Peers {
		err := newPeerApp(&chain.Peers[i])
		if err != nil {
			return errors.Wrap(err, "new peer app")
		}
		//err = newCouchDBApp(&chain.Peers[i])
		//if err != nil {
		//	return errors.Wrap(err, "new peer couchdb app")
		//}
	}
	return nil
}

func newOrdererApp(orderer *model.Orderer) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: orderer.MachineID,
		Name:      "orderer",
		Version:   "1.4",
	})
	if err != nil {
		return errors.Wrap(err, "new app")
	}
	orderer.APP = app
	grpc := &ag.Network{
		PortInfo: struct {
			Port         int         `json:"port"`
			Name         string      `json:"name"`
			ProtocolType ag.Protocol `json:"protocol_type"`
		}{
			Port:         orderer.GRPCPort,
			Name:         fmt.Sprintf("%s-grpc", orderer.UUID),
			ProtocolType: ag.TCP,
		},
	}
	err = hmd.NetworkEx(app.UUID, grpc)
	if err != nil {
		return errors.Wrap(err, "net work set ex")
	}
	marshal, _ := json.Marshal(grpc)
	log.Debugf(context.Background(), "grpc network exec result [%s]", string(marshal))
	health := &ag.Network{
		PortInfo: struct {
			Port         int         `json:"port"`
			Name         string      `json:"name"`
			ProtocolType ag.Protocol `json:"protocol_type"`
		}{
			Port:         orderer.HealthPort,
			Name:         fmt.Sprintf("%s-health", orderer.UUID),
			ProtocolType: ag.TCP,
		},
	}
	err = hmd.NetworkEx(app.UUID, health)
	if err != nil {
		return errors.Wrap(err, "net work set ex")
	}
	marshal, _ = json.Marshal(health)
	log.Debugf(context.Background(), "health network exec result [%s]", string(marshal))
	app.Networks = []ag.Network{*grpc, *health}
	for _, s := range grpc.RouteInfo {
		if s.RouteType == ag.OUT {
			split := strings.Split(s.Router, ":")
			orderer.NodeHostName = split[0]
		}
	}
	return nil
}

func newCouchDBApp(peer *model.Peer) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: peer.MachineID,
		Name:      "couchdb",
		Version:   "1.4.3",
	})
	if err != nil {
		return errors.Wrap(err, "new app")
	}
	peer.CouchDB = app
	health := &ag.Network{
		PortInfo: struct {
			Port         int         `json:"port"`
			Name         string      `json:"name"`
			ProtocolType ag.Protocol `json:"protocol_type"`
		}{
			Port:         5984,
			Name:         "http",
			ProtocolType: ag.TCP,
		},
	}
	err = hmd.NetworkEx(app.UUID, health)
	if err != nil {
		return errors.Wrap(err, "net work set ex")
	}
	app.Networks = []ag.Network{*health}
	return nil
}

func newPeerApp(peer *model.Peer) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: peer.MachineID,
		Name:      "peer",
		Version:   "1.4",
	})
	if err != nil {
		return errors.Wrap(err, "new app")
	}
	peer.APP = app
	grpc := &ag.Network{
		PortInfo: struct {
			Port         int         `json:"port"`
			Name         string      `json:"name"`
			ProtocolType ag.Protocol `json:"protocol_type"`
		}{
			Port:         peer.GRPCPort,
			Name:         fmt.Sprintf("%s-grpc", peer.UUID),
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
			Name:         fmt.Sprintf("%s-health", peer.UUID),
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
			Name:         fmt.Sprintf("%s-chaincode", peer.UUID),
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
			Name:         fmt.Sprintf("%s-event", peer.UUID),
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
			split := strings.Split(s.Router, ":")
			peer.NodeHostName = split[0]
		}
	}
	return nil
}

func uploadNodeCert(capi ag.CoreAPI, chain *model.Fabric, certMap map[string]string) error {
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
