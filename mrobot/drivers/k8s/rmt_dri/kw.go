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
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/core"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type K8sWorker struct {
	Name         string
	Namespace    string
	StorageClass string
	Token        string
	Cert         string
	URL          string
	MachineID    int
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
		Tags:    map[string]string{"uuid": uid, "machine_id": fmt.Sprintf("%d", k.MachineID)},
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

// StartApp start app
func (k *K8sWorker) StartApp(ctx context.Context, _ *worker0.App) (*worker0.Empty, error) {
	log.Debug(ctx, "Currently to start app")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}
	// 每个部分进行template之前的一些检查
	var rep = int32(1)

	// ports
	var ports []corev1.ContainerPort
	for port, info := range app.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          info.Name,
			ContainerPort: int32(port),
		})
	}
	// envs
	var envs []corev1.EnvVar
	for key, val := range app.Environments {
		envs = append(envs, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}

	// volume
	var vmes []corev1.VolumeMount
	for _, mount := range app.FileMounts {
		if len(mount.Volume) != 0 {
			if _, ok := app.Volumes[mount.Volume]; !ok {
				return nil, errors.Errorf("fail to get [%s] from app exist volume list", mount.Volume)
			}
			vmes = append(vmes, corev1.VolumeMount{
				Name:      mount.Volume,
				MountPath: mount.MountTo,
			})
		} else {
			vmes = append(vmes, corev1.VolumeMount{
				Name:      fmt.Sprintf("app-%s-pvc", app.UID),
				MountPath: mount.MountTo,
				SubPath:   fmt.Sprintf("download/%s", mount.File),
			})
		}
	}

	// health
	var readness *corev1.Probe
	var liveness *corev1.Probe

	if app.Health != nil {
		if app.Health.Readness != nil {
			readness = &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: app.Health.Readness.Path,
						Port: intstr.FromInt(app.Health.Readness.Port),
					},
				},
				InitialDelaySeconds: 3,
				TimeoutSeconds:      5,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			}
		}
		if app.Health.Liveness != nil {
			liveness = &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: app.Health.Liveness.Path,
						Port: intstr.FromInt(app.Health.Liveness.Port),
					},
				},
				InitialDelaySeconds: 20,
				TimeoutSeconds:      5,
				PeriodSeconds:       1,
				SuccessThreshold:    1,
				FailureThreshold:    5,
			}
		}
	}

	// resources
	var resourcereq corev1.ResourceRequirements
	if app.Limit != nil {
		resourcereq = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", app.Limit.CPU)),
				corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", app.Limit.Memory)),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		}
	}

	var initc []corev1.Container
	if len(app.FilePremises) != 0 {
		var commands []string
		for _, premise := range app.FilePremises {
			commands = append(commands, fmt.Sprintf("wget -O \"/work-dir/%s/%s\" -c %s", premise.Target, premise.Name, premise.AcquireAddr))
			if len(premise.Shell) != 0 {
				commands = append(commands, fmt.Sprintf("%s", premise.Shell))
			}
		}
		// get busy box image
		pkg, err := core.PackageInfo(ctx, "busybox", "latest")
		if err != nil {
			return nil, errors.Wrapf(err, "get package info")
		}
		initc = append(initc, corev1.Container{
			Name:    fmt.Sprintf("app-%s-init", app.UID),
			Image:   fmt.Sprintf("%s:%s", pkg.Image.ImageName, pkg.Image.Tag),
			Command: []string{"/bin/sh", "\"-c\"", strings.Join(commands, "\n")},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      fmt.Sprintf("app-%s-pvc", app.UID),
					MountPath: "/work-dir",
					SubPath:   "download",
				},
			},
			ImagePullPolicy: "IfNotPresent",
		})
	}

	var vols []corev1.Volume
	vols = append(vols, corev1.Volume{
		Name: fmt.Sprintf("app-%s-pvc", app.UID),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf("app-%s-pvc", app.UID),
			},
		},
	})
	vols = append(vols, corev1.Volume{
		Name: "run",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/var/run/",
			},
		},
	})

	dep := v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("app-%s-dep", app.UID),
			Namespace: k.Namespace,
			Labels:    app.Tags,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &rep,
			Selector: &metav1.LabelSelector{MatchLabels: app.Tags},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: app.Tags,
				},
				Spec: corev1.PodSpec{
					Volumes:        vols,
					InitContainers: initc,
					Containers: []corev1.Container{
						{
							Name:            fmt.Sprintf("app-%s", app.UID),
							Image:           app.Image,
							Command:         app.Command,
							Args:            nil,
							WorkingDir:      app.WorkDir,
							Ports:           ports,
							Env:             envs,
							VolumeMounts:    vmes,
							LivenessProbe:   liveness,
							ReadinessProbe:  readness,
							Resources:       resourcereq,
							ImagePullPolicy: "Always",
						},
					},
				},
			},
			Strategy:        v1.DeploymentStrategy{Type: v1.RecreateDeploymentStrategyType},
			MinReadySeconds: 10,
		},
	}

	pvc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("app-%s-pvc", app.UID),
			Namespace: k.Namespace,
			Labels:    app.Tags,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			StorageClassName: &(k.StorageClass),
		},
	}

	var srvp []corev1.ServicePort
	for _, info := range app.Ports {
		s := corev1.ServicePort{
			Name:       info.Name,
			Port:       int32(info.Port),
			TargetPort: intstr.FromInt(info.Port),
		}
		switch corev1.Protocol(strings.ToUpper(info.Protocol)) {
		case corev1.ProtocolTCP:
			s.Protocol = corev1.ProtocolTCP
		case corev1.ProtocolUDP:
			s.Protocol = corev1.ProtocolUDP
		case corev1.ProtocolSCTP:
			s.Protocol = corev1.ProtocolSCTP
		default:
			return nil, errors.Errorf("port protocol [%s] not correct", info.Protocol)
		}

		srvp = append(srvp, s)
	}
	srv := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("app-%s-service", app.UID),
			Namespace: k.Namespace,
			Labels:    app.Tags,
		},
		Spec: corev1.ServiceSpec{
			Ports:    srvp,
			Selector: app.Tags,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	marshal, err := yaml.Marshal(dep)
	if err != nil {
		return nil, errors.Wrap(err, "marshal dep")
	}
	fmt.Println(marshal)

	marshal, err = yaml.Marshal(pvc)
	if err != nil {
		return nil, errors.Wrap(err, "marshal pvc")
	}
	fmt.Println(marshal)

	marshal, err = yaml.Marshal(srv)
	if err != nil {
		return nil, errors.Wrap(err, "marshal srv")
	}
	fmt.Println(marshal)
	panic("implement me")
}

func (k *K8sWorker) StopApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	// 将对应的app副本数量减为0
	panic("implement me")
}

func (k *K8sWorker) DestroyApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	// 将对应app的所有资源删除
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

func (k *K8sWorker) TagEx(ctx context.Context, tag *worker0.App_Tag) (*worker0.App_Tag, error) {
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

// FilePremiseEx
func (k *K8sWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App, error) {

	log.Debug(ctx, "Currently start to execute set file premise")
	app, err := repo.load(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to load app from repo")
	}

	if len(file.Name) == 0 || len(file.AcquireAddr) == 0 || len(file.Target) == 0 {
		return nil, errors.Errorf("file got empty name [%s] or acquire addr [%s] or target [%s]", file.Name, file.AcquireAddr, file.Target)
	}
	key := md5.MD5(fmt.Sprintf("%s:%s:%s", file.Name, file.AcquireAddr, file.Target))

	if e, ok := app.FilePremises[key]; ok {
		return nil, errors.Errorf("file premise exist [%#+v]", e)
	}
	premise := FilePremise{
		Name:        file.Name,
		AcquireAddr: file.AcquireAddr,
		Target:      file.Target,
		Shell:       file.Shell,
	}
	app.FilePremises[key] = premise
	return nil, nil
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
	return nil, nil
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
