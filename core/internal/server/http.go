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
	"github.com/gin-gonic/gin"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/ws"
	"syscall"
)

var defaultHttpPort = 9008

var httpserver *gin.Engine

func httpServerStart(ctx context.Context) {
	httpserver = gin.Default()
	httpserver.POST("/ws", func(c *gin.Context) {
		ws.ServeWs(ctx, c.Writer, c.Request)
	})
	err := httpserver.Run(fmt.Sprintf(":%d", defaultHttpPort))
	if err != nil {
		log.Fatalf(ctx, "failed to listen: %v", err)
	}
	log.Infof(ctx, "server listening at :%d", defaultHttpPort)
}

func httpServerStop() {
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
}
