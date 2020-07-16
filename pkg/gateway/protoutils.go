/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"bytes"
	"crypto/rand"

	"github.com/gogo/protobuf/proto"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

// func (gs *GatewayServer) createProposal(txn *pb.Transaction, signer *Signer) (*peer.Proposal, error) {
// 	if txn.ChaincodeID == "" {
// 		return nil, errors.New("ChaincodeID is required")
// 	}

// 	if txn.TxnName == "" {
// 		return nil, errors.New("Fcn is required")
// 	}

// 	// Add function name to arguments
// 	argsArray := make([][]byte, len(txn.Args)+1)
// 	argsArray[0] = []byte(txn.TxnName)
// 	for i, arg := range txn.Args {
// 		argsArray[i+1] = []byte(arg)
// 	}

// 	// create invocation spec to target a chaincode with arguments
// 	ccis := &peer.ChaincodeInvocationSpec{
// 		ChaincodeSpec: &peer.ChaincodeSpec{
// 			Type:        peer.ChaincodeSpec_NODE,
// 			ChaincodeId: &peer.ChaincodeID{Name: txn.ChaincodeID},
// 			Input:       &peer.ChaincodeInput{Args: argsArray},
// 		},
// 	}

// 	creator, err := signer.Serialize()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "Failed to serialize Signer: ")
// 	}

// 	proposal, _, err := protoutil.CreateChaincodeProposal(
// 		common.HeaderType_ENDORSER_TRANSACTION,
// 		txn.Channel,
// 		ccis,
// 		creator,
// 	)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to create chaincode proposal")
// 	}

// 	return proposal, nil
// }

func (gs *GatewayServer) signProposal(proposal *peer.Proposal, signer *Signer) (*peer.SignedProposal, error) {
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

func getValueFromResponse(response *peer.ProposalResponse) (*pb.Result, error) {
	var retVal []byte

	if response.Payload != nil {
		payload, err := protoutil.UnmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of proposal response payload failed")
		}

		extension, err := protoutil.UnmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of chaincode action failed")
		}

		if extension != nil && extension.Response != nil {
			retVal = extension.Response.Payload
		}
	}

	return &pb.Result{Value: retVal}, nil
}

func createSignedTx(
	proposal *peer.Proposal,
	signer protoutil.Signer,
	resps ...*peer.ProposalResponse,
) (*common.Envelope, error) {
	if len(resps) == 0 {
		return nil, errors.New("at least one proposal response is required")
	}

	// the original header
	hdr, err := protoutil.UnmarshalHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	// // check that the Signer is the same that is referenced in the header
	// // TODO: maybe worth removing?
	// signerBytes, err := Signer.Serialize()
	// if err != nil {
	// 	return nil, err
	// }

	// shdr, err := protoutil.UnmarshalSignatureHeader(hdr.SignatureHeader)
	// if err != nil {
	// 	return nil, err
	// }

	// if !bytes.Equal(signerBytes, shdr.Creator) {
	// 	return nil, errors.New("Signer must be the same as the one referenced in the header")
	// }

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if r.Response.Status < 200 || r.Response.Status >= 400 {
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}

		if n == 0 {
			a1 = r.Payload
			continue
		}

		if !bytes.Equal(a1, r.Payload) {
			return nil, errors.New("ProposalResponsePayloads do not match")
		}
	}

	// fill endorsements
	endorsements := make([]*peer.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protoutil.GetBytesProposalPayloadForTx(pPayl)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protoutil.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := protoutil.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// generate a random nonce
	nonce, err := getRandomNonce()
	if err != nil {
		return nil, err
	}

	creator, err := signer.Serialize()

	txSig, err := proto.Marshal(&common.SignatureHeader{
		Nonce:   nonce,
		Creator: creator,
	})

	// // recompute the txid and update the ChannelHeader
	// chHdr, err := protoutil.UnmarshalChannelHeader(hdr.ChannelHeader)
	// if err != nil {
	// 	return nil, err
	// }
	// chHdr.TxId = protoutil.ComputeTxID(nonce, creator)
	// chHdrBytes, err := proto.Marshal(chHdr)
	// if err != nil {
	// 	return nil, err
	// }

	txHdr := &common.Header{
		ChannelHeader:   hdr.ChannelHeader,
		SignatureHeader: txSig,
	}

	// create the payload
	payl := &common.Payload{Header: txHdr, Data: txBytes}
	paylBytes, err := protoutil.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// sign the payload
	sig, err := signer.Sign(paylBytes)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	return &common.Envelope{Payload: paylBytes, Signature: sig}, nil
}

func createUnsignedTx(
	proposal *peer.Proposal,
	resps ...*peer.ProposalResponse,
) (*common.Envelope, error) {
	if len(resps) == 0 {
		return nil, errors.New("at least one proposal response is required")
	}

	// the original header
	hdr, err := protoutil.UnmarshalHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	// // check that the Signer is the same that is referenced in the header
	// // TODO: maybe worth removing?
	// signerBytes, err := Signer.Serialize()
	// if err != nil {
	// 	return nil, err
	// }

	// shdr, err := protoutil.UnmarshalSignatureHeader(hdr.SignatureHeader)
	// if err != nil {
	// 	return nil, err
	// }

	// if !bytes.Equal(signerBytes, shdr.Creator) {
	// 	return nil, errors.New("Signer must be the same as the one referenced in the header")
	// }

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if r.Response.Status < 200 || r.Response.Status >= 400 {
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}

		if n == 0 {
			a1 = r.Payload
			continue
		}

		if !bytes.Equal(a1, r.Payload) {
			return nil, errors.New("ProposalResponsePayloads do not match")
		}
	}

	// fill endorsements
	endorsements := make([]*peer.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protoutil.GetBytesProposalPayloadForTx(pPayl)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protoutil.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := protoutil.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protoutil.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	return &common.Envelope{Payload: paylBytes}, nil
}

func getRandomNonce() ([]byte, error) {
	key := make([]byte, 24)

	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "error getting random bytes")
	}
	return key, nil
}
