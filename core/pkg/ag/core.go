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

package ag

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	fp "path/filepath"
)

type Core struct {
	ApiVersion   APIVersion
	coreHttpAddr string
}

var coreDefaultHost = "127.0.0.1"

var coreDefaultPort = 9008

var coreDefaultHttpAddr = "http://127.0.0.1:9008"

func (c Core) DownloadFile(fileName string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", getFileURL(c.ApiVersion, c.coreHttpAddr), fileName), nil)
	if err != nil {
		return nil, errors.Wrap(err, "new http request")
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do http request")
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response code : %d", response.StatusCode)
	}
	defer response.Body.Close()

	all, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}
	return all, nil
}

func (c Core) UploadFile(fileName string) (string, error) {
	fph, err := fp.Abs(fileName)
	if err != nil {
		return "", errors.Wrap(err, "get file absolute path")
	}
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	open, err := os.Open(fph)
	if err != nil {
		return "", errors.Wrapf(err, "open file [%s]", fph)
	}
	_, tFileName := fp.Split(fileName)
	go func() {
		defer open.Close()
		fw1, _ := writer.CreateFormFile("file", tFileName)
		//h := make(textproto.MIMEHeader)
		//h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes("file"), escapeQuotes(fileName)))
		//h.Set("Content-Type", "text/plain")
		//fw1, _ := writer.CreatePart(h)
		_, _ = io.Copy(fw1, open)
		_ = writer.Close()
		_ = pw.Close()
	}()
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", getFileURL(c.ApiVersion, c.coreHttpAddr), tFileName), pr)
	if err != nil {
		return "", errors.Wrap(err, "new post request")
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Wrap(err, "do http request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http response code : %d", res.StatusCode)
	}
	all, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "read body")
	}
	var resp = struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}{}
	err = json.Unmarshal(all, &resp)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal res")
	}
	if s, ok := resp.Data.(string); !ok {
		return "", errors.Errorf("res data [%v] not string", resp.Data)
	} else {
		return s, nil
	}
}

func getFileURL(version APIVersion, coreHttpAddr string) string {
	//host := viper.GetString("CORE_HOST")
	if len(coreHttpAddr) == 0 {
		log.Debugf(context.Background(), "use default core addr [%s]", coreDefaultHttpAddr)
		coreHttpAddr = coreDefaultHttpAddr
	}

	//port := viper.GetInt("CORE_PORT")
	//if port == 0 {
	//	port = coreDefaultPort
	//}
	return fmt.Sprintf("%s/api/%s/file", coreHttpAddr, version)
}
