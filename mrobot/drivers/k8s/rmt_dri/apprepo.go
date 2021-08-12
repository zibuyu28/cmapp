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
	"google.golang.org/grpc/metadata"
	"sync"
)

type workRepository struct {
	rep sync.Map
}

var repo = workRepository{rep: sync.Map{}}

func (w *workRepository) load(ctx context.Context) (*App, error) {
	uid, err := guid(ctx)
	if err != nil {
		return nil, errors.New("load uid from context")
	}
	wsp, ok := w.rep.Load(uid)
	if !ok {
		return nil, errors.Errorf("fail to get workspace from rep by uid [%s]", uid)
	}
	return wsp.(*App), nil
}

func (w *workRepository) new(ctx context.Context, wsp *App) error {
	uid, err := guid(ctx)
	if err != nil {
		return errors.New("load uid from context")
	}
	wsp.UID = uid
	w.rep.Store(uid, wsp)
	return nil
}

func guid(ctx context.Context) (uid string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err = errors.New("fail to get metadata from context")
		return
	}
	uuid := md.Get("MA_UUID")
	if len(uuid) == 0 {
		err = errors.New("fail to get uuid from metadata")
		return
	}
	uid = uuid[0]
	return
}

type App struct {
	UID          string
	Image        string
	WorkDir      string
	Command      []string
	FileMounts   map[string]FileMount
	Volumes      map[string]Volume
	Environments map[string]string
	Ports        map[int]PortInfo
	Limit        *Limit
	Log          *Log
}

type Log struct {
	RealTimeFile    string
	CompressLogPath string
}

type Limit struct {
	// unit 'm'
	CPU int
	// unit 'MB'
	Memory int
}

type FileMount struct {
	File    string
	MountTo string
	Volume  string
}

type Volume struct {
	Name  string
	Type  string
	Param map[string]string
}

type PortInfo struct {
	Port        int
	Name        string
	Protocol    string
	ServiceName string
	IngressName string
}

var initContainer = `initContainers:
  - name: busybox
    image: {{.BusyBoxImageName}}
    imagePullPolicy: IfNotPresent
    command:
	  - /bin/sh
	  - "-c"
	  - |
	    wget -O "/tmp/msp.tar.gz"  -c {{.DownloadAddr}}/{{.ChainName}}{{.ChainID}}-bln{{.NodeID}}-msp.tar.gz
	    tar -zxvf /tmp/msp.tar.gz -C /work-dir --strip-components 1
    volumeMounts:
	  - name: peerdata
	    mountPath: "/work-dir"
	    subPath: peer{{.NodeID}}.crypt`

var defaultDeployment = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: {{.Namespace}}
  name: 
  labels:{{range $k, $v := .Labels}}
    {{$k}}: "{{$v}}"{{end}}
spec:
  selector:
    matchLabels:{{range $k, $v := .Labels}}
      {{$k}}: "{{$v}}"{{end}}
  replicas: 1
  minReadySeconds: 10
  strategy:
    type: Recreate
  template:
    metadata:
      labels:{{range $k, $v := .Labels}}
        {{$k}}: "{{$v}}"{{end}}
    spec:
      {{.InitContainer}}
      containers:
        - name: {{.Name}}
          image: {{.PeerImageName}}
          imagePullPolicy: Always
          resources:
            requests:
              cpu: "100m"
              memory: "100Mi"
            limits:
              cpu: "{{.CPULimit}}m"
              memory: "{{.MemLimit}}Gi"
          livenessProbe:
            httpGet:
              path: /healthz # 根据实际服务填写
              port: 8443     # 根据实际服务填写
            initialDelaySeconds: 300
            periodSeconds: 15
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          readinessProbe:
            httpGet:
              path: /healthz  # 9根据实际服务填写
              port: 8443      # 根据实际服务填写
            initialDelaySeconds: 3
            periodSeconds: 1
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          env:{{range $k, $v := .Envs}}
            {{$k}}: "{{$v}}"{{end}}
          workingDir: {{.WorkDir}}
          ports:{{range $p := .Ports}}
            - containerPort: {{$p}}{{end}}
          command: ["{{.Command}}"]
          volumeMounts:
            - mountPath: /var/hyperledger/production
              name: peerdata
              subPath: peer{{.NodeID}}
            - mountPath: /etc/hyperledger/fabric/msp
              name: peerdata
              subPath: peer{{.NodeID}}.crypt/msp
            - mountPath: /etc/hyperledger/fabric/tls
              name: peerdata
              subPath: peer{{.NodeID}}.crypt/tls
            - mountPath: /host/var/run
              name: run
      volumes:
        - name: peerdata
          persistentVolumeClaim:
            claimName: {{.ChainName}}{{.ChainID}}-bln{{.NodeID}}-m{{.MachineID}}-pvc
        - name: run
          hostPath:
            path: /var/run


`
