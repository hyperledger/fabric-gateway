/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Fatalf("Error creating chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Fatalf("Error starting chaincode: %v", err)
	}
}

// SmartContract provides the contract implementation
type SmartContract struct {
	contractapi.Contract
}

// Echo the argument back as the response
func (s *SmartContract) Echo(ctx contractapi.TransactionContextInterface, arg string) (string, error) {
	return arg, nil
}

// EchoTransient returns the transient data map with both keys and values as strings
func (s *SmartContract) EchoTransient(ctx contractapi.TransactionContextInterface) (map[string]string, error) {
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("failed to get transient data: %w", err)
	}

	return toStringValues(transient), nil
}

// Put a value for a given ledger key and return the value
func (s *SmartContract) Put(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
	if err := ctx.GetStub().PutState(key, []byte(value)); err != nil {
		return "", fmt.Errorf("failed to put state to ledger: %w", err)
	}

	return value, nil
}

// Get the value for a given ledger key
func (s *SmartContract) Get(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	value, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", fmt.Errorf("failed to get state from ledger: %w", err)
	}

	return string(value), nil
}

// GetPeerOrg returns the mspid of the current peer
func (s *SmartContract) GetPeerOrg(ctx contractapi.TransactionContextInterface) (string, error) {
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	return peerOrgID, nil
}

// CheckEndorsingOrg checks that the peer org is present in the given transient data
func (s *SmartContract) CheckEndorsingOrg(ctx contractapi.TransactionContextInterface) (string, error) {
	transient, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("failed to get transient data: %w", err)
	}

	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}
	// collection := fmt.Sprintf("_implicit_org_%s", peerOrgID)

	if _, ok := transient[peerOrgID]; ok {
		return "success", nil
	}

	return "", fmt.Errorf("endorser in this org (%s) should not have been invoked", peerOrgID)
}

// ErrorMessage returns an error response containing the given message
func (s *SmartContract) ErrorMessage(ctx contractapi.TransactionContextInterface, message string) error {
	return errors.New(message)
}

func toStringValues(input map[string][]byte) map[string]string {
	results := make(map[string]string)
	for key, value := range input {
		results[key] = string(value)
	}

	return results
}
