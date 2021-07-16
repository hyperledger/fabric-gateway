/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
)

func NewGatewayConnection(user string, mspID string) (*GatewayConnection, error) {
	id, err := newIdentity(mspID, certificatePath(user, mspID))
	if err != nil {
		return nil, err
	}

	connection := GatewayConnection{
		id: id,
	}
	connection.ctx, connection.cancel = context.WithCancel(context.Background())

	return &connection, nil
}

func NewGatewayConnectionWithSigner(user string, mspID string) (*GatewayConnection, error) {
	connection, err := NewGatewayConnection(user, mspID)
	if err != nil {
		return nil, err
	}

	sign, err := NewSign(PrivateKeyPath(user, mspID))
	if err != nil {
		return nil, err
	}

	connection.AddOptions(client.WithSign(sign))

	return connection, nil
}

func newIdentity(mspID string, certPath string) (*identity.X509Identity, error) {
	certificatePEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(mspID, certificate)
}

func NewSign(keyPath string) (identity.Sign, error) {
	privateKeyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}
func certificatePath(user string, mspID string) string {
	org := GetOrgForMSP(mspID)
	return credentialsDirectory(user, org) + "/signcerts/" + user + "@" + org + "-cert.pem"
}

func PrivateKeyPath(user string, mspID string) string {
	return credentialsDirectory(user, GetOrgForMSP(mspID)) + "/keystore/key.pem"
}

func credentialsDirectory(user string, org string) string {
	return fixturesDir + "/crypto-material/crypto-config/peerOrganizations/" + org + "/users/" +
		user + "@" + org + "/msp"
}

type GatewayConnection struct {
	id         identity.Identity
	options    []client.ConnectOption
	grpcClient *grpc.ClientConn
	gateway    *client.Gateway
	network    *client.Network
	contract   *client.Contract
	ctx        context.Context
	cancel     context.CancelFunc
	events     <-chan *client.ChaincodeEvent
}

func (connection *GatewayConnection) AddOptions(options ...client.ConnectOption) {
	connection.options = append(connection.options, options...)
}

func (connection *GatewayConnection) Connect(grpcClient *grpc.ClientConn) error {
	options := make([]client.ConnectOption, 0, len(connection.options)+1)
	options = append(options, connection.options...)
	options = append(options, client.WithClientConnection(grpcClient))

	gateway, err := client.Connect(connection.id, options...)
	if err != nil {
		grpcClient.Close()
		return err
	}

	connection.grpcClient = grpcClient
	connection.gateway = gateway

	return nil
}

func (connection *GatewayConnection) UseContract(contractName string) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	connection.contract = connection.network.GetContract(contractName)
	return nil
}

func (connection *GatewayConnection) UseNetwork(channelName string) error {
	if connection.gateway == nil {
		return fmt.Errorf("gateway not connected")
	}

	connection.network = connection.gateway.GetNetwork(channelName)
	return nil
}

func (connection *GatewayConnection) PrepareTransaction(txnType TransactionType, name string) (*Transaction, error) {
	if connection.network == nil {
		return nil, fmt.Errorf("no network selected")
	}
	if connection.contract == nil {
		return nil, fmt.Errorf("no contract selected")
	}

	return NewTransaction(connection.network, connection.contract, txnType, name), nil
}

func (connection *GatewayConnection) ListenForChaincodeEvents(chaincodeID string) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	events, err := connection.network.ChaincodeEvents(connection.ctx, chaincodeID)
	if err != nil {
		return err
	}

	connection.events = events
	return nil
}

func (connection *GatewayConnection) ChaincodeEvent() (*client.ChaincodeEvent, error) {
	if connection.events == nil {
		return nil, fmt.Errorf("no chaincode event listener attached")
	}

	event, ok := <-connection.events
	if !ok {
		return nil, fmt.Errorf("Event channel closed")
	}
	return event, nil
}

func (connection *GatewayConnection) Close() {
	connection.cancel()

	if connection.gateway != nil {
		connection.gateway.Close()
	}
	if connection.grpcClient != nil {
		connection.grpcClient.Close()
	}
}
