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

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Gateway struct {
	url    string
	id     identity.Identity
	sign   identity.Sign
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

type transaction struct {
	contract  *Contract
	name      string
	transient map[string][]byte
	args      []string
}

type EvaluateTransaction struct {
	*transaction
}

type SubmitTransaction struct {
	*transaction
}

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

func (ct *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return ct.Evaluate(name).WithArgs(args...).Invoke()
}

func (ct *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return ct.Submit(name).WithArgs(args...).Invoke()
}

func (ct *Contract) newTransaction(name string) *transaction {
	return &transaction{
		contract: ct,
		name:     name,
		args:     nil,
	}
}

func (ct *Contract) Evaluate(name string) *EvaluateTransaction {
	return &EvaluateTransaction{
		ct.newTransaction(name),
	}
}

func (ct *Contract) Submit(name string) *SubmitTransaction {
	return &SubmitTransaction{
		ct.newTransaction(name),
	}
}

func (tx *EvaluateTransaction) WithArgs(args ...string) *EvaluateTransaction {
	tx.args = args
	return tx
}

func (tx *SubmitTransaction) WithArgs(args ...string) *SubmitTransaction {
	tx.args = args
	return tx
}

func (tx *SubmitTransaction) SetTransient(transientData map[string][]byte) *SubmitTransaction {
	tx.transient = transientData
	return tx
}

func (tx *EvaluateTransaction) SetTransient(transientData map[string][]byte) *EvaluateTransaction {
	tx.transient = transientData
	return tx
}

func (tx *EvaluateTransaction) Invoke() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	gw := tx.contract.network.gateway
	proposal, err := createProposal(tx.transaction, gw.id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proposal: ")
	}
	signedProposal, err := signProposal(proposal, gw.sign)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign proposal: ")
	}

	result, err := gw.client.Evaluate(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to evaluate transaction: ")
	}

	return result.Value, nil
}

func (tx *SubmitTransaction) Invoke() ([]byte, error) {
	result, commit, err := tx.InvokeAsync()
	if err != nil {
		return nil, err
	}

	if err = <-commit; err != nil {
		return nil, err
	}

	return result, nil
}

func (tx *SubmitTransaction) InvokeAsync() ([]byte, chan error, error) {
	gw := tx.contract.network.gateway
	proposal, err := createProposal(tx.transaction, gw.id)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create proposal: ")
	}
	signedProposal, err := signProposal(proposal, gw.sign)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to sign proposal: ")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	preparedTxn, err := gw.client.Prepare(ctx, signedProposal)
	if err != nil {
		cancel()
		return nil, nil, errors.Wrap(err, "failed to prepare transaction: ")
	}

	err = signEnvelope(preparedTxn.Envelope, gw.sign)
	if err != nil {
		cancel()
		return nil, nil, errors.Wrap(err, "failed to sign transaction: ")
	}

	stream, err := gw.client.Commit(ctx, preparedTxn)
	if err != nil {
		cancel()
		return nil, nil, errors.Wrap(err, "failed to commit transaction: ")
	}

	commit := make(chan error)
	go func() {
		defer cancel()
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				commit <- nil
				return
			}
			if err != nil {
				commit <- errors.Wrap(err, "failed to receive event: ")
				return
			}
			fmt.Println(event)
		}
	}()

	return preparedTxn.Response.Value, commit, nil
}
