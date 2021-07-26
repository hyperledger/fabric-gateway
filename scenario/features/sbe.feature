#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: State-based endorsement
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named private at version 1.0.0 for all organizations on channel mychannel with endorsement policy OR("Org1MSP.member","Org2MSP.member","Org3MSP.member")
        And I create a gateway named gateway1 for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the private contract
        And I prepare to submit a SetStateWithEndorser transaction
        And I set the transaction arguments to ["key-001", "value-001", "Org1MSP"]
        And I invoke the transaction

    Scenario: State change must be endorsed by Org1 peer
        When I prepare to submit a ChangeState transaction
        And I set the transaction arguments to ["key-001", "value-Org1"]
        And I set the endorsing organizations to ["Org1MSP"]
        And I invoke the transaction
        And I prepare to evaluate a GetState transaction
        And I set the transaction arguments to ["key-001"]
        And I invoke the transaction
        Then the response should be "value-Org1"

    Scenario: State change fails if endorsed only by Org2 peer
        When I prepare to submit a ChangeState transaction
        And I set the transaction arguments to ["key-001", "value-Org2"]
        And I set the endorsing organizations to ["Org2MSP"]
        Then the transaction invocation should fail
        And I prepare to evaluate a GetState transaction
        And I set the transaction arguments to ["key-001"]
        And I invoke the transaction
        Then the response should be "value-001"

    Scenario: Set SBE policy for Org1 AND Org3; state change successfully when not specifying endorsing orgs
        When I prepare to submit a SetStateEndorsers transaction
        And I set the transaction arguments to ["key-001", "Org1MSP", "Org3MSP"]
        And I invoke the transaction
        And I prepare to submit a ChangeState transaction
        And I set the transaction arguments to ["key-001", "value-Org1-Org3"]
        And I invoke the transaction
        And I prepare to evaluate a GetState transaction
        And I set the transaction arguments to ["key-001"]
        And I invoke the transaction
        Then the response should be "value-Org1-Org3"

    Scenario: Set SBE policy for Org2 AND Org3 (no local org); state change successfully when not specifying endorsing orgs
        When I prepare to submit a SetStateEndorsers transaction
        And I set the transaction arguments to ["key-001", "Org2MSP", "Org3MSP"]
        And I invoke the transaction
        And I prepare to submit a ChangeState transaction
        And I set the transaction arguments to ["key-001", "value-Org2-Org3"]
        And I invoke the transaction
        And I prepare to evaluate a GetState transaction
        And I set the transaction arguments to ["key-001"]
        And I invoke the transaction
        Then the response should be "value-Org2-Org3"

    Scenario: Set SBE policy for Org1 AND Org3; state change fails when specifying incorrect endorsing orgs
        When I prepare to submit a SetStateEndorsers transaction
        And I set the transaction arguments to ["key-001", "Org1MSP", "Org3MSP"]
        And I invoke the transaction
        And I prepare to submit a ChangeState transaction
        And I set the transaction arguments to ["key-001", "value-Org1-Org2"]
        And I set the endorsing organizations to ["Org1MSP", "Org2MSP"]
        And the transaction invocation should fail
        And I prepare to evaluate a GetState transaction
        And I set the transaction arguments to ["key-001"]
        And I invoke the transaction
        Then the response should be "value-001"
