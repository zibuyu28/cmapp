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
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/core"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
)

type K8sWorker struct {
	Name string
}

func (k *K8sWorker) NewApp(ctx context.Context, req *worker0.NewAppReq) (*worker0.App, error) {
	log.Infof(ctx, "new app [%s/%s]", req.Name, req.Version)
	if len(req.Name) == 0 || len(req.Version) == 0 {
		return nil, errors.Errorf("fail to get name [%s] or version [%s] info", req.Name, req.Version)
	}
	pkg, err := core.PackageInfo(ctx, req.Name, req.Version)
	if err != nil {
		return nil, errors.Wrapf(err, "get package info")
	}
	uid, err := guid(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "get uid from ctx")
	}
	app := &App{UID: uid}
	err = repo.new(ctx, app)
	if err != nil {
		return nil, errors.Wrap(err, "k8s repo new app")
	}
	wap := &worker0.App{
		UUID: uid,
		MainP: &worker0.App_MainProcess{
			Name:     pkg.Name,
			Version:  pkg.Version,
			Type:     worker0.App_MainProcess_Image,
			Workdir:  pkg.Image.WorkDir,
			StartCMD: pkg.Image.StartCommands,
		},
		Workspace: &worker0.App_WorkspaceInfo{Workspace: "test"},
	}
	return wap, nil
}

func (k *K8sWorker) StartApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *K8sWorker) StopApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *K8sWorker) DestroyApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *K8sWorker) FileMountEx(ctx context.Context, mount *worker0.App_FileMount) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) VolumeEx(ctx context.Context, volume *worker0.App_Volume) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) EnvEx(ctx context.Context, envVar *worker0.App_EnvVar) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) NetworkEx(ctx context.Context, network *worker0.App_Network) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) LimitEx(ctx context.Context, limit *worker0.App_Limit) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) HealthEx(ctx context.Context, health *worker0.App_Health) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) LogEx(ctx context.Context, log *worker0.App_Log) (*worker0.App, error) {
	panic("implement me")
}
