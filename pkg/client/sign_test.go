// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	endorseResponse := AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network")

	t.Run("Evaluate signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(expected))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "contract", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).GetProposedTransaction().GetSignature()
		require.Equal(t, expected, actual)
	})

	t.Run("Submit signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(endorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "contract", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).GetProposedTransaction().GetSignature()
		require.Equal(t, expected, actual)
	})

	t.Run("Submit signs transaction using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(endorseResponse))
		requests := make(chan *gateway.SubmitRequest, 1)
		ExpectSubmit(mockConnection, CaptureInvokeRequest(requests))
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "contract", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).GetPreparedTransaction().GetSignature()
		require.Equal(t, expected, actual)
	})

	t.Run("Default error implementation is used if no signing implementation supplied", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)

		gateway, err := Connect(TestCredentials.Identity(), WithClientConnection(mockConnection))
		require.NoError(t, err)

		contract := gateway.GetNetwork("network").GetContract("chaincode")

		_, err = contract.EvaluateTransaction("transaction")
		require.Error(t, err)
	})
}
