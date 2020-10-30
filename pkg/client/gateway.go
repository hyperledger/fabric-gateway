/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"io"

	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	proto "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Gateway connection
type Gateway struct {
	id     identity.Identity
	sign   identity.Sign
	client proto.GatewayClient
	closer io.Closer
}

// Connect a Gateway
func Connect(id identity.Identity, sign identity.Sign, options ...ConnectOption) (*Gateway, error) {
	gateway := &Gateway{
		id:   id,
		sign: sign,
	}

	for _, option := range options {
		if err := option(gateway); err != nil {
			return nil, err
		}
	}

	if nil == gateway.client {
		return nil, errors.New("No connection details supplied")
	}

	return gateway, nil
}

// ConnectOption implements an option that can be used when connecting a Gateway.
type ConnectOption = func(gateway *Gateway) error

// WithEndpoint specifies a Gateway endpoint to which the client will connect.
func WithEndpoint(endpoint *connection.Endpoint) ConnectOption {
	return func(gateway *Gateway) error {
		clientConnection, err := endpoint.Dial()
		if err != nil {
			return errors.Wrap(err, "Failed to establish Gateway connection")
		}

		gateway.closer = clientConnection
		gateway.client = proto.NewGatewayClient(clientConnection)
		return nil
	}
}

// WithClientConnection uses a previously configured or shared gRPC client connection for the Gateway. The client
// connection will not be closed when the Gateway is closed.
func WithClientConnection(clientConnection *grpc.ClientConn) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.client = proto.NewGatewayClient(clientConnection)
		return nil
	}
}

// Close a Gateway when it is no longer required.
func (gateway *Gateway) Close() error {
	if gateway.closer != nil {
		return gateway.closer.Close()
	}
	return nil
}

// GetNetwork returns a Network for a given channel name.
func (gateway *Gateway) GetNetwork(name string) *Network {
	return &Network{
		gateway: gateway,
		name:    name,
	}
}
