// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package tlstransport

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/provider"
)

// Transport interface structure
type Transport struct {
	identity *provider.FullIdentity
}

// New returns a newly instantiated Transport Client
func New(identity *provider.FullIdentity) *Transport {
	return &Transport{identity: identity}
}

// DialNode using the authenticated mode
func (o *Transport) DialNode(ctx context.Context, node *pb.Node) (conn *grpc.ClientConn, err error) {
	defer mon.Task()(&ctx)(&err)

	if node.Address == nil || node.Address.Address == "" {
		return nil, Error.New("no address")
	}

	dialOpt, err := o.identity.DialOption()
	if err != nil {
		return nil, err
	}
	return grpc.Dial(node.Address.Address, dialOpt)
}

func (o *Transport) Listen(ctx context.Context, address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
