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

package k8s

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s/kube_engine"
	"github.com/zibuyu28/cmapp/mrobot/proto"
	"google.golang.org/grpc/metadata"
)

type DriverK8s struct {
}

func NewDriverK8s() {

}

func (d DriverK8s) DriverVersion(ctx context.Context, empty *proto.Empty) (*proto.Version, error) {
	panic("implement me")
}

func (d DriverK8s) InitMachine(ctx context.Context, empty *proto.Empty) (*proto.Machine, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to init machine")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("UUID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find uuid from metadata")
	}

	var customInfo = map[string]string{
		"config":    defaultConfig,
		"namespace": defaultNamespace,
	}

	machine := &proto.Machine{
		UUID:       datas[0],
		State:      1,
		Tags:       []string{"k8s"},
		CustomInfo: customInfo,
	}

	return machine, nil
}

func (d DriverK8s) GetCreateFlags(ctx context.Context, empty *proto.Empty) (*proto.Flags, error) {
	panic("implement me")
}

func (d DriverK8s) CreateExec(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to create machine")
	return nil, nil
}

func (d DriverK8s) InstallMRobot(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	client, err := kube_engine.NewClient(ctx, defaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}



	panic("implement me")
}

func (d DriverK8s) MRoHealthCheck(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}

func (d DriverK8s) Exit(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}

var defaultNamespace = ""
var defaultConfig = `

`

var mRobotDep = `---

`


