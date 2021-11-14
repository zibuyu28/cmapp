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

package plugin

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zibuyu28/cmapp/crobot/drivers"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc"
)

// RegisterDriver register driver build in
func RegisterDriver(d drivers.BuildInDriver) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading RPC server: %s\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	grpcserver := grpc.NewServer()

	driver.RegisterChainDriverServer(grpcserver, d.GrpcServer)

	go grpcserver.Serve(listener)

	fmt.Println(listener.Addr())

	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second)
			// driver exit
			d.Exit()
			return
		case syscall.SIGHUP:
		// TODO app reload
		default:
			return
		}
	}
}
