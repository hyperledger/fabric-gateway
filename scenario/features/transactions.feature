#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Transaction invocation
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy golang chaincode named basic at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member")
        And I create a gateway named mygateway for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the basic contract

    Scenario: Evaluate with result
        When I prepare to evaluate an echo transaction
        And I set the transaction arguments to ["conga"]
        And I invoke the transaction
        Then the response should be "conga"

    Scenario: Submit with result
        When I prepare to submit an echo transaction
        And I set the transaction arguments to ["conga"]
        And I invoke the transaction
        Then the response should be "conga"

    Scenario: Update ledger
        When I prepare to submit a put transaction
        And I set the transaction arguments to ["update", "ledger"]
        And I invoke the transaction
        And I prepare to evaluate a get transaction
        And I set the transaction arguments to ["update"]
        And I invoke the transaction
        Then the response should be "ledger"

    Scenario: Transient data
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

    Scenario: Evaluate with error response
        When I prepare to evaluate an errorMessage transaction
        And I set the transaction arguments to ["ALL_YOUR_ERROR_ARE_BELONG_TO_US"]
        Then the transaction invocation should fail
        And the error status should be UNKNOWN
        And the error message should contain "ALL_YOUR_ERROR_ARE_BELONG_TO_US"

    Scenario: Submit with error response
        When I prepare to submit an errorMessage transaction
        And I set the transaction arguments to ["ALL_YOUR_ERROR_ARE_BELONG_TO_US"]
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to endorse transaction, see attached details for more info"
        And the error details should be
            | peer0.org1.example.com:7051 | Org1MSP | ALL_YOUR_ERROR_ARE_BELONG_TO_US |
            | peer1.org1.example.com:9051 | Org1MSP | ALL_YOUR_ERROR_ARE_BELONG_TO_US |
