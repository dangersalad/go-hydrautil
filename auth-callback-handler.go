package hydrautil

import (
	"net/http"
	"regexp"

	"golang.org/x/oauth2"
)

var originParse = regexp.MustCompile(`^(https?)://([^/]+).*$`)

// AuthCallbackHandler returns an http.Handler that takes the params
// provided by the oauth server and exchanges them for an access token
func AuthCallbackHandler(oauthConf *oauth2.Config, clientConf ClientConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		// if no code, return error
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("code required"))
			return
		}
		scope := r.URL.Query().Get("scope")
		// if no challenge, return error
		if scope == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("scope required"))
			return
		}
		state := r.URL.Query().Get("state")
		// if no challenge, return error
		if scope == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("state required"))
			return
		}
		if state != makeState(r) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("state mismatch"))
			return
		}
		token, err := oauthConf.Exchange(r.Context(), code)
		if err != nil {
			if oauthErr, ok := err.(*oauth2.RetrieveError); ok {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write(oauthErr.Body)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if token == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("token is nil"))
			return
		}

		origin := r.Header.Get("origin")
		var secure bool

		matches := originParse.FindStringSubmatch(origin)
		if len(matches) == 3 {
			debug("making secure cookie")
			secure = matches[1] == "https"
		}

		debugf("assigning token to cookie %s: %#v\n", clientConf.CookieName, token)

		http.SetCookie(w, &http.Cookie{
			Name:     clientConf.CookieName,
			Value:    token.AccessToken,
			HttpOnly: true,
			Secure:   secure,
			Domain:   r.Host,
			Path:     "/",
		})

	})
}
