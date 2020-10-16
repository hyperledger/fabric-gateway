package identity

// SigningIdentity is a wrapper around an identity and signing function that implements to Fabric protoutil.Signer
// interface
type SigningIdentity struct {
	serializedID []byte
	sign         Sign
}

// newSigningIdentity creates an implementation ot the Fabric protoutil.Signer interface from an identity and signing
// function
func NewSigningIdentity(id Identity, sign Sign) (*SigningIdentity, error) {
	serializedID, err := Serialize(id)
	if err != nil {
		return nil, err
	}

	result := &SigningIdentity{
		serializedID: serializedID,
		sign:         sign,
	}
	return result, nil
}

func (signingIdentity *SigningIdentity) Sign(message []byte) ([]byte, error) {
	digest, err := Hash(message)
	if err != nil {
		return nil, err
	}

	return signingIdentity.sign(digest)
}

func (signingIdentity *SigningIdentity) Serialize() ([]byte, error) {
	return signingIdentity.serializedID, nil
}
