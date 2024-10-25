// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc/status"
)

type grpcError struct {
	error
}

func (e *grpcError) GRPCStatus() *status.Status {
	return status.Convert(e.error)
}

func (e *grpcError) Unwrap() error {
	return e.error
}

func newTransactionError(err error, transactionID string) *TransactionError {
	if err == nil {
		return nil
	}

	return &TransactionError{
		grpcError:     grpcError{err},
		TransactionID: transactionID,
	}
}

func (e *grpcError) Details() []*gateway.ErrorDetail {
	var results []*gateway.ErrorDetail

	for _, detail := range e.GRPCStatus().Details() {
		switch detail := detail.(type) {
		case *gateway.ErrorDetail:
			results = append(results, detail)
		}
	}

	return results
}

// TransactionError represents an error invoking a transaction. This is a gRPC [status] error.
type TransactionError struct {
	grpcError
	TransactionID string
}

// Error message including attached details.
func (e *TransactionError) Error() string {
	var result strings.Builder
	result.WriteString(e.grpcError.Error())

	details := e.Details()
	if len(details) == 0 {
		return result.String()
	}

	result.WriteString("\nDetails:")
	for _, detail := range details {
		result.WriteString("\n  - Address: ")
		result.WriteString(detail.GetAddress())
		result.WriteString("\n    MspId: ")
		result.WriteString(detail.GetMspId())
		result.WriteString("\n    Message: ")
		result.WriteString(detail.GetMessage())
	}

	return result.String()
}

// EndorseError represents a failure endorsing a transaction proposal.
type EndorseError struct {
	*TransactionError
}

// Error message including attached details.
func (e *EndorseError) Error() string {
	return fmt.Sprintf("endorse error: %s", e.TransactionError)
}

// Unwrap the next error in the error chain
func (e *EndorseError) Unwrap() error {
	return e.TransactionError
}

// SubmitError represents a failure submitting an endorsed transaction to the orderer.
type SubmitError struct {
	*TransactionError
}

// Error message including attached details.
func (e *SubmitError) Error() string {
	return fmt.Sprintf("submit error: %s", e.TransactionError)
}

// Unwrap the next error in the error chain
func (e *SubmitError) Unwrap() error {
	return e.TransactionError
}

// CommitStatusError represents a failure obtaining the commit status of a transaction.
type CommitStatusError struct {
	*TransactionError
}

// Error message including attached details.
func (e *CommitStatusError) Error() string {
	return fmt.Sprintf("commit status error: %s", e.TransactionError)
}

// Unwrap the next error in the error chain
func (e *CommitStatusError) Unwrap() error {
	return e.TransactionError
}

func newCommitError(transactionID string, code peer.TxValidationCode) error {
	return &CommitError{
		message:       fmt.Sprintf("transaction %s failed to commit with status code %d (%s)", transactionID, int32(code), peer.TxValidationCode_name[int32(code)]),
		TransactionID: transactionID,
		Code:          code,
	}
}

// CommitError represents a transaction that fails to commit successfully.
type CommitError struct {
	message       string
	TransactionID string
	Code          peer.TxValidationCode
}

func (e *CommitError) Error() string {
	return e.message
}
