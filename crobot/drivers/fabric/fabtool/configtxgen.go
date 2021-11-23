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
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric/model"
	"os"
	"path/filepath"
)

type ConfigtxgenTool struct {
	ctx context.Context
}

func newConfigTXGen(ctx context.Context) RT {
	return ConfigtxgenTool{ctx: ctx}
}

//func (c ConfigtxgenTool) GenerateTargetChannelTx(chain *model.FabricChain, ch model.ChainChannel, baseDir string) error {
//	pwd, _ := os.Getwd()
//	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
//	configtxPath, _ := filepath.Abs(baseDir)
//	channeltxPath := filepath.Join(configtxPath, fmt.Sprintf("%s.tx", ch.UUID))
//
//	output, err := util.CMD(10,
//		fmt.Sprintf("%s -profile OrgsChannel --configPath=%s/ -outputCreateChannelTx %s -channelID %s",
//			configtxgen, configtxPath, channeltxPath, ch.UUID))
//	fmt.Printf("output : %s\n", output)
//	if err != nil {
//		err = errors.Wrap(err, "exec configtxgen create channel tx")
//		return err
//	}
//	if !utils.FileExists(channeltxPath) {
//		err = errors.New("channel tx create failed")
//		return err
//	}
//	return nil
//}

func (c ConfigtxgenTool) GenerateAllChainChannelTx(chain *model.Fabric, baseDir string) error {
	configtxgen := filepath.Join(filepath.Dir(os.Args[0]), fmt.Sprintf("tool/%s/configtxgen", chain.Version))
	configtxPath, _ := filepath.Abs(baseDir)
	for _, ch := range chain.Channels {
		channeltxPath := filepath.Join(configtxPath, fmt.Sprintf("%s.tx", ch.UUID))
		command := fmt.Sprintf("%s -profile OrgsChannel --configPath=%s/ -outputCreateChannelTx %s -channelID %s",configtxgen, configtxPath, channeltxPath, ch.UUID)
		log.Debugf(c.ctx, "GenerateAllChainChannelTx command [%s]", command)
		output, err := cmd.NewDefaultCMD(command, []string{}, cmd.WithTimeout(10)).Run()
		log.Debugf(c.ctx, "output : %s", output)
		if err != nil {
			err = errors.Wrap(err, "exec configtxgen create channel tx")
			return err
		}
		_, err = os.Stat(channeltxPath)
		if err != nil {
			return errors.Wrapf(err, "check channel tx file [%s] exist", channeltxPath)
		}
	}
	return nil
}

func (c ConfigtxgenTool) GenerateGenesisBlock(chain *model.Fabric, baseDir string) error {
	configtxgen := filepath.Join(filepath.Dir(os.Args[0]), fmt.Sprintf("tool/%s/configtxgen", chain.Version))
	configtxPath, _ := filepath.Abs(baseDir)
	genesisBlockPath := filepath.Join(configtxPath, "orderer.genesis.block")
	command := fmt.Sprintf("%s -profile OrdererGenesis --configPath=%s/ -outputBlock %s", configtxgen, configtxPath, genesisBlockPath)
	log.Debugf(c.ctx, "GenerateGenesisBlock command [%s]", command)
	output, err := cmd.NewDefaultCMD(command, []string{}, cmd.WithTimeout(10)).Run()
	log.Debugf(c.ctx, "output : %s", output)
	if err != nil {
		return errors.Wrap(err, "exec configtxgen create genesis block")
	}
	_, err = os.Stat(genesisBlockPath)
	if err != nil {
		return errors.Wrapf(err, "check genesis block file [%s] exist", genesisBlockPath)
	}
	return nil
}

//func (c ConfigtxgenTool) GenerateChannelArtifacts(chain *model.FabricChain, neworg *model.NewOrg, baseDir string) error {
//	pwd, _ := os.Getwd()
//	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
//
//	// 二进制识别的是 configtx.yaml 这个文件
//	configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))
//
//	// 多个org一个一个生成
//	for _, org := range neworg.OrganizationInfos {
//		if org.OrganizationType != model.PeerOrganization {
//			continue
//		}
//		orgArtifactJsonName, _ := filepath.Abs(fmt.Sprintf("%s/%s.json", configtxPath, org.UUID))
//		cmd := fmt.Sprintf("%s -printOrg %s --configPath=%s/ > %s", configtxgen, org.UUID,
//			configtxPath, orgArtifactJsonName)
//		log.Debug(cmd)
//		output, err := util.CMD(10, cmd)
//		fmt.Printf("output : %s\n", output)
//		if err != nil {
//			err = errors.Wrap(err, "exec configtxgen create channel artifacts")
//			return err
//		}
//		if !utils.FileExists(orgArtifactJsonName) {
//			err = errors.New("organization artifact json create failed")
//			return err
//		}
//	}
//	return nil
//}

// configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPAnchors.tx -channelID mychannel -asOrg Org1

// GenerateAnchorPeerArtifacts generate anchor peer artifacts
func (c ConfigtxgenTool) GenerateAnchorPeerArtifacts(chain *model.Fabric, baseDir string) error {
	configtxgen := filepath.Join(filepath.Dir(os.Args[0]), fmt.Sprintf("tool/%s/configtxgen", chain.Version))

	configtxPath, _ := filepath.Abs(baseDir)
	// 二进制识别的是 configtx.yaml 这个文件
	//configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))

	var om = make(map[string]struct{})

	// 多个org一个一个生成
	for _, peer := range chain.Peers {
		orgName := peer.Organization.UUID
		if len(orgName) == 0 {
			return errors.Errorf("peer [%s] organization info is empty", peer.Name)
		}
		if _, ok := om[orgName]; ok {
			continue
		}
		om[orgName] = struct{}{}
		for _, channel := range chain.Channels {
			anchorPeerArtifactTX, _ := filepath.Abs(fmt.Sprintf("%s/%s-%sMSPAnchors.tx", configtxPath, channel.UUID, orgName))
			command := fmt.Sprintf("%s -profile OrgsChannel -outputAnchorPeersUpdate %s --configPath=%s/ -channelID %s -asOrg %s",configtxgen, anchorPeerArtifactTX, configtxPath, channel.UUID, orgName)
			log.Debugf(c.ctx, "GenerateAnchorPeerArtifacts command [%s]", command)
			output, err := cmd.NewDefaultCMD(command, []string{}, cmd.WithTimeout(10)).Run()
			log.Debugf(c.ctx, "output : %s", output)
			if err != nil {
				return errors.Wrap(err, "exec configtxgen create anchor peer artifacts")
			}
			f, err := os.Stat(anchorPeerArtifactTX)
			if err != nil {
				return errors.Wrap(err, "check anchor peer artifact tx file")
			}
			if f.Size() < 20 {
				_ = os.Remove(anchorPeerArtifactTX)
				return errors.Errorf("peer artifacts tx [%s] size is %d less than 20", anchorPeerArtifactTX, f.Size())
			}
		}
	}
	return nil
}

// GenerateTargetAnchorPeerArtifacts generate target channel anchor peer artifacts
//func (c ConfigtxgenTool) GenerateTargetAnchorPeerArtifacts(chain *model.FabricChain, organizations []model.ChainOrganization, channelInfo *model.ChainChannel, baseDir string) error {
//	pwd, _ := os.Getwd()
//	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
//
//	configtxPath, _ := filepath.Abs(baseDir)
//	// 二进制识别的是 configtx.yaml 这个文件
//	//configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))
//
//	// 多个org一个一个生成
//	for _, org := range organizations {
//		if org.OrganizationType != model.PeerOrganization {
//			continue
//		}
//		anchorPeerArtifactTX, _ := filepath.Abs(fmt.Sprintf("%s/%s-%sMSPAnchors.tx", configtxPath, channelInfo.Name, org.UUID))
//		cmd := fmt.Sprintf("%s -profile OrgsChannel -outputAnchorPeersUpdate %s --configPath=%s/ -channelID %s -asOrg %s",
//			configtxgen, anchorPeerArtifactTX, configtxPath, channelInfo.UUID, org.UUID)
//		log.Debug(cmd)
//		output, err := util.CMD(10, cmd)
//		log.Debugf("output : %s\n", output)
//		if err != nil {
//			err = errors.Wrap(err, "exec configtxgen create anchor peer artifacts")
//			return err
//		}
//		if !utils.FileExists(anchorPeerArtifactTX) {
//			err = errors.New("anchor peer artifacts generate failed")
//			return err
//		}
//		file, err := ioutil.ReadFile(anchorPeerArtifactTX)
//		if err != nil {
//			err = errors.Wrap(err, "read peer artifacts")
//			return err
//		}
//		if len(file) < 20 {
//			err = errors.Errorf("peer artifacts tx size is %d less than 20", len(file))
//			_ = os.Remove(anchorPeerArtifactTX)
//			return err
//		}
//	}
//	return nil
//}
