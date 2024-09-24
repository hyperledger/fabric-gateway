// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIdentity(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	require.NoError(t, err)

	certificate, err := test.NewCertificate(privateKey)
	require.NoError(t, err)

	id, err := identity.NewX509Identity("MSP_ID", certificate)
	require.NoError(t, err)

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID(),
		IdBytes: id.Credentials(),
	}
	creator, err := proto.Marshal(serializedIdentity)
	require.NoError(t, err)

	t.Run("Evaluate uses client identity for proposals", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))

		contract := AssertNewTestContract(t, "contract", WithClientConnection(mockConnection), WithIdentity(id))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalSignatureHeader(t, (<-requests).ProposedTransaction).Creator
		require.EqualValues(t, creator, actual)
	})

	t.Run("Submit uses client identity for proposals", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		endorseResponse := AssertNewEndorseResponse(t, "result", "channel")
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(endorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "contract", WithClientConnection(mockConnection), WithIdentity(id))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalSignatureHeader(t, (<-requests).ProposedTransaction).Creator
		require.EqualValues(t, creator, actual)
	})
}
