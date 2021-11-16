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

package core

import (
	"context"
	"fmt"
	v "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/httputil"
	"github.com/zibuyu28/cmapp/common/log"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"strings"
)

const (
	driAgentCoreHTTPAddr string = "COREHTTPADDR"
	driverPrefix         string = "DRIAGENT_"
)

var cli *client

func init() {
	environ := os.Environ()
	for _, s := range environ {
		if strings.HasPrefix(s, driverPrefix) {
			kvs := strings.SplitN(strings.TrimPrefix(s, driverPrefix), "=", 2)
			if len(kvs) != 2 {
				continue
			}
			if kvs[0] == driAgentCoreHTTPAddr {
				cli = &client{coreAddr: kvs[1]}
			}
			break
		}
	}
}

type client struct {
	coreAddr string
}

// Package package info
type Package struct {
	Name    string `json:"name" validate:"required"`
	Version string `json:"version" validate:"required"`
	Image   struct {
		ImageName     string   `json:"image_name" validate:"required"`
		Tag           string   `json:"tag" validate:"required"`
		WorkDir       string   `json:"work_dir" validate:"required"`
		StartCommands []string `json:"start_command" validate:"required"`
	}
	Binary struct {
		Download            string   `json:"download" validate:"required"`
		CheckSum            string   `json:"check_sum" validate:"required"`
		PackageHandleShells []string `json:"package_handle_shells"`
		StartCommands       []string `json:"start_command" validate:"required"`
	}
}

// PackageInfo get package info from core
func PackageInfo(ctx context.Context, name, version string) (*Package, error) {
	if cli == nil {
		return nil, errors.New("core client is nil")
	}
	// "/api/v1/package"
	packageUrl := fmt.Sprintf("http://%s/api/v1/packege/%s/%s", cli.coreAddr, name, version)

	resp, err := httputil.HTTPDoGet(packageUrl)
	if err != nil {
		return nil, errors.Wrap(err, "http do get")
	}
	log.Debugf(ctx, "Currently get package info from core. resp [%s]", string(resp))
	var res = struct {
		Code    int     `json:"code"`
		Message string  `json:"message"`
		Data    Package `json:"data"`
	}{}
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal resp to package")
	}
	p := res.Data
	validate := v.New()
	err = validate.Struct(p)
	if err != nil {
		return nil, errors.Wrap(err, "validate param")
	}
	return &p, nil
}
