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
