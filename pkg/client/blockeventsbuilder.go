/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"math"

	"github.com/hyperledger/fabric-gateway/pkg/internal/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type blockEventsBuilder struct {
	eventsBuilder *eventsBuilder
}

func seekLargestBlockNumber() *orderer.SeekPosition {
	return &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}
}

func (builder *blockEventsBuilder) payloadBytes() ([]byte, error) {
	channelHeader, err := builder.channelHeaderBytes()
	if err != nil {
		return nil, err
	}

	signatureHeader, err := builder.signatureHeaderBytes()
	if err != nil {
		return nil, err
	}

	data, err := builder.dataBytes()
	if err != nil {
		return nil, err
	}

	payload := &common.Payload{
		Header: &common.Header{
			ChannelHeader:   channelHeader,
			SignatureHeader: signatureHeader,
		},
		Data: data,
	}

	return util.Marshal(payload)
}

func (builder *blockEventsBuilder) channelHeaderBytes() ([]byte, error) {
	channelHeader := &common.ChannelHeader{
		Type:      int32(common.HeaderType_DELIVER_SEEK_INFO),
		Timestamp: timestamppb.Now(),
		ChannelId: builder.eventsBuilder.channelName,
		Epoch:     0,
	}

	return util.Marshal(channelHeader)
}

func (builder *blockEventsBuilder) signatureHeaderBytes() ([]byte, error) {
	creator, err := builder.eventsBuilder.signingID.Creator()
	if err != nil {
		return nil, err
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nil,
	}

	return util.Marshal(signatureHeader)
}

func (builder *blockEventsBuilder) dataBytes() ([]byte, error) {
	data := &orderer.SeekInfo{
		Start: builder.eventsBuilder.getStartPosition(),
		Stop:  seekLargestBlockNumber(),
	}

	return util.Marshal(data)
}

type filteredBlockEventsBuilder struct {
	blockBuilder *blockEventsBuilder
}

func (builder *filteredBlockEventsBuilder) build() (*FilteredBlockEventsRequest, error) {
	payload, err := builder.blockBuilder.payloadBytes()
	if err != nil {
		return nil, err
	}

	signedRequest := &common.Envelope{
		Payload: payload,
	}

	result := &FilteredBlockEventsRequest{
		blockEventsRequest: &blockEventsRequest{
			client:        builder.blockBuilder.eventsBuilder.client,
			signingID:     builder.blockBuilder.eventsBuilder.signingID,
			signedRequest: signedRequest,
		},
	}
	return result, nil
}

type fullBlockEventsBuilder struct {
	blockBuilder *blockEventsBuilder
}

func (builder *fullBlockEventsBuilder) build() (*FullBlockEventsRequest, error) {
	payload, err := builder.blockBuilder.payloadBytes()
	if err != nil {
		return nil, err
	}

	signedRequest := &common.Envelope{
		Payload: payload,
	}

	result := &FullBlockEventsRequest{
		blockEventsRequest: &blockEventsRequest{
			client:        builder.blockBuilder.eventsBuilder.client,
			signingID:     builder.blockBuilder.eventsBuilder.signingID,
			signedRequest: signedRequest,
		},
	}
	return result, nil
}

type blockEventsWithPrivateDataBuilder struct {
	blockBuilder *blockEventsBuilder
}

func (builder *blockEventsWithPrivateDataBuilder) build() (*BlockEventsWithPrivateDataRequest, error) {
	payload, err := builder.blockBuilder.payloadBytes()
	if err != nil {
		return nil, err
	}

	signedRequest := &common.Envelope{
		Payload: payload,
	}

	result := &BlockEventsWithPrivateDataRequest{
		blockEventsRequest: &blockEventsRequest{
			client:        builder.blockBuilder.eventsBuilder.client,
			signingID:     builder.blockBuilder.eventsBuilder.signingID,
			signedRequest: signedRequest,
		},
	}
	return result, nil
}
