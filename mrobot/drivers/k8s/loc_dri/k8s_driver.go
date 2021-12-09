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
	"strconv"
	"strings"
)

const (
	PluginEnvDriverName    = "MACHINE_PLUGIN_DRIVER_NAME"
	PluginEnvDriverVersion = "MACHINE_PLUGIN_DRIVER_VERSION"
	PluginEnvDriverID      = "MACHINE_PLUGIN_DRIVER_ID"
)

const (
	AgentPluginName    = "AGENT_PLUGIN_DRIVER_NAME"
	AgentPluginBuildIn = "AGENT_PLUGIN_BUILD_IN"
)

var defaultAgentGRPCPort = 9008
var defaultAgentHealthPort = 9009

type DriverK8s struct {
	pkg.BaseDriver
	KubeConfig string `validate:"required"`
	NodeIP     string `validate:"required,ip"`

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
			Name:   "KubeConfig",
			Usage:  "kubernetes cluster config for connection, if set this flag, the token,certificate,url can not be set",
			EnvVar: "KUBE_CONFIG",
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
	d.ImageRepository.Repository = m["Repository"]
	d.ImageRepository.StorePath = m["StorePath"]

	d.KubeConfig = m["KubeConfig"]
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

	d.DriverName = driverName
	d.DriverVersion = driverVersion
	d.DriverID = driverID
	return nil, nil
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
		"config":    d.KubeConfig,
		"namespace": d.Namespace,
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
	return nil, nil
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

	c, err := base.NewClientByConfig(ctx, []byte(defaultConfig))
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}

	// TODO: 从 core 获取 image 信息
	repository := d.ImageRepository
	image := fmt.Sprintf("%s/%s/%s:%s", repository.Repository, repository.StorePath, d.DriverName, d.DriverVersion)
	log.Debugf(ctx, "Currently image info [%v]", image)

	const (
		DriAgentNamespace    = "DRIAGENT_NAMESPACE"
		DriAgentKubeConfig   = "DRIAGENT_KUBECONFIG"
		DriAgentNodeIP       = "DRIAGENT_NODEIP"
		DriAgentStorageClass = "DRIAGENT_STORAGECLASS"
		DriAgentK8sUUID      = "DRIAGENT_K8SUUID"
		DriAgentMachineID    = "DRIAGENT_MACHINE_ID"
		DriCoreHttpAddr      = "DRIAGENT_CORE_HTTP_ADDR"
		DriCoreGrpcAddr      = "DRIAGENT_CORE_GRPC_ADDR"
	)

	mrobotEnvs := map[string]string{
		DriAgentMachineID:    fmt.Sprintf("%d", coreID),
		DriAgentNamespace:    d.Namespace,
		DriAgentNodeIP:       d.NodeIP,
		DriAgentStorageClass: d.StorageClassName,
		DriAgentKubeConfig:   base64.Encode([]byte(d.KubeConfig)),
		DriCoreHttpAddr:      d.CoreHTTPAddr,
		DriCoreGrpcAddr:      d.CoreGRPCAddr,
		DriAgentK8sUUID:      datas[0],
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

var defaultConfig = `

`

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
