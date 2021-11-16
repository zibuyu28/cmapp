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

package md5

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
)

func MD5(source string) string {
	if len(source) == 0 {
		return ""
	}
	data := []byte(source)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}


// FileMD5 获取文件的md5码
func FileMD5(filename string) (string, error) {
	// 文件全路径名
	pFile, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)

	return hex.EncodeToString(md5h.Sum(nil)), nil
}
