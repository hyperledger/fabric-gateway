/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type FileCheckpointer struct {
	path          string
	blockNumber   uint64
	transactionID string
}

type checkpointState struct {
	BlockNumber   uint64 `json:"blockNumber"`
	TransactionID string `json:"transactionId"`
}

func NewFileCheckpointer(path string) (*FileCheckpointer, error) {
	fileCheckpointer := &FileCheckpointer{
		path: path,
	}

	err := fileCheckpointer.loadFromFile()
	if err != nil {
		return nil, err
	}

	err = fileCheckpointer.saveToFile()
	if err != nil {
		return nil, err
	}

	return fileCheckpointer, err
}

func (c *FileCheckpointer) CheckpointBlock(blockNumber uint64) error {
	c.updateState(blockNumber+1, "")
	return c.saveToFile()
}

func (c *FileCheckpointer) CheckpointTransaction(blockNumber uint64, transactionID string) error {
	c.updateState(blockNumber, transactionID)
	return c.saveToFile()
}

func (c *FileCheckpointer) CheckpointChaincodeEvent(event *ChaincodeEvent) error {
	err := c.CheckpointTransaction(event.BlockNumber, event.TransactionID)
	return err
}

func (c *FileCheckpointer) BlockNumber() uint64 {
	return c.blockNumber
}

func (c *FileCheckpointer) TransactionID() string {
	return c.transactionID
}

func (c *FileCheckpointer) loadFromFile() error {
	data, err := c.readFile()
	if err != nil {
		return err
	}
	state := &checkpointState{}
	if len(data) != 0 {
		if err := json.Unmarshal(data, state); err != nil {
			return err
		}
	}
	c.updateState(state.BlockNumber, state.TransactionID)

	return nil
}

func (c *FileCheckpointer) updateState(blockNumber uint64, transactionID string) {
	c.blockNumber = blockNumber
	c.transactionID = transactionID
}

func (c *FileCheckpointer) readFile() ([]byte, error) {
	exist := c.checkFileExist()
	if !exist {
		return nil, nil
	}

	data, err := ioutil.ReadFile(c.path)
	return data, err
}

func (c *FileCheckpointer) checkFileExist() bool {
	_, err := os.Stat(c.path)
	return !os.IsNotExist(err)
}

func (c *FileCheckpointer) saveToFile() error {
	data, err := json.Marshal(checkpointState{
		BlockNumber:   c.blockNumber,
		TransactionID: c.transactionID,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0600)
}
