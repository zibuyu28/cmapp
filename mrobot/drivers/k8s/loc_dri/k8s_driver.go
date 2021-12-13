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
	"github.com/zibuyu28/cmapp/common/base64"
	"github.com/zibuyu28/cmapp/common/httputil"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/tmp"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s/kube_driver/base"
	"github.com/zibuyu28/cmapp/mrobot/pkg"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc/metadata"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

var defaultAgentGRPCPort = 9008
var defaultAgentHealthPort = 9009

type DriverK8s struct {
	pkg.BaseDriver
	KubeConfigBase64 string `validate:"required"`
	NodeIP           string `validate:"required,ip"`

	//Token       string `validate:"required"`
	//Certificate string `validate:"required"`
	//ClusterURL  string `validate:"required"`
	Namespace string `validate:"required"`

	StorageClassName string `validate:"required"`
	Labels           map[string]string
}

func NewDriverK8s() {

}

// GetCreateFlags get create flags
func (d *DriverK8s) GetCreateFlags(ctx context.Context, empty *driver.Empty) (*driver.Flags, error) {
	baseFlags := &driver.Flags{Flags: d.GetFlags()}
	flags := []*driver.Flag{
		{
			Name:   "KubeConfigBase64",
			Usage:  "kubernetes cluster config in base64 encryption for connection,  if set this flag, the token,certificate,url can not be set",
			EnvVar: "KUBE_CONFIG_BASE64",
			Value:  []string{defaultConfig},
		},
		{
			Name:   "NodeIP",
			Usage:  "kubernetes cluster ip of one node, format like \"10.1.11.11\"",
			EnvVar: "KUBE_NODEIP",
			Value:  nil,
		},
		{
			Name:   "Namespace",
			Usage:  "kubernetes cluster namespace for machine work",
			EnvVar: "KUBE_NAMESPACE",
			Value:  nil,
		},
		{
			Name:   "StorageClassName",
			Usage:  "storageClass for create pvc and pv",
			EnvVar: "KUBE_STORAGECLASS",
			Value:  nil,
		},
		{
			Name:   "Labels",
			Usage:  "union labels for resources, format like [\"key1=value1\",\"key2=value2\"]",
			EnvVar: "KUBE_LABELS",
			Value:  nil,
		},
	}
	baseFlags.Flags = append(baseFlags.Flags, flags...)
	return baseFlags, nil
}

// SetConfigFromFlags set config from flags
func (d *DriverK8s) SetConfigFromFlags(ctx context.Context, flags *driver.Flags) (*driver.Empty, error) {
	m := d.ConvertFlags(flags)
	d.CoreAddr = m["CoreAddr"]
	d.CoreHTTPAddr = m["CoreHTTPAddr"]
	d.CoreGRPCAddr = m["CoreGRPCAddr"]
	d.ImageRepository.Repository = m["Repository"]
	d.ImageRepository.StorePath = m["StorePath"]

	d.KubeConfigBase64 = m["KubeConfigBase64"]
	d.NodeIP = m["NodeIP"]
	d.Namespace = m["Namespace"]
	d.StorageClassName = m["StorageClassName"]
	labels := m["Labels"]
	d.Labels = make(map[string]string)
	kvs := strings.Split(labels, ",")
	for _, kv := range kvs {
		kav := strings.Split(kv, "=")
		if len(kav) == 2 {
			d.Labels[kav[0]] = kav[1]
		}
	}
	log.Debugf(ctx, "Currently driver get labels [%v]", d.Labels)

	validate := v.New()
	err := validate.Struct(*d)
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

	_, err = base64.Decode(d.KubeConfigBase64)
	if err != nil {
		return nil, errors.Wrap(err, "decode kube config")
	}

	d.DriverName = driverName
	d.DriverVersion = driverVersion
	d.DriverID = driverID
	return &driver.Empty{}, nil
}

func (d *DriverK8s) InitMachine(ctx context.Context, empty *driver.Empty) (*driver.Machine, error) {
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
		"kubeConfigBase64": d.KubeConfigBase64,
		"namespace":        d.Namespace,
	}
	machine := &driver.Machine{
		UUID:       datas[0],
		State:      1,
		Tags:       []string{"k8s"},
		CustomInfo: customInfo,
	}

	return machine, nil
}

func (d *DriverK8s) CreateExec(ctx context.Context, empty *driver.Empty) (*driver.Machine, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to create machine")
	return &driver.Machine{}, nil
}

func (d *DriverK8s) InstallMRobot(ctx context.Context, m *driver.Machine) (*driver.Machine, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to install mrobot")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("CoreMachineID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find core id from metadata")
	}
	coreID, err := strconv.Atoi(datas[0])
	if err != nil {
		return nil, errors.Wrapf(err, "parse core id [%s] to int", datas[0])
	}
	kubeconfig, err := base64.Decode(d.KubeConfigBase64)
	if err != nil {
		return nil, errors.Wrap(err, "decode kube config")
	}
	c, err := base.NewClientByConfig(ctx, kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}

	// TODO: 从 core 获取 image 信息
	image := fmt.Sprintf("%s:%s", d.DriverName, d.DriverVersion)
	repository := d.ImageRepository

	if len(repository.Repository) != 0 && len(repository.StorePath) != 0 {
		image = filepath.Join(repository.Repository, repository.StorePath, image)
	}

	log.Debugf(ctx, "Currently image info [%v]", image)

	const (
		DriAgentNamespace    = "DRIAGENT_NAMESPACE"
		DriAgentKubeConfig   = "DRIAGENT_KUBECONFIG"
		DriAgentNodeIP       = "DRIAGENT_NODEIP"
		DriAgentStorageClass = "DRIAGENT_STORAGECLASS"
		DriAgentK8sUUID      = "DRIAGENT_K8SUUID"
		DriAgentDomain       = "DRIAGENT_DOMAIN"
		DriAgentMachineID    = "DRIAGENT_MACHINE_ID"
		DriCoreHttpAddr      = "DRIAGENT_CORE_HTTP_ADDR"
		DriCoreGrpcAddr      = "DRIAGENT_CORE_GRPC_ADDR"
	)

	mrobotEnvs := map[string]string{
		DriAgentMachineID:    fmt.Sprintf("%d", coreID),
		DriAgentNamespace:    d.Namespace,
		DriAgentNodeIP:       d.NodeIP,
		DriAgentStorageClass: d.StorageClassName,
		DriAgentKubeConfig:   d.KubeConfigBase64,
		DriCoreHttpAddr:      d.CoreHTTPAddr,
		DriCoreGrpcAddr:      d.CoreGRPCAddr,
		DriAgentK8sUUID:      datas[0],
		DriAgentDomain:       "test-domain",
		AgentPluginBuildIn:   "true",
		AgentPluginName:      "k8s",
	}

	tempdata := struct {
		ImageName string
		MachineID int
		Env       map[string]string
		Namespace string
		UUID      string
	}{
		ImageName: image,
		MachineID: coreID,
		Env:       mrobotEnvs,
		Namespace: d.Namespace,
		UUID:      datas[0],
	}

	mrobotDepYaml, err := tmp.AdvanceTemplate(tempdata, []byte(mRobotDep))
	if err != nil {
		return nil, errors.Wrap(err, "generate mrobot deployment yaml")
	}
	log.Debugf(ctx, "%s", mrobotDepYaml)
	var mrobotDepIns = v1.Deployment{}
	err = yaml.Unmarshal(mrobotDepYaml, &mrobotDepIns)
	if err != nil {
		return nil, errors.Wrap(err, "yaml unmarshal to deployment instance")
	}
	err = c.CreateDeployment(&mrobotDepIns)
	if err != nil {
		return nil, errors.Wrap(err, "create deployment")
	}

	mrobotSvcYaml, err := tmp.AdvanceTemplate(tempdata, []byte(mRobotSvc))
	if err != nil {
		return nil, errors.Wrap(err, "generate mrobot service yaml")
	}
	log.Debugf(ctx, "%s", mrobotSvcYaml)
	var mrobotSvcIns = corev1.Service{}
	err = yaml.Unmarshal(mrobotSvcYaml, &mrobotSvcIns)
	if err != nil {
		return nil, errors.Wrap(err, "yaml unmarshal to deployment instance")
	}

	// 使用 node port 方式
	err = c.CreateService(&mrobotSvcIns)
	if err != nil {
		return nil, errors.Wrap(err, "create service")
	}
	svc, err := c.GetService(fmt.Sprintf("%s-service", datas[0]), d.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "get service")
	}
	machine := &driver.Machine{}
	for _, port := range svc.Spec.Ports {
		if port.Port == int32(defaultAgentGRPCPort) && port.NodePort != 0 {
			machine.AGGRPCAddr = fmt.Sprintf("%s:%d", d.NodeIP, port.NodePort)
		}
		if port.Port == int32(defaultAgentHealthPort) && port.NodePort != 0 {
			machine.CustomInfo["health_addr"] = fmt.Sprintf("%s:%d", d.NodeIP, port.NodePort)
		}
	}
	return machine, nil
}

func (d *DriverK8s) MRoHealthCheck(ctx context.Context, m *driver.Machine) (*driver.Machine, error) {
	// (DONE): how to check mrobot health -> deployment check ok
	log.Info(ctx, "Currently k8s machine plugin start to check robot health")
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

func (d *DriverK8s) Exit(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	log.Info(ctx, "Currently k8s machine plugin exit")
	return nil, nil
}

var defaultConfig = "YXBpVmVyc2lvbjogdjENCmNsdXN0ZXJzOg0KLSBjbHVzdGVyOg0KICAgIGNlcnRpZmljYXRlLWF1dGhvcml0eS1kYXRhOiBMUzB0TFMxQ1JVZEpUaUJEUlZKVVNVWkpRMEZVUlMwdExTMHRDazFKU1VSMGFrTkRRWEEyWjBGM1NVSkJaMGxWU25KdkswNVJhSGxRYURoalpEWllZbFZIZW5Ka1VXbFdOSGN3ZDBSUldVcExiMXBKYUhaalRrRlJSVXdLUWxGQmQxbEVSVXhOUVd0SFFURlZSVUpvVFVOUk1EUjRSVVJCVDBKblRsWkNRV2RVUWpCS1JsTlZjRXBVYTJONFEzcEJTa0puVGxaQ1FXTlVRV3hvVkFwTlVYZDNRMmRaUkZaUlVVdEZkMDV5VDBoTmVFUjZRVTVDWjA1V1FrRnpWRUpzVGpWak0xSnNZbFJGVkUxQ1JVZEJNVlZGUVhoTlMyRXpWbWxhV0VwMUNscFlVbXhqZWtGblJuY3dlVTFFUVROTlJFVjNUV3BSZUUxRVFtRkhRVGg1VFZSSmQwMUVXWGRPZWtGNVRrUkZkMDFHYjNkWlJFVk1UVUZyUjBFeFZVVUtRbWhOUTFFd05IaEZSRUZQUW1kT1ZrSkJaMVJDTUVwR1UxVndTbFJyWTNoRGVrRktRbWRPVmtKQlkxUkJiR2hVVFZGM2QwTm5XVVJXVVZGTFJYZE9jZ3BQU0UxNFJIcEJUa0puVGxaQ1FYTlVRbXhPTldNelVteGlWRVZVVFVKRlIwRXhWVVZCZUUxTFlUTldhVnBZU25WYVdGSnNZM3BEUTBGVFNYZEVVVmxLQ2t0dldrbG9kbU5PUVZGRlFrSlJRVVJuWjBWUVFVUkRRMEZSYjBOblowVkNRVW95VWxCYU9GcHJaMFZtVms5Uk5EUXdlVkZYTm1WVGRHUkZiblYzZFRjS05sVTFRWHAwV2tGVE5qUXhPRmMxZVhNclJIUnFUelpFTm01SlVFZFhWblpXUTNWcVFuRnNOa2xIZUhCR2QzTXZRVWt2VjJaclJqRlVXVVIzTkRsR01ncDFWbHBCWjBWaGNpOHhlRUZsV0N0NlpqVnNZV2g1YWs5ME4ycEhUMUl2ZHpFNU1IbHhTMFl4UWxkaWNtZHZWM1l6ZVZFeFlqZE9VbFkwYkdaaFNsVktDa2x5U1V0WVMwTXlja00xVUM5WVZVdFRNa2hHWjBNNFZGaHdSM2xHWmtkU00wSnVSV0pJTmtSUVEzcExOMDlEYWpWdmQwUk9kRkJhY1dWSWVGY3dPSFVLTUhCRmRuVnVXazVoVkRsb2JFMDBWQ3MwTTA0M05rOW9hVFpIU2s5MmJYUm9kbkJGVFRkTGFuSlhXRzA1WjNVMGFYSnlNbmR0YzJwQlVHVTNPRGg2YWdvdlVqWXhLelpDVVdGNlowNXpVRmd3UVVabGJISlZjRTlKUzBkQ1ZXZG1kRk4xV1ZoUGFXaHFlVTh6WmtOeWIyNUJjU3RXU21FNFEwRjNSVUZCWVU1dENrMUhVWGRFWjFsRVZsSXdVRUZSU0M5Q1FWRkVRV2RGUjAxQ1NVZEJNVlZrUlhkRlFpOTNVVWxOUVZsQ1FXWTRRMEZSU1hkSVVWbEVWbEl3VDBKQ1dVVUtSa3A0YVhkUksycDZiR1JPVDNadGJGbElTbmgyYTJGUFRtUldlazFDT0VkQk1WVmtTWGRSV1UxQ1lVRkdTbmhwZDFFcmFucHNaRTVQZG0xc1dVaEtlQXAyYTJGUFRtUldlazFCTUVkRFUzRkhVMGxpTTBSUlJVSkRkMVZCUVRSSlFrRlJRazFTVmpGNVJIUlNVaXRITmk4clQzSk5PR0UzVFVFNFZWTlJkM1Y1Q21vNVJYcHhORmd3Y2twQ1JqTlNRbG93ZFZCVFRrODVlV3RwYVRNdlRuWlpSRWw0YkM5cE1uaExjVVZWVEZOQ1lWRndSREpWWVRSUFpYaDBlRFlyTWxFS2NUUnFWMGdyWVdOYU1WTk9VM0U0TWpSNFEwUmpTazVyVkc5MmJGbHdNM2xTVVdSNlMycGFLMFpDYVZKcFNFdHFhSFJtYVcxUmNFRnBLM1Y1VVhGd1NBb3pabEVyY0ZSTGRFSkRiM2w0WW5kMFEzcEhaVlpHV0VKWVFYWTNWVGhVVFZONlNIcFRTVEZ5TDFJNE9XRTBjelFyTVZkUFltaDBURTQ0UVM5UGVIUm1DazVCVkhkbWMycDBhbVJvZUc5eVNqRjZNM2RVYTJ4VGRFVlRWbVJNVW5ObFVHVnZhWGd3VjBzMFdtRXdTbkF2UzJ4UGIzZFpaSEpzVmpnelJuSlpjMWNLWTNWSFRERlhabkp1WkhaR1oxWjZZemc0UmxWeU1UQTBSRFpZTjNoblZYRTNZblkyZDA1RFNFMURjV04zYWpkR1lUbEZTRll2UlVvS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPQ0KICAgIHNlcnZlcjogaHR0cHM6Ly8xNzIuMTYuMS4xMjE6NjQ0Mw0KICBuYW1lOiBjbHVzdGVyMQ0KY29udGV4dHM6DQotIGNvbnRleHQ6DQogICAgY2x1c3RlcjogY2x1c3RlcjENCiAgICB1c2VyOiB1c2VyMQ0KICBuYW1lOiBjb250ZXh0MQ0KY3VycmVudC1jb250ZXh0OiBjb250ZXh0MQ0Ka2luZDogQ29uZmlnDQpwcmVmZXJlbmNlczoge30NCnVzZXJzOg0KLSBuYW1lOiB1c2VyMQ0KICB1c2VyOg0KICAgIGNsaWVudC1jZXJ0aWZpY2F0ZS1kYXRhOiBMUzB0TFMxQ1JVZEpUaUJEUlZKVVNVWkpRMEZVUlMwdExTMHRDazFKU1VReFZFTkRRWEl5WjBGM1NVSkJaMGxWVEV0TWMzVlVRbHAwUTJkVlVIaEtVV3BaYUhSSFVXNVVia2RSZDBSUldVcExiMXBKYUhaalRrRlJSVXdLUWxGQmQxbEVSVXhOUVd0SFFURlZSVUpvVFVOUk1EUjRSVVJCVDBKblRsWkNRV2RVUWpCS1JsTlZjRXBVYTJONFEzcEJTa0puVGxaQ1FXTlVRV3hvVkFwTlVYZDNRMmRaUkZaUlVVdEZkMDV5VDBoTmVFUjZRVTVDWjA1V1FrRnpWRUpzVGpWak0xSnNZbFJGVkUxQ1JVZEJNVlZGUVhoTlMyRXpWbWxhV0VwMUNscFlVbXhqZWtGblJuY3dlVTFFUVROTlJFVjNUV3BSZUUxRVFtRkhRVGg1VFVSamQwMUVXWGhQVkVGNVRrUkZkMDFHYjNkYWFrVk1UVUZyUjBFeFZVVUtRbWhOUTFFd05IaEZSRUZQUW1kT1ZrSkJaMVJDTUVwR1UxVndTbFJyWTNoRGVrRktRbWRPVmtKQlkxUkJiR2hVVFZKamQwWlJXVVJXVVZGTFJYYzFlZ3BsV0U0d1dsY3dObUpYUm5wa1IxWjVZM3BGVUUxQk1FZEJNVlZGUTNoTlIxVXpiSHBrUjFaMFRWRTBkMFJCV1VSV1VWRkVSWGRXYUZwSE1YQmlha05EQ2tGVFNYZEVVVmxLUzI5YVNXaDJZMDVCVVVWQ1FsRkJSR2RuUlZCQlJFTkRRVkZ2UTJkblJVSkJURWdyZG5sUVduWXZOMDlzTUc0MGQwOXVaM2RHTHpJS2VTOUdlVXBTZURaWVQwVkZNekUzVjFCTGRVeDNTRUpoWjFaaGNHRklaRkZaSzBWSWJqVnJVRXAwTldKNFdraERVWFZ0VWtveFRGaHRRWE1yV2xoa1NnbzFNWGN6WVZrNWNEaFFaelkwVWpVMFpuUjNSbWdyTkU1SE5XTXpkRmxJVlVWdk9UTnpNbTlXVUdwUlUwbEVPVkJKVmpKNmMxaHlRMWxZT0dreVVISlBDbGxHTDBvMGEyVXJORTl4Y0VGdVlVSm9ka2hrWkcxQmIyMTJjbXhQUkRsV01IVXJZbEZaYzFZeFJFc3pVV05CUkN0d1dEbG1WMHMzUlVoTk5HZENlVmNLWVRGTlZYRTBZbWMwUlRSRlVEQlRTV3R1Wm00M2MwbFlZbGRSVG5GblJYY3laazVWUVV3d2NGSTFURkZyU20xR01Hc3lSbFpTV2k5ME1GVkhUMWxzUXdwaGEydE5TM04zTkhZMldrMVphVmhaUVN0eVJIWXdjaXREUlZrMFduZG5kbmR3Vmt0clkwNDVjR1F2UW5RNGRWWnBNM3BKWkV4NlN5OUtMelpJUW10RENrRjNSVUZCWVU0dlRVZ3dkMFJuV1VSV1VqQlFRVkZJTDBKQlVVUkJaMWRuVFVJd1IwRXhWV1JLVVZGWFRVSlJSME5EYzBkQlVWVkdRbmROUWtKblozSUtRbWRGUmtKUlkwUkJha0ZOUW1kT1ZraFNUVUpCWmpoRlFXcEJRVTFDTUVkQk1WVmtSR2RSVjBKQ1VTOURlVlJwVEhkS2QyRndPRzl0VlVkSWVYcEJad3BSYkZwRWVVUkJaa0puVGxaSVUwMUZSMFJCVjJkQ1UyTlpjMFZRYnpnMVdGUlVjalZ3VjBKNVkySTFSMnBxV0ZaamVrRk9RbWRyY1docmFVYzVkekJDQ2tGUmMwWkJRVTlEUVZGRlFXdDBkazlDVWtKS09HRjJlV0pFYWpNNFFqSk1RM1Z1TVdOaE1sUnphMk5DY1dsUGFsQmFkWE5ITTFWRFNtaDVWMnRxY1ZBS1ZWVlFWRWt4VjFkU2FHRlNSMlJHTTNGdE1HdDRla2xvYjA4dlRESXlNMUZXYTFGVVZuVklaWEpRVVdwSlNXdDBjSEpxZGpWbmRWUlhkREJ3WTJaQlJRcG1kVzQyVjNSSFltaHhXbmwwYzBkbEwxaEtkR1F3TUVsTUsxTXdjRWtyZUZSMFRYbFlRa2QyY2s1VGFEVlRaazAxYlVvcmJuSk1TbkpRTkUxMVdIWjFDbXd3UmxrM1ZUTnhhVWMwU0d0b1ZHUkJVR041TlRsemVqaENWRmhoVnk4eE5XZEVObkZyTDFWbmFrcDVPVGxQV1dwVE1WSkdNWGs1V25CNFZuYzFhVTRLV1U5RlJGbGFXSGN4VGxwMk1rRlNlVkZUVlhwd01sY3JURkZvZFc5RVJuQjJOVTVqUkhoamJpODVWSFU0VkdWemVXUkNjM2hGVURaRVJqVndOemhLU3dvdllWbHZTVVpVTjB0MWJrVmlkWGdyU1haWWFFMVRXa0ZFYlhkMGEwSTFhRVJuUFQwS0xTMHRMUzFGVGtRZ1EwVlNWRWxHU1VOQlZFVXRMUzB0TFFvPQ0KICAgIGNsaWVudC1rZXktZGF0YTogTFMwdExTMUNSVWRKVGlCU1UwRWdVRkpKVmtGVVJTQkxSVmt0TFMwdExRcE5TVWxGYjNkSlFrRkJTME5CVVVWQmMyWTJMMGs1YlM4dmN6WllVMlpxUVRabFJFRllMMkpNT0ZoSmJFaEljR00wVVZSbVdIUlpPSEUwZGtGalJuRkNDbFp4Ykc5a01VSnFORkZsWm0xUk9HMHpiSFpHYTJOS1F6WmFSVzVWZEdWWlEzbzFiR1F3Ym01WVJHUndhakp1ZHl0RWNtaElibWdyTTBGWFNEZG5NR0lLYkhwbE1XZGtVVk5xTTJWNllXaFZLMDVDU1dkUU1EaG9XR0pQZUdWelNtaG1lVXhaSzNNMVoxZzRibWxTTnpkbk5uRnJRMlJ2UjBjNFpERXlXVU5wWVFvcmRWVTBVREZZVXpjMWRFSnBlRmhWVFhKa1FuZEJVRFpzWmpFNVdYSnpVV042YVVGSVNscHlWWGhUY21oMVJHZFVaMUV2VWtscFUyUXJablYzYUdSMENscEJNbkZCVkVSYU9ERlJRWFpUYkVocmRFTlJiVmxZVTFSWlZsWkdiaXN6VWxGWk5XbFZTbkZUVVhkeGVrUnBMM0JyZUdsS1pHZEVObk5QTDFOMk5Fa0tVbXBvYmtOREwwTnNWWEZTZHpNeWJETTRSek41TlZkTVprMW9NSFpOY2podUwyOWpSMUZKUkVGUlFVSkJiMGxDUVVkMVNHOXhSbVEwZDJSclpIcEdVUW92Wm5CTmRFOTBSV2hYUTFCMlMzVXZiMGQ1ZDB4UFJqSlBOa1JJUWt4TVltVnNaVWxqU0haclRGQkxPVlZGVmk5VFpGQTNWRkpuU21FM1RDczFaWEVyQ2twRVVtMTBXbGRxUzBGdmJIbzNaVGhKVERsV2EzSndPWGRQV0dFd1dYVlhhalY1UkVsNmQxUnBhMHg0TlZsdGJ6STNUa3BaUVVobVNrTkZabVJhUkdJS04xWnhTa04wZHpVMFV6YzVTRFUyU1ZObmFEVnZaWFJHZUU5b1NVSjZZa3BNTDB4NU1WRnplVWRXVkhBelNXSlZVWGhFUWpnclNVUXZMMjlSUnpkNVVRcHFXamxSVkd4WVpIRlFhMkU0VHk5NFZuWTVSM2xuY2xsM1dYUXpkSEEwZURKblkzcFhkRFJHV0c1TFRWVnNNR2hCZFdrNWFrZGxVMEphTVZGSlJraEhDa0YxUlhCaVYwdENObGhDTW5OVFFqWlNkVVJ2UWt4blIzcHNPRmxQV0hoYWMwSkhabnBEUkRWVGVYbExNemRJUkVSc1lWQktiekZNUWxBM04xRnljRzRLWTFSakswTkpSVU5uV1VWQk5tMTBlRXB4WVVsU1Yyd3pibGRHY0Vsb1RFdFhWV2hoUkhSWFkyOUlheXRuYTJZNGNqQnRZVnBhUnpOMGVGTXlTMG8zTkFwNVNWUkJaMHMwYXl0bE9WbzRSMnQwVFU1elJsVmFPV1JLU2xFeGVURlRSbTFGYm5CMUsya3JPWFp0VTBGYVoyMWtjRVZTUlcxUFpuSnhSM1EwVUhaekNtTXhVRTVwZUVabFRrbDBibFJtVWk5U1lUbFdOMkpyYjB4UGNtRjVTbkpwSzNOMFkzZGxWMlZ3UlRSYVp6ZERkbXBqZUhOT2JEQkRaMWxGUVhkdFIwMEthREZoYzFCNkx6Tm9jMWw1VGpWSlRUQXlUMVY0Y1RCQkswWjNVRmxVZUN0bE1WcDBWblpxWlZweVdTOXVaVTlhVG5oUVVtaFlaekoxUkdOVVZsaG9Zd3AzYnpSRU56RXlPVXhhV2t4SlRTOVJZbU5zU2xrdlJteEVkSEJoWmtsQ1ZFOVBWeXRXVERkUFVYaDZlSGRQVm1obmNtZFpaa2sxZEZGRFlqaHpUemR4Q21WMWVtSlhXbWM1Y1dwamVsQk1RWEZITjNScVMzcE9haXN4Wm1KWVZsRk1NWFZyZFdGUE1FTm5XVUYyVWs1RlFrdFRNWEpUWWxGTU9GSjFRU3R0WTBJS1JEbDZVa0oyVUZwVEwyd3hlbU4zTDBWRmJHOHhOMUpETXpWT1NIQTFObXN5Um1Zd01uRnBjMGRVTlRabmMzSlNRV2xyVmxveGQyaDRlSG81ZW1sS1R3cE1aMFpMYldOclFVSm9NRkkwZW5SSFVqQlJOVUZCZFhZMkwzZHBNVTF0SzBNMVZIcDVUMGhHUjB0TlVrUm5PRzR6ZHpSc2NrZE9WV2N4Y25kbWIweHpDbUpXZEdJMFpFSlhZekJWYW5vNWJtMWlkVlZZU2xGTFFtZEhOM0J2U0dvNVFVc3pZV1ozS3pnclduaDRlbkpYZW5oWmJWUXpPVVJZU0hOT05UbERXRWNLVXk4NWNsa3JUM2h5UWpCWllVTXhaMDV3V2pBMFMzY3JZVTQzUmtob1F6VllaemwzUzIxcVpuTkRRMW93TjBaQ1VVbHZlR3BHZEhOU2JVNTVRMWx4VndwMFVVeEJSMUYwT0dOWmJVTnRVMEYzYTFoQ1NrTktaWEJhTVdoRFEzcE9NMEZVV21aUmVHaHJRazh4YkVNM1NHeFBOVWhFT0doSVVuTk5OMUpFVTNockNuY3lSMFpCYjBkQ1FVcDBPRVZSUWtGaVMzcEdLMnBuZEVSNmVVWjJaamRNZUdWclFqVlJNalZZVFdsVlYzQkNaamhTWkhWTGJHOUxSRWxDWm1oalFtMEtkalEzUjJ4elUwaFBMeXRhYzB0dk9WVnZWR05wUTFWSlJrOVFUbXB1TUdOT2JHMW9VVmR3VTBabFNubEJaVGhTZUZCbFJVeFlaVWRITVNzMlNHVjNad3B4TUhoR1NHZEVlWGtyZEd0clkyb3ZUbFV6TTJaU1YzUm9Ta05RVWk5WFVrbFpPSEpPVkhOTmNEUk9VMU5PUTNsa2NWWlJDaTB0TFMwdFJVNUVJRkpUUVNCUVVrbFdRVlJGSUV0RldTMHRMUzB0Q2c9PQ=="

var mRobotDep = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: {{.Namespace}}
  name: {{.UUID}}
  labels:
    app: {{.UUID}}
    machine_id: {{.MachineID}}
spec:
  replicas: 1
  strategy: {}
  selector:
    matchLabels:
      app: {{.UUID}}
      machine_id: {{.MachineID}}
  template:
    metadata:
      labels:
        app: {{.UUID}}
        machine_id: {{.MachineID}}
    spec:
      containers:
        - name: kw
          image: {{.ImageName}}
          imagePullPolicy: Always
          resources:
            requests:
              cpu: "100m"
              memory: "100Mi"
            limits:
              cpu: "100m"
              memory: "1024Mi"
          livenessProbe:
            httpGet:
              path: /healthz 
              port: 9009    
            initialDelaySeconds: 300
            periodSeconds: 15
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          readinessProbe:
            httpGet:
              path: /healthz  
              port: 9009      
            initialDelaySeconds: 10
            periodSeconds: 10
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          env:{{ range $key, $value := .Env }}
            - name: {{ $key }}
              value: {{ $value }}{{ end }}
          ports:
            - containerPort: 9009
              name: health
            - containerPort: 9008
              name: grpc
`
var mRobotSvc = `---
kind: Service
apiVersion: v1
metadata:
  labels:
    app: {{.UUID}}
    machine_id: {{.MachineID}}
  name: {{.UUID}}-service
  namespace: {{.Namespace}}
spec:
  ports:
    - port: 9009
      protocol: TCP
      name: health
      targetPort: 9009
    - port: 9008
      protocol: TCP
      name: grpc
      targetPort: 9008
  selector:
    app: {{.UUID}}
    machine_id: {{.MachineID}}
  type: NodePort
`
