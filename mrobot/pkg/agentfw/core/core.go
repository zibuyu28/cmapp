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

package core

import (
	"net/http"
	"os"
	"strings"
)

const (
	driAgentCoreHTTPAddr string = "COREHTTPADDR"
	driverPrefix         string = "DRIAGENT_"
)

var cli *client

func init() {
	environ := os.Environ()
	for _, s := range environ {
		if strings.HasPrefix(s, driverPrefix) {
			kvs := strings.SplitN(strings.TrimPrefix(s, driverPrefix), "=", 2)
			if len(kvs) != 2 {
				continue
			}
			if kvs[0] == driAgentCoreHTTPAddr {
				cli = &client{coreAddr: kvs[1], hc: http.Client{}}
			}
			break
		}
	}
}

type client struct {
	coreAddr string
	hc       http.Client
}

func PackageInfo(name, version string) {
	// "/api/v1/package"

}
