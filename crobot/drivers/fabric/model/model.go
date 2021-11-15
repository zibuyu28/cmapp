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

package model

import "github.com/zibuyu28/cmapp/core/pkg/ag"

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
	RemoteCert   string
	LogLevel     string
	APP          *ag.App
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
	RemoteCert          string
	LogLevel            string
	APP                 *ag.App
	CouchDB             *ag.App
	RMTDocker           string
}

// Fabric fabric chain info
type Fabric struct {
	Name               string
	UUID               string
	Version            string
	Consensus          ConsensusType
	CertGenType        CertGenerateType
	Channels           []Channel
	TLSEnable          bool
	RemoteGenesisBlock string
	Orderers           []Orderer
	Peers              []Peer
}
