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

package server

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/internal/api_g"
	"github.com/zibuyu28/cmapp/core/proto/ch_manager"
	"github.com/zibuyu28/cmapp/core/proto/ma_manager"
	"google.golang.org/grpc"
	"net"
)

var defaultGrpcPort = 9009

var grpcserver *grpc.Server

// grpcServerStart server start
func grpcServerStart(ctx context.Context) {
	port := viper.GetInt("grpc.port")
	if port == 0 {
		log.Fatalf(ctx, "fail to get grpc port")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf(ctx, "failed to listen: %v", err)
	}
	grpcserver = grpc.NewServer()
	ma_manager.RegisterMachineManageServer(grpcserver, &api_g.CoreMachineManager{})
	ch_manager.RegisterChainManageServer(grpcserver, &api_g.CoreChainManager{})
	log.Infof(ctx, "server listening at %v", lis.Addr())
	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf(ctx, "failed to serve: %v", err)
	}
}

// grpcServerStop server stop
func grpcServerStop() {
	if grpcserver != nil {
		grpcserver.Stop()
	}
}
