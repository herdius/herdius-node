package main

import (
	"context"
	"fmt"

	blockProtobuf "github.com/herdius/herdius-core/blockchain/protobuf"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-node/validator/service"
)

// HerdiusMessagePlugin will receive all transmitted messages.
type HerdiusMessagePlugin struct{ *network.Plugin }

func (state *HerdiusMessagePlugin) Receive(ctx *network.PluginContext) error {
	contex := network.WithSignMessage(context.Background(), true)

	switch msg := ctx.Message().(type) {

	case *blockProtobuf.ConnectionMessage:
		address := ctx.Client().ID.Address

		log.Info().Msgf("<%s> %s", address, msg.Message)

		sender, err := ctx.Network().Client(ctx.Client().Address)
		if err != nil {
			return fmt.Errorf("failed to get client network: %v", err)
		}
		nonce := 1
		err = sender.Reply(network.WithSignMessage(context.Background(), true), uint64(nonce),
			&blockProtobuf.ConnectionMessage{Message: "Connection established with Supervisor"})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to reply to client: %v", err))
		}
	case *blockProtobuf.ChildBlockMessage:
		mcb = msg
		//vote := mcb.GetVote()

		vService := service.Validator{}

		//Get all the transaction data included in the child block
		txsData := mcb.GetChildBlock().GetTxsData()
		if txsData == nil {
			fmt.Println("No txsData")
			return nil
		}
		txs := txsData.Tx

		//Get Root hash of the transactions
		cbRootHash := mcb.GetChildBlock().GetHeader().GetRootHash()
		err := vService.VerifyTxs(cbRootHash, txs)
		if err != nil {
			fmt.Println("Failed to verify transaction:", err)
			return nil
		}

		// Sign and vote the child block
		err = vService.Vote(ctx.Network(), ctx.Network().Address, mcb)
		if err != nil {
			ctx.Network().Broadcast(contex, &blockProtobuf.ConnectionMessage{Message: "Failed to get vote"})
		}

		ctx.Network().Broadcast(contex, mcb)
	}
	return nil
}
