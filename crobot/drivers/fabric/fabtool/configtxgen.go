package fabtool

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"git.hyperchain.cn/blocface/chain-drivers-go/common/utils"
	"git.hyperchain.cn/blocface/chain-drivers-go/fabric_k8s/driver/model"
	log "git.hyperchain.cn/blocface/golog"
	util "git.hyperchain.cn/blocface/goutil"
	"github.com/pkg/errors"
)

type ConfigtxgenTool struct {
}

func newConfigTXGen() RT {
	return ConfigtxgenTool{}
}

func (c ConfigtxgenTool) GenerateTargetChannelTx(chain *model.FabricChain, ch model.ChainChannel, baseDir string) error {
	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
	configtxPath, _ := filepath.Abs(baseDir)
	channeltxPath := filepath.Join(configtxPath, fmt.Sprintf("%s.tx", ch.UUID))

	output, err := util.CMD(10,
		fmt.Sprintf("%s -profile OrgsChannel --configPath=%s/ -outputCreateChannelTx %s -channelID %s",
			configtxgen, configtxPath, channeltxPath, ch.UUID))
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec configtxgen create channel tx")
		return err
	}
	if !utils.FileExists(channeltxPath) {
		err = errors.New("channel tx create failed")
		return err
	}
	return nil
}

func (c ConfigtxgenTool) GenerateAllChainChannelTx(chain *model.FabricChain, baseDir string) error {
	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
	configtxPath, _ := filepath.Abs(baseDir)
	for _, ch := range chain.ChannelInfos {
		channeltxPath := filepath.Join(configtxPath, fmt.Sprintf("%s.tx", ch.Name))
		output, err := util.CMD(10,
			fmt.Sprintf("%s -profile OrgsChannel --configPath=%s/ -outputCreateChannelTx %s -channelID %s",
				configtxgen, configtxPath, channeltxPath, ch.UUID))
		fmt.Printf("output : %s\n", output)
		if err != nil {
			err = errors.Wrap(err, "exec configtxgen create channel tx")
			return err
		}
		if !utils.FileExists(channeltxPath) {
			err = errors.New("channel tx create failed")
			return err
		}
	}
	return nil
}

func (c ConfigtxgenTool) GenerateGenesisBlock(chain *model.FabricChain, baseDir string) error {

	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))
	configtxPath, _ := filepath.Abs(baseDir)
	genesisBlockPath := filepath.Join(configtxPath, "orderer.genesis.block")
	output, err := util.CMD(10, fmt.Sprintf("%s -profile OrdererGenesis --configPath=%s/ -outputBlock %s", configtxgen,
		configtxPath, genesisBlockPath))
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec configtxgen create genesis block")
		return err
	}
	if !utils.FileExists(genesisBlockPath) {
		err = errors.New("orderer genesis block create failed")
		return err
	}
	return nil
}

func (c ConfigtxgenTool) GenerateChannelArtifacts(chain *model.FabricChain, neworg *model.NewOrg, baseDir string) error {
	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))

	// 二进制识别的是 configtx.yaml 这个文件
	configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))

	// 多个org一个一个生成
	for _, org := range neworg.OrganizationInfos {
		if org.OrganizationType != model.PeerOrganization {
			continue
		}
		orgArtifactJsonName, _ := filepath.Abs(fmt.Sprintf("%s/%s.json", configtxPath, org.UUID))
		cmd := fmt.Sprintf("%s -printOrg %s --configPath=%s/ > %s", configtxgen, org.UUID,
			configtxPath, orgArtifactJsonName)
		log.Debug(cmd)
		output, err := util.CMD(10, cmd)
		fmt.Printf("output : %s\n", output)
		if err != nil {
			err = errors.Wrap(err, "exec configtxgen create channel artifacts")
			return err
		}
		if !utils.FileExists(orgArtifactJsonName) {
			err = errors.New("organization artifact json create failed")
			return err
		}
	}
	return nil
}

// configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPAnchors.tx -channelID mychannel -asOrg Org1

// GenerateAnchorPeerArtifacts generate anchor peer artifacts
func (c ConfigtxgenTool) GenerateAnchorPeerArtifacts(chain *model.FabricChain, baseDir string) error {
	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))

	configtxPath, _ := filepath.Abs(baseDir)
	// 二进制识别的是 configtx.yaml 这个文件
	//configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))

	// 多个org一个一个生成
	for _, org := range chain.OrganizationInfos {
		if org.OrganizationType != model.PeerOrganization {
			continue
		}
		for _, channelInfo := range chain.ChannelInfos {
			anchorPeerArtifactTX, _ := filepath.Abs(fmt.Sprintf("%s/%s-%sMSPAnchors.tx", configtxPath, channelInfo.Name, org.UUID))
			cmd := fmt.Sprintf("%s -profile OrgsChannel -outputAnchorPeersUpdate %s --configPath=%s/ -channelID %s -asOrg %s",
				configtxgen, anchorPeerArtifactTX, configtxPath, channelInfo.UUID, org.UUID)
			log.Debug(cmd)
			output, err := util.CMD(10, cmd)
			log.Debugf("output : %s\n", output)
			if err != nil {
				err = errors.Wrap(err, "exec configtxgen create anchor peer artifacts")
				return err
			}
			if !utils.FileExists(anchorPeerArtifactTX) {
				err = errors.New("anchor peer artifacts generate failed")
				return err
			}
			file, err := ioutil.ReadFile(anchorPeerArtifactTX)
			if err != nil {
				err = errors.Wrap(err, "read peer artifacts")
				return err
			}
			if len(file) < 20 {
				err = errors.Errorf("peer artifacts tx size is %d less than 20", len(file))
				_ = os.Remove(anchorPeerArtifactTX)
				return err
			}
		}
	}
	return nil
}

// GenerateTargetAnchorPeerArtifacts generate target channel anchor peer artifacts
func (c ConfigtxgenTool) GenerateTargetAnchorPeerArtifacts(chain *model.FabricChain, organizations []model.ChainOrganization, channelInfo *model.ChainChannel, baseDir string) error {
	pwd, _ := os.Getwd()
	configtxgen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxgen", chain.FabricVersion))

	configtxPath, _ := filepath.Abs(baseDir)
	// 二进制识别的是 configtx.yaml 这个文件
	//configtxPath, _ := filepath.Abs(fmt.Sprintf("%s/blc%d-ch%d", baseDir, neworg.ChainID, neworg.ChannelID))

	// 多个org一个一个生成
	for _, org := range organizations {
		if org.OrganizationType != model.PeerOrganization {
			continue
		}
		anchorPeerArtifactTX, _ := filepath.Abs(fmt.Sprintf("%s/%s-%sMSPAnchors.tx", configtxPath, channelInfo.Name, org.UUID))
		cmd := fmt.Sprintf("%s -profile OrgsChannel -outputAnchorPeersUpdate %s --configPath=%s/ -channelID %s -asOrg %s",
			configtxgen, anchorPeerArtifactTX, configtxPath, channelInfo.UUID, org.UUID)
		log.Debug(cmd)
		output, err := util.CMD(10, cmd)
		log.Debugf("output : %s\n", output)
		if err != nil {
			err = errors.Wrap(err, "exec configtxgen create anchor peer artifacts")
			return err
		}
		if !utils.FileExists(anchorPeerArtifactTX) {
			err = errors.New("anchor peer artifacts generate failed")
			return err
		}
		file, err := ioutil.ReadFile(anchorPeerArtifactTX)
		if err != nil {
			err = errors.Wrap(err, "read peer artifacts")
			return err
		}
		if len(file) < 20 {
			err = errors.Errorf("peer artifacts tx size is %d less than 20", len(file))
			_ = os.Remove(anchorPeerArtifactTX)
			return err
		}
	}
	return nil
}
