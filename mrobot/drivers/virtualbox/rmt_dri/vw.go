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
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
)

type VirtualboxWorker struct {

}

func (v *VirtualboxWorker) NewApp(ctx context.Context, req *worker0.NewAppReq) (*worker0.App, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) StartApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) StopApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) DestroyApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) TagEx(ctx context.Context, tag *worker0.App_Tag) (*worker0.App_Tag, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) FileMountEx(ctx context.Context, mount *worker0.App_FileMount) (*worker0.App_FileMount, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) EnvEx(ctx context.Context, envVar *worker0.App_EnvVar) (*worker0.App_EnvVar, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) NetworkEx(ctx context.Context, network *worker0.App_Network) (*worker0.App_Network, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) LimitEx(ctx context.Context, limit *worker0.App_Limit) (*worker0.App_Limit, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) HealthEx(ctx context.Context, health *worker0.App_Health) (*worker0.App, error) {
	panic("implement me")
}

func (v *VirtualboxWorker) LogEx(ctx context.Context, log *worker0.App_Log) (*worker0.App_Log, error) {
	panic("implement me")
}
