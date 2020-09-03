/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

type EvaluateTransaction struct {
	proposal *Proposal
}

func (evaluateTransaction *EvaluateTransaction) WithArgs(args ...string) *EvaluateTransaction {
	evaluateTransaction.proposal.WithArgs(args...)
	return evaluateTransaction
}

func (evaluateTransaction *EvaluateTransaction) SetTransient(transientData map[string][]byte) *EvaluateTransaction {
	evaluateTransaction.proposal.SetTransient(transientData)
	return evaluateTransaction
}

func (evaluateTransaction *EvaluateTransaction) Invoke() ([]byte, error) {
	return evaluateTransaction.proposal.Evaluate()
}
