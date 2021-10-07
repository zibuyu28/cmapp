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
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/pkg/ag/base"
)

type NewAppReq struct {
	Name    string
	Version string
}

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
}

// NewApp new app to core
func (m *HMD) NewApp(nar *NewAppReq) (*App, error) {
	ins, err := base.Send(newApp.String(), nar)
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

func (m *HMD) StartApp(a *App) error {
	_, err := base.Send(startApp.String(), a)
	if err != nil {
		return errors.Wrap(err, "send new app request")
	}
	return nil
}

func (m *HMD) StopApp(a *App) error {
	_, err := base.Send(stopApp.String(), a)
	if err != nil {
		return errors.Wrap(err, "send stop app request")
	}
	return nil
}

func (m *HMD) DestroyApp(a *App) error {
	_, err := base.Send(destroyApp.String(), a)
	if err != nil {
		return errors.Wrap(err, "send destroy app request")
	}
	return nil
}

func (m *HMD) TagEx(t *Tag) error {
	ins, err := base.Send(tagEx.String(), t)
	if err != nil {
		return errors.Wrap(err, "send set tag request")
	}
	err = json.Unmarshal(ins, t)
	if err != nil {
		return errors.Wrap(err, "unmarshal tag")
	}
	return nil
}

func (m *HMD) FileMountEx(mount *FileMount) error {
	ins, err := base.Send(fileMountEx.String(), mount)
	if err != nil {
		return errors.Wrap(err, "send set file mount request")
	}
	err = json.Unmarshal(ins, mount)
	if err != nil {
		return errors.Wrap(err, "unmarshal file mount")
	}
	return nil
}

func (m *HMD) EnvEx(ev *EnvVar) error {
	ins, err := base.Send(envEx.String(), ev)
	if err != nil {
		return errors.Wrap(err, "send set env request")
	}
	err = json.Unmarshal(ins, ev)
	if err != nil {
		return errors.Wrap(err, "unmarshal env")
	}
	return nil
}

func (m *HMD) NetworkEx(nw *Network) error {
	ins, err := base.Send(networkEx.String(), nw)
	if err != nil {
		return errors.Wrap(err, "send set network request")
	}
	err = json.Unmarshal(ins, nw)
	if err != nil {
		return errors.Wrap(err, "unmarshal network")
	}
	return nil
}

func (m *HMD) FilePremiseEx(file *File) error {
	ins, err := base.Send(filePremiseEx.String(), file)
	if err != nil {
		return errors.Wrap(err, "send set file premise request")
	}
	err = json.Unmarshal(ins, file)
	if err != nil {
		return errors.Wrap(err, "unmarshal file")
	}
	return nil
}

func (m *HMD) LimitEx(limit *Limit) error {
	ins, err := base.Send(limitEx.String(), limit)
	if err != nil {
		return errors.Wrap(err, "send set limit request")
	}
	err = json.Unmarshal(ins, limit)
	if err != nil {
		return errors.Wrap(err, "unmarshal limit")
	}
	return nil
}

func (m *HMD) HealthEx(health *Health) error {
	ins, err := base.Send(healthEx.String(), health)
	if err != nil {
		return errors.Wrap(err, "send set health request")
	}
	err = json.Unmarshal(ins, health)
	if err != nil {
		return errors.Wrap(err, "unmarshal health")
	}
	return nil
}

func (m *HMD) LogEx(log *Log) error {
	ins, err := base.Send(logEx.String(), log)
	if err != nil {
		return errors.Wrap(err, "send set log request")
	}
	err = json.Unmarshal(ins, log)
	if err != nil {
		return errors.Wrap(err, "unmarshal log")
	}
	return nil
}
