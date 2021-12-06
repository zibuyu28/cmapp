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

package api_c

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/internal/service_c/chain"
)

type ChainAction string

const (
	CreateChain ChainAction = "create"
	StopChain   ChainAction = "stop"
)

type CWReq struct {
	DriverID int         `json:"driver_id" binding:"required"`
	ChainID  int         `json:"chain_id"`
	Action   ChainAction `json:"action" binding:"required"`
	Param    interface{} `json:"param" binding:"required"`
}

func cwExec(g *gin.Context) {
	var p = CWReq{}
	err := g.BindJSON(&p)
	if err != nil {
		fail(g, err)
		return
	}
	err = cwExecHandler(g, &p)
	if err != nil {
		fail(g, err)
		return
	}
}


func cwExecHandler(g *gin.Context, req *CWReq) error {
	switch req.Action {
	case CreateChain:
		if req.DriverID == 0 {
			return errors.New("create action: driver id is nil")
		}
		err := chain.Create(g.Request.Context(), req.DriverID, req.Param)
		if err != nil {
			return errors.Wrap(err, "do create chain")
		}
		ok(g, "ok")
	default:
		return errors.Errorf("action [%s] not support now", req.Action)
	}
	return nil
}

