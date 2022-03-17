/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type BlockAndPrivateDataEventListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	events <-chan *peer.BlockAndPrivateData
}

func NewBlockAndPrivateDataEventListener(parentCtx context.Context, network *client.Network, options ...client.BlockEventsOption) (*BlockAndPrivateDataEventListener, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	events, err := network.BlockEventsWithPrivateData(ctx, options...)
	if err != nil {
		cancel()
		return nil, err
	}

	listener := &BlockAndPrivateDataEventListener{
		ctx:    ctx,
		cancel: cancel,
		events: events,
	}
	return listener, nil
}

func (listener *BlockAndPrivateDataEventListener) Event() (*peer.BlockAndPrivateData, error) {
	select {
	case event, ok := <-listener.events:
		if !ok {
			return nil, fmt.Errorf("event channel closed")
		}
		return event, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for event")
	}
}

func (listener *BlockAndPrivateDataEventListener) Close() {
	listener.cancel()
}
