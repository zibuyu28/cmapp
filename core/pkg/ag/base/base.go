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

package base

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zibuyu28/cmapp/common/httputil"
)

var coreDefaultHost = "127.0.0.1"

var coreDefaultPort = 9008

type coreReq struct {
	Fnc   string      `json:"fnc"`
	Param interface{} `json:"param"`
}

type coreResp struct {
	Code    CoreCode    `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type CoreCode int

const (
	SUCCESS CoreCode = 200
	FAILED  CoreCode = 400
)

type APIVersion string

const (
	V1 APIVersion = "v1"
)

func SendPost(v APIVersion, req interface{}) ([]byte, error) {
	respb, err := httputil.HTTPDoPost(req, getURL(v))
	if err != nil {
		return nil, errors.Wrap(err, "send req to core")
	}
	resp := coreResp{Data: struct{}{}}
	err = json.Unmarshal(respb, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp [%s]", string(respb))
	}
	if resp.Code != SUCCESS {
		return nil, errors.Errorf("fail to call with req [%v], message [%s]", req, resp.Message)
	}
	datab, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.Wrap(err, "marshal response's data info")
	}
	return datab, nil
}

func getURL(version APIVersion) string {
	host := viper.GetString("CORE_HOST")
	if len(host) == 0 {
		host = coreDefaultHost
	}

	port := viper.GetInt("CORE_PORT")
	if port == 0 {
		port = coreDefaultPort
	}
	return fmt.Sprintf("http://%s:%d/api/%s/md/exec", host, port, version)
}
