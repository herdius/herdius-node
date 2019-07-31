package main

import (
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	nlog "log"
	"os"
	"strings"

	"github.com/herdius/herdius-core/blockchain"
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
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()
var blockchainSvc *blockchain.Service
var voteCount = 0

// Flag to check if a new child block has arrived to validator
var isChildBlockReceivedByValidator = false

// Child block message object received
var mcb = &blockProtobuf.ChildBlockMessage{}

var nodeKey = "../../nodekey.json"

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

	opcode.RegisterMessageType(opcode.Opcode(1111), &blockProtobuf.ChildBlockMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1112), &blockProtobuf.ConnectionMessage{})
	opcode.RegisterMessageType(opcode.Opcode(1113), &protoplugin.BlockHeightRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1114), &protoplugin.BlockResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1115), &protoplugin.AccountRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1116), &protoplugin.AccountResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1117), &protoplugin.TxRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1118), &protoplugin.TxResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1119), &protoplugin.TxDetailRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1120), &protoplugin.TxDetailResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1121), &protoplugin.TxsByAddressRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1122), &protoplugin.TxsResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1123), &protoplugin.TxsByAssetAndAddressRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1124), &protoplugin.TxUpdateRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1125), &protoplugin.TxUpdateResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1126), &protoplugin.TxDeleteRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1127), &protoplugin.TxLockedRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1128), &protoplugin.TxLockedResponse{})
	opcode.RegisterMessageType(opcode.Opcode(1129), &protobuf.Ping{})
	opcode.RegisterMessageType(opcode.Opcode(1130), &protobuf.Pong{})
	opcode.RegisterMessageType(opcode.Opcode(1131), &protoplugin.TxRedeemRequest{})
	opcode.RegisterMessageType(opcode.Opcode(1132), &protoplugin.TxRedeemResponse{})

	builder := network.NewBuilderWithOptions(network.Address("tcp://" + selfip + ":" + port))
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
	go func() {
		for sig := range ctrl {
			fmt.Printf("Captured %v shutting down node", sig)
			net.Close()
			os.Exit(1)
		}
	}()

	<-ctrl
}
