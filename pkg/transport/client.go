// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package transport

import (
	"context"

	"google.golang.org/grpc"

	"storj.io/storj/pkg/pb"
)

// Client defines the interface to an transport client.
type Client interface {
	DialNode(ctx context.Context, node *pb.Node) (*grpc.ClientConn, error)
}
