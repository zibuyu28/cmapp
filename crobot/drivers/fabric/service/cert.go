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

package service

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/fabtool"
)

type CertWorker struct {
	BaseWorker
}

// NewCertWorker new cert worker
func NewCertWorker(driveruuid, workDir string) *CertWorker {
	return &CertWorker{
		BaseWorker: BaseWorker{
			driveruuid: driveruuid,
			workDir:    workDir,
		},
	}
}

func (c *CertWorker) InitCert(ctx context.Context, chain *fabric.Fabric) error {
	newTool, err := fabtool.NewTool(ctx, c.driveruuid, "cryptogen", chain.Version)
	if err != nil {
		err = errors.Wrap(err, "new tool")
		return err
	}

	err = newTool.(fabtool.CryptoGenTool).GenerateInitCert(c.driveruuid, chain, c.workDir)
	if err != nil {
		err = errors.Wrap(err, "generate cert")
		return err
	}
	return nil
}

