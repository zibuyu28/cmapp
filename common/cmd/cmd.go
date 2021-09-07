package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type ShellType int

const (
	ShellTypeNone  ShellType = 1
	ShellTypeShell ShellType = 2
	ShellTypeBash  ShellType = 3
)

type Ins struct {
	shellType ShellType
	envs      map[string]string
	timeout   int
	forceKill bool
	command   string
	args      []string
	retry     int
	workdir   string
}
type CmdOption func(info *Ins)

func NewDefaultCMD(command string, args []string, opts ...CmdOption) *Ins {
	i := Ins{
		shellType: setShellType(),
		timeout:   10,
		forceKill: true,
		command:   command,
		args:      args,
		retry:     1,
	}
	for _, opt := range opts {
		opt(&i)
	}
	return &i
}

func WithRetry(retry int) CmdOption {
	if retry < 0 {
		retry = 1
	}
	return func(i *Ins) {
		i.retry = retry
	}
}

func WithEnvs(envs map[string]string) CmdOption {
	return func(i *Ins) {
		i.envs = envs
	}
}

func WithShellType(shellType ShellType) CmdOption {
	return func(i *Ins) {
		i.shellType = shellType
	}
}

func WithTimeout(timeout int) CmdOption {
	return func(i *Ins) {
		i.timeout = timeout
	}
}

func WithWorkDir(dir string) CmdOption {
	return func(i *Ins) {
		i.workdir = dir
	}
}

func WithForceKill(forceKill bool) CmdOption {
	return func(i *Ins) {
		i.forceKill = forceKill
	}
}

func (i *Ins) Run() (out string, err error) {
	for t := 0; t < i.retry; t++ {
		out, err = i.cmd()
		if err != nil {
			fmt.Printf("fail to exe : %s\n", i.command)
			fmt.Printf("output : %s\n", out)
			fmt.Printf("err : %s\n", err.Error())
			fmt.Printf("retry : %d\n", t+1)
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}
	return
}

func setShellType() ShellType {
	var shellType ShellType
	if exist, _ := pathExist("/bin/bash"); exist {
		shellType = ShellTypeBash
	} else if exist, _ = pathExist("/bin/sh"); exist {
		shellType = ShellTypeShell
	} else {
		shellType = ShellTypeNone
	}
	return shellType
}

func (i *Ins) cmd() (out string, err error) {
	var (
		stdout      bytes.Buffer
		stderr      bytes.Buffer
		cmd         *exec.Cmd
		sysProcAttr *syscall.SysProcAttr
	)

	sysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 使子进程拥有自己的 pgid，等同于子进程的 pid
	}

	// 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(i.timeout)*time.Second)
	defer cancel()

	switch i.shellType {
	case ShellTypeNone:
		cmd = exec.Command(i.command, i.args...)
	case ShellTypeShell:
		cmd = exec.Command("/bin/sh", "-c", i.command)
	case ShellTypeBash:
		cmd = exec.Command("/bin/bash", "-c", i.command)
	default:
		return "", errors.Errorf("not support shell type [%v]", i.shellType)
	}
	if len(i.workdir) != 0 {
		cmd.Dir = i.workdir
	}

	cmd.SysProcAttr = sysProcAttr
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// add env to cmd
	if i.envs != nil {
		addEnv(cmd, i.envs)
	}

	if err := cmd.Start(); err != nil {
		log.Debugf(context.Background(), "stdout: %s", stdout.String())
		log.Debugf(context.Background(), "stderr: %s", stderr.String())
		log.Debugf(context.Background(), "err: %s", err.Error())
		return "", err
	}

	waitChan := make(chan struct{}, 1)
	defer close(waitChan)

	// 超时杀掉进程组 或正常退出
	go func() {
		select {
		case <-ctx.Done():
			//fmt.Println("ctx timeout")
			if i.forceKill {
				//fmt.Printf("ctx timeout kill job ppid:%d\n", cmd.Process.Pid)
				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					//fmt.Println("syscall.Kill return err: ", err)
					return
				}
			}
		case <-waitChan:
			//fmt.Printf("normal quit job ppid:%d\n", cmd.Process.Pid)
		}
	}()

	if err := cmd.Wait(); err != nil {
		//fmt.Printf("timeout kill job ppid:%s\n%s\n", b.String(), err.Error())
		em := err.Error()
		// 超时退出，返回调用失败
		if strings.Contains(em, "signal: killed") {
			return "", err
		}
		// 未超时，被执行程序主动退出
		log.Debugf(context.Background(), "stdout: %s", stdout.String())
		log.Debugf(context.Background(), "stderr: %s", stderr.String())
		if len(stderr.String()) != 0 {
			return stdout.String(), errors.New(stderr.String())
		}
		return stdout.String() + "\n" + stderr.String(), err
		//return fmt.Sprintf("%s\n%v", b.String(), err), true
	}

	out = stdout.String()
	// 唤起正常退出
	waitChan <- struct{}{}

	return
}

func addEnv(cmd *exec.Cmd, envs map[string]string) {
	if len(cmd.Env) == 0 {
		cmd.Env = []string{}
	}
	for key, value := range envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
}

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
