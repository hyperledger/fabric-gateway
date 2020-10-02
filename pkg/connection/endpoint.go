/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package network

import (
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Endpoint for a Fabric node, such as a Gateway, Peer or Orderer.
type Endpoint struct {
	Address             string
	TLSRootCertificates []*x509.Certificate
}

func (endpoint *Endpoint) tlsRootCAs() *x509.CertPool {
	certPool := x509.NewCertPool()
	for _, certificate := range endpoint.TLSRootCertificates {
		certPool.AddCert(certificate)
	}
	return certPool
}

// Dial creates a gRPC client connection to the endpoint.
func (endpoint *Endpoint) Dial() (*grpc.ClientConn, error) {
	tlsRootCAs := endpoint.tlsRootCAs()
	transportCredentials := credentials.NewClientTLSFromCert(tlsRootCAs, "")

	return grpc.Dial(endpoint.Address, grpc.WithTransportCredentials(transportCredentials))
}
