/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type gatewayServer struct {
	ccpPath        string
	endorserClient peer.EndorserClient
}

func (gs *gatewayServer) SubmitTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	return nil, errors.New("SubmitTransaction not supported")
}

func (gs *gatewayServer) EvaluateTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	proposal, err := gs.createProposal(txn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create proposal: ")
	}

	signedProposal, err := gs.signProposal(proposal, txn.Id.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign proposal: ")
	}

	response, err := gs.endorserClient.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to process proposal: ")
	}

	var retVal []byte

	if response.Payload != nil {
		payload, err := protoutil.UnmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of proposal response payload failed")
		}

		extension, err := protoutil.UnmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of chaincode action failed")
		}

		if extension != nil && extension.Response != nil {
			retVal = extension.Response.Payload
		}
	}

	return &pb.Response{Value: retVal}, nil
}

func (gs *gatewayServer) createProposal(txn *pb.Transaction) (*peer.Proposal, error) {
	if txn.ChaincodeID == "" {
		return nil, errors.New("ChaincodeID is required")
	}

	if txn.TxnName == "" {
		return nil, errors.New("Fcn is required")
	}

	// Add function name to arguments
	argsArray := make([][]byte, len(txn.Args)+1)
	argsArray[0] = []byte(txn.TxnName)
	for i, arg := range txn.Args {
		argsArray[i+1] = []byte(arg)
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: txn.ChaincodeID},
			Input:       &peer.ChaincodeInput{Args: argsArray},
		},
	}

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   txn.Id.Msp,
		IdBytes: []byte(txn.Id.Cert),
	}
	signer, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize identity")
	}

	proposal, _, err := protoutil.CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		txn.Channel,
		ccis,
		signer,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	return proposal, nil
}

func (gs *gatewayServer) signProposal(proposal *peer.Proposal, keyPem string) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
	}

	// Before signing, we need to hash our message
	// The hash is what we actually sign
	msgHash := sha256.New()
	_, err = msgHash.Write(proposalBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to hash proposal")
	}
	msgHashSum := msgHash.Sum(nil)

	privPem, _ := pem.Decode([]byte(keyPem))

	if privPem.Type != "PRIVATE KEY" {
		return nil, errors.New("RSA key is of wrong type")
	}

	privPemBytes := privPem.Bytes

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			return nil, errors.Wrap(err, "unable to parse private key")
		}
	}

	var privateKey *ecdsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("unable to cast private key")
	}

	// In order to generate the signature, we provide a random number generator,
	// our private key, the hashing algorithm that we used, and the hash sum
	// of our message
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, msgHashSum)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign proposal")
	}

	s, err = utils.ToLowS(&privateKey.PublicKey, s)
	if err != nil {
		return nil, err
	}

	signature, err := utils.MarshalECDSASignature(r, s)
	if err != nil {
		return nil, err
	}

	sproposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return sproposal, nil
}

func newGatewayServer(ccpPath string, client peer.EndorserClient) (*gatewayServer, error) {
	return &gatewayServer{ccpPath, client}, nil
}

func main() {
	flag.Parse()

	// this is a client and server

	// make client connection to peer
	// organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem
	pemPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"tlsca",
		"tlsca.org1.example.com-cert.pem",
	)
	creds, err := credentials.NewClientTLSFromFile(pemPath, "")
	if err != nil {
		log.Fatalf("failed read credentials: %s", err)
	}
	peerConn, err := grpc.Dial("localhost:7051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("failed to connect to peer: %s", err)
	}
	endorserClient := peer.NewEndorserClient(peerConn)

	// setup server and listen
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 1234))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gwServer, _ := newGatewayServer(ccpPath, endorserClient)

	grpcServer := grpc.NewServer()
	pb.RegisterGatewayServer(grpcServer, gwServer)
	//... // determine whether to use TLS
	grpcServer.Serve(lis)
}
