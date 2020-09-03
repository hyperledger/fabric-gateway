#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Configure Fabric using SDK and submit/evaluate using a network Gateway
	Background:
		Given I have deployed a tls Fabric network
        And I have created and joined all channels from the tls connection profile
        And I have a gateway for Org1MSP
        And I deploy node chaincode named fabcar at version 1.0.0 for all organizations on channel mychannel with endorsement policy 1AdminOr2Other and arguments ["initLedger"]

 	Scenario: Using a Gateway I can submit and evaluate transactions on instantiated node chaincode
        Given I have a gateway as user User1 using the tls connection profile
        And I connect the gateway
        And I use the mychannel network
        And I use the fabcar contract
        When I prepare to submit a createCar transaction
        And I set the transaction arguments to ["CAR10", "Trabant", "601 Estate", "brown", "Simon"]
        And I invoke the transaction
        And I prepare to evaluate a queryCar transaction
        And I set the transaction argument to ["CAR10"]
        And I invoke the transaction
        Then the response should be JSON matching
            """
            {
                "color": "brown",
                "docType": "car",
                "make": "Trabant",
                "model": "601 Estate",
                "owner": "Simon"
            }
            """

    Scenario: Using a Gateway with an X509Identity I can submit and evaluate transactions on instantiated node chaincode
        Given I have a gateway as user User1 using the tls connection profile
        And I connect the gateway
        And I use the mychannel network
        And I use the fabcar contract
        When I prepare to submit a createCar transaction
        And I set the transaction arguments to ["CAR11", "Tesla", "Model X", "black", "Jon Doe"]
        And I invoke the transaction
        And I prepare to evaluate a queryCar transaction
        And I set the transaction arguments to ["CAR11"]
        And I invoke the transaction
        Then the response should be JSON matching
            """
            {
                "color": "black",
                "docType": "car",
                "make": "Tesla",
                "model": "Model X",
                "owner": "Jon Doe"
            }
            """

