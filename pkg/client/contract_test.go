/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

func AssertNewTestContract(t *testing.T, client gateway.GatewayClient, contractName string) *Contract {
	network := AssertNewTestNetwork(t, client, "network")
	return network.GetContract(contractName)
}

func AssertUnmarshall(t *testing.T, b []byte, m proto.Message) {
	if err := proto.Unmarshal(b, m); err != nil {
		t.Fatal(err)
	}
}

func bytesAsStrings(bytes [][]byte) []string {
	results := make([]string, 0, len(bytes))

	for _, v := range bytes {
		results = append(results, string(v))
	}

	return results
}

func TestContract(t *testing.T) {
	t.Run("EvaluateTransaction", func(t *testing.T) {
		t.Run("Returns gRPC invocation error", func(t *testing.T) {
			expectedError := "EVALUATE_ERROR"
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				return nil, errors.New(expectedError)
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.EvaluateTransaction("transaction")

			if nil == err || !strings.Contains(err.Error(), expectedError) {
				t.Fatalf("Expected error containing %s, got %v", expectedError, err)
			}
		})

		t.Run("Returns result", func(t *testing.T) {
			expected := []byte("TRANSACTION_RESULT")
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				value := &gateway.Result{
					Value: expected,
				}
				return value, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			actual, err := contract.EvaluateTransaction("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(actual, expected) != 0 {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Includes channel name in proposal", func(t *testing.T) {
			var actual string
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				proposal := &peer.Proposal{}
				AssertUnmarshall(t, in.Proposal.ProposalBytes, proposal)

				header := &common.Header{}
				AssertUnmarshall(t, proposal.Header, header)

				channelHeader := &common.ChannelHeader{}
				AssertUnmarshall(t, header.ChannelHeader, channelHeader)

				actual = channelHeader.ChannelId

				value := &gateway.Result{}
				return value, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.EvaluateTransaction("transaction")
			if err != nil {
				t.Fatal(err)
			}

			expected := contract.network.name
			if actual != expected {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Includes chaincode name in proposal", func(t *testing.T) {
			var actual string
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				proposal := &peer.Proposal{}
				AssertUnmarshall(t, in.Proposal.ProposalBytes, proposal)

				payload := &peer.ChaincodeProposalPayload{}
				AssertUnmarshall(t, proposal.Payload, payload)

				input := &peer.ChaincodeInvocationSpec{}
				AssertUnmarshall(t, payload.Input, input)

				actual = input.ChaincodeSpec.ChaincodeId.Name

				value := &gateway.Result{}
				return value, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.EvaluateTransaction("transaction")
			if err != nil {
				t.Fatal(err)
			}

			expected := contract.name
			if actual != expected {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Includes transaction name in proposal", func(t *testing.T) {
			var args [][]byte
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				proposal := &peer.Proposal{}
				AssertUnmarshall(t, in.Proposal.ProposalBytes, proposal)

				payload := &peer.ChaincodeProposalPayload{}
				AssertUnmarshall(t, proposal.Payload, payload)

				input := &peer.ChaincodeInvocationSpec{}
				AssertUnmarshall(t, payload.Input, input)

				args = input.ChaincodeSpec.Input.Args

				value := &gateway.Result{}
				return value, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			expected := "TRANSACTION_NAME"
			_, err := contract.EvaluateTransaction(expected)
			if err != nil {
				t.Fatal(err)
			}

			actual := string(args[0])
			if actual != expected {
				t.Fatalf("Expected Args[0] to be %s, got Args: %s", expected, args)
			}
		})

		t.Run("Includes arguments in proposal", func(t *testing.T) {
			var args [][]byte
			mockClient := mock.NewGatewayClient()
			mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
				proposal := &peer.Proposal{}
				AssertUnmarshall(t, in.Proposal.ProposalBytes, proposal)

				payload := &peer.ChaincodeProposalPayload{}
				AssertUnmarshall(t, proposal.Payload, payload)

				input := &peer.ChaincodeInvocationSpec{}
				AssertUnmarshall(t, payload.Input, input)

				args = input.ChaincodeSpec.Input.Args

				value := &gateway.Result{}
				return value, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			expected := []string{"one", "two", "three"}
			_, err := contract.EvaluateTransaction("transaction", expected...)
			if err != nil {
				t.Fatal(err)
			}

			actual := bytesAsStrings(args[1:])
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("Expected Args[1:] to be %s, got Args: %s", expected, args)
			}
		})
	})

	t.Run("SubmitTransaction", func(t *testing.T) {
		t.Run("Returns endorsement error", func(t *testing.T) {
			expectedError := "ENDORSE_ERROR"
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				return nil, errors.New(expectedError)
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.SubmitTransaction("transaction")

			if nil == err || !strings.Contains(err.Error(), expectedError) {
				t.Fatalf("Expected error containing %s, got %v", expectedError, err)
			}
		})

		t.Run("Returns error sending to orderer", func(t *testing.T) {
			expectedError := "SUBMIT_ERROR"
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				preparedTransaction := &gateway.PreparedTransaction{
					Envelope: &common.Envelope{},
					Response: &gateway.Result{
						Value: []byte("TRANSACTION_RESULT"),
					},
				}
				return preparedTransaction, nil
			}
			mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
				return nil, errors.New(expectedError)
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.SubmitTransaction("transaction")

			if nil == err || !strings.Contains(err.Error(), expectedError) {
				t.Fatalf("Expected error containing %s, got %v", expectedError, err)
			}
		})

		t.Run("Returns commit error", func(t *testing.T) {
			expectedError := "COMMIT_ERROR"
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				preparedTransaction := &gateway.PreparedTransaction{
					Envelope: &common.Envelope{},
					Response: &gateway.Result{
						Value: []byte("TRANSACTION_RESULT"),
					},
				}
				return preparedTransaction, nil
			}
			mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
				submitClient := mock.NewSubmitClient()
				submitClient.MockRecv = func() (*gateway.Event, error) {
					return nil, errors.New(expectedError)
				}
				return submitClient, nil
			}
			contract := AssertNewTestContract(t, mockClient, "contract")

			_, err := contract.SubmitTransaction("transaction")

			if nil == err || !strings.Contains(err.Error(), expectedError) {
				t.Fatalf("Expected error containing %s, got %v", expectedError, err)
			}
		})

		t.Run("Returns result for committed transaction", func(t *testing.T) {
			expected := []byte("TRANSACTION_RESULT")
			mockClient := mock.NewGatewayClient()
			mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
				preparedTransaction := &gateway.PreparedTransaction{
					Envelope: &common.Envelope{},
					Response: &gateway.Result{
						Value: expected,
					},
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
			contract := AssertNewTestContract(t, mockClient, "contract")

			actual, err := contract.SubmitTransaction("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(actual, expected) != 0 {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})
}
