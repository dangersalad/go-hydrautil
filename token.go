package hydrautil

import (
	"context"
	"fmt"
)

// ErrNoUserToken is the error returned by UserTokenFromContext when the
// user token is missing from the context
var ErrNoUserToken = fmt.Errorf("missing user token")

// ContextKeyUserToken is the context key for the user token
var ContextKeyUserToken contextKey = "token"

// UserTokenFromContext returns the userinfo on the context
func UserTokenFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(ContextKeyUserToken)
	if ui, ok := val.(string); ok {
		return ui, nil
	}
	return "", ErrNoUserToken
}
