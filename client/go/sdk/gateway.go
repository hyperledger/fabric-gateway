/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Gateway connection
type Gateway struct {
	url    string
	id     identity.Identity
	sign   identity.Sign
	conn   *grpc.ClientConn
	client pb.GatewayClient
}

// Connect a gateway
func Connect(url string, id identity.Identity, sign identity.Sign) (*Gateway, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "fail to dial: ")
	}
	client := pb.NewGatewayClient(conn)

	return &Gateway{
		url:    url,
		id:     id,
		sign:   sign,
		conn:   conn,
		client: client,
	}, nil
}

// ConnectTLS a gateway
func ConnectTLS(url string, id identity.Identity, sign identity.Sign, tlscert []byte) (*Gateway, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(tlscert) {
		return nil, errors.New("Failed to append certificate to client credentials")
	}
	creds := credentials.NewClientTLSFromCert(certPool, "localhost")
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "fail to dial: ")
	}
	client := pb.NewGatewayClient(conn)

	return &Gateway{
		url:    url,
		id:     id,
		sign:   sign,
		conn:   conn,
		client: client,
	}, nil
}

// Close a Gateway when it is no longer required
func (gw *Gateway) Close() {
	gw.conn.Close()
}

// GetNetwork returns a Network for a given channel name
func (gw *Gateway) GetNetwork(name string) *Network {
	return &Network{
		gateway: gw,
		name:    name,
	}
}
