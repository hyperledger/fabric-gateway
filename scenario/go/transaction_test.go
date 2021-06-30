/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Transaction struct {
	network     *client.Network
	contract    *client.Contract
	txType      TransactionType
	name        string
	options     []client.ProposalOption
	offlineSign identity.Sign
	result      []byte
}

func NewTransaction(network *client.Network, contract *client.Contract, txType TransactionType, name string) *Transaction {
	return &Transaction{
		network:  network,
		contract: contract,
		txType:   txType,
		name:     name,
	}
}

func (transaction *Transaction) AddOptions(options ...client.ProposalOption) {
	transaction.options = append(transaction.options, options...)
}

func (transaction *Transaction) Invoke() error {
	var result []byte
	var err error

	switch transaction.txType {
	case Submit:
		result, err = transaction.submit()
	case Evaluate:
		result, err = transaction.evaluate()
	default:
		panic(fmt.Errorf("unknown transaction type: %v", transaction.txType))
	}

	transaction.result = result
	return err
}

func (transaction *Transaction) SetOfflineSign(sign identity.Sign) {
	transaction.offlineSign = sign
}

func (transaction *Transaction) Result() []byte {
	return transaction.result
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
	if status != peer.TxValidationCode_VALID {
		return result, fmt.Errorf("commit failed with status %v (%v)", status, peer.TxValidationCode_name[int32(status)])
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

	proposal, err = transaction.contract.NewSignedProposal(bytes, signature)
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

	clientTransaction, err = transaction.contract.NewSignedTransaction(bytes, signature)
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

	commit, err = transaction.network.NewSignedCommit(bytes, signature)
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
