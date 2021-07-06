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

package worker

import (
	"os"
	"strings"
)

const DriverPrefix string = "DRIAGENT_"

type Flag struct {
	Name  string
	Value string
}

var Flags = make(map[string]Flag)

func init() {
	environ := os.Environ()
	for _, s := range environ {
		if strings.HasPrefix(s, DriverPrefix) {
			kvs := strings.SplitN(strings.TrimPrefix(s, DriverPrefix), "=", 2)
			if len(kvs) != 2 {
				continue
			}
			Flags[kvs[0]] = Flag{
				Name:  kvs[0],
				Value: kvs[1],
			}
		}
	}
}

func wsClient() {
	NewDriverWorker()
}
