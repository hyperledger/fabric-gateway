package gateway

import (
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

// signingIdentity is a wrapper around an identity and signing function that implements to Fabric protoutil.Signer
// interface
type signingIdentity struct {
	serializedID []byte
	sign         identity.Sign
}

// newSigningIdentity creates an implementation ot the Fabric protoutil.Signer interface from an identity and signing
// function
func newSigningIdentity(id *identity.Identity, sign identity.Sign) (*signingIdentity, error) {
	serializedID, err := identity.Serialize(id)
	if err != nil {
		return nil, err
	}

	result := &signingIdentity{
		serializedID: serializedID,
		sign:         sign,
	}
	return result, nil
}

func (signingIdentity *signingIdentity) Sign(message []byte) ([]byte, error) {
	digest, err := identity.Hash(message)
	if err != nil {
		return nil, err
	}

	return signingIdentity.sign(digest)
}

func (signingIdentity *signingIdentity) Serialize() ([]byte, error) {
	return signingIdentity.serializedID, nil
}
