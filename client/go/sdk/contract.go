/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

// Contract represents a chaincode smart contract
type Contract struct {
	network *Network
	name    string
}

// EvaluateTransaction invokes a query and returns its result
func (contract *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return contract.Evaluate(name).WithArgs(args...).Invoke()
}

// SubmitTransaction invokes a transaction and returns its result
func (contract *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return contract.Submit(name).WithArgs(args...).Invoke()
}

// Evaluate prepares a transaction for evalation
func (contract *Contract) Evaluate(name string) *EvaluateTransaction {
	return &EvaluateTransaction{
		contract.Proposal(name),
	}
}

// Submit prepares a transaction that will be submitted to the orderer
func (contract *Contract) Submit(name string) *SubmitTransaction {
	return &SubmitTransaction{
		contract.Proposal(name),
	}
}

// Proposal creates a proposal that can be sent to peers for endorsement. Supports off-line signing transaction flow.
func (contract *Contract) Proposal(transactionName string) *Proposal {
	return &Proposal{
		contract: contract,
		name:     transactionName,
	}
}
