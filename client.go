package hydrautil

import (
	hydra "github.com/ory/hydra/sdk/go/hydra/client"
)

// ClientConfig configures the client connection
type ClientConfig struct {

	// CookieName to be used for storing the access token in the auth
	// callback flow, if missing this and HeaderName a JSON response
	// will be generated instead, intended to be consumed by the
	// application. This setting will supercede HeaderName
	CookieName string

	// HeaderName to be used for storing the access token in the auth
	// callback flow, if missing this and CookieName a JSON response
	// will be generated instead, intended to be consumed by the
	// application. CookieName will supercede this setting
	HeaderName string

	// This is the status to be returned from the auth check handler
	// if the cookie is missing, defaults to http.StatusUnauthorized
	// if not specified
	MissingCookieStatus int

	// Bypasses will allow the auth check handler to bypass certain urls
	Bypasses []*Bypass

	// Hydra is the hydra client to use for the oauth flows
	Hydra *hydra.OryHydra

	// GetState can be defined to override the state generation. If it
	// is not specified, StateKey is required.
	GetState GetStateFunc

	// ValidateState can be defined to override the state validation. If it
	// is not specified, StateKey is required.
	ValidateState ValidateStateFunc

	// StateKey is the key used to hash the state data in the default
	// state generation function. Required unless GetState is defined.
	StateKey string
}
