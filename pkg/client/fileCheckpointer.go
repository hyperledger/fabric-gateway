
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
	path string
	blockNumber uint64
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

func (c *FileCheckpointer) CheckpointBlock(blockNumber uint64) error{
	c.blockNumber = blockNumber + 1
	c.transactionID = ""
	return c.saveToFile()
}

func (c *FileCheckpointer) CheckpointTransaction(blockNumber uint64, transactionID string) error{
	c.blockNumber = blockNumber
	c.transactionID = transactionID
	return c.saveToFile()
}

func (c *FileCheckpointer) CheckpointChaincodeEvent(event *ChaincodeEvent) error {
	return c.CheckpointTransaction(event.BlockNumber, event.TransactionID)
}

func (c *FileCheckpointer) BlockNumber() uint64 {
	return c.blockNumber
}

func (c *FileCheckpointer) TransactionID() string {
	return c.transactionID
}

func (c *FileCheckpointer) getState() *FileCheckpointer {
	return  &FileCheckpointer {
		blockNumber: c.blockNumber,
		transactionID: c.transactionID,
	}
}

func (c *FileCheckpointer) loadFromFile() error {
	fileCheckpointer :=  struct{
		BlockNumber uint64
		TransactionID string
	}{}

	data, err := c.readFile()
	if isError(err) {
		return err
	}
	if len(data) != 0 {
		 err := json.Unmarshal(data, &fileCheckpointer);
			if isError(err){
			return err
		 }
	}

	c.setState(&FileCheckpointer{blockNumber: fileCheckpointer.BlockNumber,transactionID:fileCheckpointer.TransactionID})

	return nil
}

func (c *FileCheckpointer) setState(fileCheckpointer *FileCheckpointer) {
	c.blockNumber = fileCheckpointer.blockNumber
	c.transactionID = fileCheckpointer.transactionID
}

func (c *FileCheckpointer) readFile() ([]byte, error) {
	exist := c.checkFileExist()
	if !exist {
		err := c.createFile()
		if isError(err){
			return []byte{}, err
		}
	}
	data , err := ioutil.ReadFile(c.path)

	return data ,err
}

func (c *FileCheckpointer) checkFileExist() bool {
	_, err := os.Stat(c.path)
	return !os.IsNotExist(err)
}

func (c *FileCheckpointer) createFile() error {
	file, err := os.Create(c.path)
	if isError(err) {
		return err
	}
	defer file.Close()

	return nil
}

func (c *FileCheckpointer) saveToFile() error {
	fileCheckpointer := c.getState()
	data, err := json.Marshal(struct {
		BlockNumber uint64
		TransactionID string}{
			BlockNumber:fileCheckpointer.blockNumber,
			TransactionID:fileCheckpointer.transactionID,
	})

	if isError(err){
		return err
	}
	err = os.WriteFile(c.path, data, 0755)
	if isError(err) {
		return err
	}

	return nil
}

func isError(err error ) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}