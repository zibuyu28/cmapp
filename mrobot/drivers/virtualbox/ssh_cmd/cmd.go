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

package ssh_cmd

import (
	"fmt"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"time"
)

var defaultTimeout = time.Second * 3

type SSHCli struct {
	rmtHost  string `validate:"required"`
	rmtPort  int    `validate:"required"`
	userName string `validate:"required"`
	password string `validate:"required"`

	secure  bool
	timeout time.Duration
	SSHCli     *ssh.Client
}

// NewSSHCli new ssh client
func NewSSHCli(rmtHost string, rmtPort int, username, password string) (*SSHCli, error) {
	cli := &SSHCli{
		rmtHost:  rmtHost,
		rmtPort:  rmtPort,
		userName: username,
		password: password,
		secure:   false,
		timeout:  defaultTimeout,
	}
	err := cli.initConn()
	if err != nil {
		return nil, errors.Wrap(err, "init connect")
	}
	return cli, nil
}

type Opt func(*SSHCli)

func WithTimeout(timeout time.Duration) func(*SSHCli) {
	return func(cli *SSHCli) {
		cli.timeout = timeout
	}
}

func WithSecure() func(*SSHCli) {
	return func(cli *SSHCli) {
		cli.secure = true
	}
}

func (s *SSHCli) ExecCmd(cmd string, opt ...Opt) (string, error) {
	if s.SSHCli == nil {
		return "", errors.New("fail to get client, because s.cli is nil. please exec InitConn first")
	}
	for _, op := range opt {
		op(s)
	}

	sess, err := s.SSHCli.NewSession()
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

func (s *SSHCli) Close() error {
	if s.SSHCli == nil {
		return nil
	}
	err := s.SSHCli.Close()
	if err != nil {
		return errors.Wrap(err, "close ssh client")
	}
	s.SSHCli = nil
	return nil
}

// InitConn init one client, you must to close manually
func (s *SSHCli) initConn() error {
	validate := v.New()
	err := validate.Struct(*s)
	if err != nil {
		return errors.Wrap(err, "validate client param")
	}

	config := s.config()

	// Connect to host
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.rmtHost, s.rmtPort), config)
	if err != nil {
		return errors.Wrapf(err, "dail to %s:%d", s.rmtHost, s.rmtPort)
	}
	s.SSHCli = client
	// not auto close
	//autoClose(s)
	return nil
}

func autoClose(cli *SSHCli) {
	time.AfterFunc(time.Minute*10, func() {
		_ = cli.Close()
	})
}

func (s *SSHCli) config() *ssh.ClientConfig {
	config := &ssh.ClientConfig{
		User: s.userName,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.password),
		},
		Timeout: s.timeout,
	}
	if !s.secure {
		// Non-production only
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	return config
}
