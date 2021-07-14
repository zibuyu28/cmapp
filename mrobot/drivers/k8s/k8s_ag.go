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
	"github.com/zibuyu28/cmapp/core/pkg/context"
)

type AgentK8s struct {
	Name string
}

func StartUp() {
	// TODO: start agent with ws connect to core
}

func (a AgentK8s) GetWorkspace(ctx context.Context) (string, error) {
	panic("implement me")
}

func (a AgentK8s) DestroyWorkspace(ctx context.Context, workspace string) error {
	panic("implement me")
}

func (a AgentK8s) DownloadToPath(ctx context.Context, downloadLink string, targetPath string) error {
	panic("implement me")
}

func (a AgentK8s) Upload(ctx context.Context, source string, targetLink string) error {
	panic("implement me")
}

func (a AgentK8s) Compress(ctx context.Context, dirPath string) error {
	panic("implement me")
}

func (a AgentK8s) Decompress(ctx context.Context, tarFile string) (string, error) {
	panic("implement me")
}

func (a AgentK8s) Copy(ctx context.Context, source, targetPath string) error {
	panic("implement me")
}

func (a AgentK8s) UpdateFileContent(ctx context.Context, targetFile string, newContent []byte) error {
	panic("implement me")
}

func (a AgentK8s) DeleteFile(ctx context.Context, targetFile string) error {
	panic("implement me")
}

func (a AgentK8s) CreateFile(ctx context.Context, file string, content []byte) error {
	panic("implement me")
}

func (a AgentK8s) CreateDir(ctx context.Context, dir string) error {
	panic("implement me")
}

func (a AgentK8s) RemoveDir(ctx context.Context, dir string) error {
	panic("implement me")
}

func (a AgentK8s) FetchFileContent(ctx context.Context, file string) (chan []byte, error) {
	panic("implement me")
}

func (a AgentK8s) CheckTargetPortUseful(ctx context.Context, port int) error {
	panic("implement me")
}

func (a AgentK8s) SetupApp(ctx context.Context, env, appLabels map[string]string) (string, error) {
	panic("implement me")
}

func (a AgentK8s) Done(ctx context.Context) error {
	panic("implement me")
}

func (a AgentK8s) ShutdownApp(ctx context.Context, appUniqueName string) error {
	panic("implement me")
}

func (a AgentK8s) AppHealth(ctx context.Context, appUniqueName string) error {
	panic("implement me")
}

func (a AgentK8s) TargetPortIntranetRoute(ctx context.Context, port int) (string, error) {
	panic("implement me")
}

func (a AgentK8s) TargetPortExternalRoute(ctx context.Context, port int) (string, error) {
	panic("implement me")
}

