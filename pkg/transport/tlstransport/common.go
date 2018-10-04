// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package tlstransport

import (
	"github.com/zeebo/errs"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"
)

var (
	mon = monkit.Package()
	//Error is the errs class of standard Transport Client errors
	Error = errs.Class("tlstransport error")
)
