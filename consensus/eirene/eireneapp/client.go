package eireneapp

import (
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/zenanetwork/go-zenanet/log"

	"github.com/maticnetwork/heimdall/app"
	"github.com/maticnetwork/heimdall/cmd/heimdalld/service"

	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	stateFetchLimit = 50
)

type HeimdallAppClient struct {
	hApp *app.HeimdallApp
}

func NewHeimdallAppClient() *HeimdallAppClient {
	return &HeimdallAppClient{
		hApp: service.GetHeimdallApp(),
	}
}

func (h *HeimdallAppClient) Close() {
	// Nothing to close as of now
	log.Warn("Shutdown detected, Closing Heimdall App conn")
}

func (h *HeimdallAppClient) NewContext() types.Context {
	return h.hApp.NewContext(true, abci.Header{Height: h.hApp.LastBlockHeight()})
}
