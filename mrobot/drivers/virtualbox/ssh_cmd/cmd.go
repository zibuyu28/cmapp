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
	"github.com/bramvdbogaerde/go-scp"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var defaultTimeout = time.Second * 3

type SSHCli struct {
	rmtHost  string `validate:"required"`
	rmtPort  int    `validate:"required"`
	userName string `validate:"required"`
	password string
	keyPath  string

	secure  bool
	timeout time.Duration
	SSHCli  *ssh.Client

	envs map[string]string
}

func NewSSHCliWithKey(rmtHost string, rmtPort int, username, keyPath string) (*SSHCli, error) {
	cli := &SSHCli{
		rmtHost:  rmtHost,
		rmtPort:  rmtPort,
		userName: username,
		keyPath:  keyPath,
		secure:   false,
		timeout:  defaultTimeout,
	}
	err := cli.initConn()
	if err != nil {
		return nil, errors.Wrap(err, "init connect")
	}
	return cli, nil
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

func WithEnv(envs map[string]string) func(*SSHCli) {
	return func(cli *SSHCli) {
		cli.envs = envs
	}
}

func (s *SSHCli) Scp(locFile, rmtFile string) error {
	_, err := os.Stat(locFile)
	if err != nil {
		return err
	}

	scpcli, err := scp.NewClientBySSH(s.SSHCli)
	if err != nil {
		return errors.Wrap(err, "establish connection to remote server")
	}

	// Open a file
	f, err := os.Open(locFile)
	if err != nil {
		return errors.Wrapf(err, "open local file [%s]", locFile)
	}

	// Close client connection after the file has been copied
	defer scpcli.Close()

	// Close the file after it has been copied
	defer f.Close()

	// make sure dir is created
	rdir := filepath.Dir(rmtFile)
	out, err := s.ExecCmd(fmt.Sprintf("mkdir -p %s", rdir))
	if err != nil {
		return errors.Wrapf(err, "exec create dir [%s] command", rdir)
	}
	if len(out) != 0 {
		return errors.Errorf("fail to exec create dir [%s] command, res [%s]", rdir, out)
	}

	err = scpcli.CopyFile(f, rmtFile, "0777")
	if err != nil {
		return errors.Wrapf(err, "copying file to remote [%s]", rmtFile)
	}
	return nil
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

	if s.envs != nil{
		var envs []string
		for key, val := range s.envs {
			//err = sess.Setenv(key, val)
			//if err != nil {
			//	return "", errors.Wrapf(err, "set env key [%s], val [%s]", key, val)
			//}
			envs = append(envs, fmt.Sprintf("%s=%s", key, val))
		}
		if len(envs) != 0 {
			cmd = fmt.Sprintf("%s %s",strings.Join(envs, " "), cmd)
		}
	}
	defer func() { s.envs = nil }()
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
	if config == nil {
		return errors.Errorf("fail to get config, please check param [%v]", s)
	}
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
	if len(s.password) == 0 && len(s.keyPath) == 0 {
		return nil
	}
	var authM ssh.AuthMethod
	if len(s.password) != 0 {
		authM = ssh.Password(s.password)
	} else if len(s.keyPath) != 0 {
		file, err := ioutil.ReadFile(s.keyPath)
		if err != nil {
			fmt.Printf("key [%s] not exist, err [%v]\n", s.keyPath, err)
			return nil
		}
		key, err := ssh.ParsePrivateKey(file)
		if err != nil {
			fmt.Printf("parse private key , err [%v]\n", err)
			return nil
		}
		authM = ssh.PublicKeys(key)
	}
	config := &ssh.ClientConfig{
		User: s.userName,
		Auth: []ssh.AuthMethod{
			authM,
		},
		Timeout: s.timeout,
	}
	if !s.secure {
		// Non-production only
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	return config
}
