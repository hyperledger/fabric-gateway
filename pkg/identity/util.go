package identity

import (
	"crypto/sha256"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// Serialize an identity to protobuf SerializedIdentity message bytes
func Serialize(id *Identity) ([]byte, error) {
	certPem, err := CertificateToPEM(id.Certificate)
	if err != nil {
		return nil, err
	}

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID,
		IdBytes: certPem,
	}
	return proto.Marshal(serializedIdentity)
}

// Deserialize SerializedIdentity protobuf message bytes to an Identity
func Deserialize(message []byte) (*Identity, error) {
	deserializedIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(message, deserializedIdentity); err != nil {
		return nil, err
	}
	return NewIdentity(deserializedIdentity.Mspid, deserializedIdentity.IdBytes)
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
