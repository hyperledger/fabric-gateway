/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	proto "github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
)

//go:generate mockgen -destination ./closer_mock_test.go -package ${GOPACKAGE} io Closer
//go:generate mockgen -destination ./gateway_mock_test.go -package ${GOPACKAGE} github.com/hyperledger/fabric-protos-go/gateway GatewayClient

// WithClient uses the supplied client for the Gateway. Allows a stub implementation to be used for testing.
func WithClient(client proto.GatewayClient) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.client = client
		return nil
	}
}

// WithIdentity uses the supplied identity for the Gateway.
func WithIdentity(id identity.Identity) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.signingID.id = id
		return nil
	}
}

func AssertNewTestGateway(t *testing.T, options ...ConnectOption) *Gateway {
	options = append([]ConnectOption{WithSign(TestCredentials.sign)}, options...)
	gateway, err := Connect(TestCredentials.identity, options...)
	if err != nil {
		t.Fatal(err)
	}

	return gateway
}

func TestGateway(t *testing.T) {
	id := TestCredentials.identity
	sign := TestCredentials.sign

	t.Run("Connect Gateway with no endpoint returns error", func(t *testing.T) {
		if _, err := Connect(id, WithSign(sign)); nil == err {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("Connect Gateway using existing gRPC client connection", func(t *testing.T) {
		var clientConnection *grpc.ClientConn
		gateway, err := Connect(id, WithSign(sign), WithClientConnection(clientConnection))
		if err != nil {
			t.Fatal(err)
		}
		if nil == gateway {
			t.Fatal("Expected gateway, got nil")
		}
	})

	t.Run("Close Gateway using existing gRPC client connection does not close connection", func(t *testing.T) {
		var clientConnection *grpc.ClientConn
		gateway, err := Connect(id, WithSign(sign), WithClientConnection(clientConnection))
		if err != nil {
			t.Fatal(err)
		}

		err = gateway.Close() // This would panic if clientConnection.Close() was called
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Connect Gateway with failing option returns error", func(t *testing.T) {
		expectedErr := errors.New("GATEWAY_OPTION_ERROR")
		badOption := func(gateway *Gateway) error {
			return expectedErr
		}
		_, actualErr := Connect(id, badOption)
		if !strings.Contains(actualErr.Error(), expectedErr.Error()) {
			t.Fatalf("Expected error message to contain %s, got %v", expectedErr.Error(), actualErr)
		}
	})

	t.Run("GetNetwork returns correctly named Network", func(t *testing.T) {
		networkName := "network"
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		gateway := AssertNewTestGateway(t, WithClient(mockClient))

		network := gateway.GetNetwork(networkName)

		if nil == network {
			t.Fatal("Expected network, got nil")
		}
		if network.name != networkName {
			t.Fatalf("Expected a network named %s, got %s", networkName, network.name)
		}
	})

	t.Run("Identity returns connecting identity", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		gateway := AssertNewTestGateway(t, WithIdentity(id), WithClient(mockClient))

		result := gateway.Identity()

		if result != id {
			t.Fatalf("Expected identity %v, got %v", id, result)
		}
	})
}
