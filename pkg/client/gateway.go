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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

// Gateway representing the connection of a specific client identity to a Fabric Gateway.
type Gateway struct {
	signingID *signingIdentity
	client    *gatewayClient
	cancel    context.CancelFunc
}

// Connect to a Fabric Gateway using a client identity, gRPC connection and signing implementation.
func Connect(id identity.Identity, options ...ConnectOption) (*Gateway, error) {
	ctx, cancel := context.WithCancel(context.Background())
	gw := &Gateway{
		signingID: newSigningIdentity(id),
		client: &gatewayClient{
			contexts: &contextFactory{
				ctx: ctx,
			},
		},
		cancel: cancel,
	}

	if err := gw.applyConnectOptions(options); err != nil {
		cancel()
		return nil, err
	}

	if gw.client.grpcClient == nil {
		cancel()
		return nil, errors.New("no connection details supplied")
	}

	return gw, nil
}

func (gw *Gateway) applyConnectOptions(options []ConnectOption) error {
	for _, option := range options {
		if err := option(gw); err != nil {
			return err
		}
	}

	return nil
}

// ConnectOption implements an option that can be used when connecting to a Fabric Gateway.
type ConnectOption = func(gateway *Gateway) error

// WithSign uses the supplied signing implementation to sign messages sent by the Gateway.
func WithSign(sign identity.Sign) ConnectOption {
	return func(gw *Gateway) error {
		gw.signingID.sign = sign
		return nil
	}
}

// WithHash uses the supplied hashing implementation to generate digital signatures.
func WithHash(hash hash.Hash) ConnectOption {
	return func(gw *Gateway) error {
		gw.signingID.hash = hash
		return nil
	}
}

// WithClientConnection uses the supplied gRPC client connection to a Fabric Gateway. This should be shared by all
// Gateway instances connecting to the same Fabric Gateway. The client connection will not be closed when the Gateway
// is closed.
func WithClientConnection(clientConnection *grpc.ClientConn) ConnectOption {
	return func(gw *Gateway) error {
		gw.client.grpcClient = gateway.NewGatewayClient(clientConnection)
		return nil
	}
}

// WithEvaluateTimeout specifies the default timeout for evaluating transactions.
func WithEvaluateTimeout(timeout time.Duration) ConnectOption {
	return func(gw *Gateway) error {
		gw.client.contexts.evaluate = func(parent context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, timeout)
		}
		return nil
	}
}

// WithEndorseTimeout specifies the default timeout for endorsements.
func WithEndorseTimeout(timeout time.Duration) ConnectOption {
	return func(gw *Gateway) error {
		gw.client.contexts.endorse = func(parent context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, timeout)
		}
		return nil
	}
}

// WithSubmitTimeout specifies the default timeout for submit of transactions to the orderer.
func WithSubmitTimeout(timeout time.Duration) ConnectOption {
	return func(gw *Gateway) error {
		gw.client.contexts.submit = func(parent context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, timeout)
		}
		return nil
	}
}

// WithCommitStatusTimeout specifies the default timeout for retrieving transaction commit status.
func WithCommitStatusTimeout(timeout time.Duration) ConnectOption {
	return func(gw *Gateway) error {
		gw.client.contexts.commitStatus = func(parent context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(parent, timeout)
		}
		return nil
	}
}

// Close a Gateway when it is no longer required. This releases all resources associated with Networks and Contracts
// obtained using the Gateway, including removing event listeners.
func (gw *Gateway) Close() error {
	gw.cancel()
	return nil
}

// Identity used by this Gateway.
func (gw *Gateway) Identity() identity.Identity {
	return gw.signingID.id
}

// GetNetwork returns a Network representing the named Fabric channel.
func (gw *Gateway) GetNetwork(name string) *Network {
	return &Network{
		client:    gw.client,
		signingID: gw.signingID,
		name:      name,
	}
}

// NewSignedProposal creates a transaction proposal with signature, which can be sent to peers for endorsement.
func (gw *Gateway) NewSignedProposal(bytes []byte, signature []byte) (*Proposal, error) {
	proposedTransaction := &gateway.ProposedTransaction{}
	if err := util.Unmarshal(bytes, proposedTransaction); err != nil {
		return nil, fmt.Errorf("failed to deserialize proposed transaction: %w", err)
	}

	proposal := &peer.Proposal{}
	if err := util.Unmarshal(proposedTransaction.GetProposal().GetProposalBytes(), proposal); err != nil {
		return nil, fmt.Errorf("failed to deserialize proposal: %w", err)
	}

	header := &common.Header{}
	if err := util.Unmarshal(proposal.GetHeader(), header); err != nil {
		return nil, fmt.Errorf("failed to deserialize header: %w", err)
	}

	channelHeader := &common.ChannelHeader{}
	if err := util.Unmarshal(header.GetChannelHeader(), channelHeader); err != nil {
		return nil, fmt.Errorf("failed to deserialize channel header: %w", err)
	}

	result := &Proposal{
		client:              gw.client,
		signingID:           gw.signingID,
		channelID:           channelHeader.GetChannelId(),
		proposedTransaction: proposedTransaction,
	}
	result.setSignature(signature)

	return result, nil
}

// NewSignedTransaction creates an endorsed transaction with signature, which can be submitted to the orderer for commit
// to the ledger.
func (gw *Gateway) NewSignedTransaction(bytes []byte, signature []byte) (*Transaction, error) {
	preparedTransaction := &gateway.PreparedTransaction{}
	if err := util.Unmarshal(bytes, preparedTransaction); err != nil {
		return nil, fmt.Errorf("failed to deserialize prepared transaction: %w", err)
	}

	transaction, err := newTransaction(gw.client, gw.signingID, preparedTransaction)
	if err != nil {
		return nil, err
	}

	transaction.setSignature(signature)

	return transaction, nil
}

// NewSignedCommit creates an commit with signature, which can be used to access a committed transaction.
func (gw *Gateway) NewSignedCommit(bytes []byte, signature []byte) (*Commit, error) {
	signedRequest := &gateway.SignedCommitStatusRequest{}
	if err := util.Unmarshal(bytes, signedRequest); err != nil {
		return nil, fmt.Errorf("failed to deserialize signed commit status request: %w", err)
	}

	request := &gateway.CommitStatusRequest{}
	if err := util.Unmarshal(signedRequest.Request, request); err != nil {
		return nil, fmt.Errorf("failed to deserialize commit status request: %w", err)
	}

	commit := newCommit(gw.client, gw.signingID, request.TransactionId, signedRequest)
	commit.setSignature(signature)

	return commit, nil
}

// NewSignedChaincodeEventsRequest creates a signed request to read events emitted by a specific chaincode.
func (gw *Gateway) NewSignedChaincodeEventsRequest(bytes []byte, signature []byte) (*ChaincodeEventsRequest, error) {
	request := &gateway.SignedChaincodeEventsRequest{}
	if err := util.Unmarshal(bytes, request); err != nil {
		return nil, fmt.Errorf("failed to deserialize signed chaincode events request: %w", err)
	}

	result := &ChaincodeEventsRequest{
		client:        gw.client,
		signingID:     gw.signingID,
		signedRequest: request,
	}
	return result, nil
}
