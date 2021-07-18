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
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
)

type K8SWorker struct {

}

func (k K8SWorker) GetWorkspace(ctx context.Context, empty *worker.Empty) (*worker.WorkspaceInfo, error) {
	panic("implement me")
}

func (k K8SWorker) DestroyWorkspace(ctx context.Context, info *worker.WorkspaceInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) DownloadToPath(ctx context.Context, info *worker.DownloadInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) Upload(ctx context.Context, info *worker.UploadInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) Compress(ctx context.Context, info *worker.CompressInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) Decompress(ctx context.Context, info *worker.DeCompressInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) Copy(ctx context.Context, info *worker.CopyInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) UpdateFileContent(ctx context.Context, info *worker.UpdateFileContentInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) DeleteFile(ctx context.Context, info *worker.DeleteFileInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) CreateFile(ctx context.Context, info *worker.CreateFileInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) CreateDir(ctx context.Context, info *worker.CreateDirInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) RemoveDir(ctx context.Context, info *worker.RemoveDirInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) FetchFileContent(info *worker.FetchFileContentInfo, server worker.Worker_FetchFileContentServer) error {
	panic("implement me")
}

func (k K8SWorker) CheckTargetPortUseful(ctx context.Context, info *worker.CheckTargetPortInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) SetupApp(ctx context.Context, info *worker.SetupAppInfo) (*worker.App, error) {
	panic("implement me")
}

func (k K8SWorker) Done(ctx context.Context, empty *worker.Empty) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) ShutdownApp(ctx context.Context, app *worker.App) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) AppHealth(ctx context.Context, app *worker.App) (*worker.Empty, error) {
	panic("implement me")
}

func (k K8SWorker) TargetPortIntranetRoute(ctx context.Context, info *worker.TargetPortIntranetInfo) (*worker.PortIntranetInfo, error) {
	panic("implement me")
}

func (k K8SWorker) TargetPortExternalRoute(ctx context.Context, info *worker.TargetPortExternalInfo) (*worker.PortExternalInfo, error) {
	panic("implement me")
}

