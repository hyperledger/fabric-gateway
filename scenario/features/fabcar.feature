#
# Copyright 2020 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Fabcar
	Background:
		Given I have deployed a tls Fabric network
        And I have created and joined all channels from the tls connection profile
        And I deploy node chaincode named fabcar at version 1.0.0 for all organizations on channel mychannel with endorsement policy 1AdminOr2Other
        And I create a gateway for user User1 in MSP Org1MSP
        And I connect the gateway to localhost:7053
        And I use the mychannel network
        And I use the fabcar contract

    Scenario: Submit and evaluate transactions on Node chaincode
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
