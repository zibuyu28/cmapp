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
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/common/httputil"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/ssh_cmd"
	virtualbox "github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/core"
	agfw "github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/worker"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type VirtualboxWorker struct {
	MachineID    int
	HostIP       string
	HostPort     int
	HostUsername string
	HostPassword string
	StorePath    string
	VBUUID       string
}

func NewVirtualboxWorker() *VirtualboxWorker {
	w := &VirtualboxWorker{
		MachineID:    0,
		HostIP:       agfw.Flags["HOST_IP"].Value,
		HostPort:     0,
		HostUsername: agfw.Flags["HOST_USERNAME"].Value,
		HostPassword: agfw.Flags["HOST_PASSWORD"].Value,
		StorePath:    agfw.Flags["HOST_STORE_PATH"].Value,
		VBUUID:       agfw.Flags["VBUUID"].Value,
	}
	if len(agfw.Flags["HOST_PORT"].Value) != 0 {
		port, err := strconv.Atoi(agfw.Flags["HOST_PORT"].Value)
		if err != nil {
			panic(err)
		}
		w.HostPort = port
	}
	if len(agfw.Flags["MACHINE_ID"].Value) != 0 {
		mid, err := strconv.Atoi(agfw.Flags["MACHINE_ID"].Value)
		if err != nil {
			panic(err)
		}
		w.MachineID = mid
	}
	validate := v.New()
	err := validate.Struct(*w)
	if err != nil {
		panic(err)
	}
	return w
}
func (v *VirtualboxWorker) NewApp(ctx context.Context, req *worker0.NewAppReq) (*worker0.App, error) {
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
		UID:                 uid,
		Name:                fmt.Sprintf("%s:%s", req.Name, req.Version),
		Workspace:           uid,
		InstallationPackage: pkg.Binary.Download,
		PackageMd5:          pkg.Binary.CheckSum,
		PackageHandleShells: pkg.Binary.PackageHandleShells,
		StartCMD:            pkg.Binary.StartCommands,
		Tags:                map[string]string{"uuid": uid, "machine_id": fmt.Sprintf("%d", v.MachineID)},
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
			Type:     worker0.App_MainProcess_Binary,
			Workdir:  uid,
			StartCMD: pkg.Binary.StartCommands,
		},
		Workspace: &worker0.App_WorkspaceInfo{Workspace: uid},
	}
	abs, _ := filepath.Abs(uid)
	_ = os.MkdirAll(filepath.Join(abs, uid), os.ModePerm)
	return wap, nil
}

// GetHostVolume TODO: maybe need to implement
func (v *VirtualboxWorker) GetHostVolume(ctx context.Context, _ *worker0.App) error {
	panic("implement me")
}

func (v *VirtualboxWorker) StartApp(ctx context.Context, _ *worker0.App) (*worker0.Empty, error) {
	log.Debug(ctx, "Currently to start app")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	log.Debugf(ctx, "Currently start get main package add [%s], save to dir [%s/]", app.InstallationPackage, app.Workspace)
	abs, _ := filepath.Abs(app.Workspace)
	_, err = os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(abs, os.ModePerm)
			if err != nil {
				return nil, errors.Wrapf(err, "mkdir workspace [%s]", abs)
			}
		} else {
			return nil, errors.Wrap(err, "os state")
		}
	}

	packageFile := filepath.Join(abs, "pkg.tmp")
	err = httputil.HTTPDoDownloadFile(packageFile, app.InstallationPackage)
	if err != nil {
		return nil, errors.Wrap(err, "download package")
	}

	// do package handle shell
	for _, shell := range app.PackageHandleShells {
		out, err := cmd.NewDefaultCMD(shell, []string{}, cmd.WithWorkDir(abs)).Run()
		if err != nil {
			return nil, errors.Wrapf(err, "exec package handle shell [%s], Err: [%v]", shell, err)
		}
		log.Debugf(ctx, "Currently execute shell [%s] success, out [%s]", shell, out)
	}

	log.Debug(ctx, "Currently start exec file premise")
	for _, premise := range app.FilePremises {
		log.Debugf(ctx, "start to get file [%s], addr [%s]", premise.Name, premise.AcquireAddr)
		premiseFile := filepath.Join(abs, premise.Name)
		err = httputil.HTTPDoDownloadFile(premiseFile, premise.AcquireAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "download premise [%s]", premise.Name)
		}
		out, err := cmd.NewDefaultCMD(premise.Shell, []string{}, cmd.WithWorkDir(abs)).Run()
		if err != nil {
			return nil, errors.Wrapf(err, "exec premise shell [%s], Err: [%v]", premise.Shell, err)
		}
		log.Debugf(ctx, "Currently execute premise shell [%s] success, out [%s]", premise.Shell, out)
	}

	log.Debug(ctx, "Currently start exec file mounts")

	//FileMount{
	//	File:    "uuid/cert/conf/cfg.toml",
	//	MountTo: "uuid/config/cfg.toml",
	//	Volume:  "",
	//}
	// 所以这里的 filemount 都是不需要进行操作，只需要知道用户下载的文件都在 workspace->app.UID 下即可
	//for _, mount := range app.FileMounts {
	//	mount.Volume
	//}

	// tag handle: set tag to process environment -> check same name
	var processEnvs = make(map[string]string)
	for t, v1 := range app.Tags {
		processEnvs[t] = v1
	}

	// env handle
	for k, v2 := range app.Environments {
		if old, ok := processEnvs[k]; ok {
			log.Warnf(ctx, "Currently set env got same key [%s]. Now to cover old [%s], new [%s]. Please know this", k, old, v2)
		}
		processEnvs[k] = v2
	}

	// TODO: limit process's cpu and memory in the same way like docker
	// now ignore limit

	// start app
	log.Debug(ctx, "Currently start to setup app")
	setupCommand := strings.Join(app.StartCMD, " ")
	out, err := cmd.NewDefaultCMD(setupCommand, []string{}, cmd.WithWorkDir(abs), cmd.WithEnvs(processEnvs)).Run()
	if err != nil {
		return nil, errors.Wrapf(err, "exec setup command [%s], Err: [%v]", setupCommand, err)
	}
	log.Debugf(ctx, "Currently setup app out [%s]", out)

	// health: check this app is setup success or not
	//       : by the way, the worker need to provided
	//       : app live status checking-engine (timed
	//       : check app running status, and recovery
	//       : app if not health)

	// now just to check 10 times

	if app.Health.Readness != nil {
		readnessUrl := fmt.Sprintf("http://127.0.0.1:%d%s", app.Health.Readness.Port, app.Health.Readness.Path)
		log.Debugf(ctx, "Currently read url [%s]", readnessUrl)

		// 检查1分钟
		toutctx, cancelFunc := context.WithTimeout(ctx, time.Minute)
		defer cancelFunc()
		// 间隔1秒检查一次
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				switch app.Health.Readness.Method {
				case HttpGet:
					_, err = httputil.HTTPDoGet(readnessUrl)
					if err != nil {
						log.Errorf(ctx, "Manage err when check app readness. Now to continue. Err: [%v]", err)
						continue
					}
					return nil, nil
				case HttpPost:
					_, err = httputil.HTTPDoPost("", readnessUrl)
					if err != nil {
						log.Errorf(ctx, "Manage err when check app readness. Now to continue. Err: [%v]", err)
						continue
					}
					return nil, nil
				}
			case <-toutctx.Done():
				return nil, errors.New("app readness check timeout")
			}
		}
	}

	// TODO: get app pid == info
	return nil, nil
}

func (v *VirtualboxWorker) StopApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) DestroyApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) TagEx(ctx context.Context, tag *worker0.App_Tag) (*worker0.App_Tag, error) {
	log.Debug(ctx, "Currently start to execute set app tag")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	if len(tag.Key) == 0 || len(tag.Value) == 0 {
		return nil, errors.Errorf("tag got empty key [%s] or value [%s]", tag.Key, tag.Value)
	}
	if tag.Key == "uuid" || tag.Key == "machine_id" {
		return nil, errors.New("tag named 'uid' or 'machine_id' not support to set")
	}
	app.Tags[tag.Key] = tag.Value
	return tag, nil
}

func (v *VirtualboxWorker) FileMountEx(ctx context.Context, mount *worker0.App_FileMount) (*worker0.App_FileMount, error) {
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

func (v *VirtualboxWorker) EnvEx(ctx context.Context, envVar *worker0.App_EnvVar) (*worker0.App_EnvVar, error) {
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

func (v *VirtualboxWorker) NetworkEx(ctx context.Context, network *worker0.App_Network) (*worker0.App_Network, error) {
	log.Debug(ctx, "Currently start to execute network")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	if network.PortInfo.Port == 0 {
		return nil, errors.New("env got empty port")
	}

	network.PortInfo.ProtocolType = worker0.App_Network_PortInf_TCP

	cli, err := ssh_cmd.NewSSHCli(v.HostIP, v.HostPort, v.HostUsername, v.HostPassword)
	if err != nil {
		return nil, errors.Wrap(err, "new ssh cli")
	}
	vbm := virtualbox.NewRMTDriver(ctx, v.VBUUID, v.StorePath, v.HostIP, cli)

	log.Debugf(ctx, "Currently get enum name [%s]", network.PortInfo.ProtocolType.String())
	actualPort, err := vbm.ExportPort(network.PortInfo.Name, network.PortInfo.ProtocolType.String(), int(network.PortInfo.Port))
	if err != nil {
		return nil, errors.Wrapf(err, "export port [%d]", network.PortInfo.Port)
	}

	pi := PortInfo{
		Port:            int(network.PortInfo.Port),
		Name:            network.PortInfo.Name,
		Protocol:        worker0.App_Network_PortInf_Protocol_name[int32(network.PortInfo.ProtocolType)],
		HostPortMapping: actualPort,
	}
	app.Ports[int(network.PortInfo.Port)] = pi

	localIP, err := getLocalIP()
	if err != nil {
		return nil, errors.Wrap(err, "get local ip")
	}

	inRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_IN,
		Router:    fmt.Sprintf("%s:%d", localIP, network.PortInfo.Port),
	}
	outRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_OUT,
		Router:    fmt.Sprintf("%s:%d", v.HostIP, actualPort),
	}

	network.RouteInfo = []*worker0.App_Network_RouteInf{inRoute, outRoute}

	return network, nil
}

// getLocalIP get local ip
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "net interface addrs")
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Can not find the client ip address!")
}

func (v *VirtualboxWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App_File, error) {

	log.Debug(ctx, "Currently start to execute set file premise")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	if len(file.Name) == 0 || len(file.AcquireAddr) == 0 {
		return nil, errors.Errorf("file got empty name [%s] or acquire addr [%s]", file.Name, file.AcquireAddr)
	}
	key := md5.MD5(fmt.Sprintf("%s:%s", file.Name, file.AcquireAddr))

	if e, ok := app.FilePremises[key]; ok {
		return nil, errors.Errorf("file premise exist [%#+v]", e)
	}
	premise := FilePremise{
		Name:        file.Name,
		AcquireAddr: file.AcquireAddr,
		Shell:       file.Shell,
	}
	app.FilePremises[key] = premise
	return file, nil
}

func (v *VirtualboxWorker) LimitEx(ctx context.Context, limit *worker0.App_Limit) (*worker0.App_Limit, error) {
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

func (v *VirtualboxWorker) HealthEx(ctx context.Context, health *worker0.App_Health) (*worker0.App_Health, error) {
	log.Debug(ctx, "Currently start to execute set health info")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	var healthOpt HealthOption
	if health.Readness != nil {
		log.Debugf(ctx, "Currently get read ness health info [%#+v]", health.Readness)
		read := &HealthBasic{
			Path: health.Readness.Path,
			Port: int(health.Readness.Port),
		}
		switch health.Readness.MethodType {
		case worker0.App_Health_Basic_GET:
			read.Method = HttpGet
		case worker0.App_Health_Basic_POST:
			read.Method = HttpPost
		default:
			return nil, errors.Wrapf(err, "fail to parse method [%s]", health.Readness.MethodType)
		}
		healthOpt.Readness = read
	}
	if health.Liveness != nil {
		log.Debugf(ctx, "Currently get live ness health info [%#+v]", health.Liveness)
		live := HealthBasic{
			Path: health.Liveness.Path,
			Port: int(health.Liveness.Port),
		}
		switch health.Liveness.MethodType {
		case worker0.App_Health_Basic_GET:
			live.Method = HttpGet
		case worker0.App_Health_Basic_POST:
			live.Method = HttpPost
		default:
			return nil, errors.Wrapf(err, "fail to parse method [%s]", health.Liveness.MethodType)
		}
		healthOpt.Liveness = &live
	}
	if healthOpt.Readness == nil && healthOpt.Liveness == nil {
		log.Infof(ctx, "not set health info")
		return nil, nil
	}
	app.Health = &healthOpt
	return health, nil
}

func (v *VirtualboxWorker) LogEx(ctx context.Context, appLog *worker0.App_Log) (*worker0.App_Log, error) {
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
