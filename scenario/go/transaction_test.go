/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc/status"
)

type Transaction struct {
	gateway     *client.Gateway
	contract    *client.Contract
	txType      TransactionType
	name        string
	options     []client.ProposalOption
	offlineSign identity.Sign
	result      []byte
	err         error
	blockNumber uint64
}

func NewTransaction(gateway *client.Gateway, contract *client.Contract, txType TransactionType, name string) *Transaction {
	return &Transaction{
		gateway:  gateway,
		contract: contract,
		txType:   txType,
		name:     name,
	}
}

func (transaction *Transaction) AddOptions(options ...client.ProposalOption) {
	transaction.options = append(transaction.options, options...)
}

func (transaction *Transaction) Invoke() error {
	switch transaction.txType {
	case Submit:
		transaction.result, transaction.err = transaction.submit()
	case Evaluate:
		transaction.result, transaction.err = transaction.evaluate()
	default:
		panic(fmt.Errorf("unknown transaction type: %v", transaction.txType))
	}

	return transaction.err
}

func (transaction *Transaction) SetOfflineSign(sign identity.Sign) {
	transaction.offlineSign = sign
}

func (transaction *Transaction) Result() []byte {
	return transaction.result
}

func (transaction *Transaction) Err() error {
	return transaction.err
}

func (transaction *Transaction) ErrDetails() []*gateway.ErrorDetail {
	statusErr := status.Convert(transaction.Err())
	var details []*gateway.ErrorDetail
	for _, detail := range statusErr.Details() {
		details = append(details, detail.(*gateway.ErrorDetail))
	}
	return details
}

func (transaction *Transaction) BlockNumber() uint64 {
	return transaction.blockNumber
}

func (transaction *Transaction) evaluate() ([]byte, error) {
	proposal, err := transaction.contract.NewProposal(transaction.name, transaction.options...)
	if err != nil {
		return nil, err
	}

	proposal, err = transaction.offlineSignProposal(proposal)
	if err != nil {
		return nil, err
	}

	return proposal.Evaluate()
}

func (transaction *Transaction) submit() ([]byte, error) {
	proposal, err := transaction.contract.NewProposal(transaction.name, transaction.options...)
	if err != nil {
		return nil, err
	}

	proposal, err = transaction.offlineSignProposal(proposal)
	if err != nil {
		return nil, err
	}

	unsignedTransaction, err := proposal.Endorse()
	if err != nil {
		return nil, err
	}

	signedTransaction, err := transaction.offlineSignTransaction(unsignedTransaction)
	if err != nil {
		return nil, err
	}

	result := signedTransaction.Result()

	unsignedCommit, err := signedTransaction.Submit()
	if err != nil {
		return result, err
	}

	signedCommit, err := transaction.offlineSignCommit(unsignedCommit)
	if err != nil {
		return result, err
	}

	status, err := signedCommit.Status()
	if err != nil {
		return result, err
	}

	transaction.blockNumber = status.BlockNumber

	if !status.Successful {
		return result, fmt.Errorf("commit failed with status %v (%v)", status.Code, peer.TxValidationCode_name[int32(status.Code)])
	}

	return result, nil
}

func (transaction *Transaction) offlineSignProposal(proposal *client.Proposal) (*client.Proposal, error) {
	if nil == transaction.offlineSign {
		return proposal, nil
	}

	bytes, signature, err := transaction.bytesAndSignature(proposal)
	if err != nil {
		return nil, err
	}

	proposal, err = transaction.gateway.NewSignedProposal(bytes, signature)
	if err != nil {
		return nil, err
	}

	return proposal, nil
}

func (transaction *Transaction) offlineSignTransaction(clientTransaction *client.Transaction) (*client.Transaction, error) {
	if nil == transaction.offlineSign {
		return clientTransaction, nil
	}

	bytes, signature, err := transaction.bytesAndSignature(clientTransaction)
	if err != nil {
		return nil, err
	}

	clientTransaction, err = transaction.gateway.NewSignedTransaction(bytes, signature)
	if err != nil {
		return nil, err
	}

	return clientTransaction, nil
}

func (transaction *Transaction) offlineSignCommit(commit *client.Commit) (*client.Commit, error) {
	if nil == transaction.offlineSign {
		return commit, nil
	}

	bytes, signature, err := transaction.bytesAndSignature(commit)
	if err != nil {
		return nil, err
	}

	commit, err = transaction.gateway.NewSignedCommit(bytes, signature)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

type Signable interface {
	Digest() []byte
	Bytes() ([]byte, error)
}

func (transaction *Transaction) bytesAndSignature(signable Signable) ([]byte, []byte, error) {
	digest := signable.Digest()
	bytes, err := signable.Bytes()
	if err != nil {
		return nil, nil, err
	}

	signature, err := transaction.offlineSign(digest)
	if err != nil {
		return nil, nil, err
	}

	return bytes, signature, nil
}
