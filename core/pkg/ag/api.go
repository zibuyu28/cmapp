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

import "github.com/zibuyu28/cmapp/core/pkg/context"

// MachineAPI machine api for plugin. TODO: add other api and optimization exist api
type MachineAPI interface {
	// GetWorkspace require workspace from driver,
	// please make sure this workspace exist and
	// has permission. context have unique uuid
	// for a series of operations
	GetWorkspace(ctx context.Context) (string, error)

	// DestroyWorkspace destroy workspace from driver
	DestroyWorkspace(ctx context.Context, workspace string) error

	// DownloadToPath download something by download link
	// download link maybe http or other, this is up on
	// driver, target path base on root path of driver,
	// and this target path must exist, driver may not
	// create it
	DownloadToPath(ctx context.Context, downloadLink string, targetPath string) error

	// Upload upload something to target, source file
	// must exist, target link is the remote addr to upload
	Upload(ctx context.Context, source, targetLink string) error

	// Compress compress dir to ~.tar.gz, file name
	// is same with dir name, file will be generated
	// at same level as file dir
	Compress(ctx context.Context, dirPath string) error

	// Decompress decompress file be provided, make sure tar file exist
	// return the father dir path that create by driver, maybe random name
	Decompress(ctx context.Context, tarFile string) (string, error)

	// Copy copy file to target path, make sure source and target exist
	Copy(ctx context.Context, source, targetPath string) error

	// UpdateFileContent update target file content
	UpdateFileContent(ctx context.Context, targetFile string, newContent []byte) error

	// DeleteFile delete target file
	DeleteFile(ctx context.Context, targetFile string) error

	// CreateFile create file with content, make sure the file is not
	// exit and will be created
	CreateFile(ctx context.Context, file string, content []byte) error

	// CreateDir create dir, base on workspace
	CreateDir(ctx context.Context, dir string) error

	// RemoveDir remove dir, base on workspace
	RemoveDir(ctx context.Context, dir string) error

	// FetchFileContent fetch file content, return []byte channel. TODO: check the return
	FetchFileContent(ctx context.Context, file string) (chan []byte, error)

	// CheckTargetPortUseful check target port is occupied or not
	CheckTargetPortUseful(ctx context.Context, port int) error

	// SetupApp setup app with env and labels
	// return app's unique name, maybe uuid or other
	SetupApp(ctx context.Context, env, appLabels map[string]string) (string, error)

	// Done this context has been done
	Done(ctx context.Context) error

	// ShutdownApp shutdown app with unique name
	ShutdownApp(ctx context.Context, appUniqueName string) error

	// AppHealth judge app health or not
	AppHealth(ctx context.Context, appUniqueName string) error

	// TargetPortIntranetRoute create a intranet route for given port
	TargetPortIntranetRoute(ctx context.Context, port int) (string, error)

	// TargetPortExternalRoute create a external route for given port
	TargetPortExternalRoute(ctx context.Context, port int) (string, error)
}
