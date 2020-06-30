/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGatewayClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	identity := &pb.Identity{
		Msp:  "Org1MSP",
		Cert: "-----BEGIN CERTIFICATE-----\nMIICrjCCAlWgAwIBAgIUWtO/x2zSuxj6ungGa6StbY4xqqEwCgYIKoZIzj0EAwIw\ncDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMQ8wDQYDVQQH\nEwZEdXJoYW0xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh\nLm9yZzEuZXhhbXBsZS5jb20wHhcNMjAwNjMwMTExMTAwWhcNMjEwNjMwMTExNjAw\nWjBdMQswCQYDVQQGEwJVUzEXMBUGA1UECBMOTm9ydGggQ2Fyb2xpbmExFDASBgNV\nBAoTC0h5cGVybGVkZ2VyMQ8wDQYDVQQLEwZjbGllbnQxDjAMBgNVBAMTBXVzZXIx\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEXjbY/5IipzbHM5N46Wyf15+5w8En\naHeU3YtXWRYkv7hKBmqgWGwbdFWaiehsNqNl0xlRjqfSe5vRLHXxKY32JaOB3zCB\n3DAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUYiuH7P4W\nI/MbFDOO7fkqe9aRNB0wHwYDVR0jBBgwFoAUKA7kiCq9kHqwjSTmN/Y7JcXrz1Uw\nIgYDVR0RBBswGYIXQW5kcmV3cy1NQlAtOS5icm9hZGJhbmQwWAYIKgMEBQYHCAEE\nTHsiYXR0cnMiOnsiaGYuQWZmaWxpYXRpb24iOiIiLCJoZi5FbnJvbGxtZW50SUQi\nOiJ1c2VyMSIsImhmLlR5cGUiOiJjbGllbnQifX0wCgYIKoZIzj0EAwIDRwAwRAIg\nEXJAFq8Azb08iWEYIoevf0PqTMf79zB5ABhD28Cp8s0CIG1CWuo7GgTc/YNnztGx\n/+gAhjN1WbK52FOJrHrx2J7b\n-----END CERTIFICATE-----\n",
		Key:  "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgYOHN5EJB6rldlwzU\ndzJLK0ymomZqzQfxXFamNaQaB5ihRANCAAReNtj/kiKnNsczk3jpbJ/Xn7nDwSdo\nd5Tdi1dZFiS/uEoGaqBYbBt0VZqJ6Gw2o2XTGVGOp9J7m9EsdfEpjfYl\n-----END PRIVATE KEY-----\n",
	}

	txn := &pb.Transaction{
		Id:          identity,
		Channel:     "mychannel",
		ChaincodeID: "fabcar",
		TxnName:     "queryAllCars",
		Args:        []string{},
	}

	doit(client.EvaluateTransaction(ctx, txn))

	txn = &pb.Transaction{
		Id:          identity,
		Channel:     "mychannel",
		ChaincodeID: "fabcar",
		TxnName:     "createCar",
		Args:        []string{"CAR10", "VW", "Polo", "Grey", "Mary"},
	}
	doit(client.SubmitTransaction(ctx, txn))

	txn.TxnName = "queryCar"
	txn.Args = []string{"CAR10"}
	doit(client.EvaluateTransaction(ctx, txn))

	txn.TxnName = "changeCarOwner"
	txn.Args = []string{"CAR10", "Archie"}
	doit(client.SubmitTransaction(ctx, txn))

	txn.TxnName = "queryCar"
	txn.Args = []string{"CAR10"}
	doit(client.EvaluateTransaction(ctx, txn))
}

func doit(result *pb.Response, err error) {
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	fmt.Println(string(result.Value))
}
