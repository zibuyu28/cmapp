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
	"fmt"
	"github.com/zibuyu28/cmapp/core/internal/api"
	"github.com/zibuyu28/cmapp/core/internal/log"
	"github.com/zibuyu28/cmapp/core/proto"
	"google.golang.org/grpc"
	"net"
)

var defaultGrpcPort = 9009

var grpcserver *grpc.Server

// grpcServerStart server start
func grpcServerStart() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", defaultGrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcserver = grpc.NewServer()
	proto.RegisterMachineManageServer(grpcserver, &api.CoreMachineManager{})
	log.Infof("server listening at %v", lis.Addr())
	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// grpcServerStop server stop
func grpcServerStop()  {
	if grpcserver != nil {
		grpcserver.Stop()
	}
}
