/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/client/go/sdk"
	"github.com/hyperledger/fabric-gateway/pkg/gateway"

	"github.com/hyperledger/fabric-gateway/pkg/util"
)

func main() {
	idfile := filepath.Join(
		"..",
		"..",
		"..",
		"fabric-samples",
		"fabcar",
		"javascript",
		"wallet",
		"appUser.id",
	)

	id, err := util.ReadWalletIdentity(idfile)
	if err != nil {
		log.Fatalf("failed to read gateway identity: %s", err)
	}

	signer, err := gateway.CreateSigner(
		id.MspID,
		id.Credentials.Certificate,
		id.Credentials.Key,
	)

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
