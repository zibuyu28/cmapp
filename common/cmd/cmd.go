package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
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
	stream    bool
	outStream chan string
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

func WithStream(outStream chan string) CmdOption {
	return func(i *Ins) {
		i.stream = true
		i.outStream = outStream
	}
}

func (i *Ins) Run() (out string, err error) {
	for t := 0; t < i.retry; t++ {
		if i.stream {
			err := i.cmdWithStream()
			if err != nil {
				fmt.Printf("fail to exe : %s\n", i.command)
				fmt.Printf("err : %v\n", err)
				fmt.Printf("retry : %d\n", t+1)
				time.Sleep(time.Second * 2)
			} else {
				break
			}
		} else {
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
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))

	// add env to cmd
	if i.envs != nil {
		addEnv(cmd, i.envs)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf( "stdout: %s\n", stdout.String())
		fmt.Printf( "stderr: %s\n", stderr.String())
		fmt.Printf( "err: %s\n", err.Error())
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
		fmt.Printf( "stdout: %s\n", stdout.String())
		fmt.Printf( "stderr: %s\n", stderr.String())
		if len(stderr.String()) != 0 {
			return stdout.String(), errors.New(stderr.String())
		}
		return stdout.String() + "\n" + stderr.String(), err
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

func AttachStream(scanner *bufio.Scanner, streamOutCh chan string, stopCh chan struct{}) {
	go stream(scanner, streamOutCh, stopCh)
}

func stream(scanner *bufio.Scanner, streamOutCh chan<- string, stopCh <-chan struct{}) {
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Printf( "Scanning stream: %s\n", err)
		}
		select {
		case streamOutCh <- strings.Trim(line, "\n"):
		case <-stopCh:

			return
		}
	}
}

func (i *Ins) cmdWithStream() error {
	var (
		cmd         *exec.Cmd
		stderr      bytes.Buffer
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
		return errors.Errorf("not support shell type [%v]", i.shellType)
	}
	if len(i.workdir) != 0 {
		cmd.Dir = i.workdir
	}

	cmd.SysProcAttr = sysProcAttr
	cmd.Stderr = &stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error getting cmd stdout pipe: %v", err)
	}
	outScanner := bufio.NewScanner(stdoutPipe)
	outStopChan := make(chan struct{}, 1)
	defer close(outStopChan)
	AttachStream(outScanner, i.outStream, outStopChan)

	// add env to cmd
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	if i.envs != nil {
		addEnv(cmd, i.envs)
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "cmd start")
	}

	waitChan := make(chan struct{}, 1)
	defer close(waitChan)

	// 超时杀掉进程组 或正常退出
	go func() {
		select {
		case <-ctx.Done():
			if i.forceKill {
				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					return
				}
			}
		case <-waitChan:
		}
	}()

	if err := cmd.Wait(); err != nil {
		em := err.Error()
		// 超时退出，返回调用失败
		if strings.Contains(em, "signal: killed") {
			return errors.Wrap(err, "cmd wait")
		}
		// 未超时，被执行程序主动退出
		fmt.Printf( "stderr: %s\n", stderr.String())
		if len(stderr.String()) != 0 {
			return errors.New(stderr.String())
		}
		return nil
	}
	// 唤起正常退出
	waitChan <- struct{}{}

	return nil
}
