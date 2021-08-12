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
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
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
	app := &App{
		UID:     uid,
		Image:   fmt.Sprintf("%s:%s", pkg.Image.ImageName, pkg.Image.Tag),
		WorkDir: pkg.Image.WorkDir,
		Command: pkg.Image.StartCommands,
	}
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

func (k *K8sWorker) FileMountEx(ctx context.Context, mount *worker0.App_FileMount) (*worker0.App_FileMount, error) {
	log.Debug(ctx, "Currently start to execute file mount")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	if len(mount.File) == 0 {
		return nil, errors.New("file is empty")
	}

	key := md5.MD5(fmt.Sprintf("%s:%s:%s", mount.File, mount.MountTo, mount.Volume))
	if e, ok := app.FileMounts[key]; ok {
		return nil, errors.Errorf("mount exist [%+#v]", e)
	}

	app.FileMounts[key] = FileMount{
		File:    mount.File,
		MountTo: mount.MountTo,
		Volume:  mount.Volume,
	}
	return &worker0.App_FileMount{
		File:    mount.File,
		MountTo: mount.MountTo,
		Volume:  mount.Volume,
	}, nil
}

func (k *K8sWorker) VolumeEx(ctx context.Context, volume *worker0.App_Volume) (*worker0.App_Volume, error) {
	log.Debug(ctx, "Currently start to execute volume create")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	if len(volume.Name) == 0 || len(volume.Type) == 0 {
		return nil, errors.Errorf("volume got empty Name [%s] or Type [%s]", volume.Name, volume.Type)
	}

	if e, ok := app.Volumes[volume.Name]; ok {
		return nil, errors.Errorf("volume exist [%+#v]", e)
	}

	app.Volumes[volume.Name] = Volume{
		Name:  volume.Name,
		Type:  volume.Type,
		Param: volume.Param,
	}
	return &worker0.App_Volume{
		Name:  volume.Name,
		Type:  volume.Type,
		Param: volume.Param,
	}, nil
}

func (k *K8sWorker) EnvEx(ctx context.Context, envVar *worker0.App_EnvVar) (*worker0.App_EnvVar, error) {
	log.Debug(ctx, "Currently start to execute set app env")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	if len(envVar.Key) == 0 || len(envVar.Value) == 0 {
		return nil, errors.Errorf("env got empty key [%s] or value [%s]", envVar.Key, envVar.Value)
	}

	app.Environments[envVar.Key] = envVar.Value

	return &worker0.App_EnvVar{Key: envVar.Key, Value: envVar.Value}, nil
}

func (k *K8sWorker) NetworkEx(ctx context.Context, network *worker0.App_Network) (*worker0.App_Network, error) {
	log.Debug(ctx, "Currently start to execute network")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	if network.PortInfo.Port == 0 {
		return nil, errors.New("env got empty port")
	}

	network.PortInfo.ProtocolType = worker0.App_Network_PortInf_TCP

	// 内部service
	service := fmt.Sprintf("%s-service", app.UID)

	// 外部ingress
	ingress := fmt.Sprintf("%s-%s-%d.develop.blocface.baas.hyperchain.cn", "machineinf", app.UID, network.PortInfo.Port)
	pi := PortInfo{
		Port:        int(network.PortInfo.Port),
		Name:        network.PortInfo.Name,
		Protocol:    worker0.App_Network_PortInf_Protocol_name[int32(network.PortInfo.ProtocolType)],
		ServiceName: service,
		IngressName: ingress,
	}
	app.Ports[int(network.PortInfo.Port)] = pi

	inRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_IN,
		Router:    fmt.Sprintf("%s:%d", service, network.PortInfo.Port),
	}
	outRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_OUT,
		Router:    fmt.Sprintf("%s:%d", service, 80),
	}

	network.RouteInfo = []*worker0.App_Network_RouteInf{inRoute, outRoute}

	return network, nil
}

func (k *K8sWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App, error) {

	panic("implement me")
}

func (k *K8sWorker) LimitEx(ctx context.Context, limit *worker0.App_Limit) (*worker0.App_Limit, error) {
	log.Debug(ctx, "Currently start to execute set limit")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	lm := &Limit{}

	if limit.CPU != 0 {
		lm.CPU = int(limit.CPU)
	}
	if limit.Memory != 0 {
		lm.Memory = int(limit.Memory)
	}
	app.Limit = lm
	return limit, nil
}

func (k *K8sWorker) HealthEx(ctx context.Context, health *worker0.App_Health) (*worker0.App, error) {
	panic("implement me")
}

func (k *K8sWorker) LogEx(ctx context.Context, appLog *worker0.App_Log) (*worker0.App_Log, error) {
	log.Debug(ctx, "Currently start to execute set log info")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	if len(appLog.FilePath) == 0 || len(appLog.RealTimeFile) == 0 {
		return nil, errors.Errorf("get empty param, file path [%s] or real-time [%s]", appLog.FilePath, appLog.RealTimeFile)
	}

	app.Log = &Log{RealTimeFile: appLog.RealTimeFile, CompressLogPath: appLog.FilePath}

	return appLog, nil
}
