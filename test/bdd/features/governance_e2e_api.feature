#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@governance_rest
Feature: Governance VC REST API

  Scenario Outline: Governance APIs
    Given Governance Profile "<profile>" is created with DID "<did>", privateKey "<privateKey>", keyID "<keyID>", signatureHolder "<signatureHolder>", uniRegistrar '<uniRegistrar>', didMethod "<didMethod>", signatureType "<signatureType>" and keyType "<keyType>"
    Then  Governance "<profile>" generates credential with signatureType "<signatureType>"
    Examples:
      | profile                        | did                                                              | privateKey                                                                               | keyID                                                                                                           | signatureHolder |uniRegistrar                                                                                                                                                      | didMethod |    signatureType      |  keyType  |
      |governance_local_ed25519_jws    |                                                                  |                                                                                          |                                                                                                                 | JWS             |                                                                                                                                                                  | did:orb   | Ed25519Signature2018  |  Ed25519  |
      |governance_local_p256_pv        |                                                                  |                                                                                          |                                                                                                                 | ProofValue      |                                                                                                                                                                  | did:orb   | JsonWebSignature2020  |  P256     |
      |governance_local_ed25519        |                                                                  |                                                                                          |                                                                                                                 | JWS             |                                                              		                                                                                            | did:orb   | JsonWebSignature2020  |  Ed25519  |
      |governance_local_p256           |                                                                  |                                                                                          |                                                                                                                 | JWS             |                                                              		                                                                                            | did:orb   | JsonWebSignature2020  |  P256     |
