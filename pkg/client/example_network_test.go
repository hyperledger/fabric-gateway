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
		fmt.Printf("Event: %#v\n", event)
		// Break when done reading.
	}
}
