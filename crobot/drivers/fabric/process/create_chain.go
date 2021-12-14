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
	"io/ioutil"
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
	basePath, _ := filepath.Abs(filepath.Join(c.baseDir, "chain_certs", fmt.Sprintf("%s_%s", chain.Name, chain.UUID)))
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
	// 上传配置文件 core.yaml orderer.yaml
	err = nodeConfig(ag.CoreIns, chain, basePath)
	if err != nil {
		return errors.Wrap(err, "upload node config")
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
		log.Debugf(ctx, "peer yaml addr [%s]", peer.RemoteConfig)
		err = hmd.FilePremiseEx(peer.APP.UUID, &ag.File{
			Name:        "core.yaml",
			AcquireAddr: peer.RemoteConfig,
			Shell:       "",
		})
		if err != nil {
			return errors.Wrap(err, "set core.yaml file premise")
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
		envs["CORE_PEER_CHAINCODELISTENADDRESS"] = fmt.Sprintf("0.0.0.0:%d", peer.ChainCodeListenPort) // 链码监听地址
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
		log.Debugf(ctx, "orderer yaml addr [%s]", order.RemoteConfig)
		err = hmd.FilePremiseEx(order.APP.UUID, &ag.File{
			Name:        "orderer.yaml",
			AcquireAddr: order.RemoteConfig,
			Shell:       "",
		})
		if err != nil {
			return errors.Wrap(err, "set orderer.yaml file premise")
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
		// TODO: 说明当时去除这个的原因->相对路径？
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
		err := newOrdererApp(&chain.Orderers[i], chain.Version)
		if err != nil {
			return errors.Wrap(err, "new orderer app")
		}
	}
	for i := range chain.Peers {
		err := newPeerApp(&chain.Peers[i], chain.Version)
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

func newOrdererApp(orderer *model.Orderer, version string) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: orderer.MachineID,
		Name:      "orderer",
		Version:   version,
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

func newCouchDBApp(peer *model.Peer, version string) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: peer.MachineID,
		Name:      "couchdb",
		Version:   version,
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

func newPeerApp(peer *model.Peer, version string) error {
	app, err := hmd.NewApp(&ag.NewAppReq{
		MachineID: peer.MachineID,
		Name:      "peer",
		Version:   version,
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


func nodeConfig(capi ag.CoreAPI, fabric *model.Fabric, basePath string) error {
	err := ioutil.WriteFile(filepath.Join(basePath, "core.yaml"), []byte(coreDefaultYaml), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "write core yaml")
	}
	core, err := capi.UploadFile(filepath.Join(basePath, "core.yaml"))
	if err != nil {
		return errors.Wrapf(err, "upload file [%s]", filepath.Join(basePath, "core.yaml"))
	}
	err = ioutil.WriteFile(filepath.Join(basePath, "orderer.yaml"), []byte(ordererDefaultYaml), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "write orderer yaml")
	}
	order, err := capi.UploadFile(filepath.Join(basePath, "orderer.yaml"))
	if err != nil {
		return errors.Wrapf(err, "upload file [%s]", filepath.Join(basePath, "orderer.yaml"))
	}
	for i := range fabric.Orderers {
		fabric.Orderers[i].RemoteConfig = order
	}
	for i := range fabric.Peers {
		fabric.Peers[i].RemoteConfig = core
	}
	return nil
}


var coreDefaultYaml = `# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

###############################################################################
#
#    Peer section
#
###############################################################################
peer:

    # The Peer id is used for identifying this Peer instance.
    id: jdoe

    # The networkId allows for logical seperation of networks
    networkId: dev

    # The Address at local network interface this Peer will listen on.
    # By default, it will listen on all network interfaces
    listenAddress: 0.0.0.0:7051

    # The endpoint this peer uses to listen for inbound chaincode connections.
    # If this is commented-out, the listen address is selected to be
    # the peer's address (see below) with port 7052
    # chaincodeListenAddress: 0.0.0.0:7052

    # The endpoint the chaincode for this peer uses to connect to the peer.
    # If this is not specified, the chaincodeListenAddress address is selected.
    # And if chaincodeListenAddress is not specified, address is selected from
    # peer listenAddress.
    # chaincodeAddress: 0.0.0.0:7052

    # When used as peer config, this represents the endpoint to other peers
    # in the same organization. For peers in other organization, see
    # gossip.externalEndpoint for more info.
    # When used as CLI config, this means the peer's endpoint to interact with
    address: 0.0.0.0:7051

    # Whether the Peer should programmatically determine its address
    # This case is useful for docker containers.
    addressAutoDetect: false

    # Setting for runtime.GOMAXPROCS(n). If n < 1, it does not change the
    # current setting
    gomaxprocs: -1

    # Keepalive settings for peer server and clients
    keepalive:
        # MinInterval is the minimum permitted time between client pings.
        # If clients send pings more frequently, the peer server will
        # disconnect them
        minInterval: 60s
        # Client keepalive settings for communicating with other peer nodes
        client:
            # Interval is the time between pings to peer nodes.  This must
            # greater than or equal to the minInterval specified by peer
            # nodes
            interval: 60s
            # Timeout is the duration the client waits for a response from
            # peer nodes before closing the connection
            timeout: 20s
        # DeliveryClient keepalive settings for communication with ordering
        # nodes.
        deliveryClient:
            # Interval is the time between pings to ordering nodes.  This must
            # greater than or equal to the minInterval specified by ordering
            # nodes.
            interval: 60s
            # Timeout is the duration the client waits for a response from
            # ordering nodes before closing the connection
            timeout: 20s


    # Gossip related configuration
    gossip:
        # Bootstrap set to initialize gossip with.
        # This is a list of other peers that this peer reaches out to at startup.
        # Important: The endpoints here have to be endpoints of peers in the same
        # organization, because the peer would refuse connecting to these endpoints
        # unless they are in the same organization as the peer.
        bootstrap: 127.0.0.1:7051

        # NOTE: orgLeader and useLeaderElection parameters are mutual exclusive.
        # Setting both to true would result in the termination of the peer
        # since this is undefined state. If the peers are configured with
        # useLeaderElection=false, make sure there is at least 1 peer in the
        # organization that its orgLeader is set to true.

        # Defines whenever peer will initialize dynamic algorithm for
        # "leader" selection, where leader is the peer to establish
        # connection with ordering service and use delivery protocol
        # to pull ledger blocks from ordering service. It is recommended to
        # use leader election for large networks of peers.
        useLeaderElection: true
        # Statically defines peer to be an organization "leader",
        # where this means that current peer will maintain connection
        # with ordering service and disseminate block across peers in
        # its own organization
        orgLeader: false

        # Interval for membershipTracker polling
        membershipTrackerInterval: 5s

        # Overrides the endpoint that the peer publishes to peers
        # in its organization. For peers in foreign organizations
        # see 'externalEndpoint'
        endpoint:
        # Maximum count of blocks stored in memory
        maxBlockCountToStore: 100
        # Max time between consecutive message pushes(unit: millisecond)
        maxPropagationBurstLatency: 10ms
        # Max number of messages stored until a push is triggered to remote peers
        maxPropagationBurstSize: 10
        # Number of times a message is pushed to remote peers
        propagateIterations: 1
        # Number of peers selected to push messages to
        propagatePeerNum: 3
        # Determines frequency of pull phases(unit: second)
        # Must be greater than digestWaitTime + responseWaitTime
        pullInterval: 4s
        # Number of peers to pull from
        pullPeerNum: 3
        # Determines frequency of pulling state info messages from peers(unit: second)
        requestStateInfoInterval: 4s
        # Determines frequency of pushing state info messages to peers(unit: second)
        publishStateInfoInterval: 4s
        # Maximum time a stateInfo message is kept until expired
        stateInfoRetentionInterval:
        # Time from startup certificates are included in Alive messages(unit: second)
        publishCertPeriod: 10s
        # Should we skip verifying block messages or not (currently not in use)
        skipBlockVerification: false
        # Dial timeout(unit: second)
        dialTimeout: 3s
        # Connection timeout(unit: second)
        connTimeout: 2s
        # Buffer size of received messages
        recvBuffSize: 20
        # Buffer size of sending messages
        sendBuffSize: 200
        # Time to wait before pull engine processes incoming digests (unit: second)
        # Should be slightly smaller than requestWaitTime
        digestWaitTime: 1s
        # Time to wait before pull engine removes incoming nonce (unit: milliseconds)
        # Should be slightly bigger than digestWaitTime
        requestWaitTime: 1500ms
        # Time to wait before pull engine ends pull (unit: second)
        responseWaitTime: 2s
        # Alive check interval(unit: second)
        aliveTimeInterval: 5s
        # Alive expiration timeout(unit: second)
        aliveExpirationTimeout: 25s
        # Reconnect interval(unit: second)
        reconnectInterval: 25s
        # This is an endpoint that is published to peers outside of the organization.
        # If this isn't set, the peer will not be known to other organizations.
        externalEndpoint:
        # Leader election service configuration
        election:
            # Longest time peer waits for stable membership during leader election startup (unit: second)
            startupGracePeriod: 15s
            # Interval gossip membership samples to check its stability (unit: second)
            membershipSampleInterval: 1s
            # Time passes since last declaration message before peer decides to perform leader election (unit: second)
            leaderAliveThreshold: 10s
            # Time between peer sends propose message and declares itself as a leader (sends declaration message) (unit: second)
            leaderElectionDuration: 5s

        pvtData:
            # pullRetryThreshold determines the maximum duration of time private data corresponding for a given block
            # would be attempted to be pulled from peers until the block would be committed without the private data
            pullRetryThreshold: 60s
            # As private data enters the transient store, it is associated with the peer's ledger's height at that time.
            # transientstoreMaxBlockRetention defines the maximum difference between the current ledger's height upon commit,
            # and the private data residing inside the transient store that is guaranteed not to be purged.
            # Private data is purged from the transient store when blocks with sequences that are multiples
            # of transientstoreMaxBlockRetention are committed.
            transientstoreMaxBlockRetention: 1000
            # pushAckTimeout is the maximum time to wait for an acknowledgement from each peer
            # at private data push at endorsement time.
            pushAckTimeout: 3s
            # Block to live pulling margin, used as a buffer
            # to prevent peer from trying to pull private data
            # from peers that is soon to be purged in next N blocks.
            # This helps a newly joined peer catch up to current
            # blockchain height quicker.
            btlPullMargin: 10
            # the process of reconciliation is done in an endless loop, while in each iteration reconciler tries to
            # pull from the other peers the most recent missing blocks with a maximum batch size limitation.
            # reconcileBatchSize determines the maximum batch size of missing private data that will be reconciled in a
            # single iteration.
            reconcileBatchSize: 10
            # reconcileSleepInterval determines the time reconciler sleeps from end of an iteration until the beginning
            # of the next reconciliation iteration.
            reconcileSleepInterval: 1m
            # reconciliationEnabled is a flag that indicates whether private data reconciliation is enable or not.
            reconciliationEnabled: true
            # skipPullingInvalidTransactionsDuringCommit is a flag that indicates whether pulling of invalid
            # transaction's private data from other peers need to be skipped during the commit time and pulled
            # only through reconciler.
            skipPullingInvalidTransactionsDuringCommit: false

        # Gossip state transfer related configuration
        state:
            # indicates whenever state transfer is enabled or not
            # default value is true, i.e. state transfer is active
            # and takes care to sync up missing blocks allowing
            # lagging peer to catch up to speed with rest network
            enabled: true
            # checkInterval interval to check whether peer is lagging behind enough to
            # request blocks via state transfer from another peer.
            checkInterval: 10s
            # responseTimeout amount of time to wait for state transfer response from
            # other peers
            responseTimeout: 3s
            # batchSize the number of blocks to request via state transfer from another peer
            batchSize: 10
            # blockBufferSize reflect the maximum distance between lowest and
            # highest block sequence number state buffer to avoid holes.
            # In order to ensure absence of the holes actual buffer size
            # is twice of this distance
            blockBufferSize: 100
            # maxRetries maximum number of re-tries to ask
            # for single state transfer request
            maxRetries: 3

    # TLS Settings
    # Note that peer-chaincode connections through chaincodeListenAddress is
    # not mutual TLS auth. See comments on chaincodeListenAddress for more info
    tls:
        # Require server-side TLS
        enabled:  false
        # Require client certificates / mutual TLS.
        # Note that clients that are not configured to use a certificate will
        # fail to connect to the peer.
        clientAuthRequired: false
        # X.509 certificate used for TLS server
        cert:
            file: tls/server.crt
        # Private key used for TLS server (and client if clientAuthEnabled
        # is set to true
        key:
            file: tls/server.key
        # Trusted root certificate chain for tls.cert
        rootcert:
            file: tls/ca.crt
        # Set of root certificate authorities used to verify client certificates
        clientRootCAs:
            files:
              - tls/ca.crt
        # Private key used for TLS when making client connections.  If
        # not set, peer.tls.key.file will be used instead
        clientKey:
            file:
        # X.509 certificate used for TLS when making client connections.
        # If not set, peer.tls.cert.file will be used instead
        clientCert:
            file:

    # Authentication contains configuration parameters related to authenticating
    # client messages
    authentication:
        # the acceptable difference between the current server time and the
        # client's time as specified in a client request message
        timewindow: 15m

    # Path on the file system where peer will store data (eg ledger). This
    # location must be access control protected to prevent unintended
    # modification that might corrupt the peer operations.
    fileSystemPath: /var/hyperledger/production

    # BCCSP (Blockchain crypto provider): Select which crypto implementation or
    # library to use
    BCCSP:
        Default: SW
        # Settings for the SW crypto provider (i.e. when DEFAULT: SW)
        SW:
            # TODO: The default Hash and Security level needs refactoring to be
            # fully configurable. Changing these defaults requires coordination
            # SHA2 is hardcoded in several places, not only BCCSP
            Hash: SHA2
            Security: 256
            # Location of Key Store
            FileKeyStore:
                # If "", defaults to 'mspConfigPath'/keystore
                KeyStore:
        # Settings for the PKCS#11 crypto provider (i.e. when DEFAULT: PKCS11)
        PKCS11:
            # Location of the PKCS11 module library
            Library:
            # Token Label
            Label:
            # User PIN
            Pin:
            Hash:
            Security:
            FileKeyStore:
                KeyStore:

    # Path on the file system where peer will find MSP local configurations
    mspConfigPath: msp

    # Identifier of the local MSP
    # ----!!!!IMPORTANT!!!-!!!IMPORTANT!!!-!!!IMPORTANT!!!!----
    # Deployers need to change the value of the localMspId string.
    # In particular, the name of the local MSP ID of a peer needs
    # to match the name of one of the MSPs in each of the channel
    # that this peer is a member of. Otherwise this peer's messages
    # will not be identified as valid by other nodes.
    localMspId: SampleOrg

    # CLI common client config options
    client:
        # connection timeout
        connTimeout: 3s

    # Delivery service related config
    deliveryclient:
        # It sets the total time the delivery service may spend in reconnection
        # attempts until its retry logic gives up and returns an error
        reconnectTotalTimeThreshold: 3600s

        # It sets the delivery service <-> ordering service node connection timeout
        connTimeout: 3s

        # It sets the delivery service maximal delay between consecutive retries
        reConnectBackoffThreshold: 3600s

    # Type for the local MSP - by default it's of type bccsp
    localMspType: bccsp

    # Used with Go profiling tools only in none production environment. In
    # production, it should be disabled (eg enabled: false)
    profile:
        enabled:     false
        listenAddress: 0.0.0.0:6060

    # The admin service is used for administrative operations such as
    # control over logger levels, etc.
    # Only peer administrators can use the service.
    adminService:
        # The interface and port on which the admin server will listen on.
        # If this is commented out, or the port number is equal to the port
        # of the peer listen address - the admin service is attached to the
        # peer's service (defaults to 7051).
        #listenAddress: 0.0.0.0:7055

    # Handlers defines custom handlers that can filter and mutate
    # objects passing within the peer, such as:
    #   Auth filter - reject or forward proposals from clients
    #   Decorators  - append or mutate the chaincode input passed to the chaincode
    #   Endorsers   - Custom signing over proposal response payload and its mutation
    # Valid handler definition contains:
    #   - A name which is a factory method name defined in
    #     core/handlers/library/library.go for statically compiled handlers
    #   - library path to shared object binary for pluggable filters
    # Auth filters and decorators are chained and executed in the order that
    # they are defined. For example:
    # authFilters:
    #   -
    #     name: FilterOne
    #     library: /opt/lib/filter.so
    #   -
    #     name: FilterTwo
    # decorators:
    #   -
    #     name: DecoratorOne
    #   -
    #     name: DecoratorTwo
    #     library: /opt/lib/decorator.so
    # Endorsers are configured as a map that its keys are the endorsement system chaincodes that are being overridden.
    # Below is an example that overrides the default ESCC and uses an endorsement plugin that has the same functionality
    # as the default ESCC.
    # If the 'library' property is missing, the name is used as the constructor method in the builtin library similar
    # to auth filters and decorators.
    # endorsers:
    #   escc:
    #     name: DefaultESCC
    #     library: /etc/hyperledger/fabric/plugin/escc.so
    handlers:
        authFilters:
          -
            name: DefaultAuth
          -
            name: ExpirationCheck    # This filter checks identity x509 certificate expiration
        decorators:
          -
            name: DefaultDecorator
        endorsers:
          escc:
            name: DefaultEndorsement
            library:
        validators:
          vscc:
            name: DefaultValidation
            library:

    #    library: /etc/hyperledger/fabric/plugin/escc.so
    # Number of goroutines that will execute transaction validation in parallel.
    # By default, the peer chooses the number of CPUs on the machine. Set this
    # variable to override that choice.
    # NOTE: overriding this value might negatively influence the performance of
    # the peer so please change this value only if you know what you're doing
    validatorPoolSize:

    # The discovery service is used by clients to query information about peers,
    # such as - which peers have joined a certain channel, what is the latest
    # channel config, and most importantly - given a chaincode and a channel,
    # what possible sets of peers satisfy the endorsement policy.
    discovery:
        enabled: true
        # Whether the authentication cache is enabled or not.
        authCacheEnabled: true
        # The maximum size of the cache, after which a purge takes place
        authCacheMaxSize: 1000
        # The proportion (0 to 1) of entries that remain in the cache after the cache is purged due to overpopulation
        authCachePurgeRetentionRatio: 0.75
        # Whether to allow non-admins to perform non channel scoped queries.
        # When this is false, it means that only peer admins can perform non channel scoped queries.
        orgMembersAllowedAccess: false
###############################################################################
#
#    VM section
#
###############################################################################
vm:

    # Endpoint of the vm management system.  For docker can be one of the following in general
    # unix:///var/run/docker.sock
    # http://localhost:2375
    # https://localhost:2376
    endpoint: unix:///var/run/docker.sock

    # settings for docker vms
    docker:
        tls:
            enabled: false
            ca:
                file: docker/ca.crt
            cert:
                file: docker/tls.crt
            key:
                file: docker/tls.key

        # Enables/disables the standard out/err from chaincode containers for
        # debugging purposes
        attachStdout: false
        hostConfig:
            NetworkMode: host
            Dns:
               # - 192.168.0.1
            LogConfig:
                Type: json-file
                Config:
                    max-size: "50m"
                    max-file: "5"
            Memory: 2147483648

###############################################################################
#
#    Chaincode section
#
###############################################################################
chaincode:
    id:
        path:
        name:

    # Generic builder environment, suitable for most chaincode types
    builder: $(DOCKER_NS)/fabric-ccenv:latest

    # Enables/disables force pulling of the base docker images (listed below)
    # during user chaincode instantiation.
    # Useful when using moving image tags (such as :latest)
    pull: false

    golang:
        # golang will never need more than baseos
        runtime: $(BASE_DOCKER_NS)/fabric-baseos:$(ARCH)-$(BASE_VERSION)

        # whether or not golang chaincode should be linked dynamically
        dynamicLink: false

    car:
        # car may need more facilities (JVM, etc) in the future as the catalog
        # of platforms are expanded.  For now, we can just use baseos
        runtime: $(BASE_DOCKER_NS)/fabric-baseos:$(ARCH)-$(BASE_VERSION)

    java:
        # This is an image based on java:openjdk-8 with addition compiler
        # tools added for java shim layer packaging.
        # This image is packed with shim layer libraries that are necessary
        # for Java chaincode runtime.
        runtime: $(DOCKER_NS)/fabric-javaenv:$(ARCH)-$(PROJECT_VERSION)

    node:
        # need node.js engine at runtime, currently available in baseimage
        # but not in baseos
        runtime: $(BASE_DOCKER_NS)/fabric-baseimage:$(ARCH)-$(BASE_VERSION)

    # Timeout duration for starting up a container and waiting for Register
    # to come through. 1sec should be plenty for chaincode unit tests
    startuptimeout: 300s

    # Timeout duration for Invoke and Init calls to prevent runaway.
    # This timeout is used by all chaincodes in all the channels, including
    # system chaincodes.
    # Note that during Invoke, if the image is not available (e.g. being
    # cleaned up when in development environment), the peer will automatically
    # build the image, which might take more time. In production environment,
    # the chaincode image is unlikely to be deleted, so the timeout could be
    # reduced accordingly.
    executetimeout: 30s

    # There are 2 modes: "dev" and "net".
    # In dev mode, user runs the chaincode after starting peer from
    # command line on local machine.
    # In net mode, peer will run chaincode in a docker container.
    mode: net

    # keepalive in seconds. In situations where the communiction goes through a
    # proxy that does not support keep-alive, this parameter will maintain connection
    # between peer and chaincode.
    # A value <= 0 turns keepalive off
    keepalive: 0

    # system chaincodes whitelist. To add system chaincode "myscc" to the
    # whitelist, add "myscc: enable" to the list below, and register in
    # chaincode/importsysccs.go
    system:
        cscc: enable
        lscc: enable
        escc: enable
        vscc: enable
        qscc: enable

    # System chaincode plugins:
    # System chaincodes can be loaded as shared objects compiled as Go plugins.
    # See examples/plugins/scc for an example.
    # Plugins must be white listed in the chaincode.system section above.
    systemPlugins:
      # example configuration:
      # - enabled: true
      #   name: myscc
      #   path: /opt/lib/myscc.so
      #   invokableExternal: true
      #   invokableCC2CC: true

    # Logging section for the chaincode container
    logging:
      # Default level for all loggers within the chaincode container
      level:  info
      # Override default level for the 'shim' logger
      shim:   warning
      # Format for the chaincode container logs
      format: '%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}'

###############################################################################
#
#    Ledger section - ledger configuration encompases both the blockchain
#    and the state
#
###############################################################################
ledger:

  blockchain:

  state:
    # stateDatabase - options are "goleveldb", "CouchDB"
    # goleveldb - default state database stored in goleveldb.
    # CouchDB - store state database in CouchDB
    stateDatabase: goleveldb
    # Limit on the number of records to return per query
    totalQueryLimit: 100000
    couchDBConfig:
       # It is recommended to run CouchDB on the same server as the peer, and
       # not map the CouchDB container port to a server port in docker-compose.
       # Otherwise proper security must be provided on the connection between
       # CouchDB client (on the peer) and server.
       couchDBAddress: 127.0.0.1:5984
       # This username must have read and write authority on CouchDB
       username:
       # The password is recommended to pass as an environment variable
       # during start up (eg CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD).
       # If it is stored here, the file must be access control protected
       # to prevent unintended users from discovering the password.
       password:
       # Number of retries for CouchDB errors
       maxRetries: 3
       # Number of retries for CouchDB errors during peer startup
       maxRetriesOnStartup: 12
       # CouchDB request timeout (unit: duration, e.g. 20s)
       requestTimeout: 35s
       # Limit on the number of records per each CouchDB query
       # Note that chaincode queries are only bound by totalQueryLimit.
       # Internally the chaincode may execute multiple CouchDB queries,
       # each of size internalQueryLimit.
       internalQueryLimit: 1000
       # Limit on the number of records per CouchDB bulk update batch
       maxBatchUpdateSize: 1000
       # Warm indexes after every N blocks.
       # This option warms any indexes that have been
       # deployed to CouchDB after every N blocks.
       # A value of 1 will warm indexes after every block commit,
       # to ensure fast selector queries.
       # Increasing the value may improve write efficiency of peer and CouchDB,
       # but may degrade query response time.
       warmIndexesAfterNBlocks: 1
       # Create the _global_changes system database
       # This is optional.  Creating the global changes database will require
       # additional system resources to track changes and maintain the database
       createGlobalChangesDB: false

  history:
    # enableHistoryDatabase - options are true or false
    # Indicates if the history of key updates should be stored.
    # All history 'index' will be stored in goleveldb, regardless if using
    # CouchDB or alternate database for the state.
    enableHistoryDatabase: true

###############################################################################
#
#    Operations section
#
###############################################################################
operations:
    # host and port for the operations server
    listenAddress: 127.0.0.1:9443

    # TLS configuration for the operations endpoint
    tls:
        # TLS enabled
        enabled: false

        # path to PEM encoded server certificate for the operations server
        cert:
            file:

        # path to PEM encoded server key for the operations server
        key:
            file:

        # most operations service endpoints require client authentication when TLS
        # is enabled. clientAuthRequired requires client certificate authentication
        # at the TLS layer to access all resources.
        clientAuthRequired: false

        # paths to PEM encoded ca certificates to trust for client authentication
        clientRootCAs:
            files: []

###############################################################################
#
#    Metrics section
#
###############################################################################
metrics:
    # metrics provider is one of statsd, prometheus, or disabled
    provider: disabled

    # statsd configuration
    statsd:
        # network type: tcp or udp
        network: udp

        # statsd server address
        address: 127.0.0.1:8125

        # the interval at which locally cached counters and gauges are pushed
        # to statsd; timings are pushed immediately
        writeInterval: 10s

        # prefix is prepended to all emitted statsd metrics
        prefix:
`

var ordererDefaultYaml = `# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

---
################################################################################
#
#   Orderer Configuration
#
#   - This controls the type and configuration of the orderer.
#
################################################################################
General:

    # Ledger Type: The ledger type to provide to the orderer.
    # Two non-production ledger types are provided for test purposes only:
    #  - ram: An in-memory ledger whose contents are lost on restart.
    #  - json: A simple file ledger that writes blocks to disk in JSON format.
    # Only one production ledger type is provided:
    #  - file: A production file-based ledger.
    LedgerType: file

    # Listen address: The IP on which to bind to listen.
    ListenAddress: 127.0.0.1

    # Listen port: The port on which to bind to listen.
    ListenPort: 7050

    # TLS: TLS settings for the GRPC server.
    TLS:
        Enabled: false
        # PrivateKey governs the file location of the private key of the TLS certificate.
        PrivateKey: tls/server.key
        # Certificate governs the file location of the server TLS certificate.
        Certificate: tls/server.crt
        RootCAs:
          - tls/ca.crt
        ClientAuthRequired: false
        ClientRootCAs:
    # Keepalive settings for the GRPC server.
    Keepalive:
        # ServerMinInterval is the minimum permitted time between client pings.
        # If clients send pings more frequently, the server will
        # disconnect them.
        ServerMinInterval: 60s
        # ServerInterval is the time between pings to clients.
        ServerInterval: 7200s
        # ServerTimeout is the duration the server waits for a response from
        # a client before closing the connection.
        ServerTimeout: 20s
    # Cluster settings for ordering service nodes that communicate with other ordering service nodes
    # such as Raft based ordering service.
    Cluster:
        # SendBufferSize is the maximum number of messages in the egress buffer.
        # Consensus messages are dropped if the buffer is full, and transaction
        # messages are waiting for space to be freed.
        SendBufferSize: 10
        # ClientCertificate governs the file location of the client TLS certificate
        # used to establish mutual TLS connections with other ordering service nodes.
        ClientCertificate:
        # ClientPrivateKey governs the file location of the private key of the client TLS certificate.
        ClientPrivateKey:
        # The below 4 properties should be either set together, or be unset together.
        # If they are set, then the orderer node uses a separate listener for intra-cluster
        # communication. If they are unset, then the general orderer listener is used.
        # This is useful if you want to use a different TLS server certificates on the
        # client-facing and the intra-cluster listeners.

        # ListenPort defines the port on which the cluster listens to connections.
        ListenPort:
        # ListenAddress defines the IP on which to listen to intra-cluster communication.
        ListenAddress:
        # ServerCertificate defines the file location of the server TLS certificate used for intra-cluster
        # communication.
        ServerCertificate:
        # ServerPrivateKey defines the file location of the private key of the TLS certificate.
        ServerPrivateKey:
    # Genesis method: The method by which the genesis block for the orderer
    # system channel is specified. Available options are "provisional", "file":
    #  - provisional: Utilizes a genesis profile, specified by GenesisProfile,
    #                 to dynamically generate a new genesis block.
    #  - file: Uses the file provided by GenesisFile as the genesis block.
    GenesisMethod: provisional

    # Genesis profile: The profile to use to dynamically generate the genesis
    # block to use when initializing the orderer system channel and
    # GenesisMethod is set to "provisional". See the configtx.yaml file for the
    # descriptions of the available profiles. Ignored if GenesisMethod is set to
    # "file".
    GenesisProfile: SampleInsecureSolo

    # Genesis file: The file containing the genesis block to use when
    # initializing the orderer system channel and GenesisMethod is set to
    # "file". Ignored if GenesisMethod is set to "provisional".
    GenesisFile: genesisblock

    # LocalMSPDir is where to find the private crypto material needed by the
    # orderer. It is set relative here as a default for dev environments but
    # should be changed to the real location in production.
    LocalMSPDir: msp

    # LocalMSPID is the identity to register the local MSP material with the MSP
    # manager. IMPORTANT: The local MSP ID of an orderer needs to match the MSP
    # ID of one of the organizations defined in the orderer system channel's
    # /Channel/Orderer configuration. The sample organization defined in the
    # sample configuration provided has an MSP ID of "SampleOrg".
    LocalMSPID: SampleOrg

    # Enable an HTTP service for Go "pprof" profiling as documented at:
    # https://golang.org/pkg/net/http/pprof
    Profile:
        Enabled: false
        Address: 0.0.0.0:6060

    # BCCSP configures the blockchain crypto service providers.
    BCCSP:
        # Default specifies the preferred blockchain crypto service provider
        # to use. If the preferred provider is not available, the software
        # based provider ("SW") will be used.
        # Valid providers are:
        #  - SW: a software based crypto provider
        #  - PKCS11: a CA hardware security module crypto provider.
        Default: SW

        # SW configures the software based blockchain crypto provider.
        SW:
            # TODO: The default Hash and Security level needs refactoring to be
            # fully configurable. Changing these defaults requires coordination
            # SHA2 is hardcoded in several places, not only BCCSP
            Hash: SHA2
            Security: 256
            # Location of key store. If this is unset, a location will be
            # chosen using: 'LocalMSPDir'/keystore
            FileKeyStore:
                KeyStore:

    # Authentication contains configuration parameters related to authenticating
    # client messages
    Authentication:
        # the acceptable difference between the current server time and the
        # client's time as specified in a client request message
        TimeWindow: 15m

################################################################################
#
#   SECTION: File Ledger
#
#   - This section applies to the configuration of the file or json ledgers.
#
################################################################################
FileLedger:

    # Location: The directory to store the blocks in.
    # NOTE: If this is unset, a new temporary location will be chosen every time
    # the orderer is restarted, using the prefix specified by Prefix.
    Location: /var/hyperledger/production/orderer

    # The prefix to use when generating a ledger directory in temporary space.
    # Otherwise, this value is ignored.
    Prefix: hyperledger-fabric-ordererledger

################################################################################
#
#   SECTION: RAM Ledger
#
#   - This section applies to the configuration of the RAM ledger.
#
################################################################################
RAMLedger:

    # History Size: The number of blocks that the RAM ledger is set to retain.
    # WARNING: Appending a block to the ledger might cause the oldest block in
    # the ledger to be dropped in order to limit the number total number blocks
    # to HistorySize. For example, if history size is 10, when appending block
    # 10, block 0 (the genesis block!) will be dropped to make room for block 10.
    HistorySize: 1000

################################################################################
#
#   SECTION: Kafka
#
#   - This section applies to the configuration of the Kafka-based orderer, and
#     its interaction with the Kafka cluster.
#
################################################################################
Kafka:

    # Retry: What do if a connection to the Kafka cluster cannot be established,
    # or if a metadata request to the Kafka cluster needs to be repeated.
    Retry:
        # When a new channel is created, or when an existing channel is reloaded
        # (in case of a just-restarted orderer), the orderer interacts with the
        # Kafka cluster in the following ways:
        # 1. It creates a Kafka producer (writer) for the Kafka partition that
        # corresponds to the channel.
        # 2. It uses that producer to post a no-op CONNECT message to that
        # partition
        # 3. It creates a Kafka consumer (reader) for that partition.
        # If any of these steps fail, they will be re-attempted every
        # <ShortInterval> for a total of <ShortTotal>, and then every
        # <LongInterval> for a total of <LongTotal> until they succeed.
        # Note that the orderer will be unable to write to or read from a
        # channel until all of the steps above have been completed successfully.
        ShortInterval: 5s
        ShortTotal: 10m
        LongInterval: 5m
        LongTotal: 12h
        # Affects the socket timeouts when waiting for an initial connection, a
        # response, or a transmission. See Config.Net for more info:
        # https://godoc.org/github.com/Shopify/sarama#Config
        NetworkTimeouts:
            DialTimeout: 10s
            ReadTimeout: 10s
            WriteTimeout: 10s
        # Affects the metadata requests when the Kafka cluster is in the middle
        # of a leader election.See Config.Metadata for more info:
        # https://godoc.org/github.com/Shopify/sarama#Config
        Metadata:
            RetryBackoff: 250ms
            RetryMax: 3
        # What to do if posting a message to the Kafka cluster fails. See
        # Config.Producer for more info:
        # https://godoc.org/github.com/Shopify/sarama#Config
        Producer:
            RetryBackoff: 100ms
            RetryMax: 3
        # What to do if reading from the Kafka cluster fails. See
        # Config.Consumer for more info:
        # https://godoc.org/github.com/Shopify/sarama#Config
        Consumer:
            RetryBackoff: 2s
    # Settings to use when creating Kafka topics.  Only applies when
    # Kafka.Version is v0.10.1.0 or higher
    Topic:
        # The number of Kafka brokers across which to replicate the topic
        ReplicationFactor: 3
    # Verbose: Enable logging for interactions with the Kafka cluster.
    Verbose: false

    # TLS: TLS settings for the orderer's connection to the Kafka cluster.
    TLS:

      # Enabled: Use TLS when connecting to the Kafka cluster.
      Enabled: false

      # PrivateKey: PEM-encoded private key the orderer will use for
      # authentication.
      PrivateKey:
        # As an alternative to specifying the PrivateKey here, uncomment the
        # following "File" key and specify the file name from which to load the
        # value of PrivateKey.
        #File: path/to/PrivateKey

      # Certificate: PEM-encoded signed public key certificate the orderer will
      # use for authentication.
      Certificate:
        # As an alternative to specifying the Certificate here, uncomment the
        # following "File" key and specify the file name from which to load the
        # value of Certificate.
        #File: path/to/Certificate

      # RootCAs: PEM-encoded trusted root certificates used to validate
      # certificates from the Kafka cluster.
      RootCAs:
        # As an alternative to specifying the RootCAs here, uncomment the
        # following "File" key and specify the file name from which to load the
        # value of RootCAs.
        #File: path/to/RootCAs

    # SASLPlain: Settings for using SASL/PLAIN authentication with Kafka brokers
    SASLPlain:
      # Enabled: Use SASL/PLAIN to authenticate with Kafka brokers
      Enabled: false
      # User: Required when Enabled is set to true
      User:
      # Password: Required when Enabled is set to true
      Password:

    # Kafka protocol version used to communicate with the Kafka cluster brokers
    # (defaults to 0.10.2.0 if not specified)
    Version:

################################################################################
#
#   Debug Configuration
#
#   - This controls the debugging options for the orderer
#
################################################################################
Debug:

    # BroadcastTraceDir when set will cause each request to the Broadcast service
    # for this orderer to be written to a file in this directory
    BroadcastTraceDir:

    # DeliverTraceDir when set will cause each request to the Deliver service
    # for this orderer to be written to a file in this directory
    DeliverTraceDir:

################################################################################
#
#   Operations Configuration
#
#   - This configures the operations server endpoint for the orderer
#
################################################################################
Operations:
    # host and port for the operations server
    ListenAddress: 127.0.0.1:8443

    # TLS configuration for the operations endpoint
    TLS:
        # TLS enabled
        Enabled: false

        # Certificate is the location of the PEM encoded TLS certificate
        Certificate:

        # PrivateKey points to the location of the PEM-encoded key
        PrivateKey:

        # Most operations service endpoints require client authentication when TLS
        # is enabled. ClientAuthRequired requires client certificate authentication
        # at the TLS layer to access all resources.
        ClientAuthRequired: false

        # Paths to PEM encoded ca certificates to trust for client authentication
        ClientRootCAs: []

################################################################################
#
#   Metrics  Configuration
#
#   - This configures metrics collection for the orderer
#
################################################################################
Metrics:
    # The metrics provider is one of statsd, prometheus, or disabled
    Provider: disabled

    # The statsd configuration
    Statsd:
      # network type: tcp or udp
      Network: udp

      # the statsd server address
      Address: 127.0.0.1:8125

      # The interval at which locally cached counters and gauges are pushed
      # to statsd; timings are pushed immediately
      WriteInterval: 30s

      # The prefix is prepended to all emitted statsd metrics
      Prefix:

################################################################################
#
#   Consensus Configuration
#
#   - This section contains config options for a consensus plugin. It is opaque
#     to orderer, and completely up to consensus implementation to make use of.
#
################################################################################
Consensus:
    # The allowed key-value pairs here depend on consensus plugin. For etcd/raft,
    # we use following options:

    # WALDir specifies the location at which Write Ahead Logs for etcd/raft are
    # stored. Each channel will have its own subdir named after channel ID.
    WALDir: /var/hyperledger/production/orderer/etcdraft/wal

    # SnapDir specifies the location at which snapshots for etcd/raft are
    # stored. Each channel will have its own subdir named after channel ID.
    SnapDir: /var/hyperledger/production/orderer/etcdraft/snapshot
`