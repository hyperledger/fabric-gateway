/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

func createProposal(txn *Transaction, args []string, signer *gateway.Signer) (*peer.Proposal, error) {

	// Add function name to arguments
	argsArray := make([][]byte, len(args)+1)
	argsArray[0] = []byte(txn.name)
	for i, arg := range args {
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

	creator, err := signer.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize Signer: ")
	}

	proposal, _, err := protoutil.CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		txn.contract.network.name,
		ccis,
		creator,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	return proposal, nil
}

func signProposal(proposal *peer.Proposal, signer *gateway.Signer) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
	}

	signature, err := signer.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	sproposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return sproposal, nil
}

func signEnvelope(envelope *common.Envelope, signer *gateway.Signer) error {
	signature, err := signer.Sign(envelope.Payload)
	if err != nil {
		return err
	}
	envelope.Signature = signature
	return nil
}
