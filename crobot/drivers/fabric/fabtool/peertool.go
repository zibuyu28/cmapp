package fabtool

import (
	"fmt"
	"git.hyperchain.cn/blocface/chain-drivers-go/common/utils"
	"git.hyperchain.cn/blocface/chain-drivers-go/fabric_k8s/driver/model"
	"git.hyperchain.cn/blocface/chain-drivers-go/fabric_k8s/driver/service/cmd"
	log "git.hyperchain.cn/blocface/golog"
	util "git.hyperchain.cn/blocface/goutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type PeerTool struct {
}

func newPeer() RT {
	return PeerTool{}
}

type FetchType string

const (
	FetchConfig FetchType = "config"
	FetchBlock  FetchType = "block"
)

func (p PeerTool) CreateNewChannel(chain *model.FabricChain, channel model.ChainChannel, targetPeer *model.Node, baseDir string) error {
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Dir(peer)
	abs, err := filepath.Abs(baseDir)
	if err != nil {
		return errors.Wrap(err, "base dir abs")
	}

	coreyaml := filepath.Join(toolPath, "core.yaml")
	if !util.FileExist(coreyaml) {
		// 增加core.yaml 配置文件模板
		err := ioutil.WriteFile(coreyaml, []byte(coreConfig), os.ModePerm)
		if err != nil {
			err = errors.Wrap(err, "generate core.yaml config")
			return err
		}
	}

	// peer channel create -o localhost:7050  -c nine  -f ./channel-artifacts/tttx.tx --outputBlock ./channel-artifacts/channel1.block --tls --cafile /Users/wanghengfang/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
	envs, err := getEnv(chain, targetPeer, baseDir)
	if err != nil {
		err = errors.Wrap(err, "get env by peer")
		return err
	}
	if envs != nil {
		envs["FABRIC_CFG_PATH"] = toolPath
	}

	command := fmt.Sprintf("%s channel create -o %s:7050 -c %s -f %s/%s.tx --outputBlock %s/%s.block", peer, chain.Orderers[0].NodeHostName, channel.UUID,
		abs, channel.UUID, abs, channel.UUID)
	if chain.TLSEnabled {
		// ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
		command = fmt.Sprintf("%s --tls --cafile %s/ordererOrganizations/orderer.fabric.com/orderers/orderer%d.orderer.fabric.com/msp/tlscacerts/tlsca.orderer.fabric.com-cert.pem",
			command, abs, chain.Orderers[0].ID)
	}
	log.Debugf("exec command : %s", command)

	// 增加如上环境变量
	defaultCMD := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs))

	// 执行命令
	output, err := defaultCMD.Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec create channel command")
		return err
	}
	return nil
}

func (p PeerTool) JoinChannel(chain *model.FabricChain, targetPeer *model.Node, channelBlockPath, baseDir string) error {
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Dir(peer)

	if !util.FileExist(channelBlockPath) {
		return errors.Errorf("update pb file(%s) not found", channelBlockPath)
	}

	coreyaml := filepath.Join(toolPath, "core.yaml")
	if !util.FileExist(coreyaml) {
		// 增加core.yaml 配置文件模板
		err := ioutil.WriteFile(coreyaml, []byte(coreConfig), os.ModePerm)
		if err != nil {
			err = errors.Wrap(err, "generate core.yaml config")
			return err
		}
	}

	envs, err := getEnv(chain, targetPeer, baseDir)
	if err != nil {
		err = errors.Wrap(err, "get env by peer")
		return err
	}
	if envs != nil {
		envs["FABRIC_CFG_PATH"] = toolPath
	}

	command := fmt.Sprintf("%s channel join -b %s", peer, channelBlockPath)

	log.Infof("exec command : %s", command)

	// 增加如上环境变量
	defaultCMD := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs), cmd.WithRetry(10), cmd.WithTimeout(20))

	// 执行命令
	output, err := defaultCMD.Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec join channel command")
		return err
	}
	return nil
}

func (p PeerTool) UpdateAnchorPeer(chain *model.FabricChain, targetPeer *model.Node, updatePb, baseDir string) error {
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Dir(peer)
	abs, _ := filepath.Abs(baseDir)

	coreyaml := filepath.Join(toolPath, "core.yaml")
	if !util.FileExist(coreyaml) {
		// 增加core.yaml 配置文件模板
		err := ioutil.WriteFile(coreyaml, []byte(coreConfig), os.ModePerm)
		if err != nil {
			err = errors.Wrap(err, "generate core.yaml config")
			return err
		}
	}

	// 每个通道都更新
	for _, info := range chain.ChannelInfos {
		for _, org := range chain.OrganizationInfos {
			if org.OrganizationType != model.PeerOrganization {
				continue
			}

			anchorPeerArtifactTX, _ := filepath.Abs(fmt.Sprintf("%s/%s-%sMSPAnchors.tx", abs, info.Name, org.UUID))

			if !util.FileExist(anchorPeerArtifactTX) {
				return errors.Errorf("update anchor peer actiface file(%s) not found", anchorPeerArtifactTX)
			}

			envs, err := getEnv(chain, targetPeer, baseDir)
			if err != nil {
				err = errors.Wrap(err, "get env by peer")
				return err
			}
			if envs != nil {
				envs["FABRIC_CFG_PATH"] = toolPath
			}

			// peer channel update -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
			command := fmt.Sprintf("%s channel update -f %s -o %s:7050 -c %s", peer, anchorPeerArtifactTX, chain.Orderers[0].NodeHostName, info.Name)
			if chain.TLSEnabled {
				// ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
				command = fmt.Sprintf("%s --tls --cafile %s/ordererOrganizations/orderer.fabric.com/orderers/orderer%d.orderer.fabric.com/msp/tlscacerts/tlsca.orderer.fabric.com-cert.pem",
					command, abs, chain.Orderers[0].ID)
			}
			log.Debugf("exec command : %s", command)

			// 增加如上环境变量
			defaultCMD := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs))

			// 执行命令
			output, err := defaultCMD.Run()
			fmt.Printf("output : %s\n", output)
			if err != nil {
				err = errors.Wrap(err, "exec update channel command")
				return err
			}
		}
	}
	return nil
}

func (p PeerTool) UpdateChannelConfig(chain *model.FabricChain, targetPeer *model.Node, channelName, updatePb, baseDir string) error {
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Dir(peer)
	abs, _ := filepath.Abs(baseDir)

	if len(channelName) == 0 {
		return errors.Errorf("channel id(%s) not found", channelName)
	}

	if !util.FileExist(updatePb) {
		return errors.Errorf("update pb file(%s) not found", updatePb)
	}
	coreyaml := filepath.Join(toolPath, "core.yaml")
	if !util.FileExist(coreyaml) {
		// 增加core.yaml 配置文件模板
		err := ioutil.WriteFile(coreyaml, []byte(coreConfig), os.ModePerm)
		if err != nil {
			err = errors.Wrap(err, "generate core.yaml config")
			return err
		}
	}

	envs, err := getEnv(chain, targetPeer, baseDir)
	if err != nil {
		err = errors.Wrap(err, "get env by peer")
		return err
	}
	if envs != nil {
		envs["FABRIC_CFG_PATH"] = toolPath
	}

	// peer channel update -f org3_update_in_envelope.pb -c $CHANNEL_NAME -o orderer.example.com:7050 --tls --cafile $ORDERER_CA
	command := fmt.Sprintf("%s channel update -f %s -o %s:7050 -c %s", peer, updatePb, chain.Orderers[0].NodeHostName, channelName)
	if chain.TLSEnabled {
		// ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
		command = fmt.Sprintf("%s --tls --cafile %s/ordererOrganizations/orderer.fabric.com/orderers/orderer%d.orderer.fabric.com/msp/tlscacerts/tlsca.orderer.fabric.com-cert.pem",
			command, abs, chain.Orderers[0].ID)
	}

	log.Infof("exec command : %s", command)

	// 增加如上环境变量
	defaultCMD := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs))

	// 执行命令
	output, err := defaultCMD.Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec update channel command")
		return err
	}
	return nil
}

func (p PeerTool) SignConfigTx(chain *model.FabricChain, targetPeer *model.Node, updatePb, baseDir string) error {
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/", chain.FabricVersion))

	if !util.FileExist(updatePb) {
		return errors.Errorf("update pb file(%s) not found", updatePb)
	}
	coreyaml := filepath.Join(toolPath, "core.yaml")
	if !util.FileExist(coreyaml) {
		// 增加core.yaml 配置文件模板
		err := ioutil.WriteFile(coreyaml, []byte(coreConfig), os.ModePerm)
		if err != nil {
			err = errors.Wrap(err, "generate core.yaml config")
			return err
		}
	}
	envs, err := getEnv(chain, targetPeer, baseDir)
	if err != nil {
		err = errors.Wrap(err, "get env by peer")
		return err
	}
	if envs != nil {
		envs["FABRIC_CFG_PATH"] = toolPath
	}
	// peer channel signconfigtx -f org3_update_in_envelope.pb
	command := fmt.Sprintf("%s channel signconfigtx -f %s", peer, updatePb)
	// 增加如上环境变量
	cmdWithEnv := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs))

	log.Infof("exec command : %s", command)
	output, err := cmdWithEnv.Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec sign config tx command")
		return err
	}
	return nil
}

func (p PeerTool) FetchChannelConfig(chain *model.FabricChain, channelName string, anchorPeer *model.Node, outputFile, baseDir string, fetchType FetchType) error {
	abs, _ := filepath.Abs(baseDir)
	pwd, _ := os.Getwd()
	peer := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/peer", chain.FabricVersion))
	toolPath := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/", chain.FabricVersion))

	dir := filepath.Dir(outputFile)
	exists, err := util.PathExists(dir)
	if err != nil || !exists {
		return errors.Errorf("output path(%s) not exsit", outputFile)
	}

	if len(channelName) == 0 {
		return errors.New("channel name is empty")
	}

	if len(chain.Orderers) == 0 {
		return errors.New("orderer info length is empty")
	}
	if len(chain.Peers) == 0 {
		return errors.New("peer info length is empty")
	}

	// 增加core.yaml 配置文件模板
	err = ioutil.WriteFile(filepath.Join(toolPath, "core.yaml"), []byte(coreConfig), os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "generate core.yaml config")
		return err
	}

	envs, err := getEnv(chain, anchorPeer, baseDir)
	if err != nil {
		err = errors.Wrap(err, "get env by peer")
		return err
	}

	if envs != nil {
		envs["FABRIC_CFG_PATH"] = toolPath
	}

	var command string
	if fetchType == FetchConfig {
		command = fmt.Sprintf("%s channel fetch config %s -o %s:7050 -c %s", peer, outputFile, chain.Orderers[0].NodeHostName, channelName)
	} else if fetchType == FetchBlock {
		command = fmt.Sprintf("%s channel fetch 0 %s -o %s:7050 -c %s", peer, outputFile, chain.Orderers[0].NodeHostName, channelName)
	}

	if chain.TLSEnabled {
		// ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
		command = fmt.Sprintf("%s --tls --cafile %s/ordererOrganizations/orderer.fabric.com/orderers/orderer%d.orderer.fabric.com/msp/tlscacerts/tlsca.orderer.fabric.com-cert.pem",
			command, abs, chain.Orderers[0].ID)
	}

	log.Infof("exec command : %s", command)

	// 增加如上环境变量
	defaultCMD := cmd.NewDefaultCMD(command, []string{}, cmd.WithEnvs(envs), cmd.WithRetry(3))

	// 执行如下命令
	// peer channel fetch config config_block.pb -o orderer.example.com:7050 -c mychannel --tls --cafile /Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
	output, err := defaultCMD.Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec feat channel command")
		return err
	}

	if !utils.FileExists(outputFile) {
		err = errors.New("channel config fetch failed")
		return err
	}
	return nil
}

func getEnv(chain *model.FabricChain, node *model.Node, baseDir string) (map[string]string, error) {

	//export FABRIC_CFG_PATH=/Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/config
	//export CORE_PEER_ID=Org2cli
	//export CORE_PEER_ADDRESS=127.0.0.1:9051
	//export CORE_PEER_LOCALMSPID=Org2MSP
	//export CORE_PEER_TLS_ENABLED=true
	//export CORE_PEER_TLS_CERT_FILE=/Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.crt
	//export CORE_PEER_TLS_KEY_FILE=/Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/server.key
	//export CORE_PEER_TLS_ROOTCERT_FILE=/Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
	//export CORE_PEER_MSPCONFIGPATH=/Users/wujiajia/go/src/github.com/hyperledger/fabric-samples/first-network/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp

	if node.NodeType != model.Peer {
		return nil, errors.Errorf("node type node(%s) must be peer", node.NodeType)
	}
	abs, _ := filepath.Abs(baseDir)
	var org = ""
	for _, info := range chain.OrganizationInfos {
		if info.OrganizationType != model.PeerOrganization {
			continue
		}
		if node.OrgID == info.ID {
			org = info.UUID
			break
		}
	}
	if len(org) == 0 {
		return nil, errors.Errorf("node orgid(%d) not found in chain", node.OrgID)
	}
	var envs = make(map[string]string, 0)
	envs["CORE_PEER_ID"] = fmt.Sprintf("%scli", org)
	envs["CORE_PEER_ADDRESS"] = fmt.Sprintf("%s:7051", node.NodeHostName)
	envs["CORE_PEER_LOCALMSPID"] = fmt.Sprintf("%sMSP", org)
	if chain.TLSEnabled {
		envs["CORE_PEER_TLS_ENABLED"] = "true"
		// org2.example.com/peers/peer0.org2.example.com/tls/server.crt
		tlsPath := filepath.Join(abs,
			fmt.Sprintf("peerOrganizations/%s.fabric.com/peers", org),
			fmt.Sprintf("peer%d.%s.fabric.com/tls/", node.ID, org))
		envs["CORE_PEER_TLS_CERT_FILE"] = filepath.Join(tlsPath, "signcerts/cert.pem")
		envs["CORE_PEER_TLS_KEY_FILE"] = filepath.Join(tlsPath, "keystore/key.pem")
		envs["CORE_PEER_TLS_ROOTCERT_FILE"] = filepath.Join(tlsPath, "tlscacerts/tls.pem")
		// peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
		envs["CORE_PEER_MSPCONFIGPATH"] = filepath.Join(abs,
			fmt.Sprintf("peerOrganizations/%s.fabric.com/users", org),
			fmt.Sprintf("Admin@%s.fabric.com/msp/", org))
	} else {
		envs["CORE_PEER_TLS_ENABLED"] = "false"
	}

	return envs, nil
}

var coreConfig = `peer:
    id: jdoe
    networkId: dev
    listenAddress: 0.0.0.0:7051
    address: 0.0.0.0:7051
    addressAutoDetect: false

    gomaxprocs: -1

    keepalive:
        minInterval: 60s
        client:
            interval: 60s
            timeout: 20s
        deliveryClient:
            interval: 60s
            timeout: 20s
    gossip:
        bootstrap: 127.0.0.1:7051
        useLeaderElection: true
        orgLeader: false
        membershipTrackerInterval: 5s
        endpoint:
        maxBlockCountToStore: 100
        maxPropagationBurstLatency: 10ms
        maxPropagationBurstSize: 10
        propagateIterations: 1
        propagatePeerNum: 3
        pullInterval: 4s
        pullPeerNum: 3
        requestStateInfoInterval: 4s
        publishStateInfoInterval: 4s
        stateInfoRetentionInterval:
        publishCertPeriod: 10s
        skipBlockVerification: false
        dialTimeout: 3s
        connTimeout: 2s
        recvBuffSize: 20
        sendBuffSize: 200
        digestWaitTime: 1s
        requestWaitTime: 1500ms
        responseWaitTime: 2s
        aliveTimeInterval: 5s
        aliveExpirationTimeout: 25s
        reconnectInterval: 25s
        externalEndpoint:
        election:
            startupGracePeriod: 15s
            membershipSampleInterval: 1s
            leaderAliveThreshold: 10s
            leaderElectionDuration: 5s

        pvtData:
            pullRetryThreshold: 60s
            transientstoreMaxBlockRetention: 1000
            pushAckTimeout: 3s
            btlPullMargin: 10
            reconcileBatchSize: 10
            reconcileSleepInterval: 1m
            reconciliationEnabled: true
        state:
            enabled: true
            checkInterval: 10s
            responseTimeout: 3s
            batchSize: 10
            blockBufferSize: 100
            maxRetries: 3
    tls:
        enabled:  false
        clientAuthRequired: false
        cert:
            file: tls/server.crt
        key:
            file: tls/server.key
        rootcert:
            file: tls/ca.crt
        clientRootCAs:
            files:
              - tls/ca.crt
        clientKey:
            file:
        clientCert:
            file:
    authentication:
        timewindow: 15m
    fileSystemPath: /var/hyperledger/production
    BCCSP:
        Default: SW
        SW:
            Hash: SHA2
            Security: 256
            FileKeyStore:
                KeyStore:
        PKCS11:
            Library:
            Label:
            Pin:
            Hash:
            Security:
            FileKeyStore:
                KeyStore:
    mspConfigPath: msp
    localMspId: SampleOrg
    client:
        connTimeout: 3s
    deliveryclient:
        reconnectTotalTimeThreshold: 3600s
        connTimeout: 3s
        reConnectBackoffThreshold: 3600s
    localMspType: bccsp
    profile:
        enabled:     false
        listenAddress: 0.0.0.0:6060
    adminService:
    handlers:
        authFilters:
          -
            name: DefaultAuth
          -
            name: ExpirationCheck    # This filter checks identity x509 certificate expiration
        decorators:
          -
            name: DefaultDecorator
        endorsers:
          escc:
            name: DefaultEndorsement
            library:
        validators:
          vscc:
            name: DefaultValidation
            library:
    validatorPoolSize:
    discovery:
        enabled: true
        authCacheEnabled: true
        authCacheMaxSize: 1000
        authCachePurgeRetentionRatio: 0.75
        orgMembersAllowedAccess: false
vm:
    endpoint: unix:///var/run/docker.sock
    docker:
        tls:
            enabled: false
            ca:
                file: docker/ca.crt
            cert:
                file: docker/tls.crt
            key:
                file: docker/tls.key
        attachStdout: false
        hostConfig:
            NetworkMode: host
            Dns:
            LogConfig:
                Type: json-file
                Config:
                    max-size: "50m"
                    max-file: "5"
            Memory: 2147483648
chaincode:
    id:
        path:
        name:
    builder: $(DOCKER_NS)/fabric-ccenv:latest
    pull: false

    golang:
        runtime: $(BASE_DOCKER_NS)/fabric-baseos:$(ARCH)-$(BASE_VERSION)
        dynamicLink: false

    car:
        runtime: $(BASE_DOCKER_NS)/fabric-baseos:$(ARCH)-$(BASE_VERSION)

    java:
        runtime: $(DOCKER_NS)/fabric-javaenv:$(ARCH)-$(PROJECT_VERSION)

    node:
        runtime: $(BASE_DOCKER_NS)/fabric-baseimage:$(ARCH)-$(BASE_VERSION)
    startuptimeout: 300s
    executetimeout: 30s
    mode: net
    keepalive: 0
    system:
        cscc: enable
        lscc: enable
        escc: enable
        vscc: enable
        qscc: enable
    systemPlugins:
    logging:
      level:  info
      shim:   warning
      format: '%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}'
ledger:
  blockchain:
  state:
    stateDatabase: goleveldb
    totalQueryLimit: 100000
    couchDBConfig:
       couchDBAddress: 127.0.0.1:5984
       username:
       password:
       maxRetries: 3
       maxRetriesOnStartup: 12
       requestTimeout: 35s
       internalQueryLimit: 1000
       maxBatchUpdateSize: 1000
       warmIndexesAfterNBlocks: 1
       createGlobalChangesDB: false

  history:
    enableHistoryDatabase: true
operations:
    listenAddress: 127.0.0.1:9443
    tls:
        enabled: false
        cert:
            file:
        key:
            file:
        clientAuthRequired: false
        clientRootCAs:
            files: []
metrics:
    provider: disabled
    statsd:
        network: udp
        address: 127.0.0.1:8125
        writeInterval: 10s
        prefix:

`