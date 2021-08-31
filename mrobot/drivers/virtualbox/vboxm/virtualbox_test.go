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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/mrobot/drivers/virtualbox/ssh_cmd"
	"testing"
)

func TestDriver_CreateVM(t *testing.T) {
	log.InitCus()
	machineName := uuid.New().String()[:8]
	cli, err := ssh_cmd.NewSSHCli("10.1.41.111", 22, "martin1", "zdw1")
	assert.Nil(t, err)
	assert.NotNil(t, cli)
	driver := NewRMTDriver(context.Background(), machineName, "/Users/martin1/Downloads/testvb", "10.1.41.111", cli)
	err = driver.Create()
	t.Logf("%v", err)
	assert.Nil(t, err)
}

func TestDriver_CreateVM2(t *testing.T) {
	log.InitCus()
	machineName := uuid.New().String()[:8]
	cli, err := ssh_cmd.NewSSHCli("192.168.31.63", 22, "wanghengfang", "ww1428")
	assert.Nil(t, err)
	assert.NotNil(t, cli)
	driver := NewRMTDriver(context.Background(), machineName, "/Users/wanghengfang/Downloads/testvb", "192.168.31.63", cli)
	err = driver.Create()
	t.Logf("%v", err)
	assert.Nil(t, err)
}
