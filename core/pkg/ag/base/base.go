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
	fnc   string
	param interface{}
}

type coreResp struct {
	code    CoreCode
	data    interface{}
	message string
}

type CoreCode int

const (
	SUCCESS CoreCode = 200
	FAILED CoreCode = 400
)

func Send(fnc string, req interface{}) ([]byte, error) {
	cr := coreReq{
		fnc:   fnc,
		param: req,
	}
	respb, err := httputil.HTTPDoPost(cr, getURL())
	if err != nil {
		return nil, errors.Wrap(err, "send req to core")
	}
	resp := coreResp{data: struct{}{}}
	err = json.Unmarshal(respb, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp [%s]", string(respb))
	}
	if resp.code != SUCCESS {
		return nil, errors.Errorf("fail to call [%s] to call, message [%s]", fnc, resp.message)
	}
	datab, err := json.Marshal(resp.data)
	if err != nil {
		return nil, errors.Wrap(err, "marshal response's data info")
	}
	return datab, nil
}

func getURL() string {
	host := viper.GetString("CORE_HOST")
	if len(host) == 0 {
		host = coreDefaultHost
	}

	port := viper.GetInt("CORE_PORT")
	if port == 0 {
		port = coreDefaultPort
	}
	return fmt.Sprintf("%s:%d/md", host, port)
}
