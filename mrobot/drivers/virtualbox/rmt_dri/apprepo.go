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

package rmt_dri

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"sync"
)

type workRepository struct {
	rep sync.Map
}

var repo = workRepository{rep: sync.Map{}}

func (w *workRepository) load(ctx context.Context) (*App, error) {
	uid, err := guid(ctx)
	if err != nil {
		return nil, errors.New("load uid from context")
	}
	wsp, ok := w.rep.Load(uid)
	if !ok {
		return nil, errors.Errorf("fail to get workspace from rep by uid [%s]", uid)
	}
	return wsp.(*App), nil
}

func (w *workRepository) new(ctx context.Context, wsp *App) error {
	uid, err := guid(ctx)
	if err != nil {
		return errors.New("load uid from context")
	}
	wsp.UID = uid
	w.rep.Store(uid, wsp)
	return nil
}

func guid(ctx context.Context) (uid string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("fail to get metadata from context")
		return
	}
	uuid := md.Get("MA_UUID")
	if len(uuid) == 0 {
		err = errors.New("fail to get uuid from metadata")
		return
	}
	uid = uuid[0]
	return
}

type App struct {
	UID  string `validate:"required"`
	Name string
	// relative path
	Workspace           string
	InstallationPackage string
	PackageMd5          string
	StartCMD            []string
	Tags                map[string]string
	Environments        map[string]string
	FilePremises        map[string]FilePremise
	FileMounts          map[string]FileMount
	Limit               *Limit
	Health              *HealthOption
	Log                 *Log
	Ports               map[int]PortInfo
}

type PortInfo struct {
	Port            int    `validate:"required"`
	Name            string `validate:"required"`
	Protocol        string `validate:"required"`
	HostPortMapping int    `validate:"required"`
}

type Log struct {
	RealTimeFile    string
	CompressLogPath string
}

type MethodType string

const (
	HttpGet  MethodType = "httpGet"
	HttpPost MethodType = "httpPost"
)

type HealthBasic struct {
	Method MethodType
	Path   string `validate:"required"`
	Port   int    `validate:"required"`
}

type HealthOption struct {
	Liveness *HealthBasic
	Readness *HealthBasic
}

type Limit struct {
	// unit 'm'
	CPU int
	// unit 'MB'
	Memory int
}

type FileMount struct {
	File    string `validate:"required"`
	MountTo string `validate:"required"`
	Volume  string
}

type FilePremise struct {
	Name        string
	AcquireAddr string
	Shell       string
}
