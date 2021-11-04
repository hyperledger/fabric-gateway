/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

type chaincodeEventsBuilder struct {
	client        *gatewayClient
	signingID     *signingIdentity
	channelName   string
	chaincodeName string
	startPosition *orderer.SeekPosition
}

func (builder *chaincodeEventsBuilder) build() (*ChaincodeEventsRequest, error) {
	signedRequest, err := builder.newSignedChaincodeEventsRequestProto()
	if err != nil {
		return nil, err
	}

	result := &ChaincodeEventsRequest{
		client:        builder.client,
		signingID:     builder.signingID,
		signedRequest: signedRequest,
	}
	return result, nil
}

func (builder *chaincodeEventsBuilder) newSignedChaincodeEventsRequestProto() (*gateway.SignedChaincodeEventsRequest, error) {
	request, err := builder.newChaincodeEventsRequestProto()
	if err != nil {
		return nil, err
	}

	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize chaincode events request: %w", err)
	}

	signedRequest := &gateway.SignedChaincodeEventsRequest{
		Request: requestBytes,
	}
	return signedRequest, nil
}

func (builder *chaincodeEventsBuilder) newChaincodeEventsRequestProto() (*gateway.ChaincodeEventsRequest, error) {
	creator, err := builder.signingID.Creator()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize identity: %w", err)
	}

	request := &gateway.ChaincodeEventsRequest{
		ChannelId:     builder.channelName,
		Identity:      creator,
		ChaincodeId:   builder.chaincodeName,
		StartPosition: builder.getStartPosition(),
	}
	return request, nil
}

func (builder *chaincodeEventsBuilder) getStartPosition() *orderer.SeekPosition {
	if builder.startPosition != nil {
		return builder.startPosition
	}

	return &orderer.SeekPosition{
		Type: &orderer.SeekPosition_NextCommit{
			NextCommit: &orderer.SeekNextCommit{},
		},
	}
}

// ChaincodeEventsOption implements an option for a chaincode events request.
type ChaincodeEventsOption = func(builder *chaincodeEventsBuilder) error

// WithStartBlock reads chaincode events starting at the specified block number.
func WithStartBlock(blockNumber uint64) ChaincodeEventsOption {
	return func(builder *chaincodeEventsBuilder) error {
		builder.startPosition = &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: blockNumber,
				},
			},
		}
		return nil
	}
}
