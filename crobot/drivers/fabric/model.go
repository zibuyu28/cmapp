package fabric

type ConsensusType string

const (
	Solo  ConsensusType = "solo"
	Kakfa ConsensusType = "kafka"
	Raft  ConsensusType = "etcdraft"
)

type CertGenerateType string

const (
	CAServer       CertGenerateType = "ca-server"
	CertBinaryTool CertGenerateType = "bin-tool"
)

type Channel struct {
	Name                     string
	UUID                     string
	OrdererBatchTimeout      string
	OrdererMaxMessageCount   int
	OrdererAbsoluteMaxBytes  string
	OrdererPreferredMaxBytes string
}

type Orderer struct {
	Name         string
	UUID         string
	MachineID    int
	GRPCPort     int
	HealthPort   int
	Tag          string
	NodeHostName string
}

type Organization struct {
	Name string
	UUID string
}

type Peer struct {
	Name                string
	UUID                string
	MachineID           int
	GRPCPort            int
	ChainCodeListenPort int
	EventPort           int
	HealthPort          int
	Organization        Organization
	AnchorPeer          bool
	Tag                 string
	NodeHostName        string
}

// Fabric fabric chain info
type Fabric struct {
	Name        string
	UUID        string
	Version     string
	Consensus   ConsensusType
	CertGenType CertGenerateType
	Channels    []Channel
	TLSEnable   bool
	Orderers    []Orderer
	Peers       []Peer
}