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

package cengine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/plugin/localbinary"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"

	coreproto "github.com/zibuyu28/cmapp/core/proto/ch_manager"
	"google.golang.org/grpc"
)

type InitInfo struct {
	DriverName    string `json:"driver_name" validate:"required"`
	DriverVersion string `json:"driver_version" validate:"required"`
	DriverID      int    `json:"driver_id" validate:"required"`
	CoreGRPCPort  int    `json:"core_grpc_port" validate:"required"`
}

// TODO: 如果需要穿参数，肯定是从这里传入
// CreateChain create chain action
func CreateChain(ctx context.Context, info InitInfo) error {
	log.Debugf(ctx, "create chain process")
	// 初始化链的参数
	err := validator.New().Struct(&info)
	if err != nil {
		return errors.Wrap(err, "check param")
	}
	log.Debugf(ctx, "start driver [%s] local", info.DriverName)
	var cmIns driver.ChainDriverClient
	cmIns, err = getCEnginePluginInstance(ctx, info.DriverID, info.DriverName, info.DriverVersion)
	if err != nil {
		return errors.Wrap(err, "get chain engine plugin client")
	}
	log.Debugf(ctx, "get connect with core [:%d]", info.CoreGRPCPort)
	corecli, err := getCoreGrpcClient(ctx, info.CoreGRPCPort)
	if err != nil {
		return errors.Wrap(err, "get core grpc client")
	}
	log.Debugf(ctx, "init chain")
	chain, err := cmIns.InitChain(ctx, &driver.Empty{})
	if err != nil {
		return errors.Wrap(err, "init chain")
	}
	marshal, _ := json.Marshal(chain)
	log.Debugf(ctx, "chain info [%s]", string(marshal))

	tc := coreproto.TypedChain{
		Name:       chain.Name,
		UUID:       chain.UUID,
		Type:       chain.Type,
		Version:    chain.Version,
		State:      coreproto.TypedChain_StateE(chain.State),
		DriverID:   int32(info.DriverID),
		Tags:       chain.Tags,
		CustomInfo: chain.CustomInfo,
	}

	log.Debugf(ctx, "report chain to core")
	_, err = corecli.ReportChain(ctx, &tc)
	if err != nil {
		return errors.Wrap(err, "report chain")
	}

	if chain.Nodes != nil && len(chain.Nodes) != 0 {
		var tns  []*coreproto.TypedNode
		for _, node := range chain.Nodes {
			tn := coreproto.TypedNode{
				Name:       node.Name,
				UUID:       node.UUID,
				Type:       node.Type,
				State:      coreproto.TypedNode_StateE(node.State),
				Message:    node.Message,
				MachineID:  node.MachineID,
				ChainID:    node.ChainID,
				Tags:       node.Tags,
				CustomInfo: node.CustomInfo,
			}
			tns = append(tns, &tn)
		}
		_, err := corecli.ReportNodes(ctx, &coreproto.TypedNodes{Nodes: tns})
		if err != nil {
			return errors.Wrap(err, "report nodes")
		}
	}

	log.Debugf(ctx, "driver execute create chain")
	_, err = cmIns.CreateChainExec(ctx, chain)
	if err != nil {
		return errors.Wrap(err, "create chain exec")
	}
	log.Debugf(ctx, "create chain success")
	return nil
}

func getCoreGrpcClient(ctx context.Context, grpcPort int) (coreproto.ChainManageClient, error) {
	// get grpc connect
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(10))
	defer cancel()
	// grpc.WithBlock() : use to make sure the connection is up
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("127.0.0.1:%d", grpcPort), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, "conn grpc")
	}

	return coreproto.NewChainManageClient(conn), nil
}

// getCEnginePluginInstance
func getCEnginePluginInstance(ctx context.Context, driverID int, driverName, driverVersion string) (driver.ChainDriverClient, error) {
	// 启动 plugin
	plugin, err := localbinary.NewPlugin(ctx, driverID, driverName, driverVersion)
	if err != nil {
		return nil, errors.Wrap(err, "new plugin")
	}
	go func() {
		if err = plugin.Serve(); err != nil {
			// TODO: Is this best approach?
			log.Warn(ctx, err.Error())
			return
		}

	}()

	address, err := plugin.Address()
	if err != nil {
		return nil, errors.Wrap(err, "get plugin serve address")
	}

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, "create grpc connection")
	}
	return driver.NewChainDriverClient(conn), nil
}
