// Copyright 2021/7/18 wanghengfang
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package worker

import (
	"context"
	"fmt"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)


var workerServer worker0.Worker0Server

// RegisterWorker register a instance to agentfw
func RegisterWorker(imp worker0.Worker0Server) {
	workerServer = imp
}

type plugin struct {
	addr string
	rpcClient worker0.Worker0Client
}

func pluginIns(ctx context.Context) (*plugin, error) {
	if workerServer == nil {
		log.Fatalf(ctx, "Error verify plugin, plugin is nil. Please import 'worker' package, " +
			"then register an instance that implements the 'worker.WorkerServer' interface through the 'worker.RegisterWorker' method")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf(ctx, "Error loading plugin RPC server. Err: [%v], stdErr: [%s]", err, os.Stderr)
	}

	grpcserver := grpc.NewServer()

	worker0.RegisterWorker0Server(grpcserver, workerServer)

	go func() {
		err = grpcserver.Serve(listener)
		if err != nil {
			log.Errorf(ctx, "Currently grpc server serve failed. Err: [%v]", err)
		}

		<-ctx.Done()
		listener.Close()
	}()

	fmt.Println(listener.Addr())

	return &plugin{
		addr: listener.Addr().String(),
	}, nil

}

func (p *plugin) run(ctx context.Context, sig chan bool) {
	timeout, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()
	conn, err := grpc.DialContext(timeout, p.addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf(ctx, "Error create grpc connection with [%s]", p.addr)
	}
	rpc := worker0.NewWorker0Client(conn)

	p.rpcClient = rpc

	log.Infof(ctx, "plugin running successfully [%s]", p.addr)
	sig <- true
	select {
	case <-ctx.Done():
		log.Info(ctx, "Currently worker plugin exist")
		return
	}
}
