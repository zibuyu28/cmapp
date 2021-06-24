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
	"github.com/zibuyu28/cmapp/core/internal/log"
	"os"
	"os/signal"
	"syscall"
)

// Serve all grpc and http
func Serve() {
	log.Info("start grpc and http server")
	go httpServerStart()
	go grpcServerStart()
	signalHandler()
}

// Stop stop serve both grpc and http
func Stop() {
	log.Info("stop both grpc and http server")
	httpServerStop()
	grpcServerStop()
}



func signalHandler() {
	// signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		fmt.Println(fmt.Sprintf("service get a signal %s", s.String()))
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			return
		default:
			return
		}
	}
}