// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package node

import (
	"context"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"storj.io/storj/pkg/dht/mocks"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/provider"
)

var ctx = context.Background()

func TestLookup(t *testing.T) {
	cases := []struct {
		self        pb.Node
		to          pb.Node
		find        pb.Node
		expectedErr error
	}{
		{
			self:        pb.Node{Id: "hello", Address: &pb.NodeAddress{Address: ":7070"}},
			to:          pb.Node{Id: "hello", Address: &pb.NodeAddress{Address: ":8080"}},
			find:        pb.Node{Id: "hello", Address: &pb.NodeAddress{Address: ":9090"}},
			expectedErr: nil,
		},
	}

	for _, v := range cases {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)

		v.to = pb.Node{Id: NewNodeID(t), Address: &pb.NodeAddress{Address: lis.Addr().String()}}

		srv, mock, err := newTestServer(ctx)
		assert.NoError(t, err)
		go func() { assert.NoError(t, srv.Serve(lis)) }()
		defer srv.Stop()
		ctrl := gomock.NewController(t)

		mdht := mock_dht.NewMockDHT(ctrl)
		mrt := mock_dht.NewMockRoutingTable(ctrl)

		mdht.EXPECT().GetRoutingTable(gomock.Any()).Return(mrt, nil)
		mrt.EXPECT().ConnectionSuccess(gomock.Any()).Return(nil)

		ca, err := provider.NewCA(ctx, 12, 4)
		assert.NoError(t, err)
		identity, err := ca.NewIdentity()
		assert.NoError(t, err)

		nc, err := NewNodeClient(identity, v.self, mdht)
		assert.NoError(t, err)

		_, err = nc.Lookup(ctx, v.to, v.find)
		assert.Equal(t, v.expectedErr, err)
		assert.Equal(t, 1, mock.queryCalled)
	}
}

func newTestServer(ctx context.Context) (*grpc.Server, *mockNodeServer, error) {
	ca, err := provider.NewCA(ctx, 12, 4)
	if err != nil {
		return nil, nil, err
	}
	identity, err := ca.NewIdentity()
	if err != nil {
		return nil, nil, err
	}
	identOpt, err := identity.ServerOption()
	if err != nil {
		return nil, nil, err
	}

	grpcServer := grpc.NewServer(identOpt)
	mn := &mockNodeServer{queryCalled: 0}

	pb.RegisterNodesServer(grpcServer, mn)

	return grpcServer, mn, nil

}

type mockNodeServer struct {
	queryCalled int
}

func (mn *mockNodeServer) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	mn.queryCalled++
	return &pb.QueryResponse{}, nil
}

// NewNodeID returns the string representation of a dht node ID
func NewNodeID(t *testing.T) string {
	id, err := NewID()
	assert.NoError(t, err)

	return id.String()
}
