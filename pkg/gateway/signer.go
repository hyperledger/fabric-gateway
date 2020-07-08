/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/pkg/errors"
)

type signer struct {
	mspid      string
	cert       []byte
	privateKey *ecdsa.PrivateKey
}

func (si *signer) Sign(msg []byte) ([]byte, error) {
	// Before signing, we need to hash our message
	// The hash is what we actually sign
	msgHash := sha256.New()
	_, err := msgHash.Write(msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash proposal")
	}
	msgHashSum := msgHash.Sum(nil)

	// In order to generate the signature, we provide a random number generator,
	// our private key, the hashing algorithm that we used, and the hash sum
	// of our message
	r, s, err := ecdsa.Sign(rand.Reader, si.privateKey, msgHashSum)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign proposal")
	}

	s, err = utils.ToLowS(&si.privateKey.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return utils.MarshalECDSASignature(r, s)
}

func (si *signer) Serialize() ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   si.mspid,
		IdBytes: si.cert,
	}
	return proto.Marshal(serializedIdentity)
}

func createSigner(mspid string, certPem string, keyPem string) (*signer, error) {
	privPem, _ := pem.Decode([]byte(keyPem))

	if privPem.Type != "PRIVATE KEY" {
		return nil, errors.New("RSA key is of wrong type")
	}

	privPemBytes := privPem.Bytes

	var parsedKey interface{}
	var err error
	if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
		return nil, errors.Wrap(err, "unable to parse private key")
	}

	var privateKey *ecdsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("unable to cast private key")
	}

	return &signer{
		mspid:      mspid,
		cert:       []byte(certPem),
		privateKey: privateKey,
	}, nil
}
