package hydrautil

import (
	"net/http"

	"golang.org/x/oauth2"
)

// AuthHandler returns an http.Handler that redirects the request to
// the configured OAuth server
func AuthHandler(oauthConf *oauth2.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, err := makeState(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		authURL := oauthConf.AuthCodeURL(state, oauth2.AccessTypeOnline)
		debugf("redirecting to %s\n", authURL)
		w.Header().Add("location", authURL)
		w.WriteHeader(http.StatusFound)
	})
}
