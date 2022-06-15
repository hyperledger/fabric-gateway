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
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

type BlockAndPrivateDataEventListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	events <-chan *peer.BlockAndPrivateData
}

func NewBlockAndPrivateDataEventListener(parentCtx context.Context, network *client.Network, options ...client.BlockEventsOption) (*BlockAndPrivateDataEventListener, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	events, err := network.BlockAndPrivateDataEvents(ctx, options...)
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

type CheckpointBlockAndPrivateDataEventListener struct {
	listener   BlockAndPrivateDataEvents
	checkpoint func(*peer.BlockAndPrivateData)
}

func NewCheckpointBlockAndPrivateDataEventListener(listener BlockAndPrivateDataEvents, checkpoint func(*peer.BlockAndPrivateData)) *CheckpointBlockAndPrivateDataEventListener {
	checkpointListener := &CheckpointBlockAndPrivateDataEventListener{
		listener:   listener,
		checkpoint: checkpoint,
	}
	return checkpointListener
}

func (listener *CheckpointBlockAndPrivateDataEventListener) Event() (*peer.BlockAndPrivateData, error) {
	event, err := listener.listener.Event()
	if err != nil {
		return nil, err
	}

	listener.checkpoint(event)

	return event, nil
}

func (listener *CheckpointBlockAndPrivateDataEventListener) Close() {
	listener.listener.Close()
}
