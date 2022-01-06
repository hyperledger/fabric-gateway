#
# Copyright 2021 IBM All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
Feature: Errors
    Background:
        Given I have deployed a Fabric network
        And I have created and joined all channels
        And I deploy node chaincode named errors at version 1.0.0 for all organizations on channel mychannel with endorsement policy AND("Org1MSP.member","Org2MSP.member","Org3MSP.member")
        And I create a gateway named mygateway for user User1 in MSP Org1MSP
        And I connect the gateway to peer0.org1.example.com
        And I use the mychannel network
        And I use the errors contract

    Scenario: Evaluate fails with incorrect arguments
        When I prepare to evaluate an exists transaction
        Then the transaction invocation should fail
        And the error status should be UNKNOWN
        And the error message should contain "evaluate call to endorser returned error: chaincode response 500, Expected 1 parameters, but 0 have been supplied"

    Scenario: Submit fails with incorrect chaincode name
        When I use the nonexistent contract
        And I prepare to submit an exists transaction
        Then the transaction invocation should fail
        And the error status should be FAILED_PRECONDITION
        And the error message should contain "no combination of peers can be derived which satisfy the endorsement policy: No metadata was found for chaincode nonexistent in channel mychannel"

    Scenario: Submit fails with incorrect transaction name
        When I prepare to submit a nonexistent transaction
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to endorse transaction, see attached details for more info"
        And the error details should be
            | peer0.org1.example.com:7051 | Org1MSP | chaincode response 500, You've asked to invoke a function that does not exist: nonexistent |
            | peer1.org1.example.com:9051 | Org1MSP | chaincode response 500, You've asked to invoke a function that does not exist: nonexistent |

    Scenario: Evaluate crash chaincode
        When I prepare to evaluate a crash transaction
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to evaluate transaction, see attached details for more info"
        And the error details should be
            | peer0.org1.example.com:7051 | Org1MSP | chaincode stream terminated |
            | peer1.org1.example.com:9051 | Org1MSP | chaincode stream terminated |
            | peer0.org2.example.com:8051 | Org2MSP | chaincode stream terminated |
            | peer1.org2.example.com:10051 | Org2MSP | chaincode stream terminated |
            | peer0.org3.example.com:11051 | Org3MSP | chaincode stream terminated |

    Scenario: Evaluate with signer from unauthorized MSP
        When I prepare to evaluate an exists transaction
        And I set the transaction arguments to ["123"]
        And I do off-line signing as user User1 in MSP Org3MSP
        Then the transaction invocation should fail
        And the error status should be FAILED_PRECONDITION
        And the error message should contain "evaluate call to endorser returned error: error validating proposal: access denied: channel [mychannel] creator org [Org1MSP]"
        And the error details should be
            | peer0.org1.example.com:7051 | Org1MSP | error validating proposal: access denied: channel [mychannel] creator org [Org1MSP] |

    Scenario: Org3 fails to evaluate
        When I prepare to evaluate an orgsFail transaction
        And I set the transaction arguments to ["[\"Org1MSP\"]"]
        Then the transaction invocation should fail
        And the error status should be UNKNOWN
        And the error message should contain "evaluate call to endorser returned error: chaincode response 500, Org1MSP refuses to endorse this"

    Scenario: Org3 fails to endorse
        When I prepare to submit an orgsFail transaction
        And I set the transaction arguments to ["[\"Org3MSP\"]"]
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to collect enough transaction endorsements, see attached details for more info"
        And the error details should be
            | peer0.org3.example.com:11051 | Org3MSP | Org3MSP refuses to endorse this |

    Scenario: Org2 and Org3 fail to endorse
        When I prepare to submit an orgsFail transaction
        And I set the transaction arguments to ["[\"Org2MSP\",\"Org3MSP\"]"]
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to collect enough transaction endorsements, see attached details for more info"
        And the error details should be
            | peer0.org2.example.com:8051 | Org2MSP | Org2MSP refuses to endorse this |
            | peer1.org2.example.com:10051 | Org2MSP | Org2MSP refuses to endorse this |
            | peer0.org3.example.com:11051 | Org3MSP | Org3MSP refuses to endorse this |

    Scenario: Submit non-deterministic transaction
        When I prepare to submit a nondet transaction
        Then the transaction invocation should fail
        And the error status should be ABORTED
        And the error message should contain "failed to collect enough transaction endorsements"
        And the error details should be
            | peer0.org2.example.com:8051 | Org2MSP | ProposalResponsePayloads do not match |
            | peer1.org2.example.com:10051 | Org2MSP | ProposalResponsePayloads do not match |
            | peer0.org3.example.com:11051 | Org3MSP | ProposalResponsePayloads do not match |
