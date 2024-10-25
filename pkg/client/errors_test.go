// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

func newGrpcStatus(code codes.Code, message string, details ...*gateway.ErrorDetail) (*status.Status, error) {
	result := status.New(code, message)
	if len(details) == 0 {
		return result, nil
	}

	var v1Details []protoadapt.MessageV1
	for _, detail := range details {
		v1Details = append(v1Details, protoadapt.MessageV1Of(detail))
	}

	resultWithDetails, err := result.WithDetails(v1Details...)
	if err != nil {
		return nil, err
	}

	return resultWithDetails, nil
}

func TestErrors(t *testing.T) {
	for typeName, newInstance := range map[string](func(*TransactionError) error){
		"endorse error": func(txErr *TransactionError) error {
			return &EndorseError{txErr}
		},
		"submit error": func(txErr *TransactionError) error {
			return &SubmitError{txErr}
		},
		"commit status error": func(txErr *TransactionError) error {
			return &CommitStatusError{txErr}
		},
	} {
		t.Run(typeName+" with details", func(t *testing.T) {
			details := []*gateway.ErrorDetail{
				{
					Address: "ADDRESS_1",
					MspId:   "MSP_ID_1",
					Message: "MESSAGE_1",
				},
				{
					Address: "ADDRESS_2",
					MspId:   "MSP_ID_2",
					Message: "MESSAGE_2",
				},
			}
			grpcStatus, err := newGrpcStatus(codes.Aborted, "STATUS_MESSAGE", details...)
			require.NoError(t, err)

			transactionErr := newTransactionError(grpcStatus.Err(), "TRANSACTION_ID")
			actualErr := newInstance(transactionErr)

			t.Run("Error", func(t *testing.T) {
				require.ErrorContains(t, actualErr, grpcStatus.Err().Error())
				require.ErrorContains(t, actualErr, typeName)
				for _, detail := range details {
					require.ErrorContains(t, actualErr, detail.GetAddress())
					require.ErrorContains(t, actualErr, detail.GetMspId())
					require.ErrorContains(t, actualErr, detail.GetMessage())
				}
			})

			t.Run("Details", func(t *testing.T) {
				expectedErr := new(TransactionError)
				require.ErrorAs(t, actualErr, &expectedErr)

				require.Len(t, expectedErr.Details(), len(details), "number of detail messages")
				for i, actual := range expectedErr.Details() {
					expected := details[i]
					AssertProtoEqual(t, expected, actual)
				}
			})
		})

		t.Run(typeName+" without details", func(t *testing.T) {
			grpcStatus, err := newGrpcStatus(codes.Aborted, "STATUS_MESSAGE")
			require.NoError(t, err)

			transactionErr := newTransactionError(grpcStatus.Err(), "TRANSACTION_ID")
			actualErr := newInstance(transactionErr)

			t.Run("Error", func(t *testing.T) {
				require.ErrorContains(t, actualErr, grpcStatus.Err().Error())
				require.ErrorContains(t, actualErr, typeName)
			})

			t.Run("TransactionID", func(t *testing.T) {
				var actualTransactionErr *TransactionError
				require.ErrorAs(t, actualErr, &actualTransactionErr)

				require.Equal(t, transactionErr.TransactionID, actualTransactionErr.TransactionID)
			})

			t.Run("Details", func(t *testing.T) {
				var actualTransactionErr *TransactionError
				require.ErrorAs(t, actualErr, &actualTransactionErr)

				require.Empty(t, actualTransactionErr.Details())
			})
		})
	}

	t.Run("Consistent behavior", func(t *testing.T) {
		grpcStatus, err := newGrpcStatus(codes.Aborted, "STATUS_MESSAGE")
		require.NoError(t, err)

		transactionErr := newTransactionError(grpcStatus.Err(), "TRANSACTION_ID")

		t.Run("EndorseError", func(t *testing.T) {
			actualErr := &EndorseError{transactionErr}

			var endorseErr *EndorseError
			require.ErrorAs(t, actualErr, &endorseErr)

			assert.Equal(t, transactionErr.TransactionID, endorseErr.TransactionID)
			assert.Empty(t, endorseErr.Details())
		})

		t.Run("SubmitError", func(t *testing.T) {
			actualErr := &SubmitError{transactionErr}

			var submitErr *SubmitError
			require.ErrorAs(t, actualErr, &submitErr)

			assert.Equal(t, transactionErr.TransactionID, submitErr.TransactionID)
			assert.Empty(t, submitErr.Details())
		})

		t.Run("CommitStatusError", func(t *testing.T) {
			actualErr := &CommitStatusError{transactionErr}

			var commitStatusErr *CommitStatusError
			require.ErrorAs(t, actualErr, &commitStatusErr)

			assert.Equal(t, transactionErr.TransactionID, commitStatusErr.TransactionID)
			assert.Empty(t, commitStatusErr.Details())
		})
	})
}
