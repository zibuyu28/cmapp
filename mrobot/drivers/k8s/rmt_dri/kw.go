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
	"github.com/zibuyu28/cmapp/common/base64"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/mrobot/drivers/k8s/kube_driver/base"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/core"
	agfw "github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/worker"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

type K8sWorker struct {
	Name         string
	Namespace    string `validate:"required"`
	StorageClass string `validate:"required"`
	NodeIP       string `validate:"required,ip"`
	KubeConfig   string `validate:"required"`
	MachineID    int    `validate:"required"`
	Domain       string
}

func NewK8sWorker() *K8sWorker {
	w := &K8sWorker{
		NodeIP:       agfw.Flags["NODEIP"].Value,
		KubeConfig:   agfw.Flags["KUBECONFIG"].Value,
		Namespace:    agfw.Flags["NAMESPACE"].Value,
		StorageClass: agfw.Flags["STORAGECLASS"].Value,
		Domain:       agfw.Flags["DOMAIN"].Value,
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
	decode, err := base64.Decode(w.KubeConfig)
	if err != nil {
		panic(err)
	}
	w.KubeConfig = string(decode)
	return w
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
		UID:          uid,
		Image:        fmt.Sprintf("%s:%s", pkg.Image.ImageName, pkg.Image.Tag),
		WorkDir:      pkg.Image.WorkDir,
		Command:      pkg.Image.StartCommands,
		FileMounts:   make(map[string]FileMount),
		Environments: make(map[string]string),
		Ports:        make(map[int]PortInfo),
		FilePremises: make(map[string]FilePremise),
		Tags:         map[string]string{"uuid": uid, "machine_id": fmt.Sprintf("%d", k.MachineID)},
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
		Workspace: &worker0.App_WorkspaceInfo{Workspace: app.UID},
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
			if mount.File == "*" {
				vmes = append(vmes, corev1.VolumeMount{
					Name:      mount.Volume,
					MountPath: mount.MountTo,
				})
			} else {
				vmes = append(vmes, corev1.VolumeMount{
					Name:      mount.Volume,
					MountPath: mount.MountTo,
					SubPath:   mount.File,
				})
			}
		} else {
			// 因为initcontainer 和 container 挂载pvc的路径是一致的
			//vmes = append(vmes, corev1.VolumeMount{
			//	Name:      fmt.Sprintf("%s-pvc", app.UID),
			//	MountPath: mount.MountTo,
			//	SubPath:   fmt.Sprintf("download/%s", mount.File),
			//})
		}
	}

	// 因为initcontainer 和 container 挂载 pvc 的路径是一致的，去除 initcontainer 容器中文件的 filemount
	vmes = append(vmes, corev1.VolumeMount{
		Name: fmt.Sprintf("%s-pvc", app.UID),
		// 将给定的 workspace-> app.UID 映射到 image 的 workdir 下面，这样可以使用相对路径
		MountPath: fmt.Sprintf("%s", app.WorkDir),
		SubPath:   fmt.Sprintf("download"),
	})

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
			commands = append(commands, fmt.Sprintf("wget -O \"/%s/%s\" -c %s", app.UID, premise.Name, premise.AcquireAddr))
			if len(premise.Shell) != 0 {
				commands = append(commands, premise.Shell)
			}
		}
		// get busy box image
		pkg, err := core.PackageInfo(ctx, "busybox", "latest")
		if err != nil {
			return nil, errors.Wrapf(err, "get package info")
		}
		initc = append(initc, corev1.Container{
			Name:    fmt.Sprintf("%s-init", app.UID),
			Image:   fmt.Sprintf("%s:%s", pkg.Image.ImageName, pkg.Image.Tag),
			Command: []string{"/bin/sh", "-c", strings.Join(commands, "\n")},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      fmt.Sprintf("%s-pvc", app.UID),
					MountPath: fmt.Sprintf("/%s", app.UID),
					SubPath:   "download",
				},
			},
			ImagePullPolicy: corev1.PullIfNotPresent,
			WorkingDir:      fmt.Sprintf("/%s", app.UID),
		})
	}

	var vols []corev1.Volume
	vols = append(vols, corev1.Volume{
		Name: fmt.Sprintf("%s-pvc", app.UID),
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: fmt.Sprintf("%s-pvc", app.UID),
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
			Name:      fmt.Sprintf("%s-dep", app.UID),
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
							Name:            fmt.Sprintf("%s", app.UID),
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
							ImagePullPolicy: corev1.PullIfNotPresent,
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
			Name:      fmt.Sprintf("%s-pvc", app.UID),
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


	//var srvp []corev1.ServicePort
	////var igs []v1beta1.IngressRule
	//for _, info := range app.Ports {
	//	s := corev1.ServicePort{
	//		Name:       info.Name,
	//		Port:       int32(info.Port),
	//		TargetPort: intstr.FromInt(info.Port),
	//	}
	//	switch corev1.Protocol(strings.ToUpper(info.Protocol)) {
	//	case corev1.ProtocolTCP:
	//		s.Protocol = corev1.ProtocolTCP
	//	case corev1.ProtocolUDP:
	//		s.Protocol = corev1.ProtocolUDP
	//	case corev1.ProtocolSCTP:
	//		s.Protocol = corev1.ProtocolSCTP
	//	default:
	//		return nil, errors.Errorf("port protocol [%s] not correct", info.Protocol)
	//	}
	//
	//	srvp = append(srvp, s)
	//
	//	//igs = append(igs, v1beta1.IngressRule{
	//	//	Host: info.IngressName,
	//	//	IngressRuleValue: v1beta1.IngressRuleValue{
	//	//		HTTP: &v1beta1.HTTPIngressRuleValue{
	//	//			Paths: []v1beta1.HTTPIngressPath{
	//	//				{
	//	//					Backend: v1beta1.IngressBackend{
	//	//						ServiceName: info.ServiceName,
	//	//						ServicePort: intstr.FromInt(info.Port),
	//	//					},
	//	//				},
	//	//			},
	//	//		},
	//	//	},
	//	//})
	//}
	//
	//if len(srvp) != 0 {
	//
	//}
	//srv := corev1.Service{
	//	TypeMeta: metav1.TypeMeta{
	//		Kind:       "Service",
	//		APIVersion: "v1",
	//	},
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      fmt.Sprintf("%s-service", app.UID),
	//		Namespace: k.Namespace,
	//		Labels:    app.Tags,
	//	},
	//	Spec: corev1.ServiceSpec{
	//		Ports:    srvp,
	//		Selector: app.Tags,
	//		Type:     corev1.ServiceTypeClusterIP,
	//	},
	//}

	marshal, err := yaml.Marshal(dep)
	if err != nil {
		return nil, errors.Wrap(err, "marshal dep")
	}
	fmt.Println(string(marshal))

	marshal, err = yaml.Marshal(pvc)
	if err != nil {
		return nil, errors.Wrap(err, "marshal pvc")
	}
	fmt.Println(string(marshal))

	//marshal, err = yaml.Marshal(srv)
	//if err != nil {
	//	return nil, errors.Wrap(err, "marshal srv")
	//}
	//fmt.Println(string(marshal))

	log.Debug(ctx, "Currently new k8s client")
	cli, err := base.NewClientByConfig(ctx, []byte(k.KubeConfig))
	if err != nil {
		return nil, errors.Wrap(err, "new k8s client")
	}

	// service 在network的时候会创建, 这里需要添加tag
	if len(app.Ports) != 0 {
		service := fmt.Sprintf("%s-service", app.UID)
		getService, err := cli.GetService(service, k.Namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "get service [%s]", service)
		}
		getService.ObjectMeta.Labels = app.Tags
		getService.Spec.Selector = app.Tags
		err = cli.UpdateService(getService)
		if err != nil {
			return nil, errors.Wrapf(err, "udpate service [%s]", service)
		}
	}

	//log.Debug(ctx, "Currently start to create service")
	//err = cli.CreateService(&srv)
	//if err != nil {
	//	return nil, errors.Wrap(err, "apply service")
	//}

	// 不实用ingress
	//if len(k.Domain) != 0 {
	//	igrs := v1beta1.Ingress{
	//		TypeMeta: metav1.TypeMeta{
	//			Kind:       "Ingress",
	//			APIVersion: "extensions/v1beta1",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      fmt.Sprintf("%s-ingress", app.UID),
	//			Namespace: k.Namespace,
	//			Labels:    app.Tags,
	//			Annotations: map[string]string{
	//				"kubernetes.io/ingress.class":             "nginx",
	//				"nginx.ingress.kubernetes.io/use-regex":   "true",
	//				"nginx.ingress.kubernetes.io/enable-cors": "true",
	//			},
	//		},
	//		Spec: v1beta1.IngressSpec{
	//			Rules: igs,
	//		},
	//	}
	//	log.Debug(ctx, "Currently start to create ingress")
	//	err = cli.CreateIngress(&igrs)
	//	if err != nil {
	//		return nil, errors.Wrap(err, "apply ingress")
	//	}
	//
	//	marshal, err = yaml.Marshal(igrs)
	//	if err != nil {
	//		return nil, errors.Wrap(err, "marshal ingress")
	//	}
	//	fmt.Println(string(marshal))
	//}

	log.Debug(ctx, "Currently start to create pvc")
	err = cli.CreatePersistentVolumeClaim(&pvc)
	if err != nil {
		return nil, errors.Wrap(err, "apply pvc")
	}

	log.Debug(ctx, "Currently start to create deployment")
	err = cli.CreateDeployment(&dep)
	if err != nil {
		return nil, errors.Wrap(err, "apply deployment")
	}
	log.Debug(ctx, "Currently app deploy success")
	return &worker0.Empty{}, nil
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
	//ingress := fmt.Sprintf("m%d-%s-%d.%s", k.MachineID, app.UID, network.PortInfo.Port, k.Domain)
	pi := PortInfo{
		Port:        int(network.PortInfo.Port),
		Name:        network.PortInfo.Name,
		Protocol:    worker0.App_Network_PortInf_Protocol_name[int32(network.PortInfo.ProtocolType)],
		ServiceName: service,
		//IngressName: ingress,
	}

	log.Debug(ctx, "Currently new k8s client")
	cli, err := base.NewClientByConfig(ctx, []byte(k.KubeConfig))
	//cli, err := base.NewClientInCluster(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "new k8s client")
	}

	err = svcHandle(ctx, cli, service, k.Namespace, &pi)
	if err != nil {
		return nil, errors.Wrap(err, "svc handle")
	}

	time.Sleep(time.Second*2)

	getService, err := cli.GetService(service, k.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "find serive [%s]", service)
	}

	var nodeport int
	for _, port := range getService.Spec.Ports {
		if port.Port == int32(pi.Port) {
			nodeport = int(port.NodePort)
			break
		}
	}
	if nodeport == 0 {
		return nil, errors.Errorf("find port [%d] target node port is nil", pi.Port)
	}
	pi.NodePort = nodeport

	inRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_IN,
		Router:    fmt.Sprintf("%s:%d", service, network.PortInfo.Port),
	}
	outRoute := &worker0.App_Network_RouteInf{
		RouteType: worker0.App_Network_RouteInf_OUT,
		Router:    fmt.Sprintf("%s:%d", k.NodeIP, nodeport),
	}

	network.RouteInfo = []*worker0.App_Network_RouteInf{inRoute, outRoute}
	app.Ports[int(network.PortInfo.Port)] = pi

	return network, nil
}

func svcFind(ctx context.Context, cli *base.Client, serviceName, namespace string) (*corev1.Service, error) {
	service, err := cli.GetService(serviceName, namespace)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Debugf(ctx, "Currently not found svc [%s]", serviceName)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "get svc [%s]", serviceName)
	}
	return service, nil
}

func svcHandle(ctx context.Context, cli *base.Client, serviceName, namespace string, port *PortInfo) error {
	s := corev1.ServicePort{
		Name:       port.Name,
		Port:       int32(port.Port),
		TargetPort: intstr.FromInt(port.Port),
	}
	switch corev1.Protocol(strings.ToUpper(port.Protocol)) {
	case corev1.ProtocolTCP:
		s.Protocol = corev1.ProtocolTCP
	case corev1.ProtocolUDP:
		s.Protocol = corev1.ProtocolUDP
	case corev1.ProtocolSCTP:
		s.Protocol = corev1.ProtocolSCTP
	default:
		return errors.Errorf("port protocol [%s] not correct", port.Protocol)
	}

	findSvc, err := svcFind(ctx, cli, serviceName, namespace)
	if err != nil {
		return errors.Wrap(err, "find svc")
	}
	if findSvc != nil {
		findSvc.Spec.Ports = append(findSvc.Spec.Ports, s)
		log.Debug(ctx, "Currently start to update service")
		err = cli.UpdateService(findSvc)
		if err != nil {
			return errors.Wrap(err, "update service")
		}
	} else {
		findSvc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceName,
				Namespace: namespace,
				//Labels:    tags,
			},
			Spec: corev1.ServiceSpec{
				Ports:    []corev1.ServicePort{s},
				//Selector: tags,
				Type:     corev1.ServiceTypeNodePort,
			},
		}
		log.Debug(ctx, "Currently start to create service")
		err = cli.CreateService(findSvc)
		if err != nil {
			return errors.Wrap(err, "apply service")
		}
	}
	return nil
}

// FilePremiseEx file premise
func (k *K8sWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App_File, error) {

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

func (k *K8sWorker) HealthEx(ctx context.Context, health *worker0.App_Health) (*worker0.App_Health, error) {
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
