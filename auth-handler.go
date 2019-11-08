package hydrautil

import (
	"net/http"

	"golang.org/x/oauth2"
)

// AuthHandler returns an http.Handler that redirects the request to
// the configured OAuth server
func AuthHandler(oauthConf *oauth2.Config, clientConf ClientConfig) http.Handler {
	getState := getStateFunc(clientConf)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authURL := oauthConf.AuthCodeURL(getState(r), oauth2.AccessTypeOnline)
		debugf("redirecting to %s\n", authURL)
		w.Header().Add("location", authURL)
		w.WriteHeader(http.StatusFound)
	})
}
