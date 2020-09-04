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

type Proposal struct {
	contract      *Contract
	name          string
	transient     map[string][]byte
	args          []string
	proposalBytes []byte
	signature     []byte
}

type SignedProposal struct {
	*peer.SignedProposal
}

func (proposal *Proposal) WithArgs(args ...string) *Proposal {
	proposal.args = args
	return proposal
}

func (proposal *Proposal) SetTransient(transientData map[string][]byte) *Proposal {
	proposal.transient = transientData
	return proposal
}

func (proposal *Proposal) Hash() ([]byte, error) {
	if err := proposal.createMessage(); err != nil {
		return nil, err
	}

	return identity.Hash(proposal.proposalBytes)
}

func (proposal *Proposal) Sign(signature []byte) *Proposal {
	proposal.signature = signature
	return proposal
}

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

func (proposal *Proposal) createMessage() error {
	proposalProto, err := proposal.newProposalProto()
	if err != nil {
		return errors.Wrap(err, "Failed to create Proposal protobuf")
	}

	proposal.proposalBytes, err = proto.Marshal(proposalProto)
	if err != nil {
		return errors.Wrap(err, "Failed to marshall Proposal protobuf")
	}

	return nil
}

func (proposal *Proposal) newProposalProto() (*peer.Proposal, error) {
	// Add function name to arguments
	argsArray := make([][]byte, len(proposal.args)+1)
	argsArray[0] = []byte(proposal.name)
	for i, arg := range proposal.args {
		argsArray[i+1] = []byte(arg)
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: proposal.contract.name},
			Input:       &peer.ChaincodeInput{Args: argsArray},
		},
	}

	creator, err := identity.Serialize(proposal.contract.network.gateway.id)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize identity: ")
	}

	result, _, err := protoutil.CreateChaincodeProposalWithTransient(
		common.HeaderType_ENDORSER_TRANSACTION,
		proposal.contract.network.name,
		ccis,
		creator,
		proposal.transient,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create chaincode proposal")
	}

	return result, nil
}

func (proposal *Proposal) newSignedProposal() (*peer.SignedProposal, error) {
	if err := proposal.signMessage(); err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposal.proposalBytes,
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
