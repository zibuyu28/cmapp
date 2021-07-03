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

package tmp

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"text/template"
)

// AdvanceTemplate advance go template
func AdvanceTemplate(data interface{}, tplT []byte) ([]byte, error) {
	t := template.New("temp").Funcs(sprig.TxtFuncMap())
	t, err := t.Parse(string(tplT))
	if err != nil {
		return nil, err
	}
	var buf = new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
