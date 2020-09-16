/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wallet

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

// X509Identity represents an X509 identity
type X509Identity struct {
	Version     int         `json:"version"`
	MspID       string      `json:"mspId"`
	IDType      string      `json:"type"`
	Credentials credentials `json:"credentials"`
}

type credentials struct {
	Certificate string `json:"certificate"`
	Key         string `json:"privateKey"`
}

// ReadWalletIdentity loads a user's credentials from a filesystem wallet
func ReadWalletIdentity(pathname string) (*X509Identity, error) {
	content, err := ioutil.ReadFile(filepath.Clean(pathname))
	if err != nil {
		return nil, err
	}

	id := &X509Identity{}

	if err := json.Unmarshal(content, &id); err != nil {
		return nil, errors.Wrap(err, "Invalid identity format")
	}

	return id, err
}
