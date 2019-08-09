package main

import (
	"flag"
	nlog "log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	amino "github.com/tendermint/go-amino"

	"github.com/herdius/herdius-core/blockchain/protobuf"
	blockProtobuf "github.com/herdius/herdius-core/blockchain/protobuf"
	cryptoAmino "github.com/herdius/herdius-core/crypto/encoding/amino"
	"github.com/herdius/herdius-core/hbi/message"
	protoplugin "github.com/herdius/herdius-core/hbi/protobuf"
	"github.com/herdius/herdius-core/p2p/crypto"
	keystore "github.com/herdius/herdius-core/p2p/key"
	"github.com/herdius/herdius-core/p2p/log"
	"github.com/herdius/herdius-core/p2p/network"
	"github.com/herdius/herdius-core/p2p/network/discovery"
	"github.com/herdius/herdius-core/p2p/types/opcode"
	"github.com/herdius/herdius-core/types"
)

var (
	cdc = amino.NewCodec()

	// Child block message object received
	mcb = &blockProtobuf.ChildBlockMessage{}

	nodeKey = "../../nodekey.json"
)

func init() {
	nlog.SetFlags(nlog.LstdFlags | nlog.Lshortfile)
	cryptoAmino.RegisterAmino(cdc)
}

func main() {
	// process other flags
	peersFlag := flag.String("peers", "", "peers to connect to")
	portFlag := flag.Int("port", 0, "port to bind validator to")
	selfIPFlag := flag.String("selfip", "127.0.0.1", "port to bind validator to")

	flag.Parse()

	port := *portFlag
	selfip := *selfIPFlag

	peers := []string{}
	if len(*peersFlag) == 0 {
		log.Fatal().Msg("no supervisor node address provided")
	}
	peers = strings.Split(*peersFlag, ",")

	nodekey, err := keystore.LoadOrGenNodeKey(nodeKey)
	if err != nil {
		log.Error().Msgf("Failed to create or load node key: %v", err)
	}
	privKey := nodekey.PrivKey
	pubKey := privKey.PubKey()
	keys := &crypto.KeyPair{
		PublicKey:  pubKey.Bytes(),
		PrivateKey: privKey.Bytes(),
		PrivKey:    privKey,
		PubKey:     pubKey,
	}

	opcode.RegisterMessageType(types.OpcodeChildBlockMessage, &blockProtobuf.ChildBlockMessage{})
	opcode.RegisterMessageType(types.OpcodeConnectionMessage, &blockProtobuf.ConnectionMessage{})
	opcode.RegisterMessageType(types.OpcodeBlockHeightRequest, &protoplugin.BlockHeightRequest{})
	opcode.RegisterMessageType(types.OpcodeBlockResponse, &protoplugin.BlockResponse{})
	opcode.RegisterMessageType(types.OpcodeAccountRequest, &protoplugin.AccountRequest{})
	opcode.RegisterMessageType(types.OpcodeAccountResponse, &protoplugin.AccountResponse{})
	opcode.RegisterMessageType(types.OpcodeTxRequest, &protoplugin.TxRequest{})
	opcode.RegisterMessageType(types.OpcodeTxResponse, &protoplugin.TxResponse{})
	opcode.RegisterMessageType(types.OpcodeTxDetailRequest, &protoplugin.TxDetailRequest{})
	opcode.RegisterMessageType(types.OpcodeTxDetailResponse, &protoplugin.TxDetailResponse{})
	opcode.RegisterMessageType(types.OpcodeTxsByAddressRequest, &protoplugin.TxsByAddressRequest{})
	opcode.RegisterMessageType(types.OpcodeTxsResponse, &protoplugin.TxsResponse{})
	opcode.RegisterMessageType(types.OpcodeTxsByAssetAndAddressRequest, &protoplugin.TxsByAssetAndAddressRequest{})
	opcode.RegisterMessageType(types.OpcodeTxUpdateRequest, &protoplugin.TxUpdateRequest{})
	opcode.RegisterMessageType(types.OpcodeTxUpdateResponse, &protoplugin.TxUpdateResponse{})
	opcode.RegisterMessageType(types.OpcodeTxDeleteRequest, &protoplugin.TxDeleteRequest{})
	opcode.RegisterMessageType(types.OpcodeTxLockedRequest, &protoplugin.TxLockedRequest{})
	opcode.RegisterMessageType(types.OpcodeTxLockedResponse, &protoplugin.TxLockedResponse{})
	opcode.RegisterMessageType(types.OpcodePing, &protobuf.Ping{})
	opcode.RegisterMessageType(types.OpcodePong, &protobuf.Pong{})
	opcode.RegisterMessageType(types.OpcodeTxRedeemRequest, &protoplugin.TxRedeemRequest{})
	opcode.RegisterMessageType(types.OpcodeTxRedeemResponse, &protoplugin.TxRedeemResponse{})
	opcode.RegisterMessageType(types.OpcodeTxsByBlockHeightRequest, &protoplugin.TxsByBlockHeightRequest{})
	opcode.RegisterMessageType(types.OpcodeLastBlockRequest, &protoplugin.LastBlockRequest{})

	builder := network.NewBuilderWithOptions(network.Address("tcp://" + selfip + ":" + string(port)))
	builder.SetKeys(keys)

	builder.SetAddress(network.FormatAddress("tcp", selfip, uint16(port)))

	// Register peer discovery plugin.
	builder.AddPlugin(new(discovery.Plugin))

	// Add custom Herdius plugin.
	builder.AddPlugin(new(HerdiusMessagePlugin))
	builder.AddPlugin(new(message.BlockMessagePlugin))
	builder.AddPlugin(new(message.AccountMessagePlugin))
	builder.AddPlugin(new(message.TransactionMessagePlugin))

	net, err := builder.Build()
	if err != nil {
		log.Fatal().Err(err)
		return
	}

	go net.Listen()
	defer net.Close()

	c := new(network.ConnTester)
	go func() {
		c.IsConnected(net, peers)
	}()

	ctrl := make(chan os.Signal, 1)
	signal.Notify(ctrl, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msgf("Captured %v shutting down node", <-ctrl)
}
