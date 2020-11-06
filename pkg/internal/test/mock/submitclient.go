/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

import (
	"errors"

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

// Recv mock implementation
func (mock *SubmitClient) Recv() (*proto.Event, error) {
	return mock.MockRecv()
}
