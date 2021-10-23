package fabtool

import (
	"fmt"
	"git.hyperchain.cn/blocface/chain-drivers-go/common/utils"
	"git.hyperchain.cn/blocface/chain-drivers-go/fabric_k8s/driver/service/cmd"
	log "git.hyperchain.cn/blocface/golog"
	util "git.hyperchain.cn/blocface/goutil"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type ConfigTXLatorTool struct {
}

func newConfigTXLator() RT {
	return ConfigTXLatorTool{}
}

type DecodeType string

type EncodeType string

const (
	BlockType        DecodeType = "common.Block"
	ConfigUpdateType DecodeType = "common.ConfigUpdate"
)

const (
	ConfigType   EncodeType = "common.Config"
	EnvelopeType EncodeType = "common.Envelope"
)

func (c ConfigTXLatorTool) ProtoDecodeToJson(chainVersion, pbFilePath, outputFilePath string, decodeType DecodeType) error {
	pwd, _ := os.Getwd()
	configtxlator := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxlator", chainVersion))

	if !util.FileExist(pbFilePath) {
		return errors.Errorf("pb file(%s) not exsit", pbFilePath)
	}
	dir := filepath.Dir(outputFilePath)
	exists, err := util.PathExists(dir)
	if err != nil || !exists {
		return errors.Errorf("output path(%s) not exsit", outputFilePath)
	}

	// configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json
	command := fmt.Sprintf("%s proto_decode --input %s --output %s --type %s", configtxlator, pbFilePath, outputFilePath, decodeType)
	log.Infof("command : %s", command)
	// run command
	output, err := cmd.NewDefaultCMD(command, []string{}).Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec configtxlator decode pb file")
		return err
	}
	if !utils.FileExists(outputFilePath) {
		err = errors.New("pb file decode failed")
		return err
	}
	return nil
}

func (c ConfigTXLatorTool) JsonEncodeToProto(chainVersion, jsonFilePath, outputFilePath string, encodeType EncodeType) error {
	pwd, _ := os.Getwd()
	configtxlator := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxlator", chainVersion))

	if !util.FileExist(jsonFilePath) {
		return errors.Errorf("json file(%s) not exsit", jsonFilePath)
	}
	dir := filepath.Dir(outputFilePath)
	exists, err := util.PathExists(dir)
	if err != nil || !exists {
		return errors.Errorf("output path(%s) not exsit", outputFilePath)
	}

	// configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json
	command := fmt.Sprintf("%s proto_encode --input %s --output %s --type %s", configtxlator, jsonFilePath, outputFilePath, encodeType)
	log.Infof("command : %s", command)
	// run command
	output, err := cmd.NewDefaultCMD(command, []string{}).Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec configtxlator encode json file")
		return err
	}
	if !utils.FileExists(outputFilePath) {
		err = errors.New("json file encode failed")
		return err
	}
	return nil
}

func (c ConfigTXLatorTool) ComputeUpdate(chainVersion, channelName, originPb, updatePb, outputPb string) error {

	pwd, _ := os.Getwd()
	configtxlator := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/configtxlator", chainVersion))

	if !util.FileExist(originPb) {
		return errors.Errorf("origin file(%s) not exsit", originPb)
	}
	if !util.FileExist(updatePb) {
		return errors.Errorf("update file(%s) not exsit", updatePb)
	}
	dir := filepath.Dir(outputPb)
	exists, err := util.PathExists(dir)
	if err != nil || !exists {
		return errors.Errorf("output path(%s) not exsit", outputPb)
	}

	// configtxlator compute_update --channel_id $CHANNEL_NAME --original config.pb --updated modified_config.pb --output org3_update.pb
	command := fmt.Sprintf("%s compute_update --channel_id %s --original %s --updated %s --output %s", configtxlator, channelName, originPb, updatePb, outputPb)
	log.Infof("command : %s", command)
	// run command
	output, err := cmd.NewDefaultCMD(command, []string{}).Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec configtxlator compute_update")
		return err
	}
	if !utils.FileExists(outputPb) {
		err = errors.New("update file encode failed")
		return err
	}
	return nil
}
