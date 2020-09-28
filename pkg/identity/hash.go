package identity

import (
	"crypto/sha256"
)

// Hash the supplied message bytes to create digest for signing
func Hash(message []byte) ([]byte, error) {
	hash := sha256.New()

	_, err := hash.Write(message)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}
