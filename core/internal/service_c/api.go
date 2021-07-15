package service_c

import (
	"context"
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
	"google.golang.org/grpc"
)

type MachineWorker struct {

}

func (m MachineWorker) GetWorkspace(ctx context.Context, in *worker.Empty, opts ...grpc.CallOption) (*worker.WorkspaceInfo, error) {
	panic("implement me")
}

func (m MachineWorker) DestroyWorkspace(ctx context.Context, in *worker.WorkspaceInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) DownloadToPath(ctx context.Context, in *worker.DownloadInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) Upload(ctx context.Context, in *worker.UploadInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) Compress(ctx context.Context, in *worker.CompressInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) Decompress(ctx context.Context, in *worker.DeCompressInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) Copy(ctx context.Context, in *worker.CopyInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) UpdateFileContent(ctx context.Context, in *worker.UpdateFileContentInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) DeleteFile(ctx context.Context, in *worker.DeleteFileInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) CreateFile(ctx context.Context, in *worker.CreateFileInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) CreateDir(ctx context.Context, in *worker.CreateDirInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) RemoveDir(ctx context.Context, in *worker.RemoveDirInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) FetchFileContent(ctx context.Context, in *worker.FetchFileContentInfo, opts ...grpc.CallOption) (worker.Worker_FetchFileContentClient, error) {
	panic("implement me")
}

func (m MachineWorker) CheckTargetPortUseful(ctx context.Context, in *worker.CheckTargetPortInfo, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) SetupApp(ctx context.Context, in *worker.SetupAppInfo, opts ...grpc.CallOption) (*worker.App, error) {
	panic("implement me")
}

func (m MachineWorker) Done(ctx context.Context, in *worker.Empty, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) ShutdownApp(ctx context.Context, in *worker.App, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) AppHealth(ctx context.Context, in *worker.App, opts ...grpc.CallOption) (*worker.Empty, error) {
	panic("implement me")
}

func (m MachineWorker) TargetPortIntranetRoute(ctx context.Context, in *worker.TargetPortIntranetInfo, opts ...grpc.CallOption) (*worker.PortIntranetInfo, error) {
	panic("implement me")
}

func (m MachineWorker) TargetPortExternalRoute(ctx context.Context, in *worker.TargetPortExternalInfo, opts ...grpc.CallOption) (*worker.PortExternalInfo, error) {
	panic("implement me")
}

