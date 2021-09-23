/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
)

var hsmSignerFactory *identity.HSMSignerFactory

func findSoftHSMLibrary() (string, error) {

	libraryLocations := []string{
		"/usr/lib/softhsm/libsofthsm2.so",
		"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/local/lib/softhsm/libsofthsm2.so",
		"/usr/lib/libacsp-pkcs11.so",
	}

	for _, libraryLocation := range libraryLocations {
		if _, err := os.Stat(libraryLocation); errors.Is(err, os.ErrNotExist) {
			// file does not exist
		} else {
			return libraryLocation, nil
		}
	}
	return "", fmt.Errorf("no SoftHSM Library found")
}

func NewGatewayConnection(user string, mspID string, isHSMUser bool) (*GatewayConnection, error) {
	certificatePathImpl := certificatePath
	if isHSMUser {
		certificatePathImpl = hsmCertificatePath
	}

	id, err := newIdentity(mspID, certificatePathImpl(user, mspID))
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
	connection, err := NewGatewayConnection(user, mspID, false)
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

func NewGatewayConnectionWithHSMSigner(user string, mspID string) (*GatewayConnection, error) {
	connection, err := NewGatewayConnection(user, mspID, true)
	if err != nil {
		return nil, err
	}

	hsmSign, _, err := NewHSMSigner(user)
	if err != nil {
		return nil, err
	}

	connection.AddOptions(client.WithSign(hsmSign))

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

func NewHSMSigner(user string) (identity.Sign, identity.HSMSignClose, error) {
	if hsmSignerFactory == nil {
		softHSMLibrary, err := findSoftHSMLibrary()
		if err != nil {
			return nil, nil, err
		}

		hsmSignerFactory, err = identity.NewHSMSignerFactory(softHSMLibrary)
		if err != nil {
			return nil, nil, err
		}
	}

	certificatePEM, err := ioutil.ReadFile(hsmCertificatePath(user, ""))
	if err != nil {
		return nil, nil, err
	}

	ski := getSKI(certificatePEM)
	hsmSignerOptions := identity.HSMSignerOptions{
		Label:      "ForFabric",
		Pin:        "98765432",
		Identifier: string(ski),
	}

	return hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
}

func getSKI(certPEM []byte) []byte {
	block, _ := pem.Decode(certPEM)

	x590cert, _ := x509.ParseCertificate(block.Bytes)
	pk := x590cert.PublicKey

	return skiForKey(pk.(*ecdsa.PublicKey))
}

func skiForKey(pk *ecdsa.PublicKey) []byte {
	ski := sha256.Sum256(elliptic.Marshal(pk.Curve, pk.X, pk.Y))
	return ski[:]
}

func certificatePath(user string, mspID string) string {
	org := GetOrgForMSP(mspID)
	return credentialsDirectory(user, org) + "/signcerts/" + user + "@" + org + "-cert.pem"
}

func hsmCertificatePath(user string, _ string) string {
	return fixturesDir + "/crypto-material/hsm/" + user + "/signcerts/cert.pem"
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
	return connection.receiveChaincodeEvents(chaincodeID)
}

func (connection *GatewayConnection) ReplayChaincodeEvents(chaincodeID string, startBlock uint64) error {
	return connection.receiveChaincodeEvents(chaincodeID, client.WithStartBlock(startBlock))
}

func (connection *GatewayConnection) receiveChaincodeEvents(chaincodeID string, options ...client.ChaincodeEventsOption) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	events, err := connection.network.ChaincodeEvents(connection.ctx, chaincodeID, options...)
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
