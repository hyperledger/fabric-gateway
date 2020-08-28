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
	"github.com/hyperledger/fabric-gateway/pkg/identity"

	"github.com/hyperledger/fabric-gateway/pkg/util"
)

func main() {
	idPath := flag.String("id", "", "path to the client's wallet identity")
	flag.Parse()

	walletIdentity, err := util.ReadWalletIdentity(*idPath)
	if err != nil {
		log.Fatalf("failed to read client identity: %s", err)
	}

	certificate, err := identity.CertificateFromPEM([]byte(walletIdentity.Credentials.Certificate))
	if err != nil {
		log.Fatal(err)
	}

	id, err := identity.NewX509Identity(walletIdentity.MspID, certificate)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := identity.PrivateKeyFromPEM([]byte(walletIdentity.Credentials.Key))
	if err != nil {
		log.Fatal(err)
	}

	signer, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	gw, err := sdk.Connect("localhost:1234", id, signer)
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
