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

package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"github.com/zibuyu28/cmapp/core/pkg/ag"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultGrpcPort   int    = 9009
	DefaultDriverPath string = "drv.DriverPath"
)

const (
	MachineEngineCoreGRPCPORT  = "MACHINE_ENGINE_CORE_GRPC_PORT"
	MachineEngineDriverID      = "MACHINE_ENGINE_DRIVER_ID"
	MachineEngineDriverName    = "MACHINE_ENGINE_DRIVER_NAME"
	MachineEngineDriverVersion = "MACHINE_ENGINE_DRIVER_VERSION"
)

// Create execute driver create command to initialization machine
func Create(ctx context.Context, driverid int, param interface{}) error {
	drv, err := model.GetDriverByID(driverid)
	if err != nil {
		return errors.Wrap(err, "get driver by id")
	}
	marshal, err := json.Marshal(param)
	if err != nil {
		return errors.Wrap(err, "marshal param")
	}
	err = CreateAction(ctx, DefaultDriverPath, drv, md5.MD5(fmt.Sprintf("%d", time.Now().UnixNano())), string(marshal))
	if err != nil {
		return errors.Wrap(err, "create action")
	}
	return nil
}

// CreateAction driver to create machine
func CreateAction(ctx context.Context, driverRootPath string, drv *model.Driver, uuid, param string) error {
	abs, _ := filepath.Abs(filepath.Join(driverRootPath, drv.Name, drv.Version))
	binaryPath := fmt.Sprintf("%s/driver", abs)
	_, err := os.Stat(binaryPath)
	if err != nil {
		if os.ErrNotExist == err {
			return errors.Wrapf(err, "driver executable file [%s]", binaryPath)
		}
		return err
	}
	command := fmt.Sprintf("%s ma create -u %s -p '%s'", binaryPath, uuid, param)
	newCmd := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(map[string]string{
		MachineEngineCoreGRPCPORT:  strconv.Itoa(DefaultGrpcPort),
		MachineEngineDriverName:    drv.Name,
		MachineEngineDriverVersion: drv.Version,
		MachineEngineDriverID:      strconv.Itoa(drv.ID),
		"BASE_CORE_ADDR":           "192.168.31.63:9009",
		"BASE_IMAGE_REPOSITORY":    "harbor.hyeprchain.cn",
		"BASE_IMAGE_STORE_PATH":    "platform",
	}), cmd.WithTimeout(10))
	out, err := newCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "fail to execute command [%s]", command)
	}
	log.Infof(ctx, "Currently ro create command execute result : %s", out)
	return nil
}

type RMD struct {
	agConnRepo  sync.Map
	appRepo     sync.Map
	appConnRepo sync.Map
}

type clientIns struct {
	rpcClient worker0.Worker0Client
}

var RMDIns = RMD{agConnRepo: sync.Map{}, appRepo: sync.Map{}, appConnRepo: sync.Map{}}

var defaultTimeout = 3

func contextBuild(ctx context.Context, appuid string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"UUID": appuid,
	}))
}

func connAG(ctx context.Context, addr string) (worker0.Worker0Client, error) {
	load, ok := RMDIns.agConnRepo.Load(addr)
	if !ok {
		timeout, cancelFunc := context.WithTimeout(ctx, time.Duration(10)*time.Second)
		defer cancelFunc()

		conn, err := grpc.DialContext(timeout, addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Errorf(ctx, "Error create grpc connection with [%s]", addr)
			return nil, errors.Wrapf(err, "Error create grpc connection with [%s]", addr)
		}
		client := worker0.NewWorker0Client(conn)
		RMDIns.agConnRepo.Store(addr, client)
		return client, nil
	}
	return load.(worker0.Worker0Client), nil
}

func (R *RMD) NewApp(ctx context.Context, in *ag.NewAppReq) (*ag.App, error) {
	// in.MachineID, save to repo
	// 通过id获取主机信息
	// 通过主机信息进行grpc链接, 判断rpc 客户端是否存在
	// 请求 new app 的接口
	log.Infof(ctx, "new app")
	if in.MachineID == 0 || in == nil {
		return nil, errors.New("machine id is nil")
	}
	machine, err := model.GetMachineByID(in.MachineID)
	if err != nil {
		return nil, errors.Wrap(err, "get machine by id")
	}
	log.Debugf(ctx, "machine id [%s], get agent grpc addr [%s]", in.MachineID, machine.AGGRPCAddr)
	rpc, err := connAG(ctx, machine.AGGRPCAddr)
	if err != nil {
		return nil, errors.Wrap(err, "connect to machine agent")
	}

	//outctx := contextBuild(ctx, "initapp")
	app, err := rpc.NewApp(ctx, &worker0.NewAppReq{Name: in.Name, Version: in.Version})
	if err != nil {
		return nil, errors.Wrap(err, "rpc request new app")
	}
	marshal, err := json.Marshal(app)
	if err != nil {
		return nil, errors.Wrap(err, "marshal app")
	}
	log.Debugf(ctx, "app json [%s]", string(marshal))
	if len(app.UUID) == 0 {
		return nil, errors.Errorf("app uuid is nil")
	}
	ags := appstruct(app)
	RMDIns.appRepo.Store(app.UUID, ags)
	RMDIns.appConnRepo.Store(app.UUID, &clientIns{rpcClient: rpc})
	log.Debugf(ctx, "store app [%s] to repo", app.UUID)
	return ags, nil
}

func (R *RMD) StartApp(ctx context.Context, in *ag.App) error {
	log.Infof(ctx, "start app [%s]", in.UUID)
	if len(in.UUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(in.UUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", in.UUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, in.UUID)
	_, err := ins.rpcClient.StartApp(outctx, wappstruct(in))
	if err != nil {
		return errors.Wrap(err, "rpc request start app")
	}
	log.Infof(ctx, "start app success")
	return nil
}

func (R *RMD) StopApp(ctx context.Context, in *ag.App) error {
	log.Infof(ctx, "stop app [%s]", in.UUID)
	if len(in.UUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(in.UUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", in.UUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, in.UUID)
	_, err := ins.rpcClient.StopApp(outctx, wappstruct(in))
	if err != nil {
		return errors.Wrap(err, "rpc request stop app")
	}
	log.Infof(ctx, "stop app success")
	return nil
}

func (R *RMD) DestroyApp(ctx context.Context, in *ag.App) error {
	log.Infof(ctx, "destroy app [%s]", in.UUID)
	if len(in.UUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(in.UUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", in.UUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, in.UUID)
	_, err := ins.rpcClient.DestroyApp(outctx, wappstruct(in))
	if err != nil {
		return errors.Wrap(err, "rpc request destroy app")
	}
	log.Infof(ctx, "destroy app success")
	return nil
}

func (R *RMD) TagEx(ctx context.Context, appUUID string, in *ag.Tag) error {
	log.Infof(ctx, "app exec set tag [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	t, err := ins.rpcClient.TagEx(outctx, &worker0.App_Tag{
		Key:   in.Key,
		Value: in.Value,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request set app tag")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app tag")
		app := value.(*ag.App)
		tagset(app, []*worker0.App_Tag{t})
	}
	log.Infof(ctx, "app exec set tag success")
	return nil
}

func (R *RMD) FileMountEx(ctx context.Context, appUUID string, in *ag.FileMount) error {
	log.Infof(ctx, "app exec file mount [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	r, err := ins.rpcClient.FileMountEx(outctx, &worker0.App_FileMount{
		File:    in.File,
		MountTo: in.MountTo,
		Volume:  in.Volume,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request file mount")
	}
	in.File = r.File
	in.MountTo = r.MountTo
	in.Volume = r.Volume
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app file mount")
		app := value.(*ag.App)
		filemountset(app, []*worker0.App_FileMount{r})
	}
	log.Infof(ctx, "app exec file mount success")
	return nil
}

func (R *RMD) EnvEx(ctx context.Context, appUUID string, in *ag.EnvVar) error {
	log.Infof(ctx, "app exec set env [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	env, err := ins.rpcClient.EnvEx(outctx, &worker0.App_EnvVar{
		Key:   in.Key,
		Value: in.Value,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request set env")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app envs")
		app := value.(*ag.App)
		envset(app, []*worker0.App_EnvVar{env})
	}
	log.Infof(ctx, "app exec set env success")
	return nil
}

func (R *RMD) NetworkEx(ctx context.Context, appUUID string, in *ag.Network) error {
	log.Infof(ctx, "app exec network config [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}
	if in.PortInfo.Port == 0 {
		return errors.Errorf("port info [%v] is nil", in.PortInfo)
	}

	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	net, err := ins.rpcClient.NetworkEx(outctx, &worker0.App_Network{
		PortInfo: &worker0.App_Network_PortInf{
			Port:         int32(in.PortInfo.Port),
			Name:         in.PortInfo.Name,
			ProtocolType: worker0.App_Network_PortInf_Protocol(in.PortInfo.ProtocolType),
		},
		RouteInfo: nil,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request config network")
	}

	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app network")
		app := value.(*ag.App)
		networkset(app, []*worker0.App_Network{net})
	}

	network := &ag.Network{
		PortInfo: struct {
			Port         int         `json:"port"`
			Name         string      `json:"name"`
			ProtocolType ag.Protocol `json:"protocol_type"`
		}{
			Port:         int(net.PortInfo.Port),
			Name:         net.PortInfo.Name,
			ProtocolType: ag.Protocol(int(net.PortInfo.ProtocolType)),
		},
		RouteInfo: []struct {
			RouteType ag.Route `json:"route_type"`
			Router    string   `json:"router"`
		}{},
	}
	if net.RouteInfo != nil {
		for _, inf := range net.RouteInfo {
			network.RouteInfo = append(network.RouteInfo, struct {
				RouteType ag.Route `json:"route_type"`
				Router    string   `json:"router"`
			}{RouteType: ag.Route(int(inf.RouteType)), Router: inf.Router})
		}
	}
	in = network
	log.Infof(ctx, "app exec config network success")
	return nil
}

func (R *RMD) FilePremiseEx(ctx context.Context, appUUID string, in *ag.File) error {
	log.Infof(ctx, "app exec file premise [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}

	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	f, err := ins.rpcClient.FilePremiseEx(outctx, &worker0.App_File{
		Name:        in.Name,
		AcquireAddr: in.AcquireAddr,
		Shell:       in.Shell,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request file premise")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app file premise")
		app := value.(*ag.App)
		filepremiseset(app, []*worker0.App_File{f})
	}
	log.Infof(ctx, "app exec file premise success")
	return nil
}

func (R *RMD) LimitEx(ctx context.Context, appUUID string, in *ag.Limit) error {
	log.Infof(ctx, "app exec set limit [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}

	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	l, err := ins.rpcClient.LimitEx(outctx, &worker0.App_Limit{
		CPU:    int32(in.CPU),
		Memory: int32(in.Memory),
	})
	if err != nil {
		return errors.Wrap(err, "rpc request set limit")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app limit")
		app := value.(*ag.App)
		limitset(app, l)
	}
	log.Infof(ctx, "app exec set limit success")
	return nil
}

func (R *RMD) HealthEx(ctx context.Context, appUUID string, in *ag.Health) error {
	log.Infof(ctx, "app exec set health [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}

	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	rh, err := ins.rpcClient.HealthEx(outctx, &worker0.App_Health{
		Liveness: &worker0.App_Health_Basic{
			MethodType: worker0.App_Health_Basic_Method(in.Liveness.MethodType),
			Path:       in.Liveness.Path,
			Port:       int32(in.Liveness.Port),
		},
		Readness: &worker0.App_Health_Basic{
			MethodType: worker0.App_Health_Basic_Method(in.Readness.MethodType),
			Path:       in.Readness.Path,
			Port:       int32(in.Readness.Port),
		},
	})
	if err != nil {
		return errors.Wrap(err, "rpc request set health")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app health info")
		app := value.(*ag.App)
		healthset(app, rh)
	}
	log.Infof(ctx, "app exec set health success")
	return nil
}

func (R *RMD) LogEx(ctx context.Context, appUUID string, in *ag.Log) error {
	log.Infof(ctx, "app exec set log [%v]", *in)
	if len(appUUID) == 0 || in == nil {
		return errors.New("app uuid is nil, please check")
	}

	load, ok := RMDIns.appConnRepo.Load(appUUID)
	if !ok {
		return errors.Errorf("can not found app by uuid [%s]", appUUID)
	}
	ins := load.(*clientIns)
	outctx := contextBuild(ctx, appUUID)
	lo, err := ins.rpcClient.LogEx(outctx, &worker0.App_Log{
		RealTimeFile: in.RealTimeFile,
		FilePath:     in.FilePath,
	})
	if err != nil {
		return errors.Wrap(err, "rpc request set log info")
	}
	value, tok := RMDIns.appRepo.Load(appUUID)
	if tok {
		log.Infof(ctx, "set local app log info")
		app := value.(*ag.App)
		logset(app, lo)
	}
	log.Infof(ctx, "app exec set log info success")
	return nil
}
