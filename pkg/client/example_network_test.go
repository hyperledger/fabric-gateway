/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func ExampleNetwork_ChaincodeEvents() {
	var network *client.Network // Obtained from Gateway.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := network.ChaincodeEvents(ctx, "chaincodeName", client.WithStartBlock(101))
	if err != nil {
		panic(err)
	}

	for event := range events {
		fmt.Printf("Received event: %#v\n", event)
		// Break and cancel the context when done reading.
	}
}

func ExampleNetwork_ChaincodeEvents_checkpoint() {
	var network *client.Network // Obtained from Gateway.

	checkpointer := new(client.InMemoryCheckpointer)

	for {
		func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			events, err := network.ChaincodeEvents(
				ctx,
				"chaincodeName",
				client.WithStartBlock(101), // Ignored if the checkpointer has checkpoint state
				client.WithCheckpoint(checkpointer),
			)
			if err != nil {
				panic(err)
			}

			for event := range events {
				// Process event
				checkpointer.CheckpointChaincodeEvent(event)
			}

			ctx.Err() // Reason events channel closed
		}()
	}
}

func ExampleNetwork_BlockEvents() {
	var network *client.Network // Obtained from Gateway

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := network.BlockEvents(ctx, client.WithStartBlock(101))
	if err != nil {
		panic(err)
	}

	for event := range events {
		fmt.Printf("Received block number %d\n", event.GetHeader().GetNumber())
		// Break and cancel the context when done reading.
	}
}

func ExampleNetwork_BlockEvents_checkpoint() {
	var network *client.Network // Obtained from Gateway.

	checkpointer := new(client.InMemoryCheckpointer)

	for {
		func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			events, err := network.BlockEvents(
				ctx,
				client.WithStartBlock(101), // Ignored if the checkpointer has checkpoint state
				client.WithCheckpoint(checkpointer),
			)
			if err != nil {
				panic(err)
			}

			for event := range events {
				// Process then checkpoint block
				checkpointer.CheckpointBlock(event.GetHeader().GetNumber())
			}

			ctx.Err() // Reason events channel closed
		}()
	}
}

func ExampleNetwork_FilteredBlockEvents() {
	var network *client.Network // Obtained from Gateway

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := network.FilteredBlockEvents(ctx, client.WithStartBlock(101))
	if err != nil {
		panic(err)
	}

	for event := range events {
		fmt.Printf("Received block number %d\n", event.GetNumber())
		// Break and cancel the context when done reading.
	}
}

func ExampleNetwork_BlockAndPrivateDataEvents() {
	var network *client.Network // Obtained from Gateway

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := network.BlockAndPrivateDataEvents(ctx, client.WithStartBlock(101))
	if err != nil {
		panic(err)
	}

	for event := range events {
		fmt.Printf("Received block number %d\n", event.GetBlock().GetHeader().GetNumber())
		// Break and cancel the context when done reading.
	}
}
