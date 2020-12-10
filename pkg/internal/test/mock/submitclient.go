/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

import (
	"errors"
	"io"

	proto "github.com/hyperledger/fabric-gateway/protos"
)

// SubmitClient mock implementation whose method implementations can be overridden by assigning to properties
type SubmitClient struct {
	MockRecv func() (*proto.Event, error)
	ClientStream
}

// NewSubmitClient creates a mock
func NewSubmitClient() *SubmitClient {
	return &SubmitClient{
		MockRecv: func() (*proto.Event, error) {
			return nil, errors.New("Not implemented")
		},
		ClientStream: *NewClientStream(),
	}
}

// NewFailSubmitClient creates a mock that returns the supplied commit error
func NewFailSubmitClient(err error) *SubmitClient {
	submitClient := NewSubmitClient()
	submitClient.MockRecv = func() (*proto.Event, error) {
		return nil, err
	}
	return submitClient
}

// NewSuccessSubmitClient creates a mock that returns a successful commit
func NewSuccessSubmitClient() *SubmitClient {
	submitClient := NewSubmitClient()
	submitClient.MockRecv = func() (*proto.Event, error) {
		return nil, io.EOF
	}
	return submitClient
}

// Recv mock implementation
func (mock *SubmitClient) Recv() (*proto.Event, error) {
	return mock.MockRecv()
}
