/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

//nolint:lll
package oidc4ci_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/labstack/echo/v4"
	"github.com/ory/fosite"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/trustbloc/vcs/pkg/oauth2client"
	"github.com/trustbloc/vcs/pkg/restapi/v1/common"
	"github.com/trustbloc/vcs/pkg/restapi/v1/issuer"
	"github.com/trustbloc/vcs/pkg/restapi/v1/oidc4ci"
	"github.com/trustbloc/vcs/pkg/storage/mongodb/oidc4cistatestore"
)

func TestController_OidcPushedAuthorizationRequest(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
		q                     url.Values
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)
				mockOAuthProvider.EXPECT().NewPushedAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return(&fosite.PushedAuthorizeResponse{}, nil)
				mockOAuthProvider.EXPECT().WritePushedAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

				mockInteractionClient.EXPECT().PushAuthorizationDetails(gomock.Any(), gomock.Any()).Return(
					&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", `{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "invalid pushed authorize request",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(nil, errors.New("par error"))
				q = url.Values{}
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "par error")
			},
		},
		{
			name: "fail to unmarshal authorization details",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", "invalid")
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "invalid-value[authorization_details]")
			},
		},
		{
			name: "fail to validate authorization details",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", `{"type":"invalid","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "type should be 'openid_credential'")
			},
		},
		{
			name: "fail to push authorization details",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)
				mockInteractionClient.EXPECT().PushAuthorizationDetails(gomock.Any(), gomock.Any()).Return(nil, errors.New("push authorization details error"))

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", `{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "push authorization details error")
			},
		},
		{
			name: "invalid status code for push authorization details",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)

				mockInteractionClient.EXPECT().PushAuthorizationDetails(gomock.Any(), gomock.Any()).Return(
					&http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", `{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "push authorization details: status code")
			},
		},
		{
			name: "fail to create new pushed authorize response",
			setup: func() {
				mockOAuthProvider.EXPECT().NewPushedAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{}, nil)
				mockOAuthProvider.EXPECT().NewPushedAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("new pushed authorize response error"))

				mockInteractionClient.EXPECT().PushAuthorizationDetails(gomock.Any(), gomock.Any()).Return(
					&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(nil)),
					}, nil)

				q = url.Values{}
				q.Add("op_state", "opState")
				q.Add("authorization_details", `{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "new pushed authorize response error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				IssuerInteractionClient: mockInteractionClient,
				IssuerVCSPublicHost:     "https://issuer.example.com",
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(q.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()

			err := controller.OidcPushedAuthorizationRequest(echo.New().NewContext(req, rec))
			tt.check(t, rec, err)
		})
	}
}

func TestController_OidcAuthorize(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockStateStore        = NewMockStateStore(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
		oauth2Client          = NewMockOAuth2Client(gomock.NewController(t))
		params                oidc4ci.OidcAuthorizeParams
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType:         "code",
					IssuerState:          "opState",
					AuthorizationDetails: lo.ToPtr(`{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"ldp_vc"}`),
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				oauth2Client.EXPECT().AuthCodeURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						cfg oauth2.Config,
						state string,
						opts ...oauth2client.AuthCodeOption,
					) string {
						return (&cfg).AuthCodeURL(state)
					})
				mockOAuthProvider.EXPECT().NewAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						ar fosite.AuthorizeRequester,
						session fosite.Session,
					) (fosite.AuthorizeResponder, error) {
						assert.Equal(t, params.IssuerState, ar.(*fosite.AuthorizeRequest).State)

						return &fosite.AuthorizeResponse{}, nil
					})

				b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
					AuthorizationRequest: issuer.OAuthParameters{},
				})
				require.NoError(t, err)

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil
					})

				mockStateStore.EXPECT().SaveAuthorizeState(gomock.Any(), params.IssuerState, gomock.Any()).Return(nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.NotEmpty(t, rec.Header().Get("Location"))
			},
		},
		{
			name: "success with par",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockOAuthProvider.EXPECT().NewAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AuthorizeResponse{}, nil)

				parEndpoint := "https://localhost/par"
				parResponse := "https://localhost/authorize?request_uri=gfdsgfd2341321321"

				b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
					AuthorizationRequest:               issuer.OAuthParameters{},
					PushedAuthorizationRequestEndpoint: lo.ToPtr(parEndpoint),
				})
				require.NoError(t, err)

				oauth2Client.EXPECT().AuthCodeURLWithPAR(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						cfg oauth2.Config,
						parURL string,
						state string,
						client *http.Client,
						opts ...oauth2client.AuthCodeOption,
					) (string, error) {
						assert.Equal(t, parEndpoint, parURL)
						assert.Equal(t, params.IssuerState, state)

						return parResponse, nil
					})

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil
					})

				mockStateStore.EXPECT().SaveAuthorizeState(gomock.Any(), params.IssuerState, gomock.Any()).Return(nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusSeeOther, rec.Code)
				require.NotEmpty(t, rec.Header().Get("Location"))
			},
		},
		{
			name: "success with issuer par",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockOAuthProvider.EXPECT().NewAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AuthorizeResponse{}, nil)

				parEndpoint := "https://localhost/par"

				b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
					AuthorizationRequest:               issuer.OAuthParameters{},
					PushedAuthorizationRequestEndpoint: lo.ToPtr(parEndpoint),
				})
				require.NoError(t, err)

				oauth2Client.EXPECT().AuthCodeURLWithPAR(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("issuer par error"))

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil
					})

				mockStateStore.EXPECT().SaveAuthorizeState(gomock.Any(), params.IssuerState, gomock.Any()).Return(nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err, "issuer par error")
			},
		},
		{
			name: "invalid authorize request",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(nil, errors.New("authorize error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "authorize error")
			},
		},
		{
			name: "invalid authorization_details",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType:         "code",
					IssuerState:          "opState",
					AuthorizationDetails: lo.ToPtr("invalid"),
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "authorization_details")
			},
		},
		{
			name: "fail to validate authorization_details",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType:         "code",
					IssuerState:          "opState",
					AuthorizationDetails: lo.ToPtr(`{"type":"openid_credential","credential_type":"UniversityDegreeCredential","format":"invalid"}`),
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "authorization_details.format")
			},
		},
		{
			name: "prepare claim data authorization",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return nil, errors.New("prepare claim data authorization error")
					})
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "prepare claim data authorization")
			},
		},
		{
			name: "invalid status code for prepare claim data authorization",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(bytes.NewBufferString(`{"code":"system-error","component":"OIDC4CIService","message":"unexpected transaction from 5 to 3","operation":"PrepareClaimDataAuthorizationRequest","incorrectValue":"state"}`)),
						}, nil
					})
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "prepare claim data authorization: status code")
			},
		},
		{
			name: "fail to create authorize response",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockOAuthProvider.EXPECT().NewAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						ar fosite.AuthorizeRequester,
						session fosite.Session,
					) (fosite.AuthorizeResponder, error) {
						assert.Equal(t, params.IssuerState, ar.(*fosite.AuthorizeRequest).State)

						return nil, errors.New("create authorize response error")
					})

				b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
					AuthorizationRequest: issuer.OAuthParameters{},
				})
				require.NoError(t, err)

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil
					})
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "create authorize response error")
			},
		},
		{
			name: "fail to save authorize state",
			setup: func() {
				params = oidc4ci.OidcAuthorizeParams{
					ResponseType: "code",
					IssuerState:  "opState",
				}

				scope := []string{"openid", "profile"}

				mockOAuthProvider.EXPECT().NewAuthorizeRequest(gomock.Any(), gomock.Any()).Return(&fosite.AuthorizeRequest{
					Request: fosite.Request{RequestedScope: scope},
				}, nil)

				mockOAuthProvider.EXPECT().NewAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						ar fosite.AuthorizeRequester,
						session fosite.Session,
					) (fosite.AuthorizeResponder, error) {
						assert.Equal(t, params.IssuerState, ar.(*fosite.AuthorizeRequest).State)

						return &fosite.AuthorizeResponse{}, nil
					})

				b, err := json.Marshal(&issuer.PrepareClaimDataAuthorizationResponse{
					AuthorizationRequest: issuer.OAuthParameters{},
				})
				require.NoError(t, err)

				mockInteractionClient.EXPECT().PrepareAuthorizationRequest(gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						req issuer.PrepareAuthorizationRequestJSONRequestBody,
						reqEditors ...issuer.RequestEditorFn,
					) (*http.Response, error) {
						assert.Equal(t, params.ResponseType, req.ResponseType)
						assert.Equal(t, params.IssuerState, req.OpState)
						assert.Equal(t, lo.ToPtr(scope), req.Scope)

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil
					})

				mockStateStore.EXPECT().SaveAuthorizeState(gomock.Any(), params.IssuerState, gomock.Any()).Return(
					errors.New("save state error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "save authorize state")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				StateStore:              mockStateStore,
				IssuerInteractionClient: mockInteractionClient,
				IssuerVCSPublicHost:     "https://issuer.example.com",
				OAuth2Client:            oauth2Client,
			})

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()

			err := controller.OidcAuthorize(echo.New().NewContext(req, rec), params)
			tt.check(t, rec, err)
		})
	}
}

func TestController_OidcRedirect(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockStateStore        = NewMockStateStore(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
		params                oidc4ci.OidcRedirectParams
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			setup: func() {
				params = oidc4ci.OidcRedirectParams{
					Code:  "code",
					State: "state",
				}

				redirectURI := &url.URL{Scheme: "https", Host: "example.com", Path: "redirect"}

				mockStateStore.EXPECT().GetAuthorizeState(gomock.Any(), params.State).Return(&oidc4cistatestore.AuthorizeState{
					RedirectURI: redirectURI,
				}, nil)
				mockInteractionClient.EXPECT().StoreAuthorizationCodeRequest(
					gomock.Any(),
					issuer.StoreAuthorizationCodeRequest{
						Code:    params.Code,
						OpState: params.State,
					}).Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(nil))}, nil)

				mockOAuthProvider.EXPECT().WriteAuthorizeResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(
						ctx context.Context,
						rw http.ResponseWriter,
						ar fosite.AuthorizeRequester,
						responder fosite.AuthorizeResponder,
					) {
						assert.Equal(t, redirectURI, ar.GetRedirectURI())
						assert.Equal(t, params.State, ar.GetState())
					})
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "fail to store code",
			setup: func() {
				params = oidc4ci.OidcRedirectParams{
					Code:  "code",
					State: "state",
				}

				redirectURI := &url.URL{Scheme: "https", Host: "example.com", Path: "redirect"}

				mockStateStore.EXPECT().GetAuthorizeState(gomock.Any(), params.State).Return(&oidc4cistatestore.AuthorizeState{
					RedirectURI: redirectURI,
				}, nil)
				mockInteractionClient.EXPECT().StoreAuthorizationCodeRequest(
					gomock.Any(),
					issuer.StoreAuthorizationCodeRequest{
						Code:    params.Code,
						OpState: params.State,
					}).Return(nil, errors.New("random error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "random error")
			},
		},
		{
			name: "fail to get authorize state",
			setup: func() {
				params = oidc4ci.OidcRedirectParams{
					Code:  "code",
					State: "state",
				}

				mockStateStore.EXPECT().GetAuthorizeState(gomock.Any(), params.State).Return(nil, errors.New("get state error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "get state error")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				StateStore:              mockStateStore,
				IssuerInteractionClient: mockInteractionClient,
				IssuerVCSPublicHost:     "https://issuer.example.com",
			})

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()

			err := controller.OidcRedirect(echo.New().NewContext(req, rec), params)
			tt.check(t, rec, err)
		})
	}
}

func TestController_OidcToken(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
	)

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			setup: func() {
				opState := uuid.NewString()
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{
								Extra: map[string]interface{}{
									"opState": opState,
								},
							},
						},
					}, nil)

				mockInteractionClient.EXPECT().ExchangeAuthorizationCodeRequest(gomock.Any(),
					issuer.ExchangeAuthorizationCodeRequestJSONRequestBody{
						OpState: opState,
					}).
					Return(
						&http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`{"tx_id":"txID"}`)),
						}, nil)

				mockOAuthProvider.EXPECT().NewAccessResponse(gomock.Any(), gomock.Any()).Return(
					fosite.NewAccessResponse(), nil)

				mockOAuthProvider.EXPECT().WriteAccessResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "fail to create new access request",
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil, errors.New("new access request error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "new access request error")
			},
		},
		{
			name: "fail to create new access response",
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{
								Extra: map[string]interface{}{
									"opState": "1234",
								},
							},
						},
					}, nil)

				mockInteractionClient.EXPECT().ExchangeAuthorizationCodeRequest(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`{"tx_id":"txID"}`)),
						}, nil)

				mockOAuthProvider.EXPECT().NewAccessResponse(gomock.Any(), gomock.Any()).Return(
					nil, errors.New("new access response error"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "new access response error")
			},
		},
		{
			name: "fail to exchange token",
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{
								Extra: map[string]interface{}{
									"opState": "1234",
								},
							},
						},
					}, nil)

				mockInteractionClient.EXPECT().ExchangeAuthorizationCodeRequest(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("can not exchange token"))
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "can not exchange token")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				IssuerInteractionClient: mockInteractionClient,
			})

			req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()

			err := controller.OidcToken(echo.New().NewContext(req, rec))
			tt.check(t, rec, err)
		})
	}
}

func TestController_OidcCredential(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
		accessToken           string
		requestBody           []byte
	)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	jwtVerifier, err := jwt.NewEd25519Verifier(publicKey)
	require.NoError(t, err)

	jwtSigner := jwt.NewEd25519Signer(privateKey)

	signedJWT, err := jwt.NewSigned(&oidc4ci.JWTProofClaims{
		Issuer:   clientID,
		IssuedAt: time.Now().Unix(),
		Nonce:    "c_nonce",
	}, nil, jwtSigner)
	require.NoError(t, err)

	jws, err := signedJWT.Serialize(false)
	require.NoError(t, err)

	credentialReq := oidc4ci.CredentialRequest{
		Format: lo.ToPtr(string(common.JwtVcJson)),
		Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: jws},
		Type:   "UniversityDegreeCredential",
	}

	tests := []struct {
		name  string
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				b, marshalErr := json.Marshal(issuer.PrepareCredentialResult{
					Credential: "credential in jwt format",
					Format:     string(common.JwtVcJson),
				})
				require.NoError(t, marshalErr)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(b)),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "invalid credential format",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).Times(0)
				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr("invalid"),
					Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: jws},
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "unsupported vc format")
			},
		},
		{
			name: "missing proof type",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).Times(0)
				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr(string(common.JwtVcJson)),
					Proof:  nil,
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "missing proof type")
			},
		},
		{
			name: "invalid proof type",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).Times(0)
				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr(string(common.JwtVcJson)),
					Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: ""},
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "invalid proof type")
			},
		},
		{
			name: "missing access token",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).Times(0)
				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = ""

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "missing access token")
			},
		},
		{
			name: "fail to introspect token",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(fosite.AccessToken, nil, errors.New("introspect error"))

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "introspect token")
			},
		},
		{
			name: "fail to parse proof jwt",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr(string(common.JwtVcJson)),
					Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: "invalid jws"},
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "parse jwt")
			},
		},
		{
			name: "fail to decode proof jwt claims",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				invalidClaimsJWT, jwtErr := jwt.NewSigned(`{"iat":"invalid"}`, nil, jwtSigner)
				require.NoError(t, jwtErr)

				invalidJWS, marshalErr := invalidClaimsJWT.Serialize(false)
				require.NoError(t, marshalErr)

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr(string(common.JwtVcJson)),
					Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: invalidJWS},
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "decode claims")
			},
		},
		{
			name: "nonce expired",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(-time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "nonce expired")
			},
		},
		{
			name: "invalid nonce",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).Times(0)

				accessToken = "access-token"

				invalidNonceJWT, jwtErr := jwt.NewSigned(&oidc4ci.JWTProofClaims{
					Issuer:   clientID,
					IssuedAt: time.Now().Unix(),
					Nonce:    "invalid",
				}, nil, jwtSigner)
				require.NoError(t, jwtErr)

				invalidJWS, marshalErr := invalidNonceJWT.Serialize(false)
				require.NoError(t, marshalErr)

				requestBody, err = json.Marshal(oidc4ci.CredentialRequest{
					Format: lo.ToPtr(string(common.JwtVcJson)),
					Proof:  &oidc4ci.JWTProof{ProofType: "jwt", Jwt: invalidJWS},
					Type:   "UniversityDegreeCredential",
				})
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "invalid nonce")
			},
		},
		{
			name: "fail to prepare credential",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("prepare credential error"))

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "prepare credential")
			},
		},
		{
			name: "invalid status code in prepare credential response (format)",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(strings.NewReader(`{"code" : "oidc-credential-format-not-supported"}`)),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err,
					"oidc-error[]: prepare credential: status code 500, "+
						"code: oidc-credential-format-not-supported")
			},
		},
		{
			name: "invalid status code in prepare credential response (invalid json)",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(strings.NewReader(`{`)),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "prepare credential: status code 500, {")
			},
		},
		{
			name: "invalid status code in prepare credential response (type)",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(strings.NewReader(`{"code" : "oidc-credential-type-not-supported"}`)),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err,
					"oidc-error[]: prepare credential: status code 500, "+
						"code: oidc-credential-type-not-supported")
			},
		},
		{
			name: "invalid status code in prepare credential response (random)",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       io.NopCloser(strings.NewReader(`{"code" : "random", "message": "awesome"}`)),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err,
					"prepare credential: status code 500, code: random; message: awesome")
			},
		},
		{
			name: "fail to decode prepare credential result",
			setup: func() {
				mockOAuthProvider.EXPECT().IntrospectToken(gomock.Any(), gomock.Any(), fosite.AccessToken, gomock.Any()).
					Return(
						fosite.AccessToken,
						fosite.NewAccessRequest(
							&fosite.DefaultSession{
								Extra: map[string]interface{}{
									"txID":            "tx_id",
									"cNonce":          "c_nonce",
									"cNonceExpiresAt": time.Now().Add(time.Minute).Unix(),
								},
							},
						), nil)

				mockInteractionClient.EXPECT().PrepareCredential(gomock.Any(), gomock.Any()).
					Return(
						&http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
						}, nil)

				accessToken = "access-token"

				requestBody, err = json.Marshal(credentialReq)
				require.NoError(t, err)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.ErrorContains(t, err, "decode prepare credential result")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				IssuerInteractionClient: mockInteractionClient,
				JWTVerifier:             jwtVerifier,
			})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			if accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			}

			rec := httptest.NewRecorder()

			err := controller.OidcCredential(echo.New().NewContext(req, rec))
			tt.check(t, rec, err)
		})
	}
}

func TestController_OidcPreAuthorize(t *testing.T) {
	var (
		mockOAuthProvider     = NewMockOAuth2Provider(gomock.NewController(t))
		mockInteractionClient = NewMockIssuerInteractionClient(gomock.NewController(t))
		oauthClient           = NewMockOAuth2Client(gomock.NewController(t))
		preAuthorizeClient    = NewMockHTTPClient(gomock.NewController(t))
	)

	tests := []struct {
		name  string
		body  io.Reader
		setup func()
		check func(t *testing.T, rec *httptest.ResponseRecorder, err error)
	}{
		{
			name: "success",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"123456"},
				"user_pin":            {"5678"},
			}.Encode()),
			setup: func() {
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(), gomock.Any()).
					Return(&http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"scopes" : ["a","b"], "op_state" : "opp123"}`)),
					}, nil)

				accessRq := &fosite.AccessRequest{
					Request: fosite.Request{
						Session: &fosite.DefaultSession{},
					},
				}

				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(accessRq, nil)

				mockOAuthProvider.EXPECT().NewAccessResponse(gomock.Any(), accessRq).
					Return(&fosite.AccessResponse{
						AccessToken: "123456",
						Extra: map[string]interface{}{
							"expires_in": 3600,
						},
					}, nil)

				mockOAuthProvider.EXPECT().WriteAccessResponse(gomock.Any(), gomock.Any(), accessRq, gomock.Any()).
					Do(func(ctx context.Context, rw http.ResponseWriter, requester fosite.AccessRequester, responder fosite.AccessResponder) {
						js, err := json.Marshal(responder.ToMap())
						if err != nil {
							http.Error(rw, err.Error(), http.StatusInternalServerError)
							return
						}

						rw.Header().Set("Content-Type", "application/json;charset=UTF-8")

						rw.WriteHeader(http.StatusOK)
						_, _ = rw.Write(js)
					})
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.NoError(t, err)
				var resp oidc4ci.AccessTokenResponse

				assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, "123456", resp.AccessToken)
				assert.NotEmpty(t, *resp.ExpiresIn)
			},
		},
		{
			name: "name invalid pre-auth code",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"123456"},
				"user_pin":            {"5678"},
			}.Encode()),
			setup: func() {
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(),
					issuer.ValidatePreAuthorizedCodeRequestJSONRequestBody{
						PreAuthorizedCode: "123456",
						UserPin:           lo.ToPtr("5678"),
					}).Return(nil, errors.New("invalid pin"))

				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{},
						},
					}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err, "invalid pin")
			},
		},
		{
			name: "invalid response from validator",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"321"},
			}.Encode()),
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{},
						},
					}, nil)
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(),
					issuer.ValidatePreAuthorizedCodeRequestJSONRequestBody{
						PreAuthorizedCode: "321",
						UserPin:           lo.ToPtr(""),
					}).Return(&http.Response{
					Body:       io.NopCloser(strings.NewReader("{")),
					StatusCode: http.StatusOK,
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err, "unexpected EOF")
			},
		},
		{
			name: "invalid http code from validate pre authorize",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"321"},
			}.Encode()),
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{},
						},
					}, nil)
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(),
					issuer.ValidatePreAuthorizedCodeRequestJSONRequestBody{
						PreAuthorizedCode: "321",
						UserPin:           lo.ToPtr(""),
					}).Return(&http.Response{
					Body:       io.NopCloser(strings.NewReader("{}")),
					StatusCode: http.StatusBadRequest,
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err, "validate pre-authorized code request: status code 400, code")
			},
		},
		{
			name: "invalid tx validate pre authorize",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"321"},
			}.Encode()),
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{},
						},
					}, nil)
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(),
					issuer.ValidatePreAuthorizedCodeRequestJSONRequestBody{
						PreAuthorizedCode: "321",
						UserPin:           lo.ToPtr(""),
					}).Return(&http.Response{
					Body:       io.NopCloser(strings.NewReader(`{"code": "oidc-tx-not-found"}`)),
					StatusCode: http.StatusBadRequest,
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err,
					"oidc-error[]: validate pre-authorized code request: status code 400, "+
						"code: oidc-tx-not-found")
			},
		},
		{
			name: "invalid expect pin pre authorize",
			body: strings.NewReader(url.Values{
				"grant_type":          {"urn:ietf:params:oauth:grant-type:pre-authorized_code"},
				"pre-authorized_code": {"321"},
			}.Encode()),
			setup: func() {
				mockOAuthProvider.EXPECT().NewAccessRequest(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&fosite.AccessRequest{
						Request: fosite.Request{
							Session: &fosite.DefaultSession{},
						},
					}, nil)
				mockInteractionClient.EXPECT().ValidatePreAuthorizedCodeRequest(gomock.Any(),
					issuer.ValidatePreAuthorizedCodeRequestJSONRequestBody{
						PreAuthorizedCode: "321",
						UserPin:           lo.ToPtr(""),
					}).Return(&http.Response{
					Body:       io.NopCloser(strings.NewReader(`{"code": "oidc-pre-authorize-expect-pin"}`)),
					StatusCode: http.StatusBadRequest,
				}, nil)
			},
			check: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				assert.ErrorContains(t, err,
					"oidc-error[]: validate pre-authorized code request: status code 400, "+
						"code: oidc-pre-authorize-expect-pin")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			controller := oidc4ci.NewController(&oidc4ci.Config{
				OAuth2Provider:          mockOAuthProvider,
				IssuerInteractionClient: mockInteractionClient,
				OAuth2Client:            oauthClient,
				PreAuthorizeClient:      preAuthorizeClient,
			})

			req := httptest.NewRequest(http.MethodPost, "/", tt.body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

			rec := httptest.NewRecorder()

			err := controller.OidcToken(echo.New().NewContext(req, rec))
			tt.check(t, rec, err)
		})
	}
}
