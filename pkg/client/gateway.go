/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package client enables Go developers to build client applications using the Hyperledger Fabric programming model as
// described in the Developing Applications chapter of the Fabric documentation:
//
// https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html
//
// Client applications interact with the blockchain network using a Fabric Gateway. A client connection to a Fabric
// Gateway is established by calling client.Connect() with a client identity, client signing implementation, and client
// connection details. The returned Gateway can be used to transact with smart contracts deployed to networks
// accessible through the Fabric Gateway.
package client

import (
	"io"

	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	proto "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Gateway representing the connection of a specific client identity to a Fabric Gateway.
type Gateway struct {
	signingID *signingIdentity
	client    proto.GatewayClient
	closer    io.Closer
}

func undefinedSign(digest []byte) ([]byte, error) {
	return nil, errors.New("No sign implementation supplied")
}

// Connect to a Fabric Gateway using a client identity, signing implementation, and additional options, which must
// include gRPC client connection details.
func Connect(id identity.Identity, options ...ConnectOption) (*Gateway, error) {
	gateway := &Gateway{
		signingID: newSigningIdentity(id, undefinedSign),
	}

	if err := gateway.applyConnectOptions(options); err != nil {
		return nil, err
	}

	if nil == gateway.client {
		return nil, errors.New("No connection details supplied")
	}

	return gateway, nil
}

func (gateway *Gateway) applyConnectOptions(options []ConnectOption) error {
	for _, option := range options {
		if err := option(gateway); err != nil {
			return err
		}
	}

	return nil
}

// ConnectOption implements an option that can be used when connecting a Gateway.
type ConnectOption = func(gateway *Gateway) error

// WithSign uses the supplied signing implementation for the Gateway.
func WithSign(sign identity.Sign) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.signingID.sign = sign
		return nil
	}
}

// WithEndpoint specifies a Fabric Gateway endpoint to which a gRPC client connection will be established. The client
// connection will be closed when the Gateway is closed.
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

// WithClientConnection uses a previously configured or shared gRPC client connection to a Fabric Gateway. The client
// connection will not be closed when the Gateway is closed.
func WithClientConnection(clientConnection grpc.ClientConnInterface) ConnectOption {
	return func(gateway *Gateway) error {
		gateway.closer = nil
		gateway.client = proto.NewGatewayClient(clientConnection)
		return nil
	}
}

// Close a Gateway when it is no longer required. This releases all resources associated with Networks and Contracts
// obtained using the Gateway, including removing event listeners.
func (gateway *Gateway) Close() error {
	if gateway.closer != nil {
		return gateway.closer.Close()
	}
	return nil
}

// Identity used by this Gateway
func (gateway *Gateway) Identity() identity.Identity {
	return gateway.signingID.id
}

// GetNetwork returns a Network representing the named Fabric channel.
func (gateway *Gateway) GetNetwork(name string) *Network {
	return &Network{
		client:    gateway.client,
		signingID: gateway.signingID,
		name:      name,
	}
}
