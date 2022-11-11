/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4vc_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vcsverifiable "github.com/trustbloc/vcs/pkg/doc/verifiable"
	profileapi "github.com/trustbloc/vcs/pkg/profile"
	"github.com/trustbloc/vcs/pkg/service/oidc4vc"
)

func TestService_PushAuthorizationDetails(t *testing.T) {
	var (
		mockTransactionStore = NewMockTransactionStore(gomock.NewController(t))
		ad                   *oidc4vc.AuthorizationDetails
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, err error)
	}{
		{
			name: "Success",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
					},
				}, nil)

				mockTransactionStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "universitydegreecredential",
					Format:         vcsverifiable.Ldp,
				}
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Fail to find transaction by op state",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(
					nil, errors.New("find tx error"))

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "UniversityDegreeCredential",
					Format:         vcsverifiable.Ldp,
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "find tx by op state")
			},
		},
		{
			name: "Credential template not configured",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialFormat: vcsverifiable.Ldp,
					},
				}, nil)

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "UniversityDegreeCredential",
					Format:         vcsverifiable.Ldp,
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrCredentialTemplateNotConfigured)
			},
		},
		{
			name: "Credential type not supported",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
					},
				}, nil)

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "NotSupportedCredentialType",
					Format:         vcsverifiable.Ldp,
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrCredentialTypeNotSupported)
			},
		},
		{
			name: "Credential format not supported",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
					},
				}, nil)

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "UniversityDegreeCredential",
					Format:         vcsverifiable.Jwt,
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrCredentialFormatNotSupported)
			},
		},
		{
			name: "Fail to update transaction",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
					},
				}, nil)

				mockTransactionStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

				ad = &oidc4vc.AuthorizationDetails{
					CredentialType: "UniversityDegreeCredential",
					Format:         vcsverifiable.Ldp,
				}
			},
			check: func(t *testing.T, err error) {
				require.ErrorContains(t, err, "update tx")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			svc, err := oidc4vc.NewService(&oidc4vc.Config{
				TransactionStore: mockTransactionStore,
			})
			require.NoError(t, err)

			err = svc.PushAuthorizationDetails(context.Background(), "opState", ad)
			tt.check(t, err)
		})
	}
}

func TestService_PrepareClaimDataAuthorizationRequest(t *testing.T) {
	var (
		mockTransactionStore = NewMockTransactionStore(gomock.NewController(t))
		req                  *oidc4vc.PrepareClaimDataAuthorizationRequest
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error)
	}{
		{
			name: "Success",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
						ResponseType:     "code",
						Scope:            []string{"openid", "profile", "address"},
					},
				}, nil)

				mockTransactionStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

				req = &oidc4vc.PrepareClaimDataAuthorizationRequest{
					OpState:      "opState",
					ResponseType: "code",
					Scope:        []string{"openid", "profile"},
					AuthorizationDetails: &oidc4vc.AuthorizationDetails{
						CredentialType: "UniversityDegreeCredential",
						Format:         vcsverifiable.Ldp,
					},
				}
			},
			check: func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "Response type mismatch",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						ResponseType: "code",
						Scope:        []string{"openid"},
					},
				}, nil)

				req = &oidc4vc.PrepareClaimDataAuthorizationRequest{
					ResponseType: "invalid",
					Scope:        []string{"openid"},
					OpState:      "opState",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrResponseTypeMismatch)
			},
		},
		{
			name: "Invalid scope",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						ResponseType: "code",
						Scope:        []string{"openid", "profile"},
					},
				}, nil)

				req = &oidc4vc.PrepareClaimDataAuthorizationRequest{
					ResponseType: "code",
					Scope:        []string{"openid", "profile", "address"},
					OpState:      "opState",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrInvalidScope)
			},
		},
		{
			name: "Fail to find transaction by op state",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(
					nil, errors.New("find tx error"))

				req = &oidc4vc.PrepareClaimDataAuthorizationRequest{
					OpState: "opState",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error) {
				require.ErrorContains(t, err, "find tx by op state")
				require.Nil(t, resp)
			},
		},
		{
			name: "Fail to update transaction",
			setup: func() {
				mockTransactionStore.EXPECT().FindByOpState(gomock.Any(), "opState").Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type: "UniversityDegreeCredential",
						},
						CredentialFormat: vcsverifiable.Ldp,
						ResponseType:     "code",
						Scope:            []string{"openid"},
					},
				}, nil)

				mockTransactionStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

				req = &oidc4vc.PrepareClaimDataAuthorizationRequest{
					OpState:      "opState",
					ResponseType: "code",
					Scope:        []string{"openid"},
					AuthorizationDetails: &oidc4vc.AuthorizationDetails{
						CredentialType: "UniversityDegreeCredential",
						Format:         vcsverifiable.Ldp,
					},
				}
			},
			check: func(t *testing.T, resp *oidc4vc.PrepareClaimDataAuthorizationResponse, err error) {
				require.ErrorContains(t, err, "update tx")
				require.Empty(t, resp)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			svc, err := oidc4vc.NewService(&oidc4vc.Config{
				TransactionStore: mockTransactionStore,
			})
			require.NoError(t, err)

			resp, err := svc.PrepareClaimDataAuthorizationRequest(context.Background(), req)
			tt.check(t, resp, err)
		})
	}
}

func TestValidatePreAuthCode(t *testing.T) {
	t.Run("success with pin", func(t *testing.T) {
		storeMock := NewMockTransactionStore(gomock.NewController(t))
		srv, err := oidc4vc.NewService(&oidc4vc.Config{
			TransactionStore: storeMock,
		})
		assert.NoError(t, err)

		storeMock.EXPECT().FindByOpState(gomock.Any(), "1234").Return(&oidc4vc.Transaction{
			TransactionData: oidc4vc.TransactionData{
				PreAuthCode:     "1234",
				UserPinRequired: true,
			},
		}, nil)

		resp, err := srv.ValidatePreAuthorizedCodeRequest(context.TODO(), "1234", "111")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("success without pin", func(t *testing.T) {
		storeMock := NewMockTransactionStore(gomock.NewController(t))
		srv, err := oidc4vc.NewService(&oidc4vc.Config{
			TransactionStore: storeMock,
		})
		assert.NoError(t, err)

		storeMock.EXPECT().FindByOpState(gomock.Any(), "1234").Return(&oidc4vc.Transaction{
			TransactionData: oidc4vc.TransactionData{
				PreAuthCode:     "1234",
				UserPinRequired: false,
			},
		}, nil)

		resp, err := srv.ValidatePreAuthorizedCodeRequest(context.TODO(), "1234", "")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("invalid pin", func(t *testing.T) {
		storeMock := NewMockTransactionStore(gomock.NewController(t))
		srv, err := oidc4vc.NewService(&oidc4vc.Config{
			TransactionStore: storeMock,
		})
		assert.NoError(t, err)

		storeMock.EXPECT().FindByOpState(gomock.Any(), "1234").Return(&oidc4vc.Transaction{
			TransactionData: oidc4vc.TransactionData{
				PreAuthCode:     "1234",
				UserPinRequired: true,
			},
		}, nil)

		resp, err := srv.ValidatePreAuthorizedCodeRequest(context.TODO(), "1234", "")
		assert.ErrorContains(t, err, "invalid auth credentials")
		assert.Nil(t, resp)
	})

	t.Run("fail find tx", func(t *testing.T) {
		storeMock := NewMockTransactionStore(gomock.NewController(t))
		srv, err := oidc4vc.NewService(&oidc4vc.Config{
			TransactionStore: storeMock,
		})
		assert.NoError(t, err)

		storeMock.EXPECT().FindByOpState(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

		resp, err := srv.ValidatePreAuthorizedCodeRequest(context.TODO(), "1234", "")
		assert.ErrorContains(t, err, "not found")
		assert.Nil(t, resp)
	})
}

func TestService_PrepareCredential(t *testing.T) {
	var (
		mockTransactionStore  = NewMockTransactionStore(gomock.NewController(t))
		mockHTTPClient        = NewMockHTTPClient(gomock.NewController(t))
		mockCredentialService = NewMockCredentialService(gomock.NewController(t))
		req                   *oidc4vc.CredentialRequest
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, resp *oidc4vc.CredentialResponse, err error)
	}{
		{
			name: "Success",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					ID: "txID",
					TransactionData: oidc4vc.TransactionData{
						IssuerToken: "issuer-access-token",
						CredentialTemplate: &profileapi.CredentialTemplate{
							Type:   "VerifiedEmployee",
							Issuer: "issuer",
						},
						CredentialFormat: vcsverifiable.Jwt,
					},
				}, nil)

				claimData := `{"surname":"Smith","givenName":"Pat","jobTitle":"Worker"}`

				mockHTTPClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(
					req *http.Request,
				) (*http.Response, error) {
					assert.Contains(t, req.Header.Get("Authorization"), "Bearer issuer-access-token")
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer([]byte(claimData))),
					}, nil
				})

				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&verifiable.Credential{}, nil)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "Fail to find transaction by op state",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(
					nil, errors.New("get error"))

				mockHTTPClient.EXPECT().Do(gomock.Any()).Times(0)
				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorContains(t, err, "get tx")
				require.Nil(t, resp)
			},
		},
		{
			name: "Credential template not configured",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{},
				}, nil)

				mockHTTPClient.EXPECT().Do(gomock.Any()).Times(0)
				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrCredentialTemplateNotConfigured)
				require.Nil(t, resp)
			},
		},
		{
			name: "Fail to make request to claim endpoint",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{},
					},
				}, nil)

				mockHTTPClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("http error"))
				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorContains(t, err, "http error")
				require.Nil(t, resp)
			},
		},
		{
			name: "Claim endpoint returned other than 200 OK status code",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{},
					},
				}, nil)

				mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				}, nil)

				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorContains(t, err, "claim endpoint returned status code")
				require.Nil(t, resp)
			},
		},
		{
			name: "Fail to decode claim data",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{
						CredentialTemplate: &profileapi.CredentialTemplate{},
					},
				}, nil)

				mockHTTPClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte("invalid"))),
				}, nil)

				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorContains(t, err, "decode claim data")
				require.Nil(t, resp)
			},
		},
		{
			name: "Credential format not supported",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{
						IssuerToken:        "issuer-access-token",
						CredentialTemplate: &profileapi.CredentialTemplate{},
					},
				}, nil)

				claimData := `{"surname":"Smith","givenName":"Pat","jobTitle":"Worker"}`

				mockHTTPClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(
					req *http.Request,
				) (*http.Response, error) {
					assert.Contains(t, req.Header.Get("Authorization"), "Bearer issuer-access-token")
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer([]byte(claimData))),
					}, nil
				})

				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorIs(t, err, oidc4vc.ErrCredentialFormatNotSupported)
				require.Nil(t, resp)
			},
		},
		{
			name: "Service error",
			setup: func() {
				mockTransactionStore.EXPECT().Get(gomock.Any(), oidc4vc.TxID("txID")).Return(&oidc4vc.Transaction{
					TransactionData: oidc4vc.TransactionData{
						IssuerToken:        "issuer-access-token",
						CredentialTemplate: &profileapi.CredentialTemplate{},
						CredentialFormat:   vcsverifiable.Ldp,
					},
				}, nil)

				claimData := `{"surname":"Smith","givenName":"Pat","jobTitle":"Worker"}`

				mockHTTPClient.EXPECT().Do(gomock.Any()).DoAndReturn(func(
					req *http.Request,
				) (*http.Response, error) {
					assert.Contains(t, req.Header.Get("Authorization"), "Bearer issuer-access-token")
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer([]byte(claimData))),
					}, nil
				})

				mockCredentialService.EXPECT().IssueCredential(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("issue credential error"))

				req = &oidc4vc.CredentialRequest{
					TxID: "txID",
				}
			},
			check: func(t *testing.T, resp *oidc4vc.CredentialResponse, err error) {
				require.ErrorContains(t, err, "issue credential error")
				require.Nil(t, resp)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			svc, err := oidc4vc.NewService(&oidc4vc.Config{
				TransactionStore:  mockTransactionStore,
				CredentialService: mockCredentialService,
				HTTPClient:        mockHTTPClient,
			})
			require.NoError(t, err)

			resp, err := svc.PrepareCredential(context.Background(), req)
			tt.check(t, resp, err)
		})
	}
}
