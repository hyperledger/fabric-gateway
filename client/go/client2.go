/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-gateway/client/go/sdk"
	"github.com/hyperledger/fabric-gateway/pkg/gateway"

	"github.com/hyperledger/fabric-gateway/pkg/util"
)

func main() {
	idPath := flag.String("id", "", "path to the client's wallet identity")
	flag.Parse()

	id, err := util.ReadWalletIdentity(*idPath)
	if err != nil {
		log.Fatalf("failed to read client identity: %s", err)
	}

	signer, err := gateway.CreateSigner(
		id.MspID,
		id.Credentials.Certificate,
		id.Credentials.Key,
	)

	// pem, err := ioutil.ReadFile("/Users/acoleman/gopath/src/github.com/hyperledger/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem")
	// if err != nil {
	// 	log.Fatalf("failed to read tls cert: %s", err)
	// }

	// gw, err := sdk.ConnectTLS("localhost:6051", signer, pem)
	gw, err := sdk.Connect("localhost:1234", signer)
	defer gw.Close()

	network := gw.GetNetwork("mychannel")
	contract := network.GetContract("fabcar")
	try(contract.EvaluateTransaction("queryAllCars"))
	try(contract.SubmitTransaction("createCar", "CAR10", "VW", "Polo", "Grey", "Mary"))
	try(contract.EvaluateTransaction("queryCar", "CAR10"))
	try(contract.SubmitTransaction("changeCarOwner", "CAR10", "Archie"))
	try(contract.EvaluateTransaction("queryCar", "CAR10"))

}

func try(result []byte, err error) {
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	fmt.Println(string(result))
}
