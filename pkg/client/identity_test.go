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

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = test.AssertUnmarshallSignatureHeader(t, in).Creator
			value := &gateway.Result{}
			return value, nil
		}
		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithIdentity(id))

		if _, err := contract.EvaluateTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, creator) != 0 {
			t.Fatalf("Expected identity: %v\nGot: %v", creator, actual)
		}
	})

	t.Run("Submit uses client identity for proposals", func(t *testing.T) {
		var actual []byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			actual = test.AssertUnmarshallSignatureHeader(t, in).Creator
			preparedTransaction := &gateway.PreparedTransaction{
				Envelope: &common.Envelope{},
				Response: &gateway.Result{},
			}
			return preparedTransaction, nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			submitClient := mock.NewSubmitClient()
			submitClient.MockRecv = func() (*gateway.Event, error) {
				return nil, io.EOF
			}
			return submitClient, nil
		}
		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithIdentity(id))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, creator) != 0 {
			t.Fatalf("Expected identity: %v\nGot: %v", creator, actual)
		}
	})
}
