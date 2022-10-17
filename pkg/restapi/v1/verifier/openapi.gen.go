// Package verifier provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package verifier

import (
	"fmt"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
)

// InitiateOIDC4VPData defines model for InitiateOIDC4VPData.
type InitiateOIDC4VPData struct {
	PresentationDefinitionId *string `json:"presentationDefinitionId,omitempty"`
	Purpose                  *string `json:"purpose,omitempty"`
}

// InitiateOIDC4VPResponse defines model for InitiateOIDC4VPResponse.
type InitiateOIDC4VPResponse struct {
	AuthorizationRequest string `json:"authorizationRequest"`
	TxID                 string `json:"txID"`
}

// Verify credential response containing failure check details.
type VerifyCredentialCheckResult struct {
	// Check title.
	Check string `json:"check"`

	// Error message.
	Error string `json:"error"`

	// Verification method.
	VerificationMethod string `json:"verificationMethod"`
}

// Model for credential verification.
type VerifyCredentialData struct {
	// Credential in jws(string) or jsonld(object) formats.
	Credential interface{} `json:"credential"`

	// Options for verify credential.
	Options *VerifyCredentialOptions `json:"options,omitempty"`
}

// Options for verify credential.
type VerifyCredentialOptions struct {
	// Chalange is added to the proof.
	Challenge *string `json:"challenge,omitempty"`

	// Domain is added to the proof.
	Domain *string `json:"domain,omitempty"`
}

// Model for response of credentials verification.
type VerifyCredentialResponse struct {
	Checks *[]VerifyCredentialCheckResult `json:"checks,omitempty"`
}

// Verify presentation response containing failure check details.
type VerifyPresentationCheckResult struct {
	// Check title.
	Check string `json:"check"`

	// Error message.
	Error string `json:"error"`
}

// Model for presentation verification.
type VerifyPresentationData struct {
	// Options for verify presentation.
	Options *VerifyPresentationOptions `json:"options,omitempty"`

	// Presentation in jws(string) or jsonld(object) formats.
	Presentation interface{} `json:"presentation"`
}

// Options for verify presentation.
type VerifyPresentationOptions struct {
	// Challenge is added to the proof.
	Challenge *string `json:"challenge,omitempty"`

	// Domain is added to the proof.
	Domain *string `json:"domain,omitempty"`
}

// Model for response of presentation verification.
type VerifyPresentationResponse struct {
	Checks *[]VerifyPresentationCheckResult `json:"checks,omitempty"`
}

// PostVerifyCredentialsJSONBody defines parameters for PostVerifyCredentials.
type PostVerifyCredentialsJSONBody = VerifyCredentialData

// InitiateOidcInteractionJSONBody defines parameters for InitiateOidcInteraction.
type InitiateOidcInteractionJSONBody = InitiateOIDC4VPData

// PostVerifyPresentationJSONBody defines parameters for PostVerifyPresentation.
type PostVerifyPresentationJSONBody = VerifyPresentationData

// PostVerifyCredentialsJSONRequestBody defines body for PostVerifyCredentials for application/json ContentType.
type PostVerifyCredentialsJSONRequestBody = PostVerifyCredentialsJSONBody

// InitiateOidcInteractionJSONRequestBody defines body for InitiateOidcInteraction for application/json ContentType.
type InitiateOidcInteractionJSONRequestBody = InitiateOidcInteractionJSONBody

// PostVerifyPresentationJSONRequestBody defines body for PostVerifyPresentation for application/json ContentType.
type PostVerifyPresentationJSONRequestBody = PostVerifyPresentationJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Used by verifier applications to initiate OpenID presentation flow through VCS
	// (POST /verifier/interactions/authorization-response)
	CheckAuthorizationResponse(ctx echo.Context) error
	// Used by verifier applications to get claims obtained during oidc4vp interaction.
	// (GET /verifier/interactions/{txID}/claim)
	RetrieveInteractionsClaim(ctx echo.Context, txID string) error
	// Verify credential
	// (POST /verifier/profiles/{profileID}/credentials/verify)
	PostVerifyCredentials(ctx echo.Context, profileID string) error
	// Used by verifier applications to initiate OpenID presentation flow through VCS
	// (POST /verifier/profiles/{profileID}/interactions/initiate-oidc)
	InitiateOidcInteraction(ctx echo.Context, profileID string) error
	// Verify presentation
	// (POST /verifier/profiles/{profileID}/presentations/verify)
	PostVerifyPresentation(ctx echo.Context, profileID string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// CheckAuthorizationResponse converts echo context to params.
func (w *ServerInterfaceWrapper) CheckAuthorizationResponse(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CheckAuthorizationResponse(ctx)
	return err
}

// RetrieveInteractionsClaim converts echo context to params.
func (w *ServerInterfaceWrapper) RetrieveInteractionsClaim(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "txID" -------------
	var txID string

	err = runtime.BindStyledParameterWithLocation("simple", false, "txID", runtime.ParamLocationPath, ctx.Param("txID"), &txID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter txID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RetrieveInteractionsClaim(ctx, txID)
	return err
}

// PostVerifyCredentials converts echo context to params.
func (w *ServerInterfaceWrapper) PostVerifyCredentials(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "profileID" -------------
	var profileID string

	err = runtime.BindStyledParameterWithLocation("simple", false, "profileID", runtime.ParamLocationPath, ctx.Param("profileID"), &profileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter profileID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostVerifyCredentials(ctx, profileID)
	return err
}

// InitiateOidcInteraction converts echo context to params.
func (w *ServerInterfaceWrapper) InitiateOidcInteraction(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "profileID" -------------
	var profileID string

	err = runtime.BindStyledParameterWithLocation("simple", false, "profileID", runtime.ParamLocationPath, ctx.Param("profileID"), &profileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter profileID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.InitiateOidcInteraction(ctx, profileID)
	return err
}

// PostVerifyPresentation converts echo context to params.
func (w *ServerInterfaceWrapper) PostVerifyPresentation(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "profileID" -------------
	var profileID string

	err = runtime.BindStyledParameterWithLocation("simple", false, "profileID", runtime.ParamLocationPath, ctx.Param("profileID"), &profileID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter profileID: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostVerifyPresentation(ctx, profileID)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/verifier/interactions/authorization-response", wrapper.CheckAuthorizationResponse)
	router.GET(baseURL+"/verifier/interactions/:txID/claim", wrapper.RetrieveInteractionsClaim)
	router.POST(baseURL+"/verifier/profiles/:profileID/credentials/verify", wrapper.PostVerifyCredentials)
	router.POST(baseURL+"/verifier/profiles/:profileID/interactions/initiate-oidc", wrapper.InitiateOidcInteraction)
	router.POST(baseURL+"/verifier/profiles/:profileID/presentations/verify", wrapper.PostVerifyPresentation)

}
