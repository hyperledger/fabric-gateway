/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

type proposalBuilder struct {
	contract  *Contract
	name      string
	transient map[string][]byte
	args      []string
}

func (builder *proposalBuilder) build() (*Proposal, error) {
	proposalProto, err := builder.newProposalProto()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Proposal protobuf")
	}

	proposalBytes, err := proto.Marshal(proposalProto)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshall Proposal protobuf")
	}

	proposal := &Proposal{
		contract: builder.contract,
		bytes:    proposalBytes,
	}
	return proposal, nil
}

func (builder *proposalBuilder) newProposalProto() (*peer.Proposal, error) {
	// Add function name to arguments
	argsArray := make([][]byte, len(builder.args)+1)
	argsArray[0] = []byte(builder.name)
	for i, arg := range builder.args {
		argsArray[i+1] = []byte(arg)
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: builder.contract.name},
			Input:       &peer.ChaincodeInput{Args: argsArray},
		},
	}

	creator, err := identity.Serialize(builder.contract.network.gateway.id)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize identity: ")
	}

	result, _, err := protoutil.CreateChaincodeProposalWithTransient(
		common.HeaderType_ENDORSER_TRANSACTION,
		builder.contract.network.name,
		ccis,
		creator,
		builder.transient,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create chaincode proposal")
	}

	return result, nil
}

// ProposalOption implements an option for a transaction proposal.
type ProposalOption = func(builder *proposalBuilder) error

// WithArguments specifies the arguments associated with a transaction proposal.
func WithArguments(args ...string) ProposalOption {
	return func(builder *proposalBuilder) error {
		builder.args = args
		return nil
	}
}

// WithTransient specifies the transient data associated with a transaction proposal.
func WithTransient(transient map[string][]byte) ProposalOption {
	return func(builder *proposalBuilder) error {
		builder.transient = transient
		return nil
	}
}

// Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
type Proposal struct {
	contract  *Contract
	bytes     []byte
	signature []byte
}

// Bytes of the serialized proposal.
func (proposal *Proposal) Bytes() ([]byte, error) {
	return proposal.bytes, nil
}

// Hash the proposal to obtain a digest to be signed.
func (proposal *Proposal) Hash() ([]byte, error) {
	return identity.Hash(proposal.bytes)
}

// Endorse the proposal to obtain an endorsed transaction for submission to the orderer.
func (proposal *Proposal) Endorse() (*Transaction, error) {
	signedProposal, err := proposal.newSignedProposal()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	preparedTransaction, err := proposal.contract.network.gateway.client.Prepare(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to endorse proposal")
	}

	result := &Transaction{
		contract:            proposal.contract,
		preparedTransaction: preparedTransaction,
	}
	return result, nil
}

// Evaluate the proposal to obtain a transaction result. This is effectively a query.
func (proposal *Proposal) Evaluate() ([]byte, error) {
	signedProposal, err := proposal.newSignedProposal()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := proposal.contract.network.gateway.client.Evaluate(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate transaction")
	}

	return result.Value, nil
}

func (proposal *Proposal) newSignedProposal() (*peer.SignedProposal, error) {
	if err := proposal.signMessage(); err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposal.bytes,
		Signature:     proposal.signature,
	}
	return signedProposal, nil
}

func (proposal *Proposal) signMessage() error {
	if proposal.signature != nil {
		return nil
	}

	digest, err := proposal.Hash()
	if err != nil {
		return err
	}

	proposal.signature, err = proposal.contract.network.gateway.sign(digest)
	if err != nil {
		return err
	}

	return nil
}
