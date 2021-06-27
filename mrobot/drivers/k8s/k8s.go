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
	"github.com/zibuyu28/cmapp/mrobot/proto"
)

type DriverK8s struct {

}

func (d DriverK8s) Exit(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}

func (d DriverK8s) InitMachine(ctx context.Context, empty *proto.Empty) (*proto.Machine, error) {
	panic("implement me")
}

func (d DriverK8s) CreateExec(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}

func (d DriverK8s) InstallMRobot(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}

func (d DriverK8s) MRoHealthCheck(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	panic("implement me")
}



