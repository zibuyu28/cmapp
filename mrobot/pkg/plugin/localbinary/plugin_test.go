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

package localbinary

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_addrValidate(t *testing.T) {

	t.Run("test addr validate", func(t *testing.T) {
		err := addrValidate("127.0.0.1")
		assert.NotNil(t, err)
		err = addrValidate("127.0.0.1:8088")
		assert.NotNil(t, err)
	})
}
