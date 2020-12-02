/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type signingIdentity struct {
	id   identity.Identity
	sign identity.Sign
	hash hash.Hash
}

func newSigningIdentity(id identity.Identity, sign identity.Sign) *signingIdentity {
	return &signingIdentity{
		id:   id,
		sign: sign,
		hash: hash.SHA256,
	}
}

func (signingID *signingIdentity) Identity() identity.Identity {
	return signingID.id
}

func (signingID *signingIdentity) Hash(message []byte) ([]byte, error) {
	return signingID.hash(message)
}

func (signingID *signingIdentity) Sign(digest []byte) ([]byte, error) {
	return signingID.sign(digest)
}

func (signingID *signingIdentity) Creator() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   signingID.id.MspID(),
		IdBytes: signingID.id.Credentials(),
	}
	return proto.Marshal(serializedIdentity)
}
