/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing ledger entries
type SmartContract struct {
	contractapi.Contract
}

// AddEntry adds a new entry to the world state with given details
func (s *SmartContract) AddEntry(ctx contractapi.TransactionContextInterface, name string, value string) (string, error) {
	if err := ctx.GetStub().PutState(name, []byte(value)); err != nil {
		return "", errors.Wrap(err,"Failed to write to world state")
	}

	return value, nil
}

// ReadEntry returns the entry stored in the world state with given name
func (s *SmartContract) ReadEntry(ctx contractapi.TransactionContextInterface, name string) (string, error) {
	bytes, err := ctx.GetStub().GetState(name)

	if err != nil {
		return "", errors.Wrap(err, "Failed to read from world state")
	}

	if bytes == nil {
		return "", fmt.Errorf("%s does not exist", name)
	}

	return string(bytes), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
