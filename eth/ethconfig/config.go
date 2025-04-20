// Copyright 2021 The go-zenanet Authors
// This file is part of the go-zenanet library.
//
// The go-zenanet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-zenanet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-zenanet library. If not, see <http://www.gnu.org/licenses/>.

// Package ethconfig contains the configuration of the ETH and LES protocols.
package ethconfig

import (
	"errors"
	"math/big"
	"time"

	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/consensus"
	"github.com/zenanetwork/go-zenanet/consensus/beacon"
	"github.com/zenanetwork/go-zenanet/consensus/clique"
	"github.com/zenanetwork/go-zenanet/consensus/eirene"
	"github.com/zenanetwork/go-zenanet/consensus/eirene/contract"
	"github.com/zenanetwork/go-zenanet/consensus/eirene/heimdall" //nolint:typecheck
	"github.com/zenanetwork/go-zenanet/consensus/eirene/heimdall/span"
	"github.com/zenanetwork/go-zenanet/consensus/eirene/heimdallapp"
	"github.com/zenanetwork/go-zenanet/consensus/eirene/heimdallgrpc"
	"github.com/zenanetwork/go-zenanet/consensus/ethash"
	"github.com/zenanetwork/go-zenanet/core"
	"github.com/zenanetwork/go-zenanet/core/txpool/blobpool"
	"github.com/zenanetwork/go-zenanet/core/txpool/legacypool"
	"github.com/zenanetwork/go-zenanet/eth/downloader"
	"github.com/zenanetwork/go-zenanet/eth/gasprice"
	"github.com/zenanetwork/go-zenanet/ethdb"
	"github.com/zenanetwork/go-zenanet/internal/ethapi"
	"github.com/zenanetwork/go-zenanet/log"
	"github.com/zenanetwork/go-zenanet/miner"
	"github.com/zenanetwork/go-zenanet/params"
)

// FullNodeGPO contains default gasprice oracle settings for full node.
var FullNodeGPO = gasprice.Config{
	Blocks:           20,
	Percentile:       60,
	MaxHeaderHistory: 1024,
	MaxBlockHistory:  1024,
	MaxPrice:         gasprice.DefaultMaxPrice,
	IgnorePrice:      gasprice.DefaultIgnorePrice,
}

// Defaults contains default settings for use on the Zenanet main net.
var Defaults = Config{
	SyncMode:           downloader.FullSync,
	NetworkId:          0, // enable auto configuration of networkID == chainID
	TxLookupLimit:      2350000,
	TransactionHistory: 2350000,
	StateHistory:       params.FullImmutabilityThreshold,
	LightPeers:         100,
	DatabaseCache:      512,
	TrieCleanCache:     154,
	TrieDirtyCache:     256,
	TrieTimeout:        60 * time.Minute,
	SnapshotCache:      102,
	FilterLogCacheSize: 32,
	Miner:              miner.DefaultConfig,
	TxPool:             legacypool.DefaultConfig,
	BlobPool:           blobpool.DefaultConfig,
	RPCGasCap:          50000000,
	RPCEVMTimeout:      5 * time.Second,
	GPO:                FullNodeGPO,
	RPCTxFeeCap:        1, // 1 zen
}

//go:generate go run github.com/fjl/gencodec -type Config -formats toml -out gen_config.go

// Config contains configuration options for ETH and LES protocols.
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Zenanet main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Network ID separates blockchains on the peer-to-peer networking level. When left
	// zero, the chain ID is used as network ID.
	NetworkId uint64
	SyncMode  downloader.SyncMode

	// This can be set to list of enrtree:// URLs which will be queried for
	// nodes to connect to.
	EthDiscoveryURLs  []string
	SnapDiscoveryURLs []string

	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

	// Deprecated, use 'TransactionHistory' instead.
	TxLookupLimit      uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	TransactionHistory uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	StateHistory       uint64 `toml:",omitempty"` // The maximum number of blocks from head whose state histories are reserved.

	// State scheme represents the scheme used to store zenanet states and trie
	// nodes on top. It can be 'hash', 'path', or none which means use the scheme
	// consistent with persistent state.
	StateScheme string `toml:",omitempty"`

	// RequiredBlocks is a set of block number -> hash mappings which must be in the
	// canonical chain of all remote peers. Setting the option makes gzen verify the
	// presence of these blocks for every new peer connection.
	RequiredBlocks map[uint64]common.Hash `toml:"-"`

	// Light client options
	LightServ        int  `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightIngress     int  `toml:",omitempty"` // Incoming bandwidth limit for light servers
	LightEgress      int  `toml:",omitempty"` // Outgoing bandwidth limit for light servers
	LightPeers       int  `toml:",omitempty"` // Maximum number of LES client peers
	LightNoPrune     bool `toml:",omitempty"` // Whether to disable light chain pruning
	LightNoSyncServe bool `toml:",omitempty"` // Whether to serve light clients before syncing

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	// Database - LevelDB options
	LevelDbCompactionTableSize           uint64
	LevelDbCompactionTableSizeMultiplier float64
	LevelDbCompactionTotalSize           uint64
	LevelDbCompactionTotalSizeMultiplier float64

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration
	SnapshotCache  int
	Preimages      bool
	TriesInMemory  uint64

	// This is the number of blocks for which logs will be cached in the filter system.
	FilterLogCacheSize int

	// Mining options
	Miner miner.Config

	// Transaction pool options
	TxPool   legacypool.Config
	BlobPool blobpool.Config

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Enables prefetching trie nodes for read operations too
	EnableWitnessCollection bool `toml:"-"`

	// Enables VM tracing
	VMTrace           string
	VMTraceJsonConfig string

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap uint64

	// Maximum size (in bytes) a result of an rpc request could have
	RPCReturnDataLimit uint64

	// RPCEVMTimeout is the global timeout for eth-call.
	RPCEVMTimeout time.Duration

	// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	// send-transaction variants. The unit is zen.
	RPCTxFeeCap float64

	// OverrideCancun (TODO: remove after the fork)
	OverrideCancun *big.Int `toml:",omitempty"`

	// URL to connect to Heimdall node
	HeimdallURL string

	// No heimdall service
	WithoutHeimdall bool

	// Address to connect to Heimdall gRPC server
	HeimdallgRPCAddress string

	// Run heimdall service as a child process
	RunHeimdall bool

	// Arguments to pass to heimdall service
	RunHeimdallArgs string

	// Use child heimdall process to fetch data, Only works when RunHeimdall is true
	UseHeimdallApp bool

	// Zena logs flag
	ZenaLogs bool

	// Parallel EVM (Block-STM) related config
	ParallelEVM core.ParallelEVMConfig `toml:",omitempty"`

	// Develop Fake Author mode to produce blocks without authorisation
	DevFakeAuthor bool `hcl:"devfakeauthor,optional" toml:"devfakeauthor,optional"`

	// OverrideVerkle (TODO: remove after the fork)
	OverrideVerkle *big.Int `toml:",omitempty"`

	// EnableBlockTracking allows logging of information collected while tracking block lifecycle
	EnableBlockTracking bool
}

// CreateConsensusEngine creates a consensus engine for the given chain configuration.
func CreateConsensusEngine(chainConfig *params.ChainConfig, ethConfig *Config, db ethdb.Database, blockchainAPI *ethapi.BlockChainAPI) (consensus.Engine, error) {
	// nolint:nestif
	if chainConfig.Clique != nil {
		return beacon.New(clique.New(chainConfig.Clique, db)), nil
	} else if chainConfig.Zena != nil && chainConfig.Zena.ValidatorContract != "" {
		// If Zena eirene consensus is requested, set it up
		// In order to pass the zenanet transaction tests, we need to set the burn contract which is in the eirene config
		// Then, eirene != nil will also be enabled for ethash and clique. Only enable eirene for real if there is a validator contract present.
		genesisContractsClient := contract.NewGenesisContractsClient(chainConfig, chainConfig.Zena.ValidatorContract, chainConfig.Zena.StateReceiverContract, blockchainAPI)
		spanner := span.NewChainSpanner(blockchainAPI, contract.ValidatorSet(), chainConfig, common.HexToAddress(chainConfig.Zena.ValidatorContract))

		if ethConfig.WithoutHeimdall {
			return eirene.New(chainConfig, db, blockchainAPI, spanner, nil, genesisContractsClient, ethConfig.DevFakeAuthor), nil
		} else {
			if ethConfig.DevFakeAuthor {
				log.Warn("Sanitizing DevFakeAuthor", "Use DevFakeAuthor with", "--zena.withoutheimdall")
			}

			var heimdallClient eirene.IHeimdallClient
			if ethConfig.RunHeimdall && ethConfig.UseHeimdallApp {
				heimdallClient = heimdallapp.NewHeimdallAppClient()
			} else if ethConfig.HeimdallgRPCAddress != "" {
				heimdallClient = heimdallgrpc.NewHeimdallGRPCClient(ethConfig.HeimdallgRPCAddress)
			} else {
				heimdallClient = heimdall.NewHeimdallClient(ethConfig.HeimdallURL)
			}

			return eirene.New(chainConfig, db, blockchainAPI, spanner, heimdallClient, genesisContractsClient, false), nil
		}
	}
	// If defaulting to proof-of-work, enforce an already merged network since
	// we cannot run PoW algorithms anymore, so we cannot even follow a chain
	// not coordinated by a beacon node.
	if !chainConfig.TerminalTotalDifficultyPassed {
		return nil, errors.New("ethash is only supported as a historical component of already merged networks")
	}
	return beacon.New(ethash.NewFaker()), nil
}
