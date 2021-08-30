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

package virtualbox

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/ssh_cmd"
	"strings"
)

// VBoxRemoteCmdManager communicates with VirtualBox through the commandline using `VBoxManage`.
type VBoxRemoteCmdManager struct {
	ctx context.Context
	cli *ssh_cmd.SSHCli
}

// NewVBoxRemoteCmdManager new vb remote cmd manager
func NewVBoxRemoteCmdManager(ctx context.Context, cli *ssh_cmd.SSHCli) *VBoxRemoteCmdManager {
	return &VBoxRemoteCmdManager{
		ctx: ctx,
		cli: cli,
	}
}

func (V *VBoxRemoteCmdManager) vbm(args ...string) error {
	cmd := strings.Join(args, " ")
	outORErr, err := V.cli.ExecCmd(fmt.Sprintf("/usr/local/bin/VBoxManage %s",cmd))
	if err != nil {
		return errors.Wrapf(err, "exe cmd [%s]", cmd)
	}
	log.Debugf(context.Background(), "cmd exec return [%s]", outORErr)
	return nil
}

func (V *VBoxRemoteCmdManager) vbmOut(args ...string) (string, error) {
	cmd := strings.Join(args, " ")
	outORErr, err := V.cli.ExecCmd(fmt.Sprintf("/usr/local/bin/VBoxManage %s",cmd))
	if err != nil {
		return "", errors.Wrapf(err, "exe cmd [%s]", cmd)
	}
	return outORErr, nil
}

// vbmOutErr stdout stderr err
func (V *VBoxRemoteCmdManager) vbmOutErr(args ...string) (string, string, error) {
	cmd := strings.Join(args, " ")
	outORErr, err := V.cli.ExecCmd(fmt.Sprintf("/usr/local/bin/VBoxManage %s",cmd))
	if err != nil {
		return "", "", errors.Wrapf(err, "exe cmd [%s]", cmd)
	}
	return outORErr, outORErr, nil
}
