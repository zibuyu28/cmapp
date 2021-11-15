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

package fabtool

import (
	"context"
	"fmt"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"gopkg.in/yaml.v2"
)

type HostnameData struct {
	Prefix string
	Index  int
	Domain string
}

type SpecData struct {
	Hostname   string
	Domain     string
	CommonName string
}

type NodeTemplate struct {
	Count    int      `yaml:"Count"`
	Start    int      `yaml:"Start"`
	Hostname string   `yaml:"Hostname"`
	SANS     []string `yaml:"SANS"`
}

type NodeSpec struct {
	isAdmin            bool
	Hostname           string   `yaml:"Hostname"`
	CommonName         string   `yaml:"CommonName"`
	Country            string   `yaml:"Country"`
	Province           string   `yaml:"Province"`
	Locality           string   `yaml:"Locality"`
	OrganizationalUnit string   `yaml:"OrganizationalUnit"`
	StreetAddress      string   `yaml:"StreetAddress"`
	PostalCode         string   `yaml:"PostalCode"`
	SANS               []string `yaml:"SANS"`
}

type UsersSpec struct {
	Count int `yaml:"Count"`
}

type OrgSpec struct {
	Name          string       `yaml:"Name"`
	Domain        string       `yaml:"Domain"`
	EnableNodeOUs bool         `yaml:"EnableNodeOUs"`
	CA            NodeSpec     `yaml:"CA"`
	Template      NodeTemplate `yaml:"Template"`
	Specs         []NodeSpec   `yaml:"Specs"`
	Users         UsersSpec    `yaml:"Users"`
}

type Config struct {
	OrdererOrgs []OrgSpec `yaml:"OrdererOrgs"`
	PeerOrgs    []OrgSpec `yaml:"PeerOrgs"`
}

type CryptoGenTool struct {
	ctx context.Context
}

func newCryptoGen(ctx context.Context) RT {
	return CryptoGenTool{ctx: ctx}
}

//func (c CryptoGenTool) ExtendPeerCert(chain *model.FabricChain, nodes []model.Node, baseDir string) error {
//	pwd, _ := os.Getwd()
//	cryptogen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/cryptogen", chain.FabricVersion))
//	randomCryptoFilePath, _ := filepath.Abs(fmt.Sprintf("%s/%d-addpeer.yaml", baseDir, time.Now().UnixNano()))
//	certPath, _ := filepath.Abs(baseDir)
//
//	config := constructInitConfig(chain)
//	if config == nil {
//		return errors.New("config is nil")
//	}
//	configYaml, err := yaml.Marshal(config)
//	if err != nil {
//		err = errors.Wrap(err, "marshal config")
//		return err
//	}
//
//	// generate crypto config file
//	err = ioutil.WriteFile(randomCryptoFilePath, configYaml, os.ModePerm)
//	if err != nil {
//		err = errors.Wrap(err, "build crypto config")
//		return err
//	}
//
//	output, err := util.CMD(10,
//		fmt.Sprintf("%s extend --config=%s --input=%s/", cryptogen, randomCryptoFilePath, certPath))
//	fmt.Printf("output : %s\n", output)
//	if err != nil {
//		err = errors.Wrap(err, "exec crypto extend cert")
//		return err
//	}
//
//	for _, info := range chain.OrganizationInfos {
//		if info.OrganizationType != model.PeerOrganization {
//			continue
//		}
//		for _, node := range nodes {
//			if node.OrgID != info.ID {
//				continue
//			}
//			orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.fabric.com/peers/peer%d.%s.fabric.com", info.UUID, node.ID, info.UUID))
//			if !util.FileExist(orgCertPath) {
//				err = errors.Errorf("node(%s) cert generate err", node.NodeName)
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func (c CryptoGenTool) GenerateNewOrgCert(chain *model.FabricChain, org *model.NewOrg, baseDir string) error {
//	pwd, _ := os.Getwd()
//	cryptogen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/cryptogen", chain.FabricVersion))
//	randomCryptoFilePath, _ := filepath.Abs(fmt.Sprintf("%s/%d-neworg.yaml", baseDir, time.Now().UnixNano()))
//	certPath, _ := filepath.Abs(baseDir)
//
//	config := constructNewOrgConfig(org)
//	configYaml, err := yaml.Marshal(config)
//	if err != nil {
//		err = errors.Wrap(err, "marshal config")
//		return err
//	}
//
//	// generate crypto config file
//	err = ioutil.WriteFile(randomCryptoFilePath, configYaml, os.ModePerm)
//	if err != nil {
//		err = errors.Wrap(err, "build crypto config")
//		return err
//	}
//
//	output, err := util.CMD(10,
//		fmt.Sprintf("%s generate --config=%s --output=%s/", cryptogen, randomCryptoFilePath, certPath))
//	fmt.Printf("output : %s\n", output)
//	if err != nil {
//		err = errors.Wrap(err, "exec crypto generate cert")
//		return err
//	}
//	for _, info := range org.OrganizationInfos {
//		if info.OrganizationType != model.PeerOrganization {
//			continue
//		}
//		orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.fabric.com", info.UUID))
//		if !util.FileExist(orgCertPath) {
//			err = errors.Errorf("org(%s) cert generate err", info.Name)
//			return err
//		}
//	}
//	return nil
//}

func (c CryptoGenTool) GenerateInitCert(chain *model.Fabric, baseDir string) error {
	cryptogen := filepath.Join(filepath.Dir(os.Args[0]), fmt.Sprintf("tool/%s/cryptogen", chain.Version))
	cryptogenConfigPath, _ := filepath.Abs(fmt.Sprintf("%s/crypto.yaml", baseDir))
	certPath, _ := filepath.Abs(baseDir)

	config := constructInitConfig(chain)
	if config == nil {
		return errors.New("config is nil")
	}
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		err = errors.Wrap(err, "marshal config")
		return err
	}

	// generate crypto config file
	err = ioutil.WriteFile(cryptogenConfigPath, configYaml, os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "build crypto config")
		return err
	}

	command := fmt.Sprintf("generate --config=%s --output=%s/", cryptogenConfigPath, certPath)
	log.Debugf(c.ctx, "GenerateInitCert command [%s %s]", cryptogen, command)
	output, err := cmd.NewDefaultCMD(cryptogen, []string{command}).Run()
	log.Debugf(c.ctx, "output : %s", output)
	if err != nil {
		return errors.Wrap(err, "exec crypto generate cert")
	}

	b, err := checkCertRight(chain, certPath)
	if err != nil {
		return errors.Wrap(err, "check cert exist")
	}
	if !b {
		return errors.New("cert generate fail, file not exist")
	}
	return nil
}

func checkCertRight(chain *model.Fabric, certPath string) (bool, error) {
	for _, peer := range chain.Peers {
		orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.zibuyufab.cn", peer.Organization.UUID))
		_, err := os.Stat(orgCertPath)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, errors.Wrapf(err, "os stat path [%s]", orgCertPath)
		}
	}
	return true, nil
}

//func constructNewOrgConfig(neworg *fabric.NewOrg) *Config {
//	var peerOrgSpecs []OrgSpec
//	for _, org := range neworg.OrganizationInfos {
//		if org.OrganizationType != model.PeerOrganization {
//			continue
//		}
//		peerOrgSpec := OrgSpec{
//			Name:          org.UUID,
//			Domain:        org.UUID + ".fabric.com",
//			EnableNodeOUs: true,
//		}
//		var peerNodes []NodeSpec
//		for _, peer := range neworg.Peers {
//			if peer.OrgID != org.ID {
//				continue
//			}
//			prns := NodeSpec{
//				Hostname:   fmt.Sprintf("peer%d", peer.ID),
//				CommonName: fmt.Sprintf("peer%d.%s.fabric.com", peer.ID, org.UUID),
//				SANS:       []string{peer.NodeHostName},
//			}
//
//			peerNodes = append(peerNodes, prns)
//		}
//
//		peerOrgSpec.Specs = peerNodes
//		peerOrgSpecs = append(peerOrgSpecs, peerOrgSpec)
//	}
//	return &Config{
//		PeerOrgs:    peerOrgSpecs,
//	}
//}

// constructConfig construct config from fabric chain
func constructInitConfig(chain *model.Fabric) *Config {
	// 构建orderer结构
	ordererOrgSpec := OrgSpec{
		Name:          "Orderer",
		Domain:        "orderer.zibuyufab.cn",
		EnableNodeOUs: true,
	}
	var ordererNodes []NodeSpec
	for _, orderer := range chain.Orderers {
		split := strings.Split(orderer.NodeHostName, ":")
		odns := NodeSpec{
			Hostname:   split[0],
			//CommonName: fmt.Sprintf("orderer%s.orderer.fabric.com", orderer.UUID),
			//SANS:       []string{orderer.NodeHostName},
		}

		ordererNodes = append(ordererNodes, odns)
	}

	ordererOrgSpec.Specs = ordererNodes

	// 构建peer组织结构
	var peerOrgSpecs []OrgSpec
	var peerOrgs = make(map[string][]model.Peer)
	for i, p := range chain.Peers {
		if _, ok := peerOrgs[p.Organization.Name]; ok {
			peerOrgs[p.Organization.Name] = append(peerOrgs[p.Organization.Name], chain.Peers[i])
		} else {
			peerOrgs[p.Organization.Name] = []model.Peer{chain.Peers[i]}
		}
	}

	for org, ps := range peerOrgs {
		peerOrgSpec := OrgSpec{
			Name:          org,
			Domain:        org + ".zibuyufab.cn",
			EnableNodeOUs: true,
		}
		var peerNodes []NodeSpec
		for _, peer := range ps {
			split := strings.Split(peer.NodeHostName, ":")
			prns := NodeSpec{
				Hostname:   split[0],
				//CommonName: fmt.Sprintf("peer%s.%s.fabric.com", peer.UUID, org),
				//SANS:       []string{peer.NodeHostName},
			}
			peerNodes = append(peerNodes, prns)
		}

		peerOrgSpec.Specs = peerNodes
		peerOrgSpecs = append(peerOrgSpecs, peerOrgSpec)
	}
	return &Config{
		OrdererOrgs: []OrgSpec{ordererOrgSpec},
		PeerOrgs:    peerOrgSpecs,
	}
}
