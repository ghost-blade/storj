// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package checker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"storj.io/storj/pkg/datarepair"
	"storj.io/storj/pkg/provider"
)

// Config contains configurable values for checker
type Config struct {
	QueueAddress string `help:"data repair queue address" default:"redis://localhost:6379?db=5&password=123"`
	Interval time.Duration `help:"how frequently checker should audit segments" default:"30s"`
}

// Run runs the checker with configured values
func (c Config) Run(ctx context.Context, server *provider.Provider) (err error) {
	defer datarepair.Mon.Task()(&ctx)(&err)

	zap.S().Info("Checker is starting up")

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		for {
			select {
			case <-ticker.C:
				zap.S().Info("Starting segment checker service")
				go func() {
					//get queue
					//get pointerdb
					//get overlay
					// c := NewChecker(params, pointerdb, repairQueue, overlay, logger) 
					// err := c.IdentifyInjuredSegments(ctx) 
				}()

			case <-ctx.Done():
				return
			}
		}
	}()

	return server.Run(ctx)
}
