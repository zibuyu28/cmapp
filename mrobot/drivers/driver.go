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

package drivers

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s"
	"github.com/zibuyu28/cmapp/mrobot/pkg"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
)

var buildInDrivers = []string{"k8s"}
var buildInDriversVersion = map[string]string{
	"k8s": "1.0.0",
}

type BuildInDriver struct {
	GrpcServer driver.MachineDriverServer
}

// Exit exit driver called by self
func (d BuildInDriver) Exit() {
	ctx := context.Background()
	_, err := d.GrpcServer.Exit(ctx, &driver.Empty{})
	if err != nil {
		log.Errorf(ctx, "Manage err when exit driver server, Err: [%v]", err)
	}
}

// ParseDriver parse driver by name
func ParseDriver(driverName string) (*BuildInDriver, error) {
	switch driverName {
	case "k8s":
		return &BuildInDriver{GrpcServer: &k8s.DriverK8s{
			BaseDriver: pkg.BaseDriver{
				UUID:          md5.MD5(fmt.Sprintf("%s:%s", driverName, buildInDriversVersion[driverName])),
				DriverName:    driverName,
				DriverVersion: buildInDriversVersion[driverName],
			},
		}}, nil
	default:
		return nil, errors.Errorf("fail to parse driver named [%s], build in drivers [%v], please check config or param", driverName, buildInDrivers)
	}
}

type BuildInPlugin struct {
	GrpcPluginServer worker.WorkerServer
}

func ParsePlugin(pluginName string) (*BuildInPlugin, error) {
	switch pluginName {
	case "k8s":
		return &BuildInPlugin{
			GrpcPluginServer: &k8s.Worker{},
		}, nil
	default:
		return nil, errors.Errorf("fail to parse plugin named [%s], build in drivers [%v], please check config or param", pluginName, buildInDrivers)
	}
}
