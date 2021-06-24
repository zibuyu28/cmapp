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

package service_c

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/core/internal/cmd"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"os"
	"path/filepath"
)


// Create execute driver create command to initialization machine
func Create(driverid int) error {
	drv, err := model.GetDriverByID(driverid)
	if err != nil {
		return errors.Wrap(err, "get driver by id")
	}
	err = CreateAction("drv.DriverPath",drv.Name, uuid.New().String())
	if err != nil {
		return errors.Wrap(err, "create aciton")
	}
	return nil
}
// CreateAction driver to create machine
func CreateAction(driverRootPath, driverName string, args ...string) error {
	abs, _ := filepath.Abs(filepath.Join(driverRootPath, driverName))
	binaryPath := fmt.Sprintf("%s/exec/driver", abs)
	_, err := os.Stat(binaryPath)
	if err != nil {
		if os.ErrNotExist == err  {
			return errors.Wrapf(err, "driver executable file [%s]",binaryPath)
		}
		return err
	}
	command := fmt.Sprintf("%s ro create", binaryPath)
	newCmd := cmd.NewDefaultCMD(command, args)
	out, err := newCmd.Run()
	if err != nil {
		return errors.Wrapf(err, "fail to execute command [%s]", command)
	}
	fmt.Println(out)
	return nil
}