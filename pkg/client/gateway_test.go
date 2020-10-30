/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	proto "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

// WithClient uses the supplied client for the Gateway. Allows a stub implementation to be used for testing.
func WithClient(client proto.GatewayClient) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.client = client
		return nil
	}
}

func TestGateway(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	id, err := identity.NewX509Identity("mspID", certificate)
	if err != nil {
		t.Fatal(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Connect Gateway with no endpoint returns error", func(t *testing.T) {
		if _, err := Connect(id, sign); nil == err {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("Connect Gateway using existing gRPC client connection", func(t *testing.T) {
		clientConnection := &grpc.ClientConn{}
		gateway, err := Connect(id, sign, WithClientConnection(clientConnection))
		if err != nil {
			t.Fatal(err)
		}
		if nil == gateway {
			t.Fatalf("Expected gateway, got nil")
		}
	})

	t.Run("Connect Gateway using endpoint", func(t *testing.T) {
		endpoint := &connection.Endpoint{
			Host: "example.org",
			Port: 7,
		}
		gateway, err := Connect(id, sign, WithEndpoint(endpoint))
		if err != nil {
			t.Fatal(err)
		}
		if nil == gateway {
			t.Fatalf("Expected gateway, got nil")
		}
	})

	t.Run("Close Gateway using existing gRPC client connection does not close connection", func(t *testing.T) {
		var clientConnection *grpc.ClientConn
		gateway, err := Connect(id, sign, WithClientConnection(clientConnection))
		if err != nil {
			t.Fatal(err)
		}

		err = gateway.Close() // This would panic if clientConnection.Close() was called
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Close Gateway using endpoint closes connection", func(t *testing.T) {
		closeCallCount := 0
		mockCloser := mock.NewCloser()
		mockCloser.MockClose = func() error {
			closeCallCount++
			return nil
		}

		endpoint := &connection.Endpoint{
			Host: "example.org",
			Port: 7,
		}
		gateway, err := Connect(id, sign, WithEndpoint(endpoint))
		if err != nil {
			t.Fatal(err)
		}

		if nil == gateway.closer {
			t.Fatal("Gateway closer is nil")
		}
		gateway.closer = mockCloser

		err = gateway.Close()
		if err != nil {
			t.Fatal(err)
		}

		if closeCallCount < 1 {
			t.Fatal("Close() not called")
		}
	})
}
