// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package audit

import (
	"bytes"
	"context"
	"io"

	"github.com/vivint/infectious"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/pkg/dht"
	"storj.io/storj/pkg/kademlia"
	"storj.io/storj/pkg/overlay"
	"storj.io/storj/pkg/pb"
	"storj.io/storj/pkg/piecestore/rpc/client"
	"storj.io/storj/pkg/provider"
	"storj.io/storj/pkg/transport"
)

var mon = monkit.Package()

type share struct {
	Error       error
	PieceNumber int
	Data        []byte
}

// Auditor implements the downloader interface
type Auditor struct {
	downloader downloader
}

type downloader interface {
	DownloadShares(ctx context.Context, pointer *pb.Pointer, stripeIndex int) (shares []share, nodes []*pb.Node, err error)
}

// downloader implements the downloader interface
type defaultDownloader struct {
	transport transport.Client
	overlay   overlay.Client
	identity  provider.FullIdentity
}

// newDownloader creates a new instance of a defaultDownloader struct
func newDefaultDownloader(t transport.Client, o overlay.Client, id provider.FullIdentity) *defaultDownloader {
	return &defaultDownloader{transport: t, overlay: o, identity: id}
}

func (d *defaultDownloader) dial(ctx context.Context, node *pb.Node) (ps client.PSClient, err error) {
	defer mon.Task()(&ctx)(&err)
	c, err := d.transport.DialNode(ctx, node)
	if err != nil {
		return nil, err
	}
	return client.NewPSClient(c, 0, d.identity.Key)
}

// getShare use piece store clients to download shares from a given node
func (d *defaultDownloader) getShare(ctx context.Context, stripeIndex, shareSize, pieceNumber int,
	id client.PieceID, pieceSize int64, node *pb.Node) (s share, err error) {
	defer mon.Task()(&ctx)(&err)

	ps, err := d.dial(ctx, node)
	if err != nil {
		return s, err
	}

	derivedPieceID, err := id.Derive([]byte(node.GetId()))
	if err != nil {
		return s, err
	}

	rr, err := ps.Get(ctx, derivedPieceID, pieceSize, &pb.PayerBandwidthAllocation{})
	if err != nil {
		return s, err
	}

	offset := shareSize * stripeIndex

	rc, err := rr.Range(ctx, int64(offset), int64(shareSize))
	if err != nil {
		return s, err
	}

	buf := make([]byte, shareSize)
	_, err = io.ReadFull(rc, buf)
	if err != nil {
		return s, err
	}

	s = share{
		Error:       nil,
		PieceNumber: pieceNumber,
		Data:        buf,
	}
	return s, nil
}

func (d *defaultDownloader) DownloadShares(ctx context.Context, pointer *pb.Pointer,
	stripeIndex int) (shares []share, nodes []*pb.Node, err error) {
	defer mon.Task()(&ctx)(&err)
	var nodeIds []dht.NodeID
	pieces := pointer.Remote.GetRemotePieces()
	for _, p := range pieces {
		nodeIds = append(nodeIds, kademlia.StringToNodeID(p.GetNodeId()))
	}
	nodes, err = d.overlay.BulkLookup(ctx, nodeIds)
	if err != nil {
		return nil, nodes, err
	}

	shareSize := int(pointer.Remote.Redundancy.GetErasureShareSize())
	pieceID := client.PieceID(pointer.Remote.GetPieceId())

	// this downloads shares from nodes at the given stripe index
	for i, node := range nodes {
		paddedSize := calcPadded(pointer.GetSize(), shareSize)
		pieceSize := paddedSize / int64(pointer.Remote.Redundancy.GetMinReq())

		s, err := d.getShare(ctx, stripeIndex, shareSize, i, pieceID, pieceSize, node)
		if err != nil {
			// TODO(nat): update the statdb to indicate this node failed the audit
			s = share{
				Error:       err,
				PieceNumber: i,
				Data:        nil,
			}
		}
		shares = append(shares, s)
	}
	return shares, nodes, nil
}

// auditShares takes the downloaded shares and uses infectious's Correct function to check that they
// haven't been altered. auditShares returns a slice containing the piece numbers of altered shares.
func auditShares(ctx context.Context, required, total int, originals []share) (pieceNums []int, err error) {
	defer mon.Task()(&ctx)(&err)
	f, err := infectious.NewFEC(required, total)
	if err != nil {
		return nil, err
	}
	// Have to use []infectious.Share instead of []audit.Share
	// in order to run the infectious Correct function.
	copies := make([]infectious.Share, len(originals))
	for i, original := range originals {

		// If there was an error downloading a share before,
		// this line makes it so that there will be an empty
		// infectious.Share at the copies' index (same index
		// as in the original slice).
		if original.Error != nil {
			continue
		}

		copies[i].Data = append([]byte(nil), original.Data...)
		copies[i].Number = original.PieceNumber
	}

	err = f.Correct(copies)
	if err != nil {
		return nil, err
	}

	for i, share := range copies {
		if !bytes.Equal(originals[i].Data, share.Data) {
			pieceNums = append(pieceNums, share.Number)
		}
	}
	return pieceNums, nil
}

func calcPadded(size int64, blockSize int) int64 {
	mod := size % int64(blockSize)
	if mod == 0 {
		return size
	}
	return size + int64(blockSize) - mod
}

// auditStripe gets remote segments from a pointer and runs an audit on shares
// at a given stripe index
func (a *Auditor) auditStripe(ctx context.Context, pointer *pb.Pointer, stripeIndex, required, total int) (badNodes []*pb.Node, err error) {
	defer mon.Task()(&ctx)(&err)

	shares, nodes, err := a.downloader.DownloadShares(ctx, pointer, stripeIndex)
	if err != nil {
		return nil, err
	}

	pieceNums, err := auditShares(ctx, required, total, shares)
	if err != nil {
		return nil, err
	}
	for _, pieceNum := range pieceNums {
		badNodes = append(badNodes, nodes[pieceNum])
	}

	return badNodes, nil
}
