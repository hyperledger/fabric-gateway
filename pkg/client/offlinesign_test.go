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
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
)

func TestOfflineSign(t *testing.T) {
	newContractWithNoSign := func(t *testing.T, options ...ConnectOption) *Contract {
		id, _ := GetTestCredentials()
		gateway, err := Connect(id, options...)
		if err != nil {
			t.Fatal(err)
		}

		return gateway.GetNetwork("network").GetContract("chaincode")
	}

	newPreparedTransaction := func(value string) *gateway.PreparedTransaction {
		return &gateway.PreparedTransaction{
			Envelope: &common.Envelope{},
			Response: &gateway.Result{
				Value: []byte(value),
			},
		}
	}

	newMockSubmitClient := func(controller *gomock.Controller) *MockGateway_SubmitClient {
		mock := NewMockGateway_SubmitClient(controller)
		mock.EXPECT().Recv().
			Return(nil, io.EOF).
			AnyTimes()
		return mock
	}

	t.Run("Evaluate", func(t *testing.T) {
		t.Run("Returns error with signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Return(&gateway.Result{}, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if _, err := proposal.Evaluate(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
					actual = in.Proposal.Signature
				}).
				Return(&gateway.Result{}, nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedProposal.Evaluate(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Endorse", func(t *testing.T) {
		t.Run("Returns error with signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newPreparedTransaction("result"), nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if _, err := proposal.Endorse(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
					actual = in.Proposal.Signature
				}).
				Return(newPreparedTransaction("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedProposal.Endorse(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Submit", func(t *testing.T) {
		t.Run("Returns error with signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newPreparedTransaction("result"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(newMockSubmitClient(mockController), nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			transaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			if _, err := transaction.Submit(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newPreparedTransaction("result"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.PreparedTransaction, _ ...grpc.CallOption) {
					actual = in.Envelope.Signature
				}).
				Return(newMockSubmitClient(mockController), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedTransaction.Submit(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Serialization", func(t *testing.T) {
			t.Run("Proposal keeps same digest", func(t *testing.T) {
				mockController := gomock.NewController(t)
				defer mockController.Finish()

				mockClient := NewMockGatewayClient(mockController)
				contract := newContractWithNoSign(t, WithClient(mockClient))

				unsignedProposal, err := contract.NewProposal("transaction")
				if err != nil {
					t.Fatal(err)
				}

				expected, err := unsignedProposal.Digest()
				if err != nil {
					t.Fatal(err)
				}

				proposalBytes, err := unsignedProposal.Bytes()
				if err != nil {
					t.Fatal(err)
				}

				signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
				if err != nil {
					t.Fatal(err)
				}

				actual, err := signedProposal.Digest()
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(actual, expected) {
					t.Fatalf("Expected %s, got %s", expected, actual)
				}
			})

			t.Run("Proposal keeps same transaction ID", func(t *testing.T) {
				mockController := gomock.NewController(t)
				defer mockController.Finish()

				mockClient := NewMockGatewayClient(mockController)
				contract := newContractWithNoSign(t, WithClient(mockClient))

				unsignedProposal, err := contract.NewProposal("transaction")
				if err != nil {
					t.Fatal(err)
				}

				expected := unsignedProposal.TransactionID()

				proposalBytes, err := unsignedProposal.Bytes()
				if err != nil {
					t.Fatal(err)
				}

				signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
				if err != nil {
					t.Fatal(err)
				}

				actual := signedProposal.TransactionID()

				if actual != expected {
					t.Fatalf("Expected %s, got %s", expected, actual)
				}
			})

			t.Run("Transaction keeps same digest", func(t *testing.T) {
				mockController := gomock.NewController(t)
				defer mockController.Finish()

				mockClient := NewMockGatewayClient(mockController)
				mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
					Return(newPreparedTransaction("result"), nil).
					Times(1)

				contract := newContractWithNoSign(t, WithClient(mockClient))

				unsignedProposal, err := contract.NewProposal("transaction")
				if err != nil {
					t.Fatal(err)
				}

				proposalBytes, err := unsignedProposal.Bytes()
				if err != nil {
					t.Fatal(err)
				}

				signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
				if err != nil {
					t.Fatal(err)
				}

				unsignedTransaction, err := signedProposal.Endorse()
				if err != nil {
					t.Fatal(err)
				}

				expected, err := unsignedTransaction.Digest()
				if err != nil {
					t.Fatal(err)
				}

				transactionBytes, err := unsignedTransaction.Bytes()
				if err != nil {
					t.Fatal(err)
				}

				signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
				if err != nil {
					t.Fatal(err)
				}

				actual, err := signedTransaction.Digest()
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(actual, expected) {
					t.Fatalf("Expected %s, got %s", expected, actual)
				}
			})
		})
	})
}
