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

package fabric

import (
	"context"
	"encoding/json"
	"fmt"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/process"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"os"
	"path/filepath"
	"strconv"

	"github.com/zibuyu28/cmapp/crobot/pkg"
	"google.golang.org/grpc/metadata"
)

var mockFabric = model.Fabric{
	Name:        "mock-fab",
	UUID:        "mock-uuid",
	Version:     "1.4",
	Consensus:   model.Solo,
	CertGenType: model.CertBinaryTool,
	Channels: []model.Channel{
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
	Orderers: []model.Orderer{
		{
			Name:       "orderer0",
			UUID:       "orderer0",
			MachineID:  46,
			GRPCPort:   7050,
			HealthPort: 8443,
			Tag:        "mock-tag-orderer0",
			LogLevel: "INFO",
		},
	},
	Peers: []model.Peer{
		{
			Name:                "mock-peer0",
			UUID:                "mock-peer0",
			MachineID:           46,
			GRPCPort:            7053,
			ChainCodeListenPort: 7054,
			EventPort:           7055,
			HealthPort:          8446,
			Organization: model.Organization{
				Name: "Org1",
				UUID: "Org1",
			},
			AnchorPeer: true,
			Tag:        "mock-tag-peer0",
			RMTDocker:  "tcp://192.168.0.104:2375",
			LogLevel: "INFO",
		},
	},
}

const (
	PluginEnvDriverName    = "PLUGIN_DRIVER_NAME"
	PluginEnvDriverVersion = "PLUGIN_DRIVER_VERSION"
	PluginEnvDriverID      = "PLUGIN_DRIVER_ID"
)

type FabricDriver struct {
	pkg.BaseDriver
}

func (f *FabricDriver) GetCreateFlags(ctx context.Context, empty *driver.Empty) (*driver.Flags, error) {
	baseFlags := &driver.Flags{Flags: f.GetFlags()}
	return baseFlags, nil
}

func (f *FabricDriver) SetConfigFromFlags(ctx context.Context, flags *driver.Flags) (*driver.Empty, error) {
	m := f.ConvertFlags(flags)
	f.CoreAddr = m["CoreAddr"]
	f.CoreHTTPAddr = m["CoreHTTPAddr"]
	f.CoreGRPCAddr = m["CoreGRPCAddr"]
	f.ImageRepository.Repository = m["Repository"]
	f.ImageRepository.StorePath = m["StorePath"]

	validate := v.New()
	err := validate.Struct(f)
	if err != nil {
		return nil, errors.Wrap(err, "validate param")
	}

	driverName := os.Getenv(PluginEnvDriverName)
	if len(driverName) == 0 {
		return nil, errors.Errorf("fail to get driver name from env, please check env [%s]", PluginEnvDriverName)
	}

	driverVersion := os.Getenv(PluginEnvDriverVersion)
	if len(driverVersion) == 0 {
		return nil, errors.Errorf("fail to get driver version from env, please check env [%s]", PluginEnvDriverVersion)
	}

	driverIDStr := os.Getenv(PluginEnvDriverID)
	if len(driverIDStr) == 0 {
		return nil, errors.Errorf("fail to get driver id from env, please check env [%s]", PluginEnvDriverID)
	}

	driverID, err := strconv.Atoi(driverIDStr)
	if err != nil {
		return nil, errors.Errorf("fail to parse driver id by driverStr [%s], please check env [%s]", driverIDStr, PluginEnvDriverID)
	}

	f.DriverName = driverName
	f.DriverVersion = driverVersion
	f.DriverID = driverID

	return &driver.Empty{}, nil
}

func (f *FabricDriver) InitChain(ctx context.Context, _ *driver.Empty) (*driver.Chain, error) {
	md, b := metadata.FromIncomingContext(ctx)
	if !b {
		return nil, errors.New("get context info from context")
	}

	// get info from context
	fmt.Printf("md : [%v]", md)
	data := md.Get("uuid")
	if len(data) == 0 {
		return nil, errors.New("uuid is nil")
	}

	cb, err := json.Marshal(mockFabric.Channels)
	if err != nil {
		return nil, errors.Wrap(err, "marshal channels")
	}
	rc := driver.Chain{
		Name:     mockFabric.Name,
		UUID:     data[0],
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

func (f *FabricDriver) CreateChainExec(ctx context.Context, c *driver.Chain) (*driver.Chain, error) {
	baseDir := fmt.Sprintf("%s/workspace/%s", filepath.Dir(os.Args[0]), c.UUID)
	err := os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "mkdir for basedir [%s]", baseDir)
	}
	worker := process.NewCreateChainWorker(ctx, f.CoreHTTPAddr, baseDir)

	marshal, _ := json.Marshal(mockFabric)
	log.Infof(ctx, "mock fab : [%s]", string(marshal))
	err = worker.CreateChainProcess(&mockFabric)
	if err != nil {
		return nil, errors.Wrap(err, "create chain process")
	}
	return c, nil
}

func (f *FabricDriver) Exit(ctx context.Context, c *driver.Empty) (*driver.Empty, error) {
	panic("implement me")
}
