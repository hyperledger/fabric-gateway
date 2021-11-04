/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
)

type proposalBuilder struct {
	client          *gatewayClient
	signingID       *signingIdentity
	channelName     string
	chaincodeName   string
	transactionName string
	transient       map[string][]byte
	endorsingOrgs   []string
	args            [][]byte
}

func (builder *proposalBuilder) build() (*Proposal, error) {
	proposalProto, transactionID, err := builder.newProposalProto()
	if err != nil {
		return nil, fmt.Errorf("failed to create Proposal protobuf: %w", err)
	}

	proposalBytes, err := proto.Marshal(proposalProto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall Proposal protobuf: %w", err)
	}

	signedProposalProto := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
	}

	proposedTransaction := &gateway.ProposedTransaction{
		TransactionId:          transactionID,
		Proposal:               signedProposalProto,
		EndorsingOrganizations: builder.endorsingOrgs,
	}

	proposal := &Proposal{
		client:              builder.client,
		signingID:           builder.signingID,
		channelID:           builder.channelName,
		proposedTransaction: proposedTransaction,
	}
	return proposal, nil
}

func (builder *proposalBuilder) newProposalProto() (*peer.Proposal, string, error) {
	invocationSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: builder.chaincodeName},
			Input:       &peer.ChaincodeInput{Args: builder.chaincodeArgs()},
		},
	}

	creator, err := builder.signingID.Creator()
	if err != nil {
		return nil, "", fmt.Errorf("failed to serialize identity: %w", err)
	}

	result, transactionID, err := protoutil.CreateChaincodeProposalWithTransient(
		common.HeaderType_ENDORSER_TRANSACTION,
		builder.channelName,
		invocationSpec,
		creator,
		builder.transient,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create chaincode proposal: %w", err)
	}

	return result, transactionID, nil
}

func (builder *proposalBuilder) chaincodeArgs() [][]byte {
	result := make([][]byte, len(builder.args)+1)

	result[0] = []byte(builder.transactionName)
	copy(result[1:], builder.args)

	return result
}

// ProposalOption implements an option for a transaction proposal.
type ProposalOption = func(builder *proposalBuilder) error

// WithBytesArguments appends to the transaction function arguments associated with a transaction proposal.
func WithBytesArguments(args ...[]byte) ProposalOption {
	return func(builder *proposalBuilder) error {
		builder.args = append(builder.args, args...)
		return nil
	}
}

// WithArguments appends to the transaction function arguments associated with a transaction proposal.
func WithArguments(args ...string) ProposalOption {
	return WithBytesArguments(stringsAsBytes(args)...)
}

func stringsAsBytes(strings []string) [][]byte {
	results := make([][]byte, 0, len(strings))

	for _, v := range strings {
		results = append(results, []byte(v))
	}

	return results
}

// WithTransient specifies the transient data associated with a transaction proposal.
// This is usually used in combination with WithEndorsingOrganizations for private data scenarios
func WithTransient(transient map[string][]byte) ProposalOption {
	return func(builder *proposalBuilder) error {
		builder.transient = transient
		return nil
	}
}

// WithEndorsingOrganizations specifies the organizations that should endorse the transaction proposal.
// No other organizations will be sent the proposal.  This is usually used in combination with WithTransient
// for private data scenarios, or for state-based endorsement when specific organizations have to endorse the proposal.
func WithEndorsingOrganizations(mspids ...string) ProposalOption {
	return func(builder *proposalBuilder) error {
		builder.endorsingOrgs = mspids
		return nil
	}
}
