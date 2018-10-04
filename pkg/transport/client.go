// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package transport

import (
	"context"

	"github.com/zeebo/errs"
	"google.golang.org/grpc"

	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/provider"
	"storj.io/storj/pkg/transport/tlstransport"
)

var (
	Error = errs.Class("tlstransport error")
)

// Client defines the interface for dialing a node
type Client interface {
	DialNode(ctx context.Context, node *pb.Node) (*grpc.ClientConn, error)
}

type dialer struct {
	transports map[pb.NodeTransport]Client
}

// New creates an automatic dialer that selects the correct transport
func New(identity *provider.FullIdentity) Client {
	return &dialer{
		transports: map[pb.NodeTransport]Client{
			pb.NodeTransport_TCP: tlstransport.New(identity),
		},
	}
}

func (dialer *dialer) DialNode(ctx context.Context, node *pb.Node) (*grpc.ClientConn, error) {
	if node.Address == nil || node.Address.Address == "" {
		return nil, Error.New("no address")
	}

	transport, ok := dialer.transports[node.Address.Transport]
	if transport == nil || !ok {
		return nil, Error.New("unsupported transport %v", node.Address.Transport)
	}

	return transport.DialNode(ctx, node)
}
