package network

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	network_pb "github.com/meshplus/bitxhub-kit/network/pb"
)

type ConnectCallback func(*peer.AddrInfo) error

type MessageHandler func(network.Stream, []byte)

type Network interface {
	// Start start the network service.
	Start() error

	// Stop stop the network service.
	Stop() error

	// Connect connects peer by ID.
	Connect(*peer.AddrInfo) error

	// Disconnect peer with id
	Disconnect(*peer.AddrInfo) error

	// SetConnectionCallback sets the callback after connecting
	SetConnectCallback(ConnectCallback)

	// SetMessageHandler sets message handler
	SetMessageHandler(MessageHandler)

	// AsyncSend sends message to peer with peer info.
	AsyncSend(*peer.AddrInfo, *network_pb.Message) error

	// Send message using existed stream
	SendWithStream(network.Stream, *network_pb.Message) error

	// Send sends message waiting response
	Send(*peer.AddrInfo, *network_pb.Message) (*network_pb.Message, error)

	// Broadcast message to all node
	Broadcast([]*peer.AddrInfo, *network_pb.Message) error

	// GetRemotePubKey gets remote public key
	GetRemotePubKey(id peer.ID) (crypto.PubKey, error)
}
