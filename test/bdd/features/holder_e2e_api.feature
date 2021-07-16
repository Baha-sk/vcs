#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

@all
@holder_rest
Feature: Holder VC REST API
  @holderProfileRecreation
  Scenario Outline: Delete and recreate Holder profile
    Given Client sends request to create a holder profile with ID "<profileID>"
    And   Client deletes the holder profile with ID "<profileID>"
    Then  Client can recreate the holder profile with ID "<profileID>"
    Examples:
      | profileID    |
      | test_profile |

  # TODO example 'holderwithdidv1' to be uncommented after fixing Issue[#429]
  # TODO example 'holderwithdidsov' to be uncommented after fixing Issue[#454]
  @holder_api
  Scenario Outline: Holder APIs
    Given Holder Profile "<profile>" is created with DID "<did>", privateKey "<privateKey>", keyID "<keyID>", signatureHolder "<signatureHolder>", uniRegistrar '<uniRegistrar>', didMethod "<didMethod>", signatureType "<signatureType>" and keyType "<keyType>"
    And   Holder profile "<profile>" can be retrieved with DID "<didMethod>" and signatureType "<signatureType>"
    Then  Holder "<profile>" generates presentation for the VC received from the Government
    Examples:
      | profile                        | did                                                              | privateKey                                                                               | keyID                                                                                                           | signatureHolder |uniRegistrar                                                                                                                                                      |  didMethod      |    signatureType      |  keyType  |
      | holder_local_ed25519_jws       |                                                                  |                                                                                          |                                                                                                                 | JWS             |                                                                                                                                                                  | did:orb         | Ed25519Signature2018  |  Ed25519  |
      | holder_local_p256_pv           |                                                                  |                                                                                          |                                                                                                                 | ProofValue      |                                                                                                                                                                  | did:orb         | JsonWebSignature2020  |  P256     |
      | holder_local_ed25519          |                                                                  |                                                                                          |                                                                                                                  | JWS              |                                                          		                                                                                                | did:orb         | JsonWebSignature2020  |  Ed25519  |
      | holder_local_p256             |                                                                  |                                                                                          |                                                                                                                  | JWS              |                                                            		                                                                                                | did:orb         | JsonWebSignature2020  |  P256     |
#      | holderwithdidv1                |                                                                  |                                                                                          |                                                                                                                 | JWS             |{"driverURL":"http://uni-registrar-web:9080/1.0/register?driverId=driver-universalregistrar/driver-did-v1","options": {"ledger": "test", "keytype": "ed25519"}}   | did:v1:test:nym | Ed25519Signature2018  |  Ed25519  |
#      | holderwithdidelem              | did:elem:EiAWdU2yih6NA2IGnLsDhkErZ8aQX6b8yKt7jHMi-ttFdQ          | 5AcDTQT7Cdg1gBvz8PQpnH3xEbLCE1VQxAJV5NjVHvNjsZSfn4NaLZ77mapoi4QwZeBhcAA7MQzaFYkzJLfGjNnR |  did:elem:ropsten:EiAWdU2yih6NA2IGnLsDhkErZ8aQX6b8yKt7jHMi-ttFdQ#SQ2PY2xs7NOr6B26xq_pJMNpuYk6dOeROlkzKF7909I    | JWS             |                                                                                                                                                                  | did:elem        | Ed25519Signature2018  |  Ed25519  |
#      | holderwithdidsov               |                                                                  |                                                                                          |                                                                                                                 | JWS             |{"driverURL":"https://uniregistrar.io/1.0/register?driverId=driver-universalregistrar/driver-did-sov","options": {"network":"danube"}}                            | did:sov:danube  | Ed25519Signature2018  |  Ed25519  |
      | holderwithdidkey               | did:key:z6MkjRagNiMu91DduvCvgEsqLZDVzrJzFrwahc4tXLt9DoHd         | 28xXA4NyCQinSJpaZdSuNBM4kR2GqYb8NPqAtZoGCpcRYWBcDXtzVAzpZ9BAfgV334R2FC383fiHaWWWAacRaYGs |  did:key:z6MkjRagNiMu91DduvCvgEsqLZDVzrJzFrwahc4tXLt9DoHd#z6MkjRagNiMu91DduvCvgEsqLZDVzrJzFrwahc4tXLt9DoHd      | JWS             |                                                                                                                                                                  | did:key         | Ed25519Signature2018  |  Ed25519  |

