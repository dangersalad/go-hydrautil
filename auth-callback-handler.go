package hydrautil

import (
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

// AuthCallbackHandler returns an http.Handler that takes the params
// provided by the oauth server and exchanges them for an access token
func AuthCallbackHandler(oauthConf *oauth2.Config, clientConf ClientConfig) http.Handler {

	getState := getStateFunc(clientConf)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referrer, err := url.Parse(r.Header.Get("referrer"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid referror"))
			return
		}

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
		if state != getState(r) {
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

		secure := referrer.Scheme == "https"

		cookieDomain := r.Host
		if r.Header.Get("x-cookie-domain") != "" {
			cookieDomain = r.Header.Get("x-cookie-domain")
		} else if r.Header.Get("host") != "" {
			cookieDomain = r.Header.Get("host")
		}

		debugf("assigning token to cookie %s for domain %s: %#v\n", clientConf.CookieName, cookieDomain, token)

		http.SetCookie(w, &http.Cookie{
			Name:     clientConf.CookieName,
			Value:    token.AccessToken,
			HttpOnly: true,
			Secure:   secure,
			Domain:   cookieDomain,
			Path:     "/",
		})

	})
}
