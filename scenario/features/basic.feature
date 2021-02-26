#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Basic
	Background:
		Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member", "Org2MSP.member", "Org3MSP.member")
        And I create a gateway for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the basic contract

    Scenario: Evaluate echo parameters
        When I prepare to evaluate an echo transaction
        And I set the transaction arguments to ["conga"]
        And I invoke the transaction
        Then the response should be "conga"

    Scenario: Submit a name/value pair
        When I prepare to submit a put transaction
        And I set the transaction arguments to ["foo", "bar"]
        And I invoke the transaction
        Then the response should be "bar"

    Scenario: Evaluate (query) the updated value
        When I prepare to evaluate a get transaction
        And I set the transaction arguments to ["foo"]
        And I invoke the transaction
        Then the response should be "bar"
