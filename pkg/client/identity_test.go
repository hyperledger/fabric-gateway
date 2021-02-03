/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/msp"
	"google.golang.org/grpc"
)

func TestIdentity(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	id, err := identity.NewX509Identity("MSP_ID", certificate)
	if err != nil {
		t.Fatal(err)
	}

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID(),
		IdBytes: id.Credentials(),
	}
	creator, err := proto.Marshal(serializedIdentity)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Evaluate uses client identity for proposals", func(t *testing.T) {
		var actual []byte
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallSignatureHeader(t, in).Creator
			}).
			Return(&gateway.Result{}, nil).
			Times(1)

		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithIdentity(id))

		if _, err := contract.EvaluateTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, creator) {
			t.Fatalf("Expected identity: %v\nGot: %v", creator, actual)
		}
	})

	t.Run("Submit uses client identity for proposals", func(t *testing.T) {
		var actual []byte
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		preparedTransaction := gateway.PreparedTransaction{
			Envelope: &common.Envelope{},
			Response: &gateway.Result{},
		}
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallSignatureHeader(t, in).Creator
			}).
			Return(&preparedTransaction, nil).
			Times(1)
		mockSubmitClient := NewMockGateway_SubmitClient(mockController)
		mockSubmitClient.EXPECT().Recv().Return(nil, io.EOF)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).Return(mockSubmitClient, nil)

		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithIdentity(id))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, creator) {
			t.Fatalf("Expected identity: %v\nGot: %v", creator, actual)
		}
	})
}
