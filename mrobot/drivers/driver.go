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
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s"
	"github.com/zibuyu28/cmapp/mrobot/proto"
)

var buildInDrivers = []string{"k8s"}

type Driver struct {
	GrpcServer proto.MachineDriverServer
}

// Exit exit driver called by self
func (d Driver) Exit() {
	ctx := context.Background()
	_, err := d.GrpcServer.Exit(ctx, &proto.Empty{})
	if err != nil {
		log.Errorf(ctx, "Manage err when exit driver server, Err: [%v]", err)
	}
}

// ParseDriver parse driver by name
func ParseDriver(driverName string) (*Driver, error) {
	switch driverName {
	case "k8s":
		return &Driver{k8s.DriverK8s{}}, nil
	default:
		return nil, errors.Errorf("fail to parse driver named [%s], build in drivers [%v], please check config or param",driverName, buildInDrivers)
	}
}
