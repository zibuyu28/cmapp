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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSSHCli_ExecCmd(t *testing.T) {
	//i := int64(20000) << 20
	//fmt.Println(i)
	//cli, err := NewSSHCli("wanghenangdembp", 22, "wanghengfang", "ww1428")
	//assert.Nil(t, err)
	//assert.NotNil(t, cli)
	//cmd, err := cli.ExecCmd("env")
	//assert.Nil(t, err)
	//t.Logf("%s", cmd)

	cli, err := NewSSHCliWithKey("10.1.41.185", 46452, "docker", "/Users/wanghengfang/GolandProjects/cmapp/core/sshkey/9b3f710f-8521-42fe-b39e-4eba530a7631/id_rsa")
	assert.Nil(t, err)
	cmd, err := cli.ExecCmd("env", WithEnv(map[string]string{"A": "A","C":"C","D":"D"}))
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
	t.Logf(cmd)
}
