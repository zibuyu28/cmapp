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
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/fabtool"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ConfigWorker struct {
	BaseWorker
}

// NewConfigWorker new config worker
func NewConfigWorker(workDir string) *ConfigWorker {
	return &ConfigWorker{
		BaseWorker: BaseWorker{
			workDir: workDir,
		},
	}
}

//Configtx tx实体
type Configtx struct {
	Organizations []Organization `yaml:"-"`
	Capabilities  Capabilities   `yaml:"-"`
	Application   Application    `yaml:"-"`
	Orderer       OrdererConf    `yaml:"-"`
	Channel       Channel        `yaml:"-"`
	Profiles      Profiles       `yaml:"Profiles"`
}

// Organization .
type Organization struct {
	Name        string       `yaml:"Name"`
	ID          string       `yaml:"ID"`
	MSPDir      string       `yaml:"MSPDir"`
	Policies    Policies     `yaml:"Policies"`
	AnchorPeers []AnchorPeer `yaml:"AnchorPeers"`
}

// Policies .
type Policies struct {
	Readers TypeRule `yaml:"Readers"`
	Writers TypeRule `yaml:"Writers"`
	Admins  TypeRule `yaml:"Admins"`
}

// AnchorPeer .
type AnchorPeer struct {
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

// TypeRule .
type TypeRule struct {
	Type string `yaml:"Type"`
	Rule string `yaml:"Rule"`
}

// Capabilities .
type Capabilities struct {
	Channel     ChannelCapabilities     `yaml:"Channel"`
	Orderer     OrdererCapabilities     `yaml:"Orderer"`
	Application ApplicationCapabilities `yaml:"Application"`
}

// ChannelCapabilities .
type ChannelCapabilities struct {
	V1_3 bool `yaml:"V1_3"`
}

// OrdererCapabilities .
type OrdererCapabilities struct {
	V1_1 bool `yaml:"V1_1"`
}

// ApplicationCapabilities .
type ApplicationCapabilities struct {
	V1_3 bool `yaml:"V1_3"`
	V1_2 bool `yaml:"V1_2"`
	V1_1 bool `yaml:"V1_1"`
}

// Application .
type Application struct {
	Organizations string                  `yaml:"Organizations"`
	Policies      Policies                `yaml:"Policies"`
	Capabilities  ApplicationCapabilities `yaml:"Capabilities"`
}

// Orderer .
type OrdererConf struct {
	OrdererType  string   `yaml:"OrdererType"`
	Addresses    []string `yaml:"Addresses"`
	BatchTimeout string   `yaml:"BatchTimeout"`
	BatchSize    struct {
		MaxMessageCount   int    `yaml:"MaxMessageCount"`
		AbsoluteMaxBytes  string `yaml:"AbsoluteMaxBytes"`
		PreferredMaxBytes string `yaml:"PreferredMaxBytes"`
	} `yaml:"BatchSize"`
	Kafka struct {
		Brokers []string `yaml:"Brokers"`
	} `yaml:"Kafka"`
	EtcdRaft struct {
		Consenters []Consenter `yaml:"Consenters"`
	} `yaml:"EtcdRaft"`
	Organizations string          `yaml:"Organizations"`
	Policies      OrdererPolicies `yaml:"Policies"`
}

// Consenter .
type Consenter struct {
	Host          string `yaml:"Host"`
	Port          int    `yaml:"Port"`
	ClientTLSCert string `yaml:"ClientTLSCert"`
	ServerTLSCert string `yaml:"ServerTLSCert"`
}

// OrdererPolicies .
type OrdererPolicies struct {
	Readers         TypeRule `yaml:"Readers"`
	Writers         TypeRule `yaml:"Writers"`
	Admins          TypeRule `yaml:"Admins"`
	BlockValidation TypeRule `yaml:"BlockValidation"`
}

// Channel .
type Channel struct {
	Policies     Policies            `yaml:"Policies"`
	Capabilities ChannelCapabilities `yaml:"Capabilities"`
}

// Profiles .
type Profiles struct {
	OrdererGenesis OrdererGenesis `yaml:"OrdererGenesis"`
	OrgsChannel    OrgsChannel    `yaml:"OrgsChannel"`
}

// NewChannelProfiles .
type NewChannelProfiles struct {
	OrgsChannel OrgsChannel `yaml:"OrgsChannel"`
}

// OrdererGenesis .
type OrdererGenesis struct {
	Policies     Policies            `yaml:"Policies"`
	Capabilities ChannelCapabilities `yaml:"Capabilities"`
	Orderer      OgOrderer           `yaml:"Orderer"`
	Consortiums  Consortiums         `yaml:"Consortiums"`
}

// OgOrderer .
type OgOrderer struct {
	OrdererType  string   `yaml:"OrdererType"`
	Addresses    []string `yaml:"Addresses"`
	BatchTimeout string   `yaml:"BatchTimeout"`
	BatchSize    struct {
		MaxMessageCount   int    `yaml:"MaxMessageCount"`
		AbsoluteMaxBytes  string `yaml:"AbsoluteMaxBytes"`
		PreferredMaxBytes string `yaml:"PreferredMaxBytes"`
	} `yaml:"BatchSize"`
	Kafka struct {
		Brokers []string `yaml:"Brokers"`
	} `yaml:"Kafka"`
	EtcdRaft struct {
		Consenters []Consenter `yaml:"Consenters"`
	} `yaml:"EtcdRaft"`
	Policies      OrdererPolicies     `yaml:"Policies"`
	Organizations []Organization      `yaml:"Organizations"`
	Capabilities  OrdererCapabilities `yaml:"Capabilities"`
}

// Consortiums .
type Consortiums struct {
	SampleConsortium struct {
		Organizations []Organization `yaml:"Organizations"`
	} `yaml:"SampleConsortium"`
}

// OrgsChannel .
type OrgsChannel struct {
	Consortium  string        `yaml:"Consortium"`
	Application OcApplication `yaml:"Application"`
}

// OcApplication .
type OcApplication struct {
	Policies      Policies                `yaml:"Policies"`
	Organizations []Organization          `yaml:"Organizations"`
	Capabilities  ApplicationCapabilities `yaml:"Capabilities"`
}

// CryptoConfig crypto-feconfig.yaml
type CryptoConfig struct {
	OrdererOrgs []OrdererOrg `yaml:"OrdererOrgs"`
	PeerOrgs    []PeerOrg    `yaml:"PeerOrgs"`
}

// OrdererOrg .
type OrdererOrg struct {
	Name   string `yaml:"Name"`
	Domain string `yaml:"Domain"`
	CA     struct {
		Country  string `yaml:"Country"`
		Province string `yaml:"Province"`
		Locality string `yaml:"Locality"`
	} `yaml:"CA"`
	Template struct {
		Count int `yaml:"Count"`
	} `yaml:"Template"`
}

// PeerOrg .
type PeerOrg struct {
	Name   string `yaml:"Name"`
	Domain string `yaml:"Domain"`
	CA     struct {
		Country  string `yaml:"Country"`
		Province string `yaml:"Province"`
		Locality string `yaml:"Locality"`
	} `yaml:"CA"`
	Template struct {
		Count int `yaml:"Count"`
	} `yaml:"Template"`
	Users struct {
		Count int `yaml:"Count"`
	} `yaml:"Users"`
	EnableNodeOUs bool `yaml:"EnableNodeOUs"`
}

//configtx文件下的配置
const (
	OrdererSuffix      = "orderer"
	OrdererMsp         = "OrdererMSP"
	KafkaSuffix        = "kafka"
	TypeImplicitMeta   = "ImplicitMeta"
	TypeSignature      = "Signature"
	RuleAnyReaders     = "ANY Readers"
	RuleAnyWriters     = "ANY Writers"
	RuleMajorityAdmins = "MAJORITY Admins"
	Country            = "CN"
	Province           = "GuangDong"
	Locality           = "GuangZhou"
)

type ConsensusType string

const (
	ConsensusSolo     ConsensusType = "solo"
	ConsensusKafka    ConsensusType = "kafka"
	ConsensusEtcdRaft ConsensusType = "etcdraft"
)

func (c *ConfigWorker) InitTxFile(ctx context.Context, chain *model.Fabric) error {
	caFileRootPath := c.workDir

	configtx := &Configtx{}
	err := setOrganizations(chain, configtx, caFileRootPath)
	if err != nil {
		return errors.Wrap(err, "check organization")
	}
	setCapabilities(configtx)
	setApplication(configtx)
	err = setOrderer(chain, configtx, caFileRootPath)
	if err != nil {
		return errors.Wrap(err, "set orderer")
	}
	setChannel(configtx)
	setProfiles(configtx)
	tx, err := yaml.Marshal(*configtx)
	if err != nil {
		return errors.Wrap(err, "marshal config to yaml")
	}
	err = ioutil.WriteFile(filepath.Join(caFileRootPath, "configtx.yaml"), tx, os.ModePerm)
	return err
}

func (c *ConfigWorker) GenesisBlock(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "configtxgen", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init configtxgen tool")
	}
	err = itool.(fabtool.ConfigtxgenTool).GenerateGenesisBlock(chain, c.workDir)
	if err != nil {
		return errors.Wrap(err, "generate genesis block")
	}
	return nil
}

func (c *ConfigWorker) AllChannelTxFile(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "configtxgen", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init configtxgen tool")
	}
	// 生成 channel 交易
	err = itool.(fabtool.ConfigtxgenTool).GenerateAllChainChannelTx(chain, c.workDir)
	if err != nil {
		return errors.Wrap(err, "generate channel tx")
	}
	return nil
}

func (c *ConfigWorker) AnchorPeerArtifacts(ctx context.Context, chain *model.Fabric) error {
	itool, err := fabtool.NewTool(ctx, "configtxgen", chain.Version)
	if err != nil {
		return errors.Wrap(err, "init configtxgen tool")
	}
	// 生成 anchor peer 交易
	err = itool.(fabtool.ConfigtxgenTool).GenerateAnchorPeerArtifacts(chain, c.workDir)
	if err != nil {
		return errors.Wrap(err, "generate anchor peer artifact tx")
	}
	return nil
}

//setProfiles 设置Profile
func setProfiles(configtx *Configtx) {
	//OrdererGenesis
	ordererGenesis := OrdererGenesis{}
	peerOrgs := configtx.Organizations[1:]
	sampleConsortium := struct {
		Organizations []Organization `yaml:"Organizations"`
	}{
		Organizations: peerOrgs,
	}
	consortiums := Consortiums{
		SampleConsortium: sampleConsortium,
	}
	//OrdererGenesis.Consortiums
	ordererGenesis.Consortiums = consortiums
	ordererGenesis.Policies = configtx.Channel.Policies
	ordererGenesis.Capabilities = configtx.Channel.Capabilities

	orderOrg := make([]Organization, 1)
	orderOrg[0] = configtx.Organizations[0]

	order := OgOrderer{
		OrdererType:   configtx.Orderer.OrdererType,
		Policies:      configtx.Orderer.Policies,
		Kafka:         configtx.Orderer.Kafka,
		EtcdRaft:      configtx.Orderer.EtcdRaft,
		Addresses:     configtx.Orderer.Addresses,
		BatchSize:     configtx.Orderer.BatchSize,
		BatchTimeout:  configtx.Orderer.BatchTimeout,
		Organizations: orderOrg,
		Capabilities:  configtx.Capabilities.Orderer,
	}

	//OrdererGenesis.Orderer
	ordererGenesis.Orderer = order
	//OrgsChannel
	//OrgsChannel.Application
	application := OcApplication{
		Policies:      configtx.Application.Policies,
		Capabilities:  configtx.Capabilities.Application,
		Organizations: peerOrgs,
	}

	orgsChannel := OrgsChannel{
		Consortium:  "SampleConsortium",
		Application: application,
	}
	// Profiles
	profiles := Profiles{
		OrdererGenesis: ordererGenesis,
		OrgsChannel:    orgsChannel,
	}
	configtx.Profiles = profiles
}

//setChannel 设置channel
func setChannel(configtx *Configtx) {
	channel := Channel{
		Policies: Policies{
			Readers: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleAnyReaders,
			},
			Writers: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleAnyWriters,
			},
			Admins: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleMajorityAdmins,
			},
		},
		Capabilities: configtx.Capabilities.Channel,
	}
	configtx.Channel = channel
}

//setOrderer 设置Orderer
func setOrderer(chain *model.Fabric, configtx *Configtx, caFileRootPath string) error {
	if len(chain.Channels) == 0 {
		return errors.New("channel info is empty")
	}
	orderer := OrdererConf{
		OrdererType:  string(chain.Consensus),
		BatchTimeout: chain.Channels[0].OrdererBatchTimeout,
		BatchSize: struct {
			MaxMessageCount   int    `yaml:"MaxMessageCount"`
			AbsoluteMaxBytes  string `yaml:"AbsoluteMaxBytes"`
			PreferredMaxBytes string `yaml:"PreferredMaxBytes"`
		}{
			MaxMessageCount:   chain.Channels[0].OrdererMaxMessageCount,
			AbsoluteMaxBytes:  chain.Channels[0].OrdererAbsoluteMaxBytes,
			PreferredMaxBytes: chain.Channels[0].OrdererPreferredMaxBytes,
		},
	}
	orderer.Policies = OrdererPolicies{
		Readers: TypeRule{
			Type: TypeImplicitMeta,
			Rule: RuleAnyReaders,
		},
		Writers: TypeRule{
			Type: TypeImplicitMeta,
			Rule: RuleAnyWriters,
		},
		Admins: TypeRule{
			Type: TypeImplicitMeta,
			Rule: RuleMajorityAdmins,
		},
		BlockValidation: TypeRule{
			Type: TypeImplicitMeta,
			Rule: RuleAnyWriters,
		},
	}
	switch ConsensusType(chain.Consensus) {
	case ConsensusSolo:
		if len(chain.Orderers) == 0 {
			return errors.New("solo consensus type must has one orderer")
		}
		//var domain string
		for _, net := range chain.Orderers[0].APP.Networks {
			if net.PortInfo.Port == chain.Orderers[0].GRPCPort {
				for _, s := range net.RouteInfo {
					if s.RouteType == ag.OUT {
						orderer.Addresses = []string{s.Router}
					}
				}
			}
		}
		//domain := fmt.Sprintf("%s:%d", chain.Orderers[0].NodeHostName, chain.Orderers[0].GRPCPort)
		//domain := chain.Orderers[0].NodeHostName + ":7050"
		//orderer.Addresses = []string{domain}
	case ConsensusKafka:
		return errors.New("not support consensus type kafka")
	case ConsensusEtcdRaft:
		domains := make([]string, 0)
		consenters := make([]Consenter, 0)
		for _, o := range chain.Orderers {
			//ordererAddr := fmt.Sprintf("%s:%d", o.NodeHostName, o.GRPCPort)
			var port int = 80
			var host string = o.NodeHostName
			for _, net := range o.APP.Networks {
				if net.PortInfo.Port == o.GRPCPort {
					for _, s := range net.RouteInfo {
						if s.RouteType == ag.OUT {
							host = s.Router
							domains = append(domains, s.Router)
							split := strings.Split(s.Router, ":")
							if len(split) == 2 {
								atoi, err := strconv.Atoi(split[1])
								if err != nil {
									return errors.Errorf("grpc router [%s] is not correct", s.Router)
								}
								port = atoi
								host = split[0]
							}
						}
					}
				}
			}
			//domains = append(domains, ordererAddr)
			tls := fmt.Sprintf("%s/ordererOrganizations/orderer.zibuyufab.cn/orderers/%s.orderer.zibuyufab.cn/tls/signcerts/cert.pem", caFileRootPath, o.NodeHostName)
			consenter := Consenter{
				Host:          host,
				Port:          port,
				ClientTLSCert: tls,
				ServerTLSCert: tls,
			}
			consenters = append(consenters, consenter)
		}
		orderer.Addresses = domains
		orderer.EtcdRaft = struct {
			Consenters []Consenter `yaml:"Consenters"`
		}{
			Consenters: consenters,
		}
	}
	configtx.Orderer = orderer
	return nil
}

//setApplication 设置Application
func setApplication(configtx *Configtx) {
	application := Application{
		Organizations: "",
		Policies: Policies{
			Readers: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleAnyReaders,
			},
			Writers: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleAnyWriters,
			},
			Admins: TypeRule{
				Type: TypeImplicitMeta,
				Rule: RuleMajorityAdmins,
			},
		},
		Capabilities: configtx.Capabilities.Application,
	}
	configtx.Application = application
}

//setCapabilities 设置capabilitie
func setCapabilities(configtx *Configtx) {
	capabilities := Capabilities{
		Channel: ChannelCapabilities{
			V1_3: true,
		},
		Orderer: OrdererCapabilities{
			V1_1: true,
		},
		Application: ApplicationCapabilities{
			V1_3: true,
			V1_2: false,
			V1_1: false,
		},
	}
	configtx.Capabilities = capabilities
}

//func setTargetOrganizations(organizations []ChainOrganization, peers []Node, configtx *Configtx, caFileRootPath string) error {
//	orgs := make([]Organization, 0)
//	for _, info := range organizations {
//		if info.OrganizationType != PeerOrganization {
//			continue
//		}
//		for _, peer := range peers {
//			if peer.OrgID != info.ID || !peer.AnchorPeer {
//				continue
//			}
//			name := info.UUID
//			if peer.NodeHostName == "" {
//				err := fmt.Errorf("peer(%d) Host ip is empty", peer.ID)
//				log.Error(err.Error())
//				return err
//			}
//			peerOrg := Organization{
//				Name:   name,
//				ID:     name + MspSuf,
//				MSPDir: filepath.Join(caFileRootPath, "peerOrganizations/"+info.UUID+".fabric.com", "msp"),
//				Policies: Policies{
//					Readers: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin', '" + name + MspSuf + ".peer', '" + name + MspSuf + ".client')",
//					},
//					Writers: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin', '" + name + MspSuf + ".client')",
//					},
//					Admins: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin')",
//					},
//				},
//				AnchorPeers: []AnchorPeer{
//					{
//						Host: peer.NodeHostName,
//						Port: 7051,
//					},
//				},
//			}
//			orgs = append(orgs, peerOrg)
//		}
//	}
//	configtx.Organizations = orgs
//	return nil
//}
//
//func setNewOrganizations(newOrg *NewOrg, configtx *Configtx, caFileRootPath string) error {
//	orgs := make([]Organization, 0)
//	for _, info := range newOrg.OrganizationInfos {
//		if info.OrganizationType != PeerOrganization {
//			continue
//		}
//		for _, peer := range newOrg.Peers {
//			if peer.OrgID != info.ID || !peer.AnchorPeer {
//				continue
//			}
//			name := info.UUID
//			if peer.NodeHostName == "" {
//				err := fmt.Errorf("peer(%d) Host ip is empty", peer.ID)
//				log.Error(err.Error())
//				return err
//			}
//			peerOrg := Organization{
//				Name:   name,
//				ID:     name + MspSuf,
//				MSPDir: filepath.Join(caFileRootPath, "peerOrganizations/"+info.UUID+".fabric.com", "msp"),
//				Policies: Policies{
//					Readers: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin', '" + name + MspSuf + ".peer', '" + name + MspSuf + ".client')",
//					},
//					Writers: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin', '" + name + MspSuf + ".client')",
//					},
//					Admins: TypeRule{
//						Type: TypeSignature,
//						Rule: "OR('" + name + MspSuf + ".admin')",
//					},
//				},
//				AnchorPeers: []AnchorPeer{
//					{
//						Host: peer.NodeHostName,
//						Port: 7051,
//					},
//				},
//			}
//			orgs = append(orgs, peerOrg)
//		}
//	}
//	configtx.Organizations = orgs
//	return nil
//}

//setOrganizations 设置organization
func setOrganizations(chain *model.Fabric, configtx *Configtx, caFileRootPath string) error {
	orgs := make([]Organization, 0)
	orderOrg := Organization{
		Name:   "Orderer",
		ID:     OrdererMsp,
		MSPDir: caFileRootPath + "/ordererOrganizations/orderer.zibuyufab.cn/msp",
		Policies: Policies{
			Readers: TypeRule{
				Type: TypeSignature,
				Rule: "OR('" + OrdererMsp + ".member')",
			},
			Writers: TypeRule{
				Type: TypeSignature,
				Rule: "OR('" + OrdererMsp + ".member')",
			},
			Admins: TypeRule{
				Type: TypeSignature,
				Rule: "OR('" + OrdererMsp + ".admin')",
			},
		},
	}

	orgs = append(orgs, orderOrg)

	var po = make(map[string]struct{})
	for _, peer := range chain.Peers {
		if len(peer.Organization.UUID) == 0 {
			return errors.Errorf("peer [%s] organization uuid is empty", peer.Organization.UUID)
		}
		if _, ok := po[peer.Organization.UUID]; !ok && peer.AnchorPeer {
			po[peer.Organization.UUID] = struct{}{}
			if len(peer.NodeHostName) == 0 {
				return errors.Errorf("peer[%s] host name is empty", peer.Name)
			}
			var mspT = fmt.Sprintf("%sMSP", peer.Organization.UUID)
			readRule := fmt.Sprintf("OR('%s.admin', '%s.peer', '%s.client')", mspT, mspT, mspT)
			writeRule := fmt.Sprintf("OR('%s.admin', '%s.client')", mspT, mspT)
			adminRule := fmt.Sprintf("OR('%s.admin')", mspT)
			var port int = 80
			var host string = peer.NodeHostName
			for _, network := range peer.APP.Networks {
				if network.PortInfo.Port == peer.GRPCPort {
					for _, s := range network.RouteInfo {
						if s.RouteType == ag.OUT {
							host = s.Router
							split := strings.Split(s.Router, ":")
							if len(split) == 2 {
								atoi, err := strconv.Atoi(split[1])
								if err != nil {
									return errors.Errorf("grpc router [%s] is not correct", s.Router)
								}
								port = atoi
								host = split[0]
							}
						}
					}
				}
			}
			peerOrg := Organization{
				Name:   peer.Organization.UUID,
				ID:     fmt.Sprintf("%sMSP", peer.Organization.UUID),
				MSPDir: filepath.Join(caFileRootPath, fmt.Sprintf("peerOrganizations/%s.zibuyufab.cn", peer.Organization.UUID), "msp"),
				Policies: Policies{
					Readers: TypeRule{
						Type: TypeSignature,
						Rule: readRule,
					},
					Writers: TypeRule{
						Type: TypeSignature,
						Rule: writeRule,
					},
					Admins: TypeRule{
						Type: TypeSignature,
						Rule: adminRule,
					},
				},
				AnchorPeers: []AnchorPeer{
					{
						Host: host,
						Port: port,
					},
				},
			}
			orgs = append(orgs, peerOrg)
		}
	}
	configtx.Organizations = orgs
	return nil
}
