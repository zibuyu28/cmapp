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

package pkg

import (
	"github.com/zibuyu28/cmapp/plugin/proto/driver"
	"os"
)

// BaseDriver base driver
type BaseDriver struct {
	UUID          string
	DriverName    string
	DriverVersion string
	DriverID      int

	CoreAddr        string          `validate:"required"`
	ImageRepository ImageRepository `validate:"required"`
}

type ImageRepository struct {
	Repository string `validate:"required"`
	StorePath  string `validate:"required"`
}

func (b *BaseDriver) GetFlags() []*driver.Flag {
	return []*driver.Flag{
		{
			Name:   "CoreAddr",
			Usage:  "core http addr",
			EnvVar: "BASE_CORE_ADDR",
			Value:  nil,
		},
		{
			Name:   "Repository",
			Usage:  "image repository",
			EnvVar: "BASE_IMAGE_REPOSITORY",
			Value:  nil,
		},
		{
			Name:   "StorePath",
			Usage:  "image repository",
			EnvVar: "BASE_IMAGE_STORE_PATH",
			Value:  nil,
		},
	}
}


func (b *BaseDriver) ConvertFlags(flags *driver.Flags) map[string]string {
	var fls = make(map[string]string)
	for _, flag := range flags.Flags {
		if len(flag.Value) != 0 {
			fls[flag.Name] = flag.Value[0]
			continue
		}
		envVal := os.Getenv(flag.EnvVar)
		if len(envVal) != 0 {
			fls[flag.Name] = envVal
		}
	}
	return fls
}