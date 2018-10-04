// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package transport

import (
	"context"
	"net"

	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/provider"
	"storj.io/storj/pkg/transport/tlstransport"
)

var (
	Error = errs.Class("tlstransport error")
)

// Transport defines the interface for dialing a node
type Transport interface {
	DialNode(ctx context.Context, node *pb.Node) (*grpc.ClientConn, error)
	Listen(ctx context.Context, address string) (net.Listener, error)
}

type dispatch struct {
	transports map[pb.NodeTransport]Transport
}

// New creates a dispatch that selects the correct transport
func New(identity *provider.FullIdentity) Transport {
	return &dispatch{
		transports: map[pb.NodeTransport]Transport{
			pb.NodeTransport_TCP_TLS_GRPC: tlstransport.New(identity),
		},
	}
}

func (dispatch *dispatch) DialNode(ctx context.Context, node *pb.Node) (*grpc.ClientConn, error) {
	if node.Address == nil || node.Address.Address == "" {
		return nil, Error.New("no address")
	}

	specific, ok := dispatch.transports[node.Address.Transport]
	if specific == nil || !ok {
		return nil, Error.New("unsupported transport %v", node.Address.Transport)
	}

	return specific.DialNode(ctx, node)
}

func (dispatch *dispatch) Listen(ctx context.Context, address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
