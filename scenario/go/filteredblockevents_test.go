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

type FilteredBlockEventListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	events <-chan *peer.FilteredBlock
}

func NewFilteredBlockEventListener(parentCtx context.Context, network *client.Network, options ...client.BlockEventsOption) (*FilteredBlockEventListener, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	events, err := network.FilteredBlockEvents(ctx, options...)
	if err != nil {
		cancel()
		return nil, err
	}

	listener := &FilteredBlockEventListener{
		ctx:    ctx,
		cancel: cancel,
		events: events,
	}
	return listener, nil
}

func (listener *FilteredBlockEventListener) Event() (*peer.FilteredBlock, error) {
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

func (listener *FilteredBlockEventListener) Close() {
	listener.cancel()
}

type CheckpointFilteredBlockEventListener struct {
	listener   FilteredBlockEvents
	checkpoint func(*peer.FilteredBlock)
}

func NewCheckpointFilteredBlockEventListener(listener FilteredBlockEvents, checkpoint func(*peer.FilteredBlock)) *CheckpointFilteredBlockEventListener {
	checkpointListener := &CheckpointFilteredBlockEventListener{
		listener:   listener,
		checkpoint: checkpoint,
	}
	return checkpointListener
}

func (listener *CheckpointFilteredBlockEventListener) Event() (*peer.FilteredBlock, error) {
	event, err := listener.listener.Event()
	if err != nil {
		return nil, err
	}

	listener.checkpoint(event)

	return event, nil
}

func (listener *CheckpointFilteredBlockEventListener) Close() {
	listener.listener.Close()
}
