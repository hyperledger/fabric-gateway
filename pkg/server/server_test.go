/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/server/mocks"
)

func TestNewGatewayServer(t *testing.T) {

	_, err := NewGatewayServer(mocks.NewMockRegistry())

	if err != nil {
		t.Fatalf("Failed to create gateway server: %s", err)
	}

}
