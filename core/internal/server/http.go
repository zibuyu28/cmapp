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
	"github.com/spf13/viper"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/internal/api_c"
	"github.com/zibuyu28/cmapp/core/internal/server/mid"
	"strings"
	"syscall"
)

var defaultHttpPort = 9008

var httpserver = func() *gin.Engine {
	engine := gin.New()
	engine.Use(mid.GinLogger(false))
	engine.Use(mid.RecoveryWithLogger(false))
	return engine
}()

func httpServerStart(ctx context.Context) {
	//httpserver.POST("/ws", func(c *gin.Context) {
	//	ws.ServeWs(ctx, c.Writer, c.Request)
	//})
	log.Infof(ctx, "register api")
	registerApi()
	port := viper.GetInt("http.port")
	if port == 0 {
		log.Fatalf(ctx, "fail to get http port")
	}
	log.Infof(ctx, "server listening at :%d", port)
	err := httpserver.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf(ctx, "failed to listen: %v", err)
	}
}

func httpServerStop() {
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
}

// Group new group
func registerApi() {
	for group, routers := range api_c.GMR {
		gp := strings.TrimPrefix(string(group), "/")
		routerGroup := httpserver.Group(fmt.Sprintf("/api/%s", gp))
		for path, f := range routers {
			split := strings.Split(string(path), "@")
			if len(split) != 2 {
				panic(fmt.Sprintf("error path [%s]", path))
			}
			routerGroup.Handle(split[0], split[1], f)
		}
	}
}
