/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

// ClientStream mock implementation whose method implementations can be overridden by assigning to properties
type ClientStream struct {
	MockHeader    func() (metadata.MD, error)
	MockTrailer   func() metadata.MD
	MockCloseSend func() error
	MockContext   func() context.Context
	MockSendMsg   func(m interface{}) error
	MockRecvMsg   func(m interface{}) error
}

// NewClientStream creates a mock
func NewClientStream() *ClientStream {
	return &ClientStream{
		MockHeader: func() (metadata.MD, error) {
			return nil, errors.New("Not implemented")
		},
		MockTrailer: func() metadata.MD {
			return nil
		},
		MockCloseSend: func() error {
			return errors.New("Not implemented")
		},
		MockContext: func() context.Context {
			return nil
		},
		MockSendMsg: func(m interface{}) error {
			return errors.New("Not implemented")
		},
		MockRecvMsg: func(m interface{}) error {
			return errors.New("Not implemented")
		},
	}
}

// Header mock implementation
func (mock *ClientStream) Header() (metadata.MD, error) {
	return mock.MockHeader()
}

// Trailer mock implementation
func (mock *ClientStream) Trailer() metadata.MD {
	return mock.MockTrailer()
}

// CloseSend mock implementation
func (mock *ClientStream) CloseSend() error {
	return mock.MockCloseSend()
}

// Context mock implementation
func (mock *ClientStream) Context() context.Context {
	return mock.MockContext()
}

// SendMsg mock implementation
func (mock *ClientStream) SendMsg(m interface{}) error {
	return mock.MockSendMsg(m)
}

// RecvMsg mock implementation
func (mock *ClientStream) RecvMsg(m interface{}) error {
	return mock.MockRecvMsg(m)
}
