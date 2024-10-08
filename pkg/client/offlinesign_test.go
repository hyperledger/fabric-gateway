// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"io"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
)

func TestOfflineSign(t *testing.T) {
	for testName, testCase := range map[string]struct {
		New         func(*signingTest)
		Sign        func(*signingTest, []byte)
		Invocations map[string](func(*signingTest) ([]byte, error))
		Recreate    func(*signingTest)
		State       func(*signingTest) any
	}{
		"Proposal": {
			New: func(s *signingTest) {
				s.NewProposal()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignProposal(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Evaluate": func(s *signingTest) ([]byte, error) {
					return s.Evaluate()
				},
				"Endorse": func(s *signingTest) ([]byte, error) {
					return s.Endorse()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateProposal()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest        []byte
					TransactionID string
					EndorsingOrgs []string
				}{
					Digest:        s.proposal.Digest(),
					TransactionID: s.proposal.TransactionID(),
					EndorsingOrgs: s.proposal.proposedTransaction.GetEndorsingOrganizations(),
				}
			},
		},
		"Transaction": {
			New: func(s *signingTest) {
				s.NewTransaction()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignTransaction(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Submit": func(s *signingTest) ([]byte, error) {
					return s.Submit()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateTransaction()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest        []byte
					TransactionID string
					Result        []byte
				}{
					Digest:        s.transaction.Digest(),
					TransactionID: s.transaction.TransactionID(),
					Result:        s.transaction.Result(),
				}
			},
		},
		"Commit": {
			New: func(s *signingTest) {
				s.NewCommit()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignCommit(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Status": func(s *signingTest) ([]byte, error) {
					return s.CommitStatus()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateCommit()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest        []byte
					TransactionID string
				}{
					Digest:        s.transaction.Digest(),
					TransactionID: s.transaction.TransactionID(),
				}
			},
		},
		"ChaincodeEvents": {
			New: func(s *signingTest) {
				s.NewChaincodeEvents()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignChaincodeEvents(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Events": func(s *signingTest) ([]byte, error) {
					return s.ChaincodeEvents()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateChaincodeEvents()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest []byte
				}{
					Digest: s.chaincodeEvents.Digest(),
				}
			},
		},
		"BlockEvents": {
			New: func(s *signingTest) {
				s.NewBlockEvents()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignBlockEvents(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Events": func(s *signingTest) ([]byte, error) {
					return s.BlockEvents()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateBlockEvents()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest []byte
				}{
					Digest: s.blockEvents.Digest(),
				}
			},
		},
		"FilteredBlockEvents": {
			New: func(s *signingTest) {
				s.NewFilteredBlockEvents()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignFilteredBlockEvents(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Events": func(s *signingTest) ([]byte, error) {
					return s.FilteredBlockEvents()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateFilteredBlockEvents()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest []byte
				}{
					Digest: s.filteredBlockEvents.Digest(),
				}
			},
		},
		"BlockAndPrivateDataEvents": {
			New: func(s *signingTest) {
				s.NewBlockAndPrivateDataEvents()
			},
			Sign: func(s *signingTest, signature []byte) {
				s.SignBlockAndPrivateDataEvents(signature)
			},
			Invocations: map[string](func(*signingTest) ([]byte, error)){
				"Events": func(s *signingTest) ([]byte, error) {
					return s.BlockAndPrivateDataEvents()
				},
			},
			Recreate: func(s *signingTest) {
				s.RecreateBlockAndPrivateDataEvents()
			},
			State: func(s *signingTest) any {
				return struct {
					Digest []byte
				}{
					Digest: s.blockAndPrivateDataEvents.Digest(),
				}
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			for invocationName, invoke := range testCase.Invocations {
				t.Run(invocationName, func(t *testing.T) {
					t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
						s := NewSigningTest(t)
						testCase.New(s)
						_, err := invoke(s)
						require.Error(t, err)
					})

					t.Run("Uses off-line signature", func(t *testing.T) {
						expected := []byte("SIGNATURE")
						s := NewSigningTest(t)

						testCase.New(s)
						testCase.Sign(s, expected)
						actual, err := invoke(s)
						require.NoError(t, err)

						require.Equal(t, expected, actual)
					})

					t.Run("retains signature", func(t *testing.T) {
						expected := []byte("SIGNATURE")
						s := NewSigningTest(t)

						testCase.New(s)
						testCase.Sign(s, expected)

						testCase.Recreate(s)
						actual, err := invoke(s)
						require.NoError(t, err)

						require.Equal(t, expected, actual)
					})
				})
			}

			t.Run("Retains state after signing", func(t *testing.T) {
				s := NewSigningTest(t)

				testCase.New(s)
				expected := testCase.State(s)

				testCase.Sign(s, []byte("SIGNATURE"))
				actual := testCase.State(s)

				require.Equal(t, expected, actual)
			})
		})
	}
}

type signingTest struct {
	t                         *testing.T
	mockConnection            *MockClientConnInterface
	gateway                   *Gateway
	proposal                  *Proposal
	transaction               *Transaction
	commit                    *Commit
	chaincodeEvents           *ChaincodeEventsRequest
	blockEvents               *BlockEventsRequest
	filteredBlockEvents       *FilteredBlockEventsRequest
	blockAndPrivateDataEvents *BlockAndPrivateDataEventsRequest
}

func NewSigningTest(t *testing.T) *signingTest {
	mockConnection := NewMockClientConnInterface(t)
	gateway, err := Connect(TestCredentials.Identity(), WithClientConnection(mockConnection))
	require.NoError(t, err, "Connect")

	return &signingTest{
		t:              t,
		mockConnection: mockConnection,
		gateway:        gateway,
	}
}

func (s *signingTest) NewProposal() {
	result, err := s.gateway.GetNetwork("channel").GetContract("chaincode").NewProposal("transaction")
	require.NoError(s.t, err, "NewProposal")
	s.proposal = result
}

func (s *signingTest) SignProposal(signature []byte) {
	bytes := s.getBytes(s.proposal)
	result, err := s.gateway.NewSignedProposal(bytes, signature)
	require.NoError(s.t, err, "NewSignedProposal")
	s.proposal = result
}

func (s *signingTest) getBytes(serializable interface {
	Bytes() ([]byte, error)
}) []byte {
	bytes, err := serializable.Bytes()
	require.NoError(s.t, err, "Bytes")
	return bytes
}

func (s *signingTest) RecreateProposal() {
	bytes := s.getBytes(s.proposal)
	result, err := s.gateway.NewProposal(bytes)
	require.NoError(s.t, err, "NewProposal")
	s.proposal = result
}

func (s *signingTest) Evaluate() ([]byte, error) {
	requests := make(chan *gateway.EvaluateRequest, 1)
	ExpectEvaluate(s.mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil)).Maybe()

	_, err := s.proposal.Evaluate()
	if err != nil {
		return nil, err
	}

	return (<-requests).GetProposedTransaction().GetSignature(), nil
}

func (s *signingTest) endorse() (*Transaction, []byte, error) {
	requests := make(chan *gateway.EndorseRequest, 1)
	response := AssertNewEndorseResponse(s.t, "TRANSACTION_RESULT", "network")
	ExpectEndorse(s.mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(response)).Maybe()

	transaction, err := s.proposal.Endorse()
	if err != nil {
		return nil, nil, err
	}

	return transaction, (<-requests).ProposedTransaction.GetSignature(), nil
}

func (s *signingTest) Endorse() ([]byte, error) {
	_, signature, err := s.endorse()
	return signature, err
}

func (s *signingTest) NewTransaction() {
	s.NewProposal()
	s.SignProposal([]byte("SIGNATURE"))
	transaction, _, err := s.endorse()
	require.NoError(s.t, err, "Endorse")

	s.transaction = transaction
}

func (s *signingTest) SignTransaction(signature []byte) {
	bytes := s.getBytes(s.transaction)
	result, err := s.gateway.NewSignedTransaction(bytes, signature)
	require.NoError(s.t, err, "NewSignedTransaction")

	s.transaction = result
}

func (s *signingTest) RecreateTransaction() {
	bytes := s.getBytes(s.transaction)
	result, err := s.gateway.NewTransaction(bytes)
	require.NoError(s.t, err, "NewTransaction")

	s.transaction = result
}

func (s *signingTest) submit() (*Commit, []byte, error) {
	requests := make(chan *gateway.SubmitRequest, 1)
	ExpectSubmit(s.mockConnection, CaptureInvokeRequest(requests)).Maybe()

	commit, err := s.transaction.Submit()
	if err != nil {
		return nil, nil, err
	}

	return commit, (<-requests).GetPreparedTransaction().GetSignature(), nil
}

func (s *signingTest) Submit() ([]byte, error) {
	_, signature, err := s.submit()
	return signature, err
}

func (s *signingTest) NewCommit() {
	s.NewTransaction()
	s.SignTransaction([]byte("SIGNATURE"))
	commit, _, err := s.submit()
	require.NoError(s.t, err, "Submit")

	s.commit = commit
}

func (s *signingTest) SignCommit(signature []byte) {
	bytes := s.getBytes(s.commit)
	result, err := s.gateway.NewSignedCommit(bytes, signature)
	require.NoError(s.t, err, "NewSignedCommit")

	s.commit = result
}

func (s *signingTest) RecreateCommit() {
	bytes := s.getBytes(s.commit)
	result, err := s.gateway.NewCommit(bytes)
	require.NoError(s.t, err, "NewCommit")
	s.commit = result
}

func (s *signingTest) CommitStatus() ([]byte, error) {
	requests := make(chan *gateway.SignedCommitStatusRequest, 1)
	ExpectCommitStatus(s.mockConnection, CaptureInvokeRequest(requests), WithCommitStatusResponse(peer.TxValidationCode_VALID, 1)).Maybe()

	_, err := s.commit.Status()
	if err != nil {
		return nil, err
	}

	return (<-requests).GetSignature(), nil
}

func (s *signingTest) NewChaincodeEvents() {
	result, err := s.gateway.GetNetwork("channel").NewChaincodeEventsRequest("chaincode")
	require.NoError(s.t, err, "NewChaincodeEventsRequest")
	s.chaincodeEvents = result
}

func (s *signingTest) SignChaincodeEvents(signature []byte) {
	bytes := s.getBytes(s.chaincodeEvents)
	result, err := s.gateway.NewSignedChaincodeEventsRequest(bytes, signature)
	require.NoError(s.t, err, "NewSignedChaincodeEventsRequest")
	s.chaincodeEvents = result
}

func (s *signingTest) RecreateChaincodeEvents() {
	bytes := s.getBytes(s.chaincodeEvents)
	result, err := s.gateway.NewChaincodeEventsRequest(bytes)
	require.NoError(s.t, err, "NewChaincodeEventsRequest")
	s.chaincodeEvents = result
}

func (s *signingTest) ChaincodeEvents() ([]byte, error) {
	mockStream := NewMockClientStream(s.t)
	ExpectChaincodeEvents(s.mockConnection, WithNewStreamResult(mockStream)).Maybe()

	messages := make(chan *gateway.SignedChaincodeEventsRequest, 1)
	ExpectSendMsg(mockStream, CaptureSendMsg(messages)).Maybe()
	mockStream.EXPECT().CloseSend().Return(nil).Maybe()
	ExpectRecvMsg(mockStream).Return(io.EOF).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := s.chaincodeEvents.Events(ctx)
	if err != nil {
		return nil, err
	}

	return (<-messages).GetSignature(), nil
}

func (s *signingTest) NewBlockEvents() {
	result, err := s.gateway.GetNetwork("channel").NewBlockEventsRequest()
	require.NoError(s.t, err, "NewBlockEventsRequest")
	s.blockEvents = result
}

func (s *signingTest) SignBlockEvents(signature []byte) {
	bytes := s.getBytes(s.blockEvents)
	result, err := s.gateway.NewSignedBlockEventsRequest(bytes, signature)
	require.NoError(s.t, err, "NewSignedBlockEventsRequest")
	s.blockEvents = result
}

func (s *signingTest) RecreateBlockEvents() {
	bytes := s.getBytes(s.blockEvents)
	result, err := s.gateway.NewBlockEventsRequest(bytes)
	require.NoError(s.t, err, "NewBlockEventsRequest")
	s.blockEvents = result
}

func (s *signingTest) BlockEvents() ([]byte, error) {
	return s.deliverEvents(func(ctx context.Context) error {
		_, err := s.blockEvents.Events(ctx)
		return err
	})
}

func (s *signingTest) deliverEvents(invoke func(context.Context) error) ([]byte, error) {
	mockStream := NewMockClientStream(s.t)
	ExpectDeliver(s.mockConnection, WithNewStreamResult(mockStream)).Maybe()
	ExpectDeliverFiltered(s.mockConnection, WithNewStreamResult(mockStream)).Maybe()
	ExpectDeliverWithPrivateData(s.mockConnection, WithNewStreamResult(mockStream)).Maybe()

	messages := make(chan *common.Envelope, 1)
	ExpectSendMsg(mockStream, CaptureSendMsg(messages)).Maybe()
	mockStream.EXPECT().CloseSend().Return(nil).Maybe()
	ExpectRecvMsg(mockStream).Return(io.EOF).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := invoke(ctx)
	if err != nil {
		return nil, err
	}

	return (<-messages).GetSignature(), nil
}

func (s *signingTest) NewFilteredBlockEvents() {
	result, err := s.gateway.GetNetwork("channel").NewFilteredBlockEventsRequest()
	require.NoError(s.t, err, "NewFilteredBlockEventsRequest")
	s.filteredBlockEvents = result
}

func (s *signingTest) SignFilteredBlockEvents(signature []byte) {
	bytes := s.getBytes(s.filteredBlockEvents)
	result, err := s.gateway.NewSignedFilteredBlockEventsRequest(bytes, signature)
	require.NoError(s.t, err, "NewSignedFilteredBlockEventsRequest")
	s.filteredBlockEvents = result
}

func (s *signingTest) RecreateFilteredBlockEvents() {
	bytes := s.getBytes(s.filteredBlockEvents)
	result, err := s.gateway.NewFilteredBlockEventsRequest(bytes)
	require.NoError(s.t, err, "NewFilteredBlockEventsRequest")
	s.filteredBlockEvents = result
}

func (s *signingTest) FilteredBlockEvents() ([]byte, error) {
	return s.deliverEvents(func(ctx context.Context) error {
		_, err := s.filteredBlockEvents.Events(ctx)
		return err
	})
}

func (s *signingTest) NewBlockAndPrivateDataEvents() {
	result, err := s.gateway.GetNetwork("channel").NewBlockAndPrivateDataEventsRequest()
	require.NoError(s.t, err, "NewBlockAndPrivateDataEventsRequest")
	s.blockAndPrivateDataEvents = result
}

func (s *signingTest) SignBlockAndPrivateDataEvents(signature []byte) {
	bytes := s.getBytes(s.blockAndPrivateDataEvents)
	result, err := s.gateway.NewSignedBlockAndPrivateDataEventsRequest(bytes, signature)
	require.NoError(s.t, err, "NewSignedBlockAndPrivateDataEventsRequest")
	s.blockAndPrivateDataEvents = result
}

func (s *signingTest) RecreateBlockAndPrivateDataEvents() {
	bytes := s.getBytes(s.blockAndPrivateDataEvents)
	result, err := s.gateway.NewBlockAndPrivateDataEventsRequest(bytes)
	require.NoError(s.t, err, "NewBlockAndPrivateDataEventsRequest")
	s.blockAndPrivateDataEvents = result
}

func (s *signingTest) BlockAndPrivateDataEvents() ([]byte, error) {
	return s.deliverEvents(func(ctx context.Context) error {
		_, err := s.blockAndPrivateDataEvents.Events(ctx)
		return err
	})
}
