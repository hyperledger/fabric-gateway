/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	proto1 "github.com/golang/protobuf/proto"
	proto2 "google.golang.org/protobuf/proto"
)

// Marshal returns the wire-format encoding of a message. The message can be either a protobuf v1 or v2 message.
func Marshal(message proto1.GeneratedMessage) ([]byte, error) {
	messageV2 := proto1.MessageV2(message)
	return proto2.Marshal(messageV2)
}

// Unmarshal parses the wire-format message bytes and places the result in the provided message. The provided message
// must be mutable (e.g., a non-nil pointer to a message). The message can be either a protobuf v1 or v2 message.
func Unmarshal(bytes []byte, message proto1.GeneratedMessage) error {
	messageV2 := proto1.MessageV2(message)
	return proto2.Unmarshal(bytes, messageV2)
}

// ProtoEqual reports whether two messages are equal, as with the standard proto.Equal function. The messages can be
// either protobuf v1 or v2 message.
func ProtoEqual(x proto1.GeneratedMessage, y proto1.GeneratedMessage) bool {
	return proto2.Equal(proto1.MessageV2(x), proto1.MessageV2(y))
}
