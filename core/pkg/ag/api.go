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

// MachineDriverAPI machine driver api
type MachineDriverAPI interface {
	NewApp(in *NewAppReq) (*App, error)
	StartApp(appuid string, in *App) error
	StopApp(appuid string, in *App) error
	DestroyApp(appuid string, in *App) error

	// --- construct App ---

	TagEx(appuid string, in *Tag) error
	FileMountEx(appuid string, in *FileMount) error
	EnvEx(appuid string, in *EnvVar) error
	NetworkEx(appuid string, in *Network) error
	FilePremiseEx(appuid string, in *File) error
	LimitEx(appuid string, in *Limit) error
	HealthEx(appuid string, in *Health) error
	LogEx(appuid string, in *Log) error
}

type CoreAPI interface {
	DownloadFile(fileName string) ([]byte, error)
	// UploadFile upload file, fileName is full path of the file, than return the download path of this file
	UploadFile(fileName string) (string, error)
}
