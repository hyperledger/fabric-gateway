/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Gateway struct {
	url    string
	signer *gateway.Signer
	conn   *grpc.ClientConn
	client pb.GatewayClient
}

type Network struct {
	gateway *Gateway
	name    string
}

type Contract struct {
	network *Network
	name    string
}

type Transaction struct {
	contract *Contract
	name     string
}

func Connect(url string, signer *gateway.Signer) (*Gateway, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "fail to dial: ")
	}
	client := pb.NewGatewayClient(conn)

	return &Gateway{
		url:    url,
		signer: signer,
		conn:   conn,
		client: client,
	}, nil
}

func ConnectTLS(url string, signer *gateway.Signer, tlscert []byte) (*Gateway, error) {
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
		signer: signer,
		conn:   conn,
		client: client,
	}, nil
}

func (gw *Gateway) Close() {
	gw.conn.Close()
}

func (gw *Gateway) GetNetwork(name string) *Network {
	return &Network{
		gateway: gw,
		name:    name,
	}
}

func (nw *Network) GetContract(name string) *Contract {
	return &Contract{
		network: nw,
		name:    name,
	}
}

func (ct *Contract) CreateTransaction(name string) *Transaction {
	return &Transaction{
		contract: ct,
		name:     name,
	}
}

func (ct *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return ct.CreateTransaction(name).Evaluate(args...)
}

func (ct *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return ct.CreateTransaction(name).Submit(args...)
}

func (tx *Transaction) Evaluate(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	gw := tx.contract.network.gateway
	proposal, err := createProposal(tx, args, gw.signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proposal: ")
	}
	signedProposal, err := signProposal(proposal, gw.signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign proposal: ")
	}

	result, err := gw.client.Evaluate(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to evaluate transaction: ")
	}

	return result.Value, nil
}

func (tx *Transaction) Submit(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	gw := tx.contract.network.gateway
	proposal, err := createProposal(tx, args, gw.signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proposal: ")
	}
	signedProposal, err := signProposal(proposal, gw.signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign proposal: ")
	}

	preparedTxn, err := gw.client.Prepare(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare transaction: ")
	}

	err = signEnvelope(preparedTxn.Envelope, gw.signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign transaction: ")
	}

	stream, err := gw.client.Commit(ctx, preparedTxn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction: ")
	}

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "failed to receive event: ")
		}
		fmt.Println(event)
	}

	return preparedTxn.Response.Value, nil
}
