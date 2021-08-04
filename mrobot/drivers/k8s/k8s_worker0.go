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

package k8s

import (
	"context"
	"github.com/zibuyu28/cmapp/plugin/proto/worker0"
)

func (k *Worker) NewApp(ctx context.Context, req *worker0.NewAppReq) (*worker0.App, error) {
	panic("implement me")
}

func (k *Worker) StartApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *Worker) StopApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *Worker) DestroyApp(ctx context.Context, app *worker0.App) (*worker0.Empty, error) {
	panic("implement me")
}

func (k *Worker) FileMountEx(ctx context.Context, mount *worker0.App_FileMount) (*worker0.App, error) {
	panic("implement me")
}

func (k *Worker) StorageMountEx(ctx context.Context, mount *worker0.App_StorageMount) (*worker0.App, error) {
	panic("implement me")
}

func (k *Worker) SetEnvEx(ctx context.Context, envVar *worker0.App_EnvVar) (*worker0.App, error) {
	panic("implement me")
}

func (k *Worker) NetworkEx(ctx context.Context, app_network *worker0.App_Network) (*worker0.App, error) {
	panic("implement me")
}

func (k *Worker) FilePremiseEx(ctx context.Context, file *worker0.App_File) (*worker0.App, error) {
	panic("implement me")
}

