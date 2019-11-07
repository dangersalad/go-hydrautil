package hydrautil

import (
	hydra "github.com/ory/hydra/sdk/go/hydra/client"
)

// ClientConfig configures the client connection
type ClientConfig struct {
	CookieName          string
	Bypasses            []*Bypass
	MissingCookieStatus int
	Hydra               *hydra.OryHydra
}
