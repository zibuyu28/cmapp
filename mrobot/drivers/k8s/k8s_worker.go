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
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
	"google.golang.org/grpc/metadata"
)

type Worker struct {
}

func (k *Worker) GetWorkspace(ctx context.Context, empty *worker.Empty) (*worker.WorkspaceInfo, error) {
	_, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("fail to get metadata from context")
	}


	panic("implement me")
}

func (k *Worker) DestroyWorkspace(ctx context.Context, info *worker.WorkspaceInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) DownloadToPath(ctx context.Context, info *worker.DownloadInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) Upload(ctx context.Context, info *worker.UploadInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) Compress(ctx context.Context, info *worker.CompressInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) Decompress(ctx context.Context, info *worker.DeCompressInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) Copy(ctx context.Context, info *worker.CopyInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) UpdateFileContent(ctx context.Context, info *worker.UpdateFileContentInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) DeleteFile(ctx context.Context, info *worker.DeleteFileInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) CreateFile(ctx context.Context, info *worker.CreateFileInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) CreateDir(ctx context.Context, info *worker.CreateDirInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) RemoveDir(ctx context.Context, info *worker.RemoveDirInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) FetchFileContent(info *worker.FetchFileContentInfo, server worker.Worker_FetchFileContentServer) error {
	panic("implement me")
}

func (k *Worker) CheckTargetPortUseful(ctx context.Context, info *worker.CheckTargetPortInfo) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) SetupApp(ctx context.Context, info *worker.SetupAppInfo) (*worker.App, error) {
	panic("implement me")
}

func (k *Worker) Done(ctx context.Context, empty *worker.Empty) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) ShutdownApp(ctx context.Context, app *worker.App) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) AppHealth(ctx context.Context, app *worker.App) (*worker.Empty, error) {
	panic("implement me")
}

func (k *Worker) TargetPortIntranetRoute(ctx context.Context, info *worker.TargetPortIntranetInfo) (*worker.PortIntranetInfo, error) {
	panic("implement me")
}

func (k *Worker) TargetPortExternalRoute(ctx context.Context, info *worker.TargetPortExternalInfo) (*worker.PortExternalInfo, error) {
	panic("implement me")
}