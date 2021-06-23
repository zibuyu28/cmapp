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

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// Command run command
// shellMode 是否是shell脚本
// timeout 设置超时时间
// forceKill 达到超时时间，是否强制杀死进程（包括子进程和孙进程）
// command shell脚本或二进制全路径
// arg 二进制运行参数，shell模式无效
func Command(shellMode bool, timeout int, forceKill bool, command string, arg ...string) (out string, res bool) {
	var (
		b           bytes.Buffer
		cmd         *exec.Cmd
		sysProcAttr *syscall.SysProcAttr
	)

	sysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 使子进程拥有自己的 pgid，等同于子进程的 pid
	}

	// 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if shellMode {
		cmd = exec.Command("/bin/bash", "-c", command)
	} else {
		cmd = exec.Command(command, arg...)
	}

	cmd.SysProcAttr = sysProcAttr
	cmd.Stdout = &b
	cmd.Stderr = &b

	if err := cmd.Start(); err != nil {
		fmt.Printf("%s\n%s\n", b.String(), err.Error())
		return
	}

	waitChan := make(chan struct{}, 1)
	defer close(waitChan)

	// 超时杀掉进程组 或正常退出
	go func() {
		select {
		case <-ctx.Done():
			fmt.Println("ctx timeout")
			if forceKill {
				fmt.Printf("ctx timeout kill job ppid:%d\n", cmd.Process.Pid)
				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					fmt.Println("syscall.Kill return err: ", err)
					return
				}
			}
		case <-waitChan:
			fmt.Printf("normal quit job ppid:%d\n", cmd.Process.Pid)
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Printf("timeout kill job ppid:%s\n%s\n", b.String(), err.Error())
		return
	}

	out = b.String()
	res = true
	// 唤起正常退出
	waitChan <- struct{}{}

	return
}
