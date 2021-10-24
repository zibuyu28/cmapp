package fabric

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"

	"github.com/zibuyu28/cmapp/crobot/pkg"
	"google.golang.org/grpc/metadata"
)

var mockFabric = Fabric{
	Name:        "mock-fab",
	UUID:        "mock-uuid",
	Version:     "1.4.1",
	Consensus:   "etcdraft",
	CertGenType: "bin-tool",
	Channels: []Channel{
		{
			Name:                     "mock-channel",
			UUID:                     "channel",
			OrdererBatchTimeout:      "2s",
			OrdererMaxMessageCount:   500,
			OrdererAbsoluteMaxBytes:  "99MB",
			OrdererPreferredMaxBytes: "512KB",
		},
	},
	TLSEnable: true,
	Orderers: []Orderer{
		{
			Name:       "orderer0",
			UUID:       "orderer0",
			MachineID:  1,
			GRPCPort:   7050,
			HealthPort: 8443,
			Tag:        "mock-tag-orderer0",
		},
		{
			Name:       "orderer1",
			UUID:       "orderer1",
			MachineID:  1,
			GRPCPort:   7051,
			HealthPort: 8444,
			Tag:        "mock-tag-orderer1",
		},
		{
			Name:       "orderer2",
			UUID:       "orderer2",
			MachineID:  1,
			GRPCPort:   7052,
			HealthPort: 8445,
			Tag:        "mock-tag-orderer2",
		},
	},
	Peers: []Peer{
		{
			Name:                "mock-peer0",
			UUID:                "mock-peer0",
			MachineID:           1,
			GRPCPort:            7053,
			ChainCodeListenPort: 7054,
			EventPort:           7055,
			HealthPort:          8446,
			Organization: Organization{
				Name: "Org1",
				UUID: "Org1",
			},
			AnchorPeer: true,
			Tag:        "mock-tag-peer0",
		},
	},
}

type FabricDriver struct {
	BaseDriver pkg.BaseDriver
}

func (f *FabricDriver) InitChain(ctx context.Context, _ *driver.Empty) (*driver.Chain, error) {
	md, b := metadata.FromIncomingContext(ctx)
	if !b {
		return nil, errors.New("get context info from context")
	}

	// get info from context
	fmt.Printf("md : [%v]", md)

	cb, err := json.Marshal(mockFabric.Channels)
	if err != nil {
		return nil, errors.Wrap(err, "marshal channels")
	}
	rc := driver.Chain{
		Name:     mockFabric.Name,
		UUID:     mockFabric.UUID,
		Type:     "fabric",
		Version:  mockFabric.Version,
		State:    driver.Chain_Handling,
		DriverID: 1,
		Tags: []string{
			"mock-tag-chain",
		},
		CustomInfo: map[string]string{
			"Consensus":   string(mockFabric.Consensus),
			"TLSEnable":   fmt.Sprintf("%t", mockFabric.TLSEnable),
			"CertGenType": string(mockFabric.CertGenType),
			"Channels":    string(cb),
		},
		Nodes: []*driver.Node{},
	}
	var ns []*driver.Node
	for _, o := range mockFabric.Orderers {
		n := &driver.Node{
			Name:      o.Name,
			UUID:      o.UUID,
			Type:      "ORDERER",
			State:     driver.Node_Handling,
			Message:   fmt.Sprintf("init node %s", o.Name),
			MachineID: int32(o.MachineID),
			Tags:      []string{o.Tag},
			CustomInfo: map[string]string{
				"GRPCPort":   fmt.Sprintf("%d", o.GRPCPort),
				"HealthPort": fmt.Sprintf("%d", o.HealthPort),
			},
		}
		ns = append(ns, n)
	}
	for _, p := range mockFabric.Peers {
		ob, err := json.Marshal(p.Organization)
		if err != nil {
			return nil, errors.Wrap(err, "marshal organizaton")
		}
		n := &driver.Node{
			Name:      p.Name,
			UUID:      p.UUID,
			Type:      "PEER",
			State:     driver.Node_Handling,
			Message:   fmt.Sprintf("init node %s", p.Name),
			MachineID: int32(p.MachineID),
			Tags:      []string{p.Tag},
			CustomInfo: map[string]string{
				"GRPCPort":            fmt.Sprintf("%d", p.GRPCPort),
				"HealthPort":          fmt.Sprintf("%d", p.HealthPort),
				"ChainCodeListenPort": fmt.Sprintf("%d", p.ChainCodeListenPort),
				"EventPort":           fmt.Sprintf("%d", p.EventPort),
				"Organitaion":         string(ob),
				"AnchorPeer":          fmt.Sprintf("%t", p.AnchorPeer),
			},
		}
		ns = append(ns, n)
	}
	rc.Nodes = ns

	return &rc, nil
}

func (f *FabricDriver) CreateChainExec(ctx context.Context, c *driver.Chain) (*driver.Empty, error) {
	// TODO: 开始创建链
	/*
			0. 生成证书
			0.5. 将证书上传到core
			1. 新建一个app
			2. app 增加tag
			3. app 增加预置文件
			4. app 设置cpu和memory
			5. app 设置健康检查
			6. app 设置环境变量
			7. app 设置端口
			8. 启动app

		  --- 节点启动之后

			0. 创建通道
			1. 加入通道
	*/


	panic("implement me")
}

func (f *FabricDriver) Exit(ctx context.Context, c *driver.Empty) (*driver.Empty, error) {
	panic("implement me")
}
