/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
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

	t.Run("Evaluate", func(t *testing.T) {
		t.Run("Returns error with signer and no explicit signing", func(t *testing.T) {
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				return &gateway.Result{}, nil
			}
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
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				actual = in.Proposal.Signature
				return &gateway.Result{}, nil
			}
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
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				return newPreparedTransaction("result"), nil
			}
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
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				actual = in.Proposal.Signature
				return newPreparedTransaction("result"), nil
			}
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
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				return newPreparedTransaction("result"), nil
			}
			mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
				return mock.NewSuccessSubmitClient(), nil
			}
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
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				return newPreparedTransaction("result"), nil
			}
			mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
				actual = in.Envelope.Signature
				return mock.NewSuccessSubmitClient(), nil
			}
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
				mockClient := mock.NewGatewayClient()
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
				mockClient := mock.NewGatewayClient()
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
				mockClient := mock.NewGatewayClient()
				mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
					return newPreparedTransaction("result"), nil
				}
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
