/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestSign(t *testing.T) {
	evaluateResponse := &gateway.EvaluateResponse{
		Result: &peer.Response{
			Payload: nil,
		},
	}

	statusResponse := &gateway.CommitStatusResponse{
		Result: peer.TxValidationCode_VALID,
	}

	t.Run("Evaluate signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = in.ProposedTransaction.Signature
			}).
			Return(evaluateResponse, nil).
			Times(1)

		contract := AssertNewTestContract(t, "contract", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
	})

	t.Run("Submit signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.ProposedTransaction.Signature
			}).
			Return(AssertNewEndorseResponse(t, "result", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(statusResponse, nil)

		contract := AssertNewTestContract(t, "contract", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
	})

	t.Run("Submit signs transaction using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "result", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
				actual = in.PreparedTransaction.Signature
			}).
			Return(nil, nil).
			Times(1)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(statusResponse, nil)

		contract := AssertNewTestContract(t, "contract", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
	})

	t.Run("Default error implementation is used if no signing implementation supplied", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Return(evaluateResponse, nil).
			AnyTimes()

		mockDeliver := NewMockDeliverClient(gomock.NewController(t))

		gateway, err := Connect(TestCredentials.Identity(), WithGatewayClient(mockClient), WithDeliverClient(mockDeliver))
		require.NoError(t, err)

		contract := gateway.GetNetwork("network").GetContract("chaincode")

		_, err = contract.EvaluateTransaction("transaction")
		require.Error(t, err)
	})
}
