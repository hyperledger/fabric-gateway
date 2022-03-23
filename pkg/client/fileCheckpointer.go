/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FileCheckpointer struct {
	path          string
	blockNumber   uint64
	transactionID string
}

func NewFileCheckpointer(path string) (*FileCheckpointer, error) {
	fileCheckpointer := new(FileCheckpointer)
	fileCheckpointer.path = path

	err := fileCheckpointer.loadFromFile()
	if err != nil {
		return fileCheckpointer, err
	}

	err = fileCheckpointer.saveToFile()
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

func (c *FileCheckpointer) State() *FileCheckpointer {
	return &FileCheckpointer{
		blockNumber:   c.blockNumber,
		transactionID: c.transactionID,
	}
}

func (c *FileCheckpointer) loadFromFile() error {
	fileCheckpointer := struct {
		BlockNumber   uint64
		TransactionID string
	}{}

	data, err := c.readFile()
	if isError(err) {
		return err
	}
	if len(data) != 0 {
		err := json.Unmarshal(data, &fileCheckpointer)
		if isError(err) {
			return err
		}
	}

	c.updateState(fileCheckpointer.BlockNumber, fileCheckpointer.TransactionID)

	return nil
}

func (c *FileCheckpointer) updateState(blockNumber uint64, transactionID string) {
	c.blockNumber = blockNumber
	c.transactionID = transactionID
}

func (c *FileCheckpointer) readFile() ([]byte, error) {
	exist := c.checkFileExist()
	if !exist {
		err := c.createFile()
		if isError(err) {
			return []byte{}, err
		}
	}
	data, err := ioutil.ReadFile(c.path)

	return data, err
}

func (c *FileCheckpointer) checkFileExist() bool {
	_, err := os.Stat(c.path)
	return !os.IsNotExist(err)
}

func (c *FileCheckpointer) createFile() error {
	file, err := os.Create(c.path)
	if isError(err) {
		_ = file.Close()
		return err
	}

	return nil
}

func (c *FileCheckpointer) saveToFile() error {
	fileCheckpointer := c.State()
	data, err := json.Marshal(struct {
		BlockNumber   uint64
		TransactionID string
	}{
		BlockNumber:   fileCheckpointer.blockNumber,
		TransactionID: fileCheckpointer.transactionID,
	})

	if isError(err) {
		return err
	}
	err = os.WriteFile(c.path, data, 0600)
	if isError(err) {
		return err
	}

	return nil
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
