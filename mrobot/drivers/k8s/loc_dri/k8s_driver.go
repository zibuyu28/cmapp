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
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/tmp"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s/kube_driver/base"
	"github.com/zibuyu28/cmapp/mrobot/pkg"
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"google.golang.org/grpc/metadata"
	v1 "k8s.io/api/apps/v1"
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

type DriverK8s struct {
	pkg.BaseDriver
	KubeConfig string

	Token       string `validate:"required"`
	Certificate string `validate:"required"`
	ClusterURL  string `validate:"required"`
	Namespace   string `validate:"required"`

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
			Name:   "Token",
			Usage:  "kubernetes cluster token for connection",
			EnvVar: "KUBE_TOKEN",
			Value:  nil,
		},
		{
			Name:   "Certificate",
			Usage:  "kubernetes cluster certificate for connection",
			EnvVar: "KUBE_CERTIFICATE",
			Value:  nil,
		},
		{
			Name:   "Url",
			Usage:  "kubernetes cluster url for connection",
			EnvVar: "KUBE_URL",
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
	m := convertFlags(flags)
	d.CoreAddr = m["CoreAddr"]
	d.ImageRepository.Repository = m["Repository"]
	d.ImageRepository.StorePath = m["StorePath"]

	d.KubeConfig = m["KubeConfig"]
	d.Token = m["Token"]
	d.Certificate = m["Certificate"]
	d.ClusterURL = m["Url"]
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

func convertFlags(flags *driver.Flags) map[string]string {
	var fls = make(map[string]string)
	for _, flag := range flags.Flags {
		if len(flag.Value) != 0 {
			fls[flag.Name] = flag.Value[0]
			continue
		}
		envVal := os.Getenv(flag.EnvVar)
		if len(envVal) != 0 {
			fls[flag.Name] = envVal
		}
	}
	return fls
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

func (d *DriverK8s) CreateExec(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to create machine")
	return nil, nil
}

func (d *DriverK8s) InstallMRobot(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	log.Debug(ctx, "Currently k8s machine plugin start to install mrobot")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to parse metadata info from context")
	}

	datas := md.Get("CoreID")
	if len(datas) != 1 {
		return nil, errors.New("fail to find core id from metadata")
	}
	coreID, err := strconv.Atoi(datas[0])
	if err != nil {
		return nil, errors.Wrapf(err, "parse core id [%s] to int", datas[0])
	}

	c, err := base.NewClient(ctx, defaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new kubernetes client")
	}

	repository := d.ImageRepository
	image := fmt.Sprintf("%s/%s/mrobot:%s", repository.Repository, repository.StorePath, d.DriverVersion)
	log.Debugf(ctx, "Currently image info [%v]", image)

	tempdata := struct {
		Image     string
		CoreAddr  string
		MachineID int
	}{
		Image:     image,
		CoreAddr:  d.CoreAddr,
		MachineID: coreID,
		// TODO: 将驱动中的一些参数也传到 ag 里面
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
	return nil, nil
}

func (d *DriverK8s) MRoHealthCheck(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	// TODO: how to check mrobot health ?
	log.Info(ctx, "Currently k8s machine plugin start to check robot health")
	return nil, nil
}

func (d *DriverK8s) Exit(ctx context.Context, empty *driver.Empty) (*driver.Empty, error) {
	log.Info(ctx, "Currently k8s machine plugin exit")
	return nil, nil
}

var defaultNamespace = ""
var defaultConfig = `

`

var mRobotDep = `---

`
