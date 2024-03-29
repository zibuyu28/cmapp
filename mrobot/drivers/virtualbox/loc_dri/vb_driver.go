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

package loc_dri

import (
	"context"
	"fmt"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/httputil"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/ssh_cmd"
	virtualbox "github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/vboxm"
	"github.com/zibuyu28/cmapp/mrobot/pkg"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/worker"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	PluginEnvDriverName    = "PLUGIN_DRIVER_NAME"
	PluginEnvDriverVersion = "PLUGIN_DRIVER_VERSION"
	PluginEnvDriverID      = "PLUGIN_DRIVER_ID"
)

const (
	AgentPluginName    = "AGENT_PLUGIN_DRIVER_NAME"
	AgentPluginBuildIn = "AGENT_PLUGIN_BUILD_IN"
)

type DriverVB struct {
	pkg.BaseDriver
	ServerSSHHost     string `validate:"required"`
	ServerSSHPort     int    `validate:"required"`
	ServerSSHUsername string `validate:"required"`
	ServerSSHPassword string `validate:"required"`
	ServerVMStorePath string `validate:"required"`
}

func NewDriverVB() *DriverVB {
	return &DriverVB{}
}

func (d *DriverVB) GetCreateFlags(ctx context.Context, empty *driver.Empty) (*driver.Flags, error) {
	baseFlags := &driver.Flags{Flags: d.GetFlags()}
	flags := []*driver.Flag{
		{
			Name:   "VBServerSSHHost",
			Usage:  "virtualbox server ssh host",
			EnvVar: "VB_SERVER_SSH_HOST",
			Value:  nil,
		},
		{
			Name:   "VBServerSSHPort",
			Usage:  "virtualbox server ssh port",
			EnvVar: "VB_SERVER_SSH_PORT",
			Value:  nil,
		},
		{
			Name:   "VBServerSSHUserName",
			Usage:  "virtualbox server ssh username",
			EnvVar: "VB_SERVER_SSH_USERNAME",
			Value:  nil,
		},
		{
			Name:   "VBServerSSHPassword",
			Usage:  "virtualbox server ssh password",
			EnvVar: "VB_SERVER_SSH_PASSWORD",
			Value:  nil,
		},
		{
			Name:   "VBServerVMStorePath",
			Usage:  "virtualbox server ssh password",
			EnvVar: "VB_SERVER_VM_STORE_PATH",
			Value:  nil,
		},
	}
	baseFlags.Flags = append(baseFlags.Flags, flags...)
	return baseFlags, nil
}

func (d *DriverVB) SetConfigFromFlags(ctx context.Context, flags *driver.Flags) (*driver.Empty, error) {
	m := d.ConvertFlags(flags)
	d.CoreAddr = m["CoreAddr"]
	d.CoreHTTPAddr = m["CoreHTTPAddr"]
	d.CoreGRPCAddr = m["CoreGRPCAddr"]
	d.ImageRepository.Repository = m["Repository"]
	d.ImageRepository.StorePath = m["StorePath"]

	d.ServerSSHHost = m["VBServerSSHHost"]
	portStr := m["VBServerSSHPort"]
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.Wrapf(err, "parse ssh port info [%s]", portStr)
	}
	d.ServerSSHPort = portInt

	d.ServerSSHUsername = m["VBServerSSHUserName"]
	d.ServerSSHPassword = m["VBServerSSHPassword"]
	d.ServerVMStorePath = m["VBServerVMStorePath"]

	validate := v.New()
	err = validate.Struct(d)
	if err != nil {
		return nil, errors.Wrap(err, "validate param")
	}

	driverName := os.Getenv(PluginEnvDriverName)
	if len(driverName) == 0 {
		return nil, errors.Errorf("fail to get driver name from env, please check env [%s]", PluginEnvDriverName)
	}

	driverVersion := os.Getenv(PluginEnvDriverVersion)
	if len(driverVersion) == 0 {
		return nil, errors.Errorf("fail to get driver version from env, please check env [%s]", PluginEnvDriverVersion)
	}

	driverIDStr := os.Getenv(PluginEnvDriverID)
	if len(driverIDStr) == 0 {
		return nil, errors.Errorf("fail to get driver id from env, please check env [%s]", PluginEnvDriverID)
	}

	driverID, err := strconv.Atoi(driverIDStr)
	if err != nil {
		return nil, errors.Errorf("fail to parse driver id by driverStr [%s], please check env [%s]", driverIDStr, PluginEnvDriverID)
	}

	d.DriverName = driverName
	d.DriverVersion = driverVersion
	d.DriverID = driverID

	// check params
	cli, err := ssh_cmd.NewSSHCli(d.ServerSSHHost, d.ServerSSHPort, d.ServerSSHUsername, d.ServerSSHPassword)
	if err != nil {
		return nil, errors.Wrap(err, "test ssh flags")
	}
	_ = cli.Close()
	return &driver.Empty{}, nil
}

func (d *DriverVB) InitMachine(ctx context.Context, empty *driver.Empty) (*driver.Machine, error) {
	log.Debug(ctx, "Currently virtualbox machine plugin start to init machine")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("UUID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find uuid from metadata")
	}
	var customInfo = map[string]string{
		"server_ssh_host":      d.ServerSSHHost,
		"server_ssh_port":      strconv.Itoa(d.ServerSSHPort),
		"server_ssh_username":  d.ServerSSHUsername,
		"server_ssh_password":  d.ServerSSHPassword,
		"server_vm_store_path": d.ServerVMStorePath,
	}
	return &driver.Machine{
		UUID:       datas[0],
		State:      1,
		Tags:       []string{"virtualbox"},
		CustomInfo: customInfo,
	}, nil
}

// CreateExec 将empty参数变更为machine -> 解决：不需要
func (d *DriverVB) CreateExec(ctx context.Context, empty *driver.Empty) (*driver.Machine, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("UUID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find uuid from metadata")
	}

	// 1. 使用sdk请求远程的vbox webserver 创建一个主机 ----> 这个方式很复杂，主要vb支持的是webservice，soap协议。
	//    go目前没有完善的配套，需要从零开发，所以放弃
	// 2. 使用远程ssh的方式，使用shell创建主机， 直接调用 vboxManage create, 并且开启ssh端口映射
	log.Debug(ctx, "Currently start to create machine exec")
	cli, err := ssh_cmd.NewSSHCli(d.ServerSSHHost, d.ServerSSHPort, d.ServerSSHUsername, d.ServerSSHPassword)
	if err != nil {
		return nil, errors.Wrap(err, "new ssh cli")
	}
	defer cli.Close()
	rmtDriver := virtualbox.NewRMTDriver(ctx, datas[0], d.ServerVMStorePath, d.ServerSSHHost, cli)
	err = rmtDriver.Create()
	if err != nil {
		return nil, errors.Wrap(err, "create vm")
	}
	port, err := rmtDriver.GetSSHPort()
	if err != nil {
		return nil, errors.Wrap(err, "get virtualbox ssh port")
	}
	if port == 0 {
		return nil, errors.New("get virtualbox ssh port is nil")
	}

	file, err := ioutil.ReadFile(rmtDriver.GetSSHKeyPath())
	if err != nil {
		return nil, errors.Wrap(err, "read ssh key")
	}
	var customInfo = map[string]string{
		"virtualbox_ssh_port":     fmt.Sprintf("%d", port),
		"virtualbox_ssh_host":     d.ServerSSHHost,
		"virtualbox_ssh_username": rmtDriver.GetSSHUsername(),
		"virtualbox_ssh_password": "",
		"virtualbox_ssh_key_path": rmtDriver.GetSSHKeyPath(),
		"virtualbox_ssh_key":      string(file),
	}
	return &driver.Machine{
		UUID:       datas[0],
		State:      1,
		CustomInfo: customInfo,
	}, nil
}

func (d *DriverVB) InstallMRobot(ctx context.Context, m *driver.Machine) (*driver.Machine, error) {
	// 1. 请求远程vb webserver 安装 ha ----> 调研后发现不支持

	// 2. 远程shell的方式可以直接创建 ----> 还是使用远程ssh的方式安装
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("UUID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find uuid from metadata")
	}

	log.Debug(ctx, "Currently start to install mrobot")
	// 远程ssh copy mrobot 到 virtualbox 内部
	vbSSHPort := m.CustomInfo["virtualbox_ssh_port"]
	vbSSHHost := m.CustomInfo["virtualbox_ssh_host"]
	atoi, err := strconv.Atoi(vbSSHPort)
	if err != nil {
		return nil, errors.Wrapf(err, "parse virtualbox ssh port [%s]", vbSSHPort)
	}
	vbSSHUsername := m.CustomInfo["virtualbox_ssh_username"]
	//vbSSHPassword := m.CustomInfo["virtualbox_ssh_password"]
	vbSSHKeypath := m.CustomInfo["virtualbox_ssh_key_path"]

	cli, err := ssh_cmd.NewSSHCliWithKey(vbSSHHost, atoi, vbSSHUsername, vbSSHKeypath)
	if err != nil {
		return nil, errors.Wrap(err, "new ssh cli")
	}
	defer cli.Close()
	s := os.Args[0]
	goos := runtime.GOOS
	if goos != "linux" {
		s = s + "-linux"
	}
	abs, err := filepath.Abs(s)
	if err != nil {
		return nil, errors.Wrapf(err, "get abs path of [%s]", s)
	}

	var rmtPath = "/home/docker/virtualbox-mrobot"
	err = cli.Scp(abs, rmtPath)
	if err != nil {
		return nil, errors.Wrap(err, "scp local plugin to remote")
	}

	mids := md.Get("CoreMachineID")
	if len(mids) != 1 {
		return nil, errors.New("fail to find core id from metadata")
	}
	coreID, err := strconv.Atoi(mids[0])
	if err != nil {
		return nil, errors.Wrapf(err, "parse core id [%s] to int", mids[0])
	}


	const (
		DriAgentHostIP        = "DRIAGENT_HOST_IP"
		DriAgentHostPort      = "DRIAGENT_HOST_PORT"
		DriAgentHostUsername  = "DRIAGENT_HOST_USERNAME"
		DriAgentHostPassword  = "DRIAGENT_HOST_PASSWORD"
		DriAgentHostStorePath = "DRIAGENT_HOST_STORE_PATH"
		DriAgentVBUUID        = "DRIAGENT_VBUUID"
		DriAgentMachineID     = "DRIAGENT_MACHINE_ID"

		DriCoreHttpAddr = "DRIAGENT_CORE_HTTP_ADDR"
		DriCoreGrpcAddr = "DRIAGENT_CORE_GRPC_ADDR"
	)

	mrobotEnvs := map[string]string{
		DriAgentMachineID:     fmt.Sprintf("%d", coreID),
		DriAgentHostIP:        d.ServerSSHHost,
		DriAgentHostPort:      fmt.Sprintf("%d", d.ServerSSHPort),
		DriAgentHostUsername:  d.ServerSSHUsername,
		DriAgentHostPassword:  d.ServerSSHPassword,
		DriAgentHostStorePath: d.ServerVMStorePath,
		DriCoreHttpAddr:       d.CoreHTTPAddr,
		DriCoreGrpcAddr:       d.CoreGRPCAddr,
		DriAgentVBUUID:        datas[0],
		AgentPluginBuildIn:    "true",
		AgentPluginName:       "virtualbox",
	}

	// 远程执行启动命令
	_, err = cli.ExecCmd("nohup /home/docker/virtualbox-mrobot ag >hostagent.log 2>&1 &", ssh_cmd.WithEnv(mrobotEnvs))
	if err != nil {
		return nil, errors.Wrap(err, "run cmd to start mrobot")
	}
	//log.Debugf(ctx, "")
	//if !strings.Contains(out, "agent start ok") {
	//	return nil, errors.New("fail to start mrobot")
	//}
	log.Debug(ctx, "Currently install mrobot success")

	// 将 9008 端口映射出来，并且返回外部可以访问的地址
	sshCli, err := ssh_cmd.NewSSHCli(d.ServerSSHHost, d.ServerSSHPort, d.ServerSSHUsername, d.ServerSSHPassword)
	if err != nil {
		return nil, errors.Wrap(err, "new vb server client")
	}
	defer sshCli.Close()
	rmtDriver := virtualbox.NewRMTDriver(ctx, datas[0], d.ServerVMStorePath, d.ServerSSHHost, sshCli)
	grpcPort, err := rmtDriver.ExportPort("ag_grpc_port", "tcp", worker.AGGRPCDefaultPort)
	if err != nil {
		return nil, errors.Wrap(err, "export grpc port")
	}
	healthPort, err := rmtDriver.ExportPort("ag_health_port", "tcp", worker.AGHealthDefaultPort)
	if err != nil {
		return nil, errors.Wrap(err, "export health port")
	}
	m.AGGRPCAddr = fmt.Sprintf("%s:%d", d.ServerSSHHost, grpcPort)
	m.CustomInfo["health_addr"] = fmt.Sprintf("%s:%d", d.ServerSSHHost, healthPort)
	m.State = 1
	m.UUID = datas[0]
	return m, nil
}

func (d *DriverVB) MRoHealthCheck(ctx context.Context, m *driver.Machine) (*driver.Machine, error) {
	// 使用远程ssh的方式调用version接口
	s, ok := m.CustomInfo["health_addr"]
	if !ok {
		return nil, errors.New("health addr is nil")
	}
	get, err := httputil.HTTPDoGet(fmt.Sprintf("http://%s/healthz", s))
	if err != nil {
		return nil, errors.Wrap(err, "request to agent healthz interface")
	}
	if string(get) == "ok" {
		m.State = 2
	}
	return m, nil
}

func (d *DriverVB) Exit(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	panic("implement me")
}
