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
	"google.golang.org/grpc/metadata"
)

type ActionPhase string

const (
	InstallPhase ActionPhase = "install"
	SetupPhase   ActionPhase = "setup"
	RunningPhase ActionPhase = "running"
)

func phaseParse(ctx context.Context) (ap ActionPhase, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("fail to get metadata from context")
		return
	}
	ps := md.Get("MA_PHASE")
	if len(ps) == 0 {
		err = errors.New("fail to get phase from metadata")
		return
	}
	p := ps[0]
	switch p {
	case "install":
		ap = InstallPhase
	case "setup":
		ap = SetupPhase
	case "running":
		ap = RunningPhase
	default:
		err = errors.Errorf("not support phase [%s] for action", p)
	}
	return
}
