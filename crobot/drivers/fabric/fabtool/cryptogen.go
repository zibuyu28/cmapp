package fabtool

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"git.hyperchain.cn/blocface/chain-drivers-go/fabric_k8s/driver/model"
	util "git.hyperchain.cn/blocface/goutil"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/cmd"
	"github.com/zibuyu28/cmapp/crobot/drivers/fabric"
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
}

func newCryptoGen() RT {
	return CryptoGenTool{}
}

func (c CryptoGenTool) ExtendPeerCert(chain *model.FabricChain, nodes []model.Node, baseDir string) error {
	pwd, _ := os.Getwd()
	cryptogen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/cryptogen", chain.FabricVersion))
	randomCryptoFilePath, _ := filepath.Abs(fmt.Sprintf("%s/%d-addpeer.yaml", baseDir, time.Now().UnixNano()))
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
	err = ioutil.WriteFile(randomCryptoFilePath, configYaml, os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "build crypto config")
		return err
	}

	output, err := util.CMD(10,
		fmt.Sprintf("%s extend --config=%s --input=%s/", cryptogen, randomCryptoFilePath, certPath))
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec crypto extend cert")
		return err
	}

	for _, info := range chain.OrganizationInfos {
		if info.OrganizationType != model.PeerOrganization {
			continue
		}
		for _, node := range nodes {
			if node.OrgID != info.ID {
				continue
			}
			orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.fabric.com/peers/peer%d.%s.fabric.com", info.UUID, node.ID, info.UUID))
			if !util.FileExist(orgCertPath) {
				err = errors.Errorf("node(%s) cert generate err", node.NodeName)
				return err
			}
		}
	}

	return nil
}

func (c CryptoGenTool) GenerateNewOrgCert(chain *model.FabricChain, org *model.NewOrg, baseDir string) error {
	pwd, _ := os.Getwd()
	cryptogen := filepath.Join(pwd, "drivers", os.Args[3], fmt.Sprintf("tool/%s/cryptogen", chain.FabricVersion))
	randomCryptoFilePath, _ := filepath.Abs(fmt.Sprintf("%s/%d-neworg.yaml", baseDir, time.Now().UnixNano()))
	certPath, _ := filepath.Abs(baseDir)

	config := constructNewOrgConfig(org)
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		err = errors.Wrap(err, "marshal config")
		return err
	}

	// generate crypto config file
	err = ioutil.WriteFile(randomCryptoFilePath, configYaml, os.ModePerm)
	if err != nil {
		err = errors.Wrap(err, "build crypto config")
		return err
	}

	output, err := util.CMD(10,
		fmt.Sprintf("%s generate --config=%s --output=%s/", cryptogen, randomCryptoFilePath, certPath))
	fmt.Printf("output : %s\n", output)
	if err != nil {
		err = errors.Wrap(err, "exec crypto generate cert")
		return err
	}
	for _, info := range org.OrganizationInfos {
		if info.OrganizationType != model.PeerOrganization {
			continue
		}
		orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.fabric.com", info.UUID))
		if !util.FileExist(orgCertPath) {
			err = errors.Errorf("org(%s) cert generate err", info.Name)
			return err
		}
	}
	return nil
}

func (c CryptoGenTool) GenerateInitCert(driveruuid string, chain *fabric.Fabric, baseDir string) error {
	pwd, _ := os.Getwd()
	cryptogen := filepath.Join(pwd, "drivers", driveruuid, fmt.Sprintf("tool/%s/cryptogen", chain.FabricVersion))
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
	coutput, err := cmd.NewDefaultCMD(cryptogen, command).Run()
	fmt.Printf("output : %s\n", output)
	if err != nil {
		return errors.Wrap(err, "exec crypto generate cert")
	}
	for _, info := range chain.OrganizationInfos {
		if info.OrganizationType != model.PeerOrganization {
			continue
		}
		orgCertPath := filepath.Join(certPath, fmt.Sprintf("peerOrganizations/%s.fabric.com", info.UUID))
		if !util.FileExist(orgCertPath) {
			err = errors.Errorf("org(%s) cert generate err", info.Name)
			return err
		}
	}
	return nil
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
func constructInitConfig(chain *fabric.Fabric) *Config {
	// 构建orderer结构
	ordererOrgSpec := OrgSpec{
		Name:          "Orderer",
		Domain:        "orderer.fabric.com",
		EnableNodeOUs: true,
	}
	var ordererNodes []NodeSpec
	for _, orderer := range chain.Orderers {

		odns := NodeSpec{
			Hostname:   fmt.Sprintf("orderer%d", orderer.ID),
			CommonName: fmt.Sprintf("orderer%d.orderer.fabric.com", orderer.ID),
			SANS:       []string{orderer.NodeHostName},
		}

		ordererNodes = append(ordererNodes, odns)
	}

	ordererOrgSpec.Specs = ordererNodes

	// 构建peer组织结构
	var peerOrgSpecs []OrgSpec
	var peerOrgs = make(map[string][]fabric.Peer)
	for i, p := range chain.Peers {
		if _, ok := peerOrgs[p.Organization.Name]; ok {
			peerOrgs[p.Organization.Name] = append(peerOrgs[p.Organization.Name], chain.Peers[i])
		} else {
			peerOrgs[p.Organization.Name] = []fabric.Peer{chain.Peers[i]}
		}
	}

	for org, ps := range peerOrgs {
		peerOrgSpec := OrgSpec{
			Name:          org,
			Domain:        org + ".fabric.com",
			EnableNodeOUs: true,
		}
		var peerNodes []NodeSpec
		for _, peer := range ps {
			prns := NodeSpec{
				Hostname:   fmt.Sprintf("peer%d", peer.ID),
				CommonName: fmt.Sprintf("peer%d.%s.fabric.com", peer.ID, org),
				SANS:       []string{peer.NodeHostName},
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
