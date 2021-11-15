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
	"fmt"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/file"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/fabtool"
)

type CertWorker struct {
	BaseWorker
	nodeCertMap         map[string]string
	organizationCertMap map[string]string
}

// NewCertWorker new cert worker
func NewCertWorker(workDir string) *CertWorker {
	return &CertWorker{
		BaseWorker: BaseWorker{
			workDir: workDir,
		},
	}
}

func (c *CertWorker) InitCert(ctx context.Context, chain *model.Fabric) error {
	newTool, err := fabtool.NewTool(ctx, "cryptogen", chain.Version)
	if err != nil {
		return errors.Wrap(err, "new tool")
	}

	err = newTool.(fabtool.CryptoGenTool).GenerateInitCert(chain, c.workDir)
	if err != nil {
		return errors.Wrap(err, "generate cert")
	}
	err = c.certCompress(chain)
	if err != nil {
		return errors.Wrap(err, "compress cert")
	}
	c.PathMap(chain)
	return nil
}

func (c *CertWorker) GetNodeCertMap() (map[string]string, error) {
	if c.nodeCertMap == nil {
		return nil, errors.New("node cert map is nil, please exec PathMap first")
	}
	return c.nodeCertMap, nil
}

func (c *CertWorker) GetOrganizationCertMap() (map[string]string, error) {
	if c.organizationCertMap == nil {
		return nil, errors.New("organization cert map is nil, please exec PathMap first")
	}
	return c.organizationCertMap, nil
}

func (c *CertWorker) PathMap(chain *model.Fabric) {
	c.nodeCertMap = make(map[string]string)
	c.organizationCertMap = make(map[string]string)
	for _, orderer := range chain.Orderers {
		orderCertTarPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com/orderers/orderer%s.orderer.fabric.com.tar.gz", c.workDir, orderer.UUID)
		orderOrgCertTarPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com.tar.gz", c.workDir)
		c.nodeCertMap[orderer.UUID] = orderCertTarPath
		c.organizationCertMap["orderer"] = orderOrgCertTarPath
	}
	for _, peer := range chain.Peers {
		peerCertTarPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com/peers/peer%s.%s.fabric.com.tar.gz",
			c.workDir, peer.Organization.UUID, peer.UUID, peer.Organization.UUID)
		peerOrgCertTarPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com.tar.gz",
			c.workDir, peer.Organization.UUID)
		c.nodeCertMap[peer.UUID] = peerCertTarPath
		c.organizationCertMap[peer.Organization.UUID] = peerOrgCertTarPath
	}
}

func (c *CertWorker) certCompress(chain *model.Fabric) error {
	for _, orderer := range chain.Orderers {
		orderOrgCertPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com", c.workDir)
		orderOrgCertTarPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com.tar.gz", c.workDir)
		err := file.TarGzDir(orderOrgCertPath, orderOrgCertTarPath, true)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("compress file(%s)", orderOrgCertPath))
		}

		orderCertPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com/orderers/orderer%s.orderer.fabric.com", c.workDir, orderer.UUID)
		orderCertTarPath := fmt.Sprintf("%s/ordererOrganizations/orderer.fabric.com/orderers/orderer%s.orderer.fabric.com.tar.gz", c.workDir, orderer.UUID)
		err = file.TarGzDir(orderCertPath, orderCertTarPath, true)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("compress file(%s)", orderCertPath))
		}
	}
	for _, peer := range chain.Peers {
		peerOrgCertPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com",
			c.workDir, peer.Organization.UUID)
		peerOrgCertTarPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com.tar.gz",
			c.workDir, peer.Organization.UUID)
		err := file.TarGzDir(peerOrgCertPath, peerOrgCertTarPath, true)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("compress file(%s)", peerOrgCertPath))
		}

		peerCertPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com/peers/peer%s.%s.fabric.com",
			c.workDir, peer.Organization.UUID, peer.UUID, peer.Organization.UUID)
		peerCertTarPath := fmt.Sprintf("%s/peerOrganizations/%s.fabric.com/peers/peer%s.%s.fabric.com.tar.gz",
			c.workDir, peer.Organization.UUID, peer.UUID, peer.Organization.UUID)
		err = file.TarGzDir(peerCertPath, peerCertTarPath, true)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("compress file(%s)", peerCertPath))
		}
	}
	return nil
}
