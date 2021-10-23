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
	StartApp(in *App) error
	StopApp(in *App) error
	DestroyApp(in *App) error

	// --- construct App ---

	TagEx(in *Tag) error
	FileMountEx(in *FileMount) error
	EnvEx(in *EnvVar) error
	NetworkEx(in *Network) error
	FilePremiseEx(in *File) error
	LimitEx(in *Limit) error
	HealthEx(in *Health) error
	LogEx(in *Log) error
}

type CoreAPI interface {
	DownloadFile(fileName string) ([]byte, error)
}
