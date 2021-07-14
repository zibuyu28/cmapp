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

package worker

import "github.com/zibuyu28/cmapp/core/pkg/context"

type DriverWorker struct {

}


func NewDriverWorker() *DriverWorker {
	//var m = make(map[string]string)
	//et := reflect.TypeOf(DriverWorker{}).Elem()
	//for i := 0; i < et.NumField(); i++ {
	//	name := et.Field(i).Name
	//	jsonName := et.Field(i).Tag.Get("json")
	//	kind := et.Field(i).Type.Kind()
	//	if fg, ok := Flags[name]; ok {
	//		m
	//	}
	//
	//}
	return nil

}

func (d DriverWorker) GetWorkspace(ctx context.Context) (string, error) {
	panic("implement me")
}

func (d DriverWorker) DestroyWorkspace(ctx context.Context, workspace string) error {
	panic("implement me")
}

func (d DriverWorker) DownloadToPath(ctx context.Context, downloadLink string, targetPath string) error {
	panic("implement me")
}

func (d DriverWorker) Upload(ctx context.Context, source, targetLink string) error {
	panic("implement me")
}

func (d DriverWorker) Compress(ctx context.Context, dirPath string) error {
	panic("implement me")
}

func (d DriverWorker) Decompress(ctx context.Context, tarFile string) (string, error) {
	panic("implement me")
}

func (d DriverWorker) Copy(ctx context.Context, source, targetPath string) error {
	panic("implement me")
}

func (d DriverWorker) UpdateFileContent(ctx context.Context, targetFile string, newContent []byte) error {
	panic("implement me")
}

func (d DriverWorker) DeleteFile(ctx context.Context, targetFile string) error {
	panic("implement me")
}

func (d DriverWorker) CreateFile(ctx context.Context, file string, content []byte) error {
	panic("implement me")
}

func (d DriverWorker) CreateDir(ctx context.Context, dir string) error {
	panic("implement me")
}

func (d DriverWorker) RemoveDir(ctx context.Context, dir string) error {
	panic("implement me")
}

func (d DriverWorker) FetchFileContent(ctx context.Context, file string) (chan []byte, error) {
	panic("implement me")
}

func (d DriverWorker) CheckTargetPortUseful(ctx context.Context, port int) error {
	panic("implement me")
}

func (d DriverWorker) SetupApp(ctx context.Context, env, appLabels map[string]string) (string, error) {
	panic("implement me")
}

func (d DriverWorker) Done(ctx context.Context) error {
	panic("implement me")
}

func (d DriverWorker) ShutdownApp(ctx context.Context, appUniqueName string) error {
	panic("implement me")
}

func (d DriverWorker) AppHealth(ctx context.Context, appUniqueName string) error {
	panic("implement me")
}

func (d DriverWorker) TargetPortIntranetRoute(ctx context.Context, port int) (string, error) {
	panic("implement me")
}

func (d DriverWorker) TargetPortExternalRoute(ctx context.Context, port int) (string, error) {
	panic("implement me")
}



