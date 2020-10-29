/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"bytes"
	"crypto/x509"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

func TestEndpoint(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Creates an empty CertPool if no root certificates", func(t *testing.T) {
		endpoint := Endpoint{}

		certPool := endpoint.tlsRootCAs()

		subjects := certPool.Subjects()
		if len(subjects) != 0 {
			t.Fatalf("Expected no subject, got: %v", subjects)
		}
	})

	t.Run("Creates a CertPool containing root certificates", func(t *testing.T) {
		caCerts := []*x509.Certificate{certificate}
		endpoint := Endpoint{
			TLSRootCertificates: caCerts,
		}

		certPool := endpoint.tlsRootCAs()

		subjects := certPool.Subjects()
		if len(subjects) != len(caCerts) {
			t.Fatalf("Expected %d subjects, got: %v", len(caCerts), subjects)
		}

		for i, actual := range subjects {
			expected := caCerts[i].RawSubject
			if !bytes.Equal(expected, actual) {
				t.Fatalf("Subjects did not match. Expected: %v, got: %v", expected, actual)
			}
		}
	})

	t.Run("Endpoint with no root certificates is not TLS", func(t *testing.T) {
		endpoint := Endpoint{}

		if endpoint.isTLS() {
			t.Fatal("Expected isTLS to be false, got true")
		}
	})

	t.Run("Endpoint with root certificates is TLS", func(t *testing.T) {
		caCerts := []*x509.Certificate{certificate}
		endpoint := Endpoint{
			TLSRootCertificates: caCerts,
		}

		if !endpoint.isTLS() {
			t.Fatal("Expected isTLS to be false, got true")
		}
	})

	t.Run("Endpoint address string", func(t *testing.T) {
		endpoint := Endpoint{
			Host: "host",
			Port: 418,
		}

		actual := endpoint.String()
		expected := "host:418"

		if actual != expected {
			t.Fatalf("Incorrect endpoint address string. Expected: %s, got: %s", expected, actual)
		}
	})

	t.Run("Dial non-TLS", func(t *testing.T) {
		endpoint := Endpoint{
			Host: "example.org",
			Port: 7,
		}

		clientConnection, err := endpoint.Dial()
		if err != nil {
			t.Fatal(err)
		}
		if nil == clientConnection {
			t.Fatal("Client connection is nil")
		}
	})

	t.Run("Dial TLS", func(t *testing.T) {
		caCerts := []*x509.Certificate{certificate}
		endpoint := Endpoint{
			Host:                "example.org",
			Port:                7,
			TLSRootCertificates: caCerts,
		}

		clientConnection, err := endpoint.Dial()
		if err != nil {
			t.Fatal(err)
		}
		if nil == clientConnection {
			t.Fatal("Client connection is nil")
		}
	})
}
