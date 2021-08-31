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
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
)

var defaultISOPath = "./bin/boo2docker.iso"

func NewRMTB2DUpdater(ctx context.Context, client *ssh.Client, isoPath string) B2DUpdater {
	return &rmtB2DUpdater{
		ctx:     ctx,
		cli:     client,
		isoPath: isoPath,
	}
}

type rmtB2DUpdater struct {
	ctx     context.Context
	cli     *ssh.Client
	isoPath string
}

func (u *rmtB2DUpdater) CopyIsoToMachineDir(storePath, machineName, isoURL string) error {
	client, err := scp.NewClientBySSH(u.cli)
	if err != nil {
		log.Errorf(u.ctx, "Couldn't establish a connection to the remote server, Err: [%v]", err)
		return errors.Wrap(err, "establish connection to remote server")
	}

	if len(u.isoPath) == 0 {
		u.isoPath = defaultISOPath
	}
	// Open a file
	f, err := os.Open(u.isoPath)
	if err != nil {
		return errors.Wrap(err, "open boot2docker.iso")
	}

	// Close client connection after the file has been copied
	defer client.Close()

	// Close the file after it has been copied
	defer f.Close()

	// make sure dir is created
	dir := filepath.Join(storePath, "machines", machineName)
	out, err := u.execCmd(fmt.Sprintf("mkdir -p %s", dir))
	if err != nil {
		return errors.Wrapf(err, "exec create dir [%s] command", dir)
	}
	if len(out) != 0 {
		return errors.Errorf("fail to exec create dir [%s] command, res [%s]", dir, out)
	}

	rmtFile := filepath.Join(dir, "boot2docker.iso")

	err = client.CopyFile(f, rmtFile, "0777")
	if err != nil {
		return errors.Wrap(err, "copying file to remote")
	}

	return nil
}

func (u *rmtB2DUpdater) UpdateISOCache(storePath, isoURL string) error {
	log.Infof(u.ctx, "Currently update iso cache storePath [%s], iso url [%s]", storePath, isoURL)
	return nil
}

func (u *rmtB2DUpdater) execCmd(cmd string) (string, error) {
	sess, err := u.cli.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "new session with ssh server")
	}
	defer sess.Close()

	// exe command
	res, err := sess.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "exec command [%s]", cmd)
	}
	return string(res), nil
}
