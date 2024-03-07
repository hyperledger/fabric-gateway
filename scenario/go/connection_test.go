/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
)

var hsmSignerFactory *identity.HSMSignerFactory

func findSoftHSMLibrary() (string, error) {
	libraryLocations := []string{
		"/usr/lib/softhsm/libsofthsm2.so",
		"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
		"/usr/local/lib/softhsm/libsofthsm2.so",
		"/usr/lib/libacsp-pkcs11.so",
		"/opt/homebrew/lib/softhsm/libsofthsm2.so",
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

type ChaincodeEvents interface {
	Event() (*client.ChaincodeEvent, error)
	Close()
}

type BlockEvents interface {
	Event() (*common.Block, error)
	Close()
}

type FilteredBlockEvents interface {
	Event() (*peer.FilteredBlock, error)
	Close()
}

type BlockAndPrivateDataEvents interface {
	Event() (*peer.BlockAndPrivateData, error)
	Close()
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

	connection := &GatewayConnection{
		id:                                id,
		chaincodeEventListeners:           make(map[string]ChaincodeEvents),
		blockEventListeners:               make(map[string]BlockEvents),
		filteredBlockEventListeners:       make(map[string]FilteredBlockEvents),
		blockAndPrivateDataEventListeners: make(map[string]BlockAndPrivateDataEvents),
	}
	connection.ctx, connection.cancel = context.WithCancel(context.Background())

	return connection, nil
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
	certificatePEM, err := os.ReadFile(certPath)
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
	privateKeyPEM, err := os.ReadFile(keyPath)
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

	certificatePEM, err := os.ReadFile(hsmCertificatePath(user, ""))
	if err != nil {
		return nil, nil, err
	}

	ski, err := getSKI(certificatePEM)
	if err != nil {
		return nil, nil, err
	}

	hsmSignerOptions := identity.HSMSignerOptions{
		Label:      "ForFabric",
		Pin:        "98765432",
		Identifier: string(ski),
	}

	return hsmSignerFactory.NewHSMSigner(hsmSignerOptions)
}

func getSKI(certPEM []byte) ([]byte, error) {
	block, _ := pem.Decode(certPEM)

	x590cert, _ := x509.ParseCertificate(block.Bytes)
	pk := x590cert.PublicKey

	return skiForKey(pk.(*ecdsa.PublicKey))
}

func skiForKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	ecdhPublicKey, err := publicKey.ECDH()
	if err != nil {
		return nil, err
	}

	ski := sha256.Sum256(ecdhPublicKey.Bytes())
	return ski[:], nil
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
	id                                identity.Identity
	options                           []client.ConnectOption
	grpcClient                        *grpc.ClientConn
	gateway                           *client.Gateway
	network                           *client.Network
	contract                          *client.Contract
	ctx                               context.Context
	cancel                            context.CancelFunc
	checkpointer                      *client.InMemoryCheckpointer
	chaincodeEventListeners           map[string]ChaincodeEvents
	blockEventListeners               map[string]BlockEvents
	filteredBlockEventListeners       map[string]FilteredBlockEvents
	blockAndPrivateDataEventListeners map[string]BlockAndPrivateDataEvents
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

	return NewTransaction(connection.gateway, connection.contract, txnType, name), nil
}

func (connection *GatewayConnection) ListenForChaincodeEvents(listenerName string, chaincodeName string) error {
	listener, err := connection.newChaincodeEventListener(chaincodeName)
	if err != nil {
		return err
	}

	connection.setChaincodeEventListener(listenerName, listener)
	return nil
}
func (connection *GatewayConnection) createCheckpointer() {
	connection.checkpointer = new(client.InMemoryCheckpointer)
}

func (connection *GatewayConnection) ListenForChaincodeEventsUsingCheckpointer(listenerName string, chaincodeName string) error {
	if connection.checkpointer == nil {
		return fmt.Errorf("no checkpointer")
	}

	listener, err := connection.newChaincodeEventListener(chaincodeName, client.WithCheckpoint(connection.checkpointer))
	if err != nil {
		return err
	}

	checkpointListener := NewCheckpointChaincodeEventListener(listener, func(event *client.ChaincodeEvent) {
		connection.checkpointer.CheckpointChaincodeEvent(event)
	})
	connection.setChaincodeEventListener(listenerName, checkpointListener)
	return nil
}

func (connection *GatewayConnection) ReplayChaincodeEvents(listenerName string, chaincodeName string, startBlock uint64) error {
	listener, err := connection.newChaincodeEventListener(chaincodeName, client.WithStartBlock(startBlock))
	if err != nil {
		return err
	}

	connection.setChaincodeEventListener(listenerName, listener)
	return nil
}

func (connection *GatewayConnection) setChaincodeEventListener(listenerName string, listener ChaincodeEvents) {
	connection.CloseChaincodeEvents(listenerName)
	connection.chaincodeEventListeners[listenerName] = listener
}

func (connection *GatewayConnection) setBlockEventListener(listenerName string, listener BlockEvents) {
	connection.CloseBlockEvents(listenerName)
	connection.blockEventListeners[listenerName] = listener
}

func (connection *GatewayConnection) setFilteredBlockEventListener(listenerName string, listener FilteredBlockEvents) {
	connection.CloseFilteredBlockEvents(listenerName)
	connection.filteredBlockEventListeners[listenerName] = listener
}

func (connection *GatewayConnection) setBlockAndPrivateDataEventListener(listenerName string, listener BlockAndPrivateDataEvents) {
	connection.CloseBlockAndPrivateDataEvents(listenerName)
	connection.blockAndPrivateDataEventListeners[listenerName] = listener
}

func (connection *GatewayConnection) newChaincodeEventListener(chaincodeName string, options ...client.ChaincodeEventsOption) (*ChaincodeEventListener, error) {
	if connection.network == nil {
		return nil, fmt.Errorf("no network selected")
	}

	return NewChaincodeEventListener(connection.ctx, connection.network, chaincodeName, options...)
}

func (connection *GatewayConnection) ChaincodeEvent(listenerName string) (*client.ChaincodeEvent, error) {
	listener := connection.chaincodeEventListeners[listenerName]
	if listener == nil {
		return nil, fmt.Errorf("no chaincode event listener attached")
	}

	return listener.Event()
}

func (connection *GatewayConnection) ListenForBlockEvents(listenerName string) error {
	return connection.receiveBlockEvents(listenerName)
}

func (connection *GatewayConnection) ListenForBlockEventsUsingCheckpointer(listenerName string) error {
	if connection.checkpointer == nil {
		return fmt.Errorf("no checkpointer")
	}

	listener, err := NewBlockEventListener(connection.ctx, connection.network, client.WithCheckpoint(connection.checkpointer))
	if err != nil {
		return err
	}

	checkpointListener := NewCheckpointBlockEventListener(listener, func(event *common.Block) {
		connection.checkpointer.CheckpointBlock(event.Header.Number)
	})
	connection.setBlockEventListener(listenerName, checkpointListener)
	return nil
}

func (connection *GatewayConnection) ListenForFilteredBlockEventsUsingCheckpointer(listenerName string) error {
	if connection.checkpointer == nil {
		return fmt.Errorf("no checkpointer")
	}

	listener, err := NewFilteredBlockEventListener(connection.ctx, connection.network, client.WithCheckpoint(connection.checkpointer))
	if err != nil {
		return err
	}

	checkpointListener := NewCheckpointFilteredBlockEventListener(listener, func(event *peer.FilteredBlock) {
		connection.checkpointer.CheckpointBlock(event.Number)
	})
	connection.setFilteredBlockEventListener(listenerName, checkpointListener)
	return nil
}

func (connection *GatewayConnection) ListenForBlockAndPrivateDataEventsUsingCheckpointer(listenerName string) error {
	if connection.checkpointer == nil {
		return fmt.Errorf("no checkpointer")
	}

	listener, err := NewBlockAndPrivateDataEventListener(connection.ctx, connection.network, client.WithCheckpoint(connection.checkpointer))
	if err != nil {
		return err
	}

	checkpointListener := NewCheckpointBlockAndPrivateDataEventListener(listener, func(event *peer.BlockAndPrivateData) {
		connection.checkpointer.CheckpointBlock(event.Block.Header.Number)
	})
	connection.setBlockAndPrivateDataEventListener(listenerName, checkpointListener)
	return nil
}

func (connection *GatewayConnection) ReplayBlockEvents(listenerName string, startBlock uint64) error {
	return connection.receiveBlockEvents(listenerName, client.WithStartBlock(startBlock))
}

func (connection *GatewayConnection) receiveBlockEvents(listenerName string, options ...client.BlockEventsOption) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	listener, err := NewBlockEventListener(connection.ctx, connection.network, options...)
	if err != nil {
		return err
	}

	connection.CloseBlockEvents(listenerName)
	connection.blockEventListeners[listenerName] = listener
	return nil
}

func (connection *GatewayConnection) BlockEvent(listenerName string) (*common.Block, error) {
	listener := connection.blockEventListeners[listenerName]
	if listener == nil {
		return nil, fmt.Errorf("no block event listener attached")
	}

	return listener.Event()
}

func (connection *GatewayConnection) ListenForFilteredBlockEvents(listenerName string) error {
	return connection.receiveFilteredBlockEvents(listenerName)
}

func (connection *GatewayConnection) ReplayFilteredBlockEvents(listenerName string, startBlock uint64) error {
	return connection.receiveFilteredBlockEvents(listenerName, client.WithStartBlock(startBlock))
}

func (connection *GatewayConnection) receiveFilteredBlockEvents(listenerName string, options ...client.BlockEventsOption) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	listener, err := NewFilteredBlockEventListener(connection.ctx, connection.network, options...)
	if err != nil {
		return err
	}

	connection.CloseFilteredBlockEvents(listenerName)
	connection.filteredBlockEventListeners[listenerName] = listener
	return nil
}

func (connection *GatewayConnection) FilteredBlockEvent(listenerName string) (*peer.FilteredBlock, error) {
	listener := connection.filteredBlockEventListeners[listenerName]
	if listener == nil {
		return nil, fmt.Errorf("no filtered block event listener attached")
	}

	return listener.Event()
}

func (connection *GatewayConnection) ListenForBlockAndPrivateDataEvents(listenerName string) error {
	return connection.receiveBlockAndPrivateDataEvents(listenerName)
}

func (connection *GatewayConnection) ReplayBlockAndPrivateDataEvents(listenerName string, startBlock uint64) error {
	return connection.receiveBlockAndPrivateDataEvents(listenerName, client.WithStartBlock(startBlock))
}

func (connection *GatewayConnection) receiveBlockAndPrivateDataEvents(listenerName string, options ...client.BlockEventsOption) error {
	if connection.network == nil {
		return fmt.Errorf("no network selected")
	}

	listener, err := NewBlockAndPrivateDataEventListener(connection.ctx, connection.network, options...)
	if err != nil {
		return err
	}

	connection.CloseBlockAndPrivateDataEvents(listenerName)
	connection.blockAndPrivateDataEventListeners[listenerName] = listener
	return nil
}

func (connection *GatewayConnection) BlockAndPrivateDataEvent(listenerName string) (*peer.BlockAndPrivateData, error) {
	listener := connection.blockAndPrivateDataEventListeners[listenerName]
	if listener == nil {
		return nil, fmt.Errorf("no block and private data event listener attached")
	}

	return listener.Event()
}

func (connection *GatewayConnection) Close() {
	connection.cancel() // Closes all listener contexts

	if connection.gateway != nil {
		connection.gateway.Close()
	}
	if connection.grpcClient != nil {
		connection.grpcClient.Close()
	}
}

func (connection *GatewayConnection) CloseChaincodeEvents(listenerName string) {
	if listener := connection.chaincodeEventListeners[listenerName]; listener != nil {
		listener.Close()
		delete(connection.chaincodeEventListeners, listenerName)
	}
}

func (connection *GatewayConnection) CloseBlockEvents(listenerName string) {
	if listener := connection.blockEventListeners[listenerName]; listener != nil {
		listener.Close()
		delete(connection.blockEventListeners, listenerName)
	}
}

func (connection *GatewayConnection) CloseFilteredBlockEvents(listenerName string) {
	if listener := connection.filteredBlockEventListeners[listenerName]; listener != nil {
		listener.Close()
		delete(connection.filteredBlockEventListeners, listenerName)
	}
}

func (connection *GatewayConnection) CloseBlockAndPrivateDataEvents(listenerName string) {
	if listener := connection.blockAndPrivateDataEventListeners[listenerName]; listener != nil {
		listener.Close()
		delete(connection.blockAndPrivateDataEventListeners, listenerName)
	}
}
