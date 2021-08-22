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
	"github.com/agiledragon/gomonkey"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/core"
	"google.golang.org/grpc/metadata"
	"testing"
)

var mockApp = App{
	UID:     md5.MD5("whf")[:12],
	Image:   "harbor.hyperchain.cn/platform/fabric-peer:1.4.1",
	WorkDir: "/opt/gopath/src/github.com/hyperledger/fabric/peer",
	Command: []string{"peer", "node", "start"},
	FileMounts: map[string]FileMount{
		"1": {
			File:    "*",
			MountTo: "/var/hyperledger/production",
		},
		"2": {
			File:    "msp",
			MountTo: "/etc/hyperledger/fabric/msp",
		},
		"3": {
			File:    "tls",
			MountTo: "/etc/hyperledger/fabric/tls",
		},
	},
	Environments: map[string]string{
		"CORE_CHAINCODE_BUILDER":             "harbor.hyperchain.cn/platform/fabric-ccend:latest",
		"CORE_CHAINCODE_GOLANG_RUNTIME":      "harbor.hyperchain.cn/platform/fabric-baseos:latest",
		"CORE_VM_ENDPOINT":                   "unix:///host/var/run/docker.sock",
		"FABRIC_LOGGING_SPEC":                "debug",
		"CORE_PEER_TLS_ENABLED":              "true",
		"CORE_OPERATIONS_LISTENADDRESS":      "0.0.0.0:8443",
		"CORE_PEER_GOSSIP_USELEADERELECTION": "true",
		"CORE_PEER_GOSSIP_ORGLEADER":         "false",
		"CORE_PEER_PROFILE_ENABLED":          "true",
		"CORE_PEER_TLS_CERT_FILE":            fmt.Sprintf("%s/tls/signcerts/cert.pem", "/"+md5.MD5("whf")[:12]),
		"CORE_PEER_TLS_KEY_FILE":             fmt.Sprintf("%s/tls/keystore/key.pem", "/"+md5.MD5("whf")[:12]),
		"CORE_PEER_TLS_ROOTCERT_FILE":        fmt.Sprintf("%s/tls/tlscacerts/tls.pem", "/"+md5.MD5("whf")[:12]),
		"CORE_PEER_MSPCONFIGPATH":            fmt.Sprintf("%s/msp", "/"+md5.MD5("whf")[:12]),
		"CORE_PEER_ADDRESSAUTODETECT":        "true",
		"CORE_PEER_CHAINCODELISTENADDRESS":   "0.0.0.0:7052",
		"CORE_PEER_GOSSIP_BOOTSTRAP":         fmt.Sprintf("app-%s-service:7051", md5.MD5("whf")[:12]),
		"CORE_PEER_ID":                       fmt.Sprintf("app-%s-service", md5.MD5("whf")[:12]),
		"CORE_PEER_ADDRESS":                  fmt.Sprintf("app-%s-service:7051", md5.MD5("whf")[:12]),
		"CORE_PEER_GOSSIP_EXTERNALENDPOINT":  fmt.Sprintf("app-%s-service:7051", md5.MD5("whf")[:12]),
		"CORE_PEER_LOCALMSPID":               "Org1MSP",
	},
	Ports: map[int]PortInfo{
		7051: {
			Port:        7051,
			Name:        "endpoint",
			Protocol:    "tcp",
			ServiceName: fmt.Sprintf("app-%s-service", md5.MD5("whf")[:12]),
			IngressName: fmt.Sprintf("m1-%s-7051.test.hyperchain.cn", md5.MD5("whf")[:12]),
		},
		7052: {
			Port:        7052,
			Name:        "chaincode-listen",
			Protocol:    "tcp",
			ServiceName: fmt.Sprintf("app-%s-service", md5.MD5("whf")[:12]),
			IngressName: fmt.Sprintf("m1-%s-7052.test.hyperchain.cn", md5.MD5("whf")[:12]),
		},
		7053: {
			Port:        7053,
			Name:        "event-port",
			Protocol:    "tcp",
			ServiceName: fmt.Sprintf("app-%s-service", md5.MD5("whf")[:12]),
			IngressName: fmt.Sprintf("m1-%s-7053.test.hyperchain.cn", md5.MD5("whf")[:12]),
		},
		8443: {
			Port:        8443,
			Name:        "heath",
			Protocol:    "tcp",
			ServiceName: fmt.Sprintf("app-%s-service", md5.MD5("whf")[:12]),
			IngressName: fmt.Sprintf("m1-%s-8443.test.hyperchain.cn", md5.MD5("whf")[:12]),
		},
	},
	Limit: &Limit{
		CPU:    1000,
		Memory: 1024,
	},
	Log: nil,
	Health: &HealthOption{
		Liveness: &HealthBasic{
			Method: "get",
			Path:   "/healthz",
			Port:   8443,
		},
		Readness: &HealthBasic{
			Method: "get",
			Path:   "/healthz",
			Port:   8443,
		},
	},
	FilePremises: map[string]FilePremise{
		"1": {
			Name:        "msp.tar.gz",
			AcquireAddr: "https://core.hyperchain.cn/msp.tar.gz",
			Shell:       fmt.Sprintf("tar -zxvf %s/msp.tar.gz -C %s --strip-components 1", "/"+md5.MD5("whf")[:12], "/"+md5.MD5("whf")[:12]),
		},
	},
	Tags: map[string]string{
		"chain": "mockchain",
		"app":   "hyperledger",
		"role":  "peer",
		"node":  "bln111",
		"org":   "org1",
	},
}

func TestK8sWorker_StartApp(t *testing.T) {
	p := gomonkey.ApplyFunc(core.PackageInfo, func(ctx context.Context, name, version string) (*core.Package, error) {
		t.Logf("package info name:version [%s:%s]", name, version)
		return &core.Package{
			Name:    "baseos",
			Version: "latest",
			Image: struct {
				ImageName     string   `json:"image_name" validate:"required"`
				Tag           string   `json:"tag" validate:"required"`
				WorkDir       string   `json:"work_dir" validate:"required"`
				StartCommands []string `json:"start_command" validate:"required"`
			}{
				ImageName: "harbor.hyperchain.cn/platform/library/busybox",
				Tag:       "latest",
				WorkDir:   "/",
			},
			Binary: struct {
				Download      string   `json:"download" validate:"required"`
				CheckSum      string   `json:"check_sum" validate:"required"`
				StartCommands []string `json:"start_command" validate:"required"`
			}{},
		}, nil
	})
	defer p.Reset()
	worker := K8sWorker{
		Name:         "mock-k8s-worker",
		Namespace:    "mock-namespace",
		StorageClass: "mock-storageclass",
		Token:        "xxx",
		Certificate:  "xxx",
		ClusterURL:   "xxx",
		MachineID:    1,
		Domain:       "test.hyperchain.cn",
	}
	background := context.Background()
	outgoingContext := metadata.NewIncomingContext(background, metadata.New(map[string]string{
		"MA_UUID": md5.MD5("whf")[:12],
	}))
	err := repo.new(outgoingContext, &mockApp)
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = worker.StartApp(outgoingContext, nil)
	if err != nil {
		t.Log(err.Error())
	}
}
