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
	"fmt"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	agfw "github.com/zibuyu28/cmapp/mrobot/pkg/agentfw/worker"
	"github.com/zibuyu28/cmapp/plugin/proto/worker"
	"strings"
	"sync"
)

type Worker struct {
	wrp *workRepository

	Token       string `validate:"required"`
	Certificate string `validate:"required"`
	ClusterURL  string `validate:"required"`
	Namespace   string `validate:"required"`

	StorageClassName string `validate:"required"`
}

func NewWorker() *Worker {
	w := &Worker{
		wrp:              &workRepository{rep: sync.Map{}},
		Token:            agfw.Flags["Token"].Value,
		Certificate:      agfw.Flags["Certificate"].Value,
		ClusterURL:       agfw.Flags["ClusterURL"].Value,
		Namespace:        agfw.Flags["Namespace"].Value,
		StorageClassName: agfw.Flags["StorageClassName"].Value,
	}
	validate := v.New()
	err := validate.Struct(*w)
	if err != nil {
		panic(err)
	}
	return w
}

func (k *Worker) GetWorkspace(ctx context.Context, empty *worker.Empty) (*worker.WorkspaceInfo, error) {
	// 检查 phase
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		// TODO: check exist
		wsp := &work{
			init:       &initcmd{},
			deployment: "",
			net:        map[int]network{},
		}
		err = k.wrp.new(ctx, wsp)
		if err != nil {
			return nil, errors.Wrap(err, "new workspace")
		}
		return &worker.WorkspaceInfo{Workspace: wsp.UID}, nil
	case SetupPhase, RunningPhase:
		wsp, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "load exist workspace")
		}
		return &worker.WorkspaceInfo{Workspace: wsp.UID}, nil
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) DestroyWorkspace(ctx context.Context, info *worker.WorkspaceInfo) (*worker.Empty, error) {
	//md, ok := metadata.FromIncomingContext(ctx)
	//if !ok {
	//	return nil, errors.New("fail to get metadata from context")
	//}
	//info.Workspace

	// 删除所有资源，包括deployment，service，pvc， ingress 相关部分

	panic("implement me")
}

func (k *Worker) DownloadToPath(ctx context.Context, info *worker.DownloadInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		download := fmt.Sprintf(`wget -O "%s" -c %s`, info.TargetPath, info.DownloadLink)
		wk, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get work from worker repository")
		}
		wk.init.append(download)
		return nil, nil
	case RunningPhase, SetupPhase:
		return nil, nil
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) Upload(ctx context.Context, info *worker.UploadInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		upload := fmt.Sprintf(`wget %s --post-file=%s`, info.TargetLink, info.SourceFile)
		wk, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get work from worker repository")
		}
		wk.init.append(upload)
		return nil, nil
	case RunningPhase, SetupPhase:
		return nil, nil
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) Compress(ctx context.Context, info *worker.CompressInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		log.Infof(ctx, "Currently Compress got dir path [%s]", info.DirPath)
		if len(info.DirPath) == 0 {
			return nil, errors.New("fail to get dir path, got empty path")
		}

		index := strings.LastIndex(info.DirPath, "/")

		dir := info.DirPath[index+1:]
		log.Infof(ctx, "Currently get dir [%s]", dir)

		if len(dir) == 0 {
			return nil, errors.New("fail to parse dir, got empty dir")
		}
		fatherPath := info.DirPath[:index]

		compress := fmt.Sprintf("tar -zcvf %s/%s.tar.gz %s", fatherPath, dir, info.DirPath)
		wk, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get work from worker repository")
		}
		wk.init.append(compress)
		return nil, nil
	case SetupPhase, RunningPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) Decompress(ctx context.Context, info *worker.DeCompressInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		log.Infof(ctx, "Currently Decompress got dir path [%s]", info.TarFile)
		if len(info.TarFile) == 0 {
			return nil, errors.New("fail to get tar file, got empty tar file")
		}

		decompress := fmt.Sprintf("tar -zxvf %s", info.TarFile)

		wk, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get work from worker repository")
		}
		wk.init.append(decompress)
		return nil, nil
	case SetupPhase, RunningPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) Copy(ctx context.Context, info *worker.CopyInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		log.Infof(ctx, "Currently copy, source [%s], target [%s]", info.SourceFile, info.TargetPath)
		if len(info.SourceFile) == 0 || len(info.TargetPath) == 0 {
			return nil, errors.Errorf("fail to get source [%s] and target [%s] file", info.SourceFile, info.TargetPath)
		}

		cp := fmt.Sprintf("cp -r %s %s", info.SourceFile, info.TargetPath)
		wk, err := k.wrp.load(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "get work from worker repository")
		}
		wk.init.append(cp)
		return nil, nil
	case SetupPhase, RunningPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) UpdateFileContent(ctx context.Context, info *worker.UpdateFileContentInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case InstallPhase:
		log.Infof(ctx, "Currently update file content, target File", info.TargetFile)
		log.Debugf(ctx, "Currently update file content, new content [%s]", string(info.NewContent))
		if len(info.TargetFile) == 0 || len(info.NewContent) == 0 {
			return nil, errors.Errorf("fail to get target file [%s]", info.TargetFile)
		}

		return nil, nil
	case SetupPhase, RunningPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) DeleteFile(ctx context.Context, info *worker.DeleteFileInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case RunningPhase, InstallPhase:
		return nil, nil
	case SetupPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) CreateFile(ctx context.Context, info *worker.CreateFileInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case RunningPhase, InstallPhase:
		return nil, nil
	case SetupPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) CreateDir(ctx context.Context, info *worker.CreateDirInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case RunningPhase, InstallPhase:
		return nil, nil
	case SetupPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
}

func (k *Worker) RemoveDir(ctx context.Context, info *worker.RemoveDirInfo) (*worker.Empty, error) {
	ap, err := phaseParse(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get phase from context")
	}
	switch ap {
	case RunningPhase, InstallPhase:
		return nil, nil
	case SetupPhase:
		panic("implement me")
	default:
		return nil, errors.Wrapf(err, "not support phase [%s]", ap)
	}
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
