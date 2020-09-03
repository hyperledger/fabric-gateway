/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

func createProposal(txn *transaction, id identity.Identity) (*peer.Proposal, error) {

	// Add function name to arguments
	argsArray := make([][]byte, len(txn.args)+1)
	argsArray[0] = []byte(txn.name)
	for i, arg := range txn.args {
		argsArray[i+1] = []byte(arg)
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: txn.contract.name},
			Input:       &peer.ChaincodeInput{Args: argsArray},
		},
	}

	creator, err := identity.Serialize(id)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize identity: ")
	}

	proposal, _, err := protoutil.CreateChaincodeProposalWithTransient(
		common.HeaderType_ENDORSER_TRANSACTION,
		txn.contract.network.name,
		ccis,
		creator,
		txn.transient,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	return proposal, nil
}

func signProposal(proposal *peer.Proposal, sign identity.Sign) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
	}

	signature, err := signMessage(proposalBytes, sign)
	if err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return signedProposal, nil
}

func signEnvelope(envelope *common.Envelope, sign identity.Sign) error {
	signature, err := signMessage(envelope.Payload, sign)
	if err != nil {
		return err
	}
	envelope.Signature = signature
	return nil
}

func signMessage(message []byte, sign identity.Sign) ([]byte, error) {
	digest, err := identity.Hash(message)
	if err != nil {
		return nil, err
	}

	return sign(digest)
}
