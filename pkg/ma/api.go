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

package ma

import "context"

// create 这个命令，包含了如何调用api的流程，
// 也就是说create这个命令是如何调用 mengine
// 这个接口的里面的方法的，是由create这个命令
// 决定的

type MEngine interface {
	// create 相关
	InitMachine(engineContext MEngineContext) (Machine, error)

	CreateExec(engineContext MEngineContext) error

	InstallMRobot(engineContext MEngineContext) error

	MRoHealthCheck(engineContext MEngineContext, timeoutSecond int) error

	// delete 相关

	// startUp 相关

	// shutdown 相关

}

type Machine struct {
	UUID       string // 由core提供
	State      int
	DriverID   int
	Tags       []string
	CustomInfo map[string]string
}

type MEngineContext struct {
	context.Context
	UUID   string
	CoreID int
}
