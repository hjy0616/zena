// Copyright 2023 The go-zenanet Authors
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

package catalyst

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/zenanetwork/go-zenanet/common"
	"github.com/zenanetwork/go-zenanet/core"
	"github.com/zenanetwork/go-zenanet/core/types"
	"github.com/zenanetwork/go-zenanet/crypto"
	"github.com/zenanetwork/go-zenanet/eth"
	"github.com/zenanetwork/go-zenanet/eth/downloader"
	"github.com/zenanetwork/go-zenanet/eth/ethconfig"
	"github.com/zenanetwork/go-zenanet/miner"
	"github.com/zenanetwork/go-zenanet/node"
	"github.com/zenanetwork/go-zenanet/p2p"
	"github.com/zenanetwork/go-zenanet/params"
)

func startSimulatedBeaconEthService(t *testing.T, genesis *core.Genesis) (*node.Node, *eth.Zenanet, *SimulatedBeacon) {
	t.Helper()

	n, err := node.New(&node.Config{
		P2P: p2p.Config{
			ListenAddr:  "127.0.0.1:8545",
			NoDiscovery: true,
			MaxPeers:    0,
		},
	})
	if err != nil {
		t.Fatal("can't create node:", err)
	}

	ethcfg := &ethconfig.Config{Genesis: genesis, SyncMode: downloader.FullSync, TrieTimeout: time.Minute, TrieDirtyCache: 256, TrieCleanCache: 256, Miner: miner.DefaultConfig}
	ethservice, err := eth.New(n, ethcfg)
	if err != nil {
		t.Fatal("can't create eth service:", err)
	}

	simBeacon, err := NewSimulatedBeacon(1, ethservice)
	if err != nil {
		t.Fatal("can't create simulated beacon:", err)
	}

	n.RegisterLifecycle(simBeacon)

	if err := n.Start(); err != nil {
		t.Fatal("can't start node:", err)
	}

	ethservice.SetSynced()
	return n, ethservice, simBeacon
}

// send 20 transactions, >10 withdrawals and ensure they are included in order
// send enough transactions to fill multiple blocks
func TestSimulatedBeaconSendWithdrawals(t *testing.T) {
	t.Skip()
	var withdrawals []types.Withdrawal
	txs := make(map[common.Hash]*types.Transaction)

	var (
		// testKey is a private key to use for funding a tester account.
		testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

		// testAddr is the Zenanet address of the tester account.
		testAddr = crypto.PubkeyToAddress(testKey.PublicKey)
	)

	// short period (1 second) for testing purposes
	var gasLimit uint64 = 10_000_000
	genesis := core.DeveloperGenesisBlock(gasLimit, &testAddr)
	node, ethService, mock := startSimulatedBeaconEthService(t, genesis)
	_ = mock
	defer node.Close()

	chainHeadCh := make(chan core.ChainHeadEvent, 10)
	subscription := ethService.BlockChain().SubscribeChainHeadEvent(chainHeadCh)
	defer subscription.Unsubscribe()

	// generate some withdrawals
	for i := 0; i < 20; i++ {
		withdrawals = append(withdrawals, types.Withdrawal{Index: uint64(i)})
		if err := mock.withdrawals.add(&withdrawals[i]); err != nil {
			t.Fatal("addWithdrawal failed", err)
		}
	}

	// generate a bunch of transactions
	signer := types.NewEIP155Signer(ethService.BlockChain().Config().ChainID)
	for i := 0; i < 20; i++ {
		tx, err := types.SignTx(types.NewTransaction(uint64(i), common.Address{}, big.NewInt(1000), params.TxGas, big.NewInt(params.InitialBaseFee), nil), signer, testKey)
		if err != nil {
			t.Fatalf("error signing transaction, err=%v", err)
		}
		txs[tx.Hash()] = tx

		if err := ethService.APIBackend.SendTx(context.Background(), tx); err != nil {
			t.Fatal("SendTx failed", err)
		}
	}

	includedTxs := make(map[common.Hash]struct{})
	var includedWithdrawals []uint64

	timer := time.NewTimer(12 * time.Second)
	for {
		select {
		case evt := <-chainHeadCh:
			for _, includedTx := range evt.Block.Transactions() {
				includedTxs[includedTx.Hash()] = struct{}{}
			}
			for _, includedWithdrawal := range evt.Block.Withdrawals() {
				includedWithdrawals = append(includedWithdrawals, includedWithdrawal.Index)
			}

			// ensure all withdrawals/txs included. this will take two blocks b/c number of withdrawals > 10
			if len(includedTxs) == len(txs) && len(includedWithdrawals) == len(withdrawals) && evt.Block.Number().Cmp(big.NewInt(2)) == 0 {
				return
			}
		case <-timer.C:
			t.Fatal("timed out without including all withdrawals/txs")
		}
	}
}
