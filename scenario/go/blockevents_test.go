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
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
)

type BlockEventListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	events <-chan *common.Block
}

func NewBlockEventListener(parentCtx context.Context, network *client.Network, options ...client.BlockEventsOption) (*BlockEventListener, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	events, err := network.BlockEvents(ctx, options...)
	if err != nil {
		cancel()
		return nil, err
	}

	listener := &BlockEventListener{
		ctx:    ctx,
		cancel: cancel,
		events: events,
	}
	return listener, nil
}

func (listener *BlockEventListener) Event() (*common.Block, error) {
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

func (listener *BlockEventListener) Close() {
	listener.cancel()
}

type CheckpointBlockEventListener struct {
	listener   BlockEvents
	checkpoint func(*common.Block)
}

func NewCheckpointBlockEventListener(listener BlockEvents, checkpoint func(*common.Block)) *CheckpointBlockEventListener {
	checkpointListener := &CheckpointBlockEventListener{
		listener:   listener,
		checkpoint: checkpoint,
	}
	return checkpointListener
}

func (listener *CheckpointBlockEventListener) Event() (*common.Block, error) {
	event, err := listener.listener.Event()
	if err != nil {
		return nil, err
	}

	listener.checkpoint(event)

	return event, nil
}

func (listener *CheckpointBlockEventListener) Close() {
	listener.listener.Close()
}
