package network

import (
	"github.com/meshplus/bitxhub-kit/network/pb"
)

func Message(data []byte) *pb.Message {
	return &pb.Message{
		Data: data,
	}
}
