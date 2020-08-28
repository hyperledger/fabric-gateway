package identity

import (
	"bytes"
	"crypto/sha256"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// Serialize an identity to protobuf SerializedIdentity message bytes
func Serialize(id *Identity) ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID,
		IdBytes: id.IDBytes,
	}
	return proto.Marshal(serializedIdentity)
}

// Deserialize SerializedIdentity protobuf message bytes to an Identity
func Deserialize(message []byte) (*Identity, error) {
	serializedIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(message, serializedIdentity); err != nil {
		return nil, err
	}

	result := &Identity{
		MspID:   serializedIdentity.Mspid,
		IDBytes: serializedIdentity.IdBytes,
	}
	return result, nil
}

// Hash the supplied message bytes to create digest for signing
func Hash(message []byte) ([]byte, error) {
	hash := sha256.New()

	_, err := hash.Write(message)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func pemEncode(block *pem.Block) ([]byte, error) {
	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, block); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil

}
