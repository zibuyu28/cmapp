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

import "github.com/zibuyu28/cmapp/mrobot/proto"

// BaseDriver base driver
type BaseDriver struct {
	MachineName string
	UUID        string

	CoreAddr        string
	ImageRepository ImageRepository
}

type ImageRepository struct {
	Repository string
	StorePath  string
}

func (b *BaseDriver) GetFlags() []*proto.Flag {
	return []*proto.Flag{
		{
			Name:   "CoreAddr",
			Usage:  "core http addr",
			EnvVar: "BASE_CORE_ADDR",
			Value:  nil,
		},
		{
			Name:   "CoreAddr",
			Usage:  "core http addr",
			EnvVar: "BASE_CORE_ADDR",
			Value:  nil,
		},
		{
			Name:   "ImageRepository",
			Usage:  "image repository",
			EnvVar: "BASE_IMAGE_REPOSITORY",
			Value:  nil,
		},
		{
			Name:   "ImageStorePath",
			Usage:  "image repository",
			EnvVar: "BASE_IMAGE_STORE_PATH",
			Value:  nil,
		},
	}
}
