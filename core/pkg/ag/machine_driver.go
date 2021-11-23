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

package ag

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/httputil"
	"github.com/zibuyu28/cmapp/common/log"
)

type function string

const (
	newApp        function = "NewApp"
	startApp      function = "StartApp"
	stopApp       function = "StopApp"
	destroyApp    function = "DestroyApp"
	tagEx         function = "TagEx"
	fileMountEx   function = "FileMountEx"
	envEx         function = "EnvEx"
	networkEx     function = "NetworkEx"
	filePremiseEx function = "FilePremiseEx"
	limitEx       function = "LimitEx"
	healthEx      function = "HealthEx"
	logEx         function = "LogEx"
)

func (f function) String() string {
	return string(f)
}

type PType int

const (
	Binary PType = iota
	Image
)

type App struct {
	UUID  string `json:"uuid"`
	MainP struct {
		CheckSum string   `json:"check_sum"`
		Name     string   `json:"name"`
		Version  string   `json:"version"`
		Type     PType    `json:"type"`
		WorkDir  string   `json:"work_dir"`
		StartCMD []string `json:"start_cmd"`
	}

	FileMounts      []FileMount   `json:"file_mounts"`
	EnvironmentVars []EnvVar      `json:"environment_vars"`
	Networks        []Network     `json:"networks"`
	Workspace       WorkspaceInfo `json:"workspace"`
	FilePremise     []File        `json:"file_premise"`
	LimitInfo       Limit         `json:"limit_info"`
	HealthInfo      Health        `json:"health_info"`
	LogInfo         Log           `json:"log_info"`
	Tags            []Tag         `json:"tags"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Log struct {
	RealTimeFile string `json:"real_time_file"`
	FilePath     string `json:"file_path"`
}

type Method int

const (
	GET Method = iota
	POST
)

type Basic struct {
	MethodType Method `json:"method_type"`
	Path       string `json:"path"`
	Port       int    `json:"port"`
}

type Health struct {
	Liveness Basic `json:"liveness"`
	Readness Basic `json:"readness"`
}

type Limit struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
}

type File struct {
	Name        string `json:"name"`
	AcquireAddr string `json:"acquire_addr"`
	Shell       string `json:"shell"`
}

type WorkspaceInfo struct {
	Workspace string `json:"workspace"`
}

type Protocol int

const (
	TCP Protocol = iota
	UDP
)

type Route int

const (
	IN Route = iota
	OUT
)

type Network struct {
	PortInfo struct {
		Port         int      `json:"port"`
		Name         string   `json:"name"`
		ProtocolType Protocol `json:"protocol_type"`
	}
	RouteInfo []struct {
		RouteType Route  `json:"route_type"`
		Router    string `json:"router"`
	}
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FileMount struct {
	File    string `json:"file"`
	MountTo string `json:"mount_to"`
	Volume  string `json:"volume"`
}

// HMD MachineDriver implement with HTTP
type HMD struct {
	V            APIVersion
	CoreHttpAddr string
}

type NewAppReq struct {
	MachineID int
	Name      string
	Version   string
}

type Req struct {
	AppUUID string      `json:"app_uuid"`
	Fnc     string      `json:"fnc"`
	Param   interface{} `json:"param"`
}

// NewApp new app to core
func (h *HMD) NewApp(nar *NewAppReq) (*App, error) {
	req := Req{
		AppUUID: "init",
		Fnc:     newApp.String(),
		Param:   nar,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return nil, errors.Wrap(err, "send new app request")
	}
	app := App{}
	err = json.Unmarshal(ins, &app)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal app")
	}
	return &app, nil
}

func (h *HMD) StartApp(appUUID string, a *App) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     startApp.String(),
		Param:   a,
	}
	_, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send new app request")
	}
	return nil
}

func (h *HMD) StopApp(appUUID string, a *App) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     stopApp.String(),
		Param:   a,
	}
	_, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send stop app request")
	}
	return nil
}

func (h *HMD) DestroyApp(appUUID string, a *App) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     destroyApp.String(),
		Param:   a,
	}
	_, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send destroy app request")
	}
	return nil
}

func (h *HMD) TagEx(appUUID string, t *Tag) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     tagEx.String(),
		Param:   t,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set tag request")
	}
	err = json.Unmarshal(ins, t)
	if err != nil {
		return errors.Wrap(err, "unmarshal tag")
	}
	return nil
}

func (h *HMD) FileMountEx(appUUID string, mount *FileMount) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     fileMountEx.String(),
		Param:   mount,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set file mount request")
	}
	err = json.Unmarshal(ins, mount)
	if err != nil {
		return errors.Wrap(err, "unmarshal file mount")
	}
	return nil
}

func (h *HMD) EnvEx(appUUID string, ev *EnvVar) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     envEx.String(),
		Param:   ev,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set env request")
	}
	err = json.Unmarshal(ins, ev)
	if err != nil {
		return errors.Wrap(err, "unmarshal env")
	}
	return nil
}

func (h *HMD) NetworkEx(appUUID string, nw *Network) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     networkEx.String(),
		Param:   nw,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set network request")
	}
	err = json.Unmarshal(ins, nw)
	if err != nil {
		return errors.Wrap(err, "unmarshal network")
	}
	return nil
}

func (h *HMD) FilePremiseEx(appUUID string, file *File) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     filePremiseEx.String(),
		Param:   file,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set file premise request")
	}
	err = json.Unmarshal(ins, file)
	if err != nil {
		return errors.Wrap(err, "unmarshal file")
	}
	return nil
}

func (h *HMD) LimitEx(appUUID string, limit *Limit) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     limitEx.String(),
		Param:   limit,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set limit request")
	}
	err = json.Unmarshal(ins, limit)
	if err != nil {
		return errors.Wrap(err, "unmarshal limit")
	}
	return nil
}

func (h *HMD) HealthEx(appUUID string, health *Health) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     healthEx.String(),
		Param:   health,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set health request")
	}
	err = json.Unmarshal(ins, health)
	if err != nil {
		return errors.Wrap(err, "unmarshal health")
	}
	return nil
}

func (h *HMD) LogEx(appUUID string, log *Log) error {
	req := Req{
		AppUUID: appUUID,
		Fnc:     logEx.String(),
		Param:   log,
	}
	ins, err := h.SendPost(req)
	if err != nil {
		return errors.Wrap(err, "send set log request")
	}
	err = json.Unmarshal(ins, log)
	if err != nil {
		return errors.Wrap(err, "unmarshal log")
	}
	return nil
}

type coreReq struct {
	Fnc   string      `json:"fnc"`
	Param interface{} `json:"param"`
}

type coreResp struct {
	Code    CoreCode    `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type CoreCode int

const (
	SUCCESS CoreCode = 200
	FAILED  CoreCode = 400
)

type APIVersion string

const (
	V1 APIVersion = "v1"
)

func (h *HMD) SendPost(req interface{}) ([]byte, error) {
	respb, err := httputil.HTTPDoPost(req, getURL(h.V, h.CoreHttpAddr))
	if err != nil {
		return nil, errors.Wrap(err, "send req to core")
	}
	resp := coreResp{Data: struct{}{}}
	err = json.Unmarshal(respb, &resp)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp [%s]", string(respb))
	}
	if resp.Code != SUCCESS {
		return nil, errors.Errorf("fail to call with req [%v], message [%s]", req, resp.Message)
	}
	datab, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.Wrap(err, "marshal response's data info")
	}
	log.Debugf(context.Background(),"post res [%s]", string(datab))
	return datab, nil
}

func getURL(version APIVersion, addr string) string {
	if len(addr) == 0 {
		log.Debugf(context.Background(), "use default core addr [%s]", coreDefaultHttpAddr)
		addr = coreDefaultHttpAddr
	}
	return fmt.Sprintf("%s/api/%s/md/exec", addr, version)
}
