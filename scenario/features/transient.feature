#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Transient data
	Background:
		Given I have deployed a tls Fabric network
        And I have created and joined all channels from the tls connection profile
        And I deploy node chaincode named fabcar at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member")
        And I create a gateway for user User1 in MSP Org1MSP
        And I connect the gateway to localhost:7053
        And I use the mychannel network
        And I use the fabcar contract

 	Scenario: Evaluate transaction with transient data on Node chaincode
		When I prepare to evaluate an echoTransient transaction
		And I set transient data on the transaction to
			| key1 | value1 |
			| key2 | value2 |
		And I invoke the transaction
		Then the response should be JSON matching
		    """
		    {
		    	"key1": "value1",
		    	"key2": "value2"
		    }
		    """
