/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

// Closer mock implementation whose method implementations can be overridden by assigning to properties
type Closer struct {
	MockClose func() error
}

// NewCloser creates mock
func NewCloser() *Closer {
	return &Closer{
		MockClose: func() error {
			return nil
		},
	}
}

// Close mock implementation
func (mock *Closer) Close() error {
	return mock.MockClose()
}
