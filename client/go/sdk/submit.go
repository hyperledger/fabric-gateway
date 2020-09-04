/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

type SubmitTransaction struct {
	proposal *Proposal
}

func (submitTransaction *SubmitTransaction) WithArgs(args ...string) *SubmitTransaction {
	submitTransaction.proposal.WithArgs(args...)
	return submitTransaction
}

func (submitTransaction *SubmitTransaction) SetTransient(transientData map[string][]byte) *SubmitTransaction {
	submitTransaction.proposal.SetTransient(transientData)
	return submitTransaction
}

func (submitTransaction *SubmitTransaction) Invoke() ([]byte, error) {
	result, commit, err := submitTransaction.InvokeAsync()
	if err != nil {
		return nil, err
	}

	if err = <-commit; err != nil {
		return nil, err
	}

	return result, nil
}

func (submitTransaction *SubmitTransaction) InvokeAsync() ([]byte, chan error, error) {
	transaction, err := submitTransaction.proposal.Endorse()
	if err != nil {
		return nil, nil, err
	}

	result := transaction.Result()

	commit, err := transaction.Submit()
	if err != nil {
		return nil, nil, err
	}

	return result, commit, nil
}
