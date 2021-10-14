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
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/plugin/localbinary"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"

	"google.golang.org/grpc"
)

type InitInfo struct {
	DriverName    string `json:"driver_name" validate:"required"`
	DriverVersion string `json:"driver_version" validate:"required"`
	DriverID      int    `json:"driver_id" validate:"required"`
	CoreGRPCPort  int    `json:"core_grpc_port" validate:"required"`
}

// CreateChain create chain action
func CreateChain(ctx context.Context, info InitInfo) error {
	// 初始化链的参数
	err := validator.New().Struct(&info)
	if err != nil {
		return errors.Wrap(err, "check param")
	}
	ins, err := getCEnginePluginInstance(ctx, info.DriverID, info.DriverName, info.DriverVersion)
	if err != nil {
		return errors.Wrap(err, "get chain engine plugin client")
	}
	fmt.Println(ins)

	return errors.New("implement me")
}

// TODODone: 这个是一个很大的问题, 该怎么嵌入驱动 ----> 使用grpc嵌入
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
