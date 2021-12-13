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
	"github.com/stretchr/testify/assert"
	"github.com/zibuyu28/cmapp/common/base64"
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
				Download            string   `json:"download" validate:"required"`
				CheckSum            string   `json:"check_sum" validate:"required"`
				PackageHandleShells []string `json:"package_handle_shells"`
				StartCommands       []string `json:"start_command" validate:"required"`
			}{},
		}, nil
	})
	defer p.Reset()
	worker := K8sWorker{
		Name:         "mock-k8s-worker",
		Namespace:    "mock-namespace",
		StorageClass: "mock-storageclass",
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


func TestBase64(t *testing.T) {
	bs := "dGhpcyBpcyBhIGV4YW1wbGVhcGlWZXJzaW9uOiB2MQpjbHVzdGVyczoKLSBjbHVzdGVyOgogICAgY2VydGlmaWNhdGUtYXV0aG9yaXR5LWRhdGE6IExTMHRMUzFDUlVkSlRpQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENrMUpTVU12YWtORFFXVmhaMEYzU1VKQlowbENRVVJCVGtKbmEzRm9hMmxIT1hjd1FrRlJjMFpCUkVGV1RWSk5kMFZSV1VSV1VWRkVSWGR3Y21SWFNtd0tZMjAxYkdSSFZucE5RalJZUkZSSmVFMVVTWGhOZWtVd1RWUmpNRTlXYjFoRVZFMTRUVlJKZUUxVVJUQk5WR013VDFadmQwWlVSVlJOUWtWSFFURlZSUXBCZUUxTFlUTldhVnBZU25WYVdGSnNZM3BEUTBGVFNYZEVVVmxLUzI5YVNXaDJZMDVCVVVWQ1FsRkJSR2RuUlZCQlJFTkRRVkZ2UTJkblJVSkJUM1lyQ25CdFVIaEJkMjlWUmpsYWNTOHpWMmRzY2pkMGJEZzRXWFJTYjBWMVJuaExLekJXVVRsWlJVOHlabGR6TUZaYVFXSnhka2xPYUd0V1JWTkxaVFJ1YXpRS1duVnFja2xDSzFoYWVGWmFTME5WT0ZRd1UydFVUMnRSVm1wd2FrazBaRE5oTkhOc2J6WjRNWGhsZGtWblpsVkdibkphVVhoc01reEtiRmhQVGl0eFRRcFdaVTVJT1hoR05EbEROMnhCWjFWQ0syNUtWWGRVVEZCaWQwUk1SRXBXVG5ONmRrcHFTbWd6WTJOakswSjRWVzV4U3pCTWNUa3diakExUjBGaFdHdHRDbXc1Y0VkRU5rWm9NVVpLUW05cU5ISldOWFpQYjBVM2VGZDRXbGhqSzJSVlQwNDNUbGREVjFoVVIwOTFNWGxRZDNoVVVrSnpka2hqYjJka1dHRjRPV29LYkVGR1V6SkZVMHB0VUZKNmQxRlBWSEZWWjJsUVdXNDVTQ3RaZFZSaGJGUldZVFptY21jclJDdE1iVEpCVFM5U2VUaGpUelJETDNwNFUzRmFWelV4YWdwd1F6TlJVa2hhVVVrdmExRTBRbll4ZDNnNFEwRjNSVUZCWVU1YVRVWmpkMFJuV1VSV1VqQlFRVkZJTDBKQlVVUkJaMHRyVFVFNFIwRXhWV1JGZDBWQ0NpOTNVVVpOUVUxQ1FXWTRkMGhSV1VSV1VqQlBRa0paUlVaRlRDOXhNRkZzVmxFeFVYbEJORlZGUWpkWGIyWlVPRTlhTURWTlFsVkhRVEZWWkVWUlVVOEtUVUY1UTBOdGRERlpiVlo1WW0xV01GcFlUWGRFVVZsS1MyOWFTV2gyWTA1QlVVVk1RbEZCUkdkblJVSkJURVpEVFRCM2FVdFdSMDUxTW01QlJEaGFlZ28yTUdScGMwWTRZVlpOV1VSSFYxRnlUMUJxUXpaTFV5OW9jMDFEVlRScWFHTnBUMUZhZVZwaWNWWlVTWHBYWm13MVlXMUtWMDloUTIxblJuSTJPVFJYQ2paaFpIVmlPV2RXTmxOR015ODBjWEZ3U0U1b1N6bEJhM05vYVRkdFZXSTNXV1pDV25VemVXc3lWQ3RQVWtSakwyeG9VMmxCYm5aTGVGWkJSM0I0T1hnS1RGRjNabFZhVFhGbE5WUlBSWFpaU3preVNteElia2wyWVZabU1sSldla3hHSzNwaVQwdGlSRE5QVTI1VlJrUkxkR2RoVUhwdlJFSlNXVWNyTVdoSmRRcE1hblJMUjJWT2JGTkRTSE4xVWxveFYzWnZkRWhqZUV3eVpqbDNTelZPUW1ScFRtbDBlbkZEYWxNelEzUkJlR2hUZDBSWVZqZHJWVXd2ZHlzNWRqWnJDblZPWVZWbFZtdHdRMHh3UWtsSlZYSTJaRGxEVjB0S1JVRjZNVFIwUTJSeGNrMWtWbVpPVDFObVMyeFpjVGRDY0hSTk1XVTJNVXhJUkdkdGRUZFZWMGdLUm1JMFBRb3RMUzB0TFVWT1JDQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENnPT0KICAgIHNlcnZlcjogaHR0cHM6Ly9rdWJlcm5ldGVzLmRvY2tlci5pbnRlcm5hbDo2NDQzCiAgbmFtZTogZG9ja2VyLWRlc2t0b3AKY29udGV4dHM6Ci0gY29udGV4dDoKICAgIGNsdXN0ZXI6IGRvY2tlci1kZXNrdG9wCiAgICB1c2VyOiBkb2NrZXItZGVza3RvcAogIG5hbWU6IGRvY2tlci1kZXNrdG9wCmN1cnJlbnQtY29udGV4dDogZG9ja2VyLWRlc2t0b3AKa2luZDogQ29uZmlnCnByZWZlcmVuY2VzOiB7fQp1c2VyczoKLSBuYW1lOiBkb2NrZXItZGVza3RvcAogIHVzZXI6CiAgICBjbGllbnQtY2VydGlmaWNhdGUtZGF0YTogTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVUlJha05EUVdseFowRjNTVUpCWjBsSlRFY3daVEptU0c5VWRuZDNSRkZaU2t0dldrbG9kbU5PUVZGRlRFSlJRWGRHVkVWVVRVSkZSMEV4VlVVS1FYaE5TMkV6Vm1sYVdFcDFXbGhTYkdONlFXVkdkekI1VFZSRmVVMVVUWGhPUkVVelRrUnNZVVozTUhsTmFrVjVUVlJOZUU1RVJUTk9WRVpoVFVSWmVBcEdla0ZXUW1kT1ZrSkJiMVJFYms0MVl6TlNiR0pVY0hSWldFNHdXbGhLZWsxU2MzZEhVVmxFVmxGUlJFVjRTbXRpTWs1eVdsaEpkRnB0T1hsTVYxSnNDbU15ZERCaU0wRjNaMmRGYVUxQk1FZERVM0ZIVTBsaU0wUlJSVUpCVVZWQlFUUkpRa1IzUVhkblowVkxRVzlKUWtGUlJFaEhkU3QwUTNJdlZWbG9SamNLZGpWNFVtNVJkemxZVFVFNGNYazBhakpFYVd4bGJEWnJSelpSTms1SVZWRjNLekJZUWxFek0zZzNZWFJ2UTBOSE1rZzNkVUo1Ym5OeldHMWlZMDVNTWdwcVNFaHhSek01U1Zwek5YZDNibmR0ZHpSNWNpdHFTakkyVVhkblNqa3llVzVqYTJwRlJqaFJOVlpUVDFaMGFVbHZabk5CY0dwVWRuQnRaV2RsZGpsRkNqUk9UVlEzVTFWQ1pGcFVURlI0VDBrMU0zTjFPVWc0T1U4emFWaFlka0p5YkhOWFVXdHlhMngxY0VscFNYQkhTUzlKVGpBMWVVNXdSemRKVUZWNVpWTUtielYzV2tWdGNFTkZValpoVlUxbVZXWTFXRXB4YUdkd2VtTlVXVkJuWnpKUU9UVmhjbFZySzNsTmVra3lLMEpHYVRkR1JtYzRkQ3R3UTJOeVMxY3hhZ3BSV0RaM2MyUndhR3gxYm1OSE5sRlBiWE40UlVKcGNVeHVkVXhoVGtaQ1NHdHhTMHhpVW1neWJuTkdOVXRGY1dWNE16ZEVTVFpLZW5wc2VWZFZlSFV3Q2tOU1pWbEdlVVZRUVdkTlFrRkJSMnBrVkVKNlRVRTBSMEV4VldSRWQwVkNMM2RSUlVGM1NVWnZSRUZVUW1kT1ZraFRWVVZFUkVGTFFtZG5ja0puUlVZS1FsRmpSRUZxUVUxQ1owNVdTRkpOUWtGbU9FVkJha0ZCVFVJNFIwRXhWV1JKZDFGWlRVSmhRVVpGVEM5eE1GRnNWbEV4VVhsQk5GVkZRamRYYjJaVU9BcFBXakExVFVJd1IwRXhWV1JGVVZGWFRVSlRRMFZ0VW5aWk1uUnNZMmt4YldJelNYUmFSMVo2WVROU2RtTkVRVTVDWjJ0eGFHdHBSemwzTUVKQlVYTkdDa0ZCVDBOQlVVVkJaSEk0VGpKSVYxZHBkVEZYZUhKT04xUk9UR3BrYUdSd1QwdFFiVmxGZWpFNVJYazJkbkJqZUc5TWVuZzVORFY0U0hsT2NYbFNTeThLUm1kNlNGVkpTeXRNVjBSdFVHNW5LMFZDZHpCWGJEQjNiWE5TTlhCNVozSktUVGRGYkRONVNteGhLMnRNZWxBMlZ6Tk9PRFJGZG5GUlpVNHlkWEpRYVFwWE4wYzBURmRyTnpoT2NrZExjR1oyWlRCbVprNHpNVFp1UkVGR1dXZE5hekZMVms1VlV5OTZUVWxtYm1aaFFtbHlXbmg0U2xWUFNFMVRaWEJEU1RFeUNpOXpNVzl1VFd0WlRrNTNWSE15ZEVSb2NXNUVWM2RKVkVsa2MwWTRVRkJ1U0dsV2IyNW9NVzFaVVhSNE9GSjBlbGgwV0djeWMwdHFlRTFPWkhoWUszWUtXVmRCUkdoNFREbENSekZ4Ym1sMVdqUTNTRTVZTWpVNVNFd3hVek5sVTJWUVFqUXZRWEkzTHpVemRUQklaalIzVjJJMVpWcFNSbHBuYkhOcFdYVndRZ3B6UlZJeWJUZzRkRGw1TUU0M2F6QkhVMkppUWtwMVpuTmFUemhWSzJjOVBRb3RMUzB0TFVWT1JDQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENnPT0KICAgIGNsaWVudC1rZXktZGF0YTogTFMwdExTMUNSVWRKVGlCU1UwRWdVRkpKVmtGVVJTQkxSVmt0TFMwdExRcE5TVWxGYjNkSlFrRkJTME5CVVVWQmVIaHlkbkpSY1M4eFIwbFNaVGNyWTFWYU1FMVFWbnBCVUV0emRVazVaelJ3V0hCbGNFSjFhMDlxVWpGRlRWQjBDa1ozVlU0NU9HVXljbUZCWjJoMGFDczNaMk53TjB4R05XMHpSRk01YjNoNE5taDBMMU5IWWs5alRVbzRTbk5QVFhFdmIzbGtkV3ROU1VObVpITndNMG9LU1hoQ1prVlBWbFZxYkdKWmFVdElOMEZMV1RBM05scHViMGh5TDFKUFJGUkZLekJzUVZoWFZYa3dPRlJwVDJRM1RIWlNMMUJVZERSc01UZDNZVFZpUmdwclNrczFTbUp4VTBscFMxSnBVSGxFWkU5amFtRlNkWGxFTVUxdWEzRlBZMGRTU25GUmFFVmxiV3hFU0RGSUsxWjVZVzlaUzJNelJUSkVORWxPYWk5bENsZHhNVXBRYzJwTmVVNTJaMUpaZFhoU1dWQk1abkZSYmt0NWJIUlpNRVlyYzB4SVlWbGFZbkF6UW5WclJIQnlUVkpCV1hGcE5UZHBNbXBTVVZJMVMya0thVEl3V1dSd04wSmxVMmhMYm5Oa0szZDVUMmxqT0RWamJHeE5ZblJCYTFodFFtTm9SSGRKUkVGUlFVSkJiMGxDUVVVMmJHSnBlbVJDVWpoQlUyTlhiUXA2ZW14U2JtOWpVMlozUXpNMFZHSmFZbkUyTW05dVIxTldXa2RGVEZGWFpGUnhXamhKWVU1U1MxQkNNbEZ6VEdjck1uQk5VRUozTld0cU5UZEdNbTAxQ2pKcWJFNHJWRFJDUm1WU2JXWnBTRlowYkdkelRWRnFWbFI1ZGxsak0weGtZbkJxUTBjM2MxSjVkeXRoWTBabGQwY3ZVMHR6Ym1oRVRUaFhjRXd2U0cwS1IwcHdWRmgyTW5ack1FZHVhalpFVWxKMFFYRldZVkUzVUVsM1ZtbDFRMnczVml0cGRUVXZXWGQxWkVwQ1VqZG1WazFqU2pCVFFXZGFXbTR6TUV4T1NRcHlaR2x0TjA1RFpqZ3JUM0JTY1dKSmVVOUhUMUJrZGl0MlQxWkZTMWNyTlhrM2FVaGhXbXAyTlRWb04xcGxjWFZWZEUxdlVUVlNRbEV5WjFsTlpqWndDbGwwZERkeWNFVnlSMFpUVmtOdFdGQnZVU3R3VXl0UE9IQlJOWEZIZVRadWNsRTVjWGs1U1dSS1JXTm5SMnN6Tkd0QllVOTFTWEZOUTI5elIyRkVNa29LY2l0NVZXWm9hME5uV1VWQk0yRnlUbkZXU0Rsc2FFeGliRnBvUVVGQmNqbDBWV2RMUmpGbGVuVXJNQ3N6VjFwa05YbHJUM0UwYUhndlFVSm1WbTFNY3dveFFYRnRUM2RyVEVwVU4xSmxWV0pwU0RBeU1IVndLMmQyVVRONlFWVldSRzAwT0ZOMFdIQkRlbkpZUVN0WVJYUkJaSE4zT0RscWRHdG5VamhYZVVsVUNtUk1aa05tUmxKRVFVRjNaM2hsVW5OQ2EzZG9kV3B6Y0RaVFZIbHNkSFozTm10aGJXZDJUMGRoUnpGM1JtdHNhVlowTTNrNFRrMURaMWxGUVRWbVIwc0taVE5QVkhwNU1uWjNhVzFtTDFwNGVrTTNUMjlDTmxoeFpYZzFZbGRNUW5Zd2MxVlNaMWswY25sM2JIaENVVGRwYjI1cmRGRXZXakZaVjJkU1J6Qm9ad3BSVUU0NFVGWlNaWE5hU21zeVQwOUxiVUpwWTFCRFNqQmFWRnBMU2tVNVUwcHlSRzlxUlhkSFZVNXhWakZNT1RCTFNXY3pWaXRVU0doWVdEZGxXRTFZQ2poQmRqbHphM28wWWpWVGFVOTVObWhxWmxaU1JUSjRMMUJxVmpVMVlWTjZhVkJEWWxOV1ZVTm5XVUl6YVhCb1ZYVnpia1lyUXl0Q1ZXeFZXbU5PVURNS2QyNTZiWEkzTlZoR05DdDVaVWhWYkhBMFpXOTVMMGRyYkVaTVZqRkpNbWh2U1dsaVFqZHhRMEZKYlZwSVRHWjJXamh0VTFveWJWazJkaTlyTjBsVFZRcFVTMnR5VVRFM2NDdFRlWHB0VDNoTWRIWnBSVmRNVm5RM1NYUm1NMnhYWkhWVlptbEpkRXh5U1VvNEx5dFljV3RRYzFKc2JWZG9NWFZvTkRGaGVrWlhDazV1YlhZek1FeGFOa1ZNZEZWbEx6bHdjalZSVkhkTFFtZEVVR0pqVm5VclExRjFhMUpuYzBnd1EydHNOakYzYzFCRGVtUlpha0UxTDBkbVdVRnhRVXNLUTFwVGNVNHpORzEwZGtwd1JsQjFTRlZKYjFCUWVFZHJiM0JqUVdvMVUxZFdUMHRRT0VoemFtaE9URlpJYVVOSVJFWlZWR2REVTBSU1pERkRaR3RvTndwb1ZGb3Jkbkl6YkZSbU1GazBSakY2ZFhWa04yeFFUMjlWTDNCcU1XNVpkMlpvY1VRclZrNVJkR1pYWVRSd2VHUlJkR1ZoYkVkbE0wdzNTakp1Ym1FekNuSkpNbmhCYjBkQ1FVNVhNbnBFTkRGQmEzRTJWWFJ3WTA5U2NYYzNNRGMyZG1sVGRWcDBUbEpUYURCUVlXdzRjekkzY1c5TGNrNXRhRWhZYzNCbFZFSUtPSEJXVXpCQmVtcFNWek5PTWtNeVMyUnRSVGROZUcxd1ZWRXZNWEYzZDFCcWJtTkRNbTAwU2swM05ISm5XVFlyY0daR00ybzJNVWh5WkhKUEt6ZFJhUXBZY2pKM2JVUmtVRnAxUzBZeU5GQjFNMWMzV2xBM1ZWTXJlR2hOZUdSeU5HVm5lU3RPU0U1M1owNHpWR1p0UTBwUFl6UkZDaTB0TFMwdFJVNUVJRkpUUVNCUVVrbFdRVlJGSUV0RldTMHRMUzB0Q2c9PQ=="
	decode, err := base64.Decode(bs)
	assert.Nil(t, err)
	assert.NotNil(t, decode)
}