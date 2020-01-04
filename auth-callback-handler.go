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
	validateState := validateStateFunc(clientConf)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		// if no code, return error
		if code == "" {
			sendErrorMessage(w, http.StatusBadRequest, "code required")
			return
		}
		scope := r.URL.Query().Get("scope")
		// if no challenge, return error
		if scope == "" {
			sendErrorMessage(w, http.StatusBadRequest, "scope required")
			return
		}
		state := r.URL.Query().Get("state")
		// if no challenge, return error
		if scope == "" {
			sendErrorMessage(w, http.StatusBadRequest, "state required")
			return
		}
		if validateState(state, getState(r)) {
			sendErrorMessage(w, http.StatusUnauthorized, "state mismatch")
			return
		}
		token, err := oauthConf.Exchange(r.Context(), code)
		if err != nil {
			if oauthErr, ok := err.(*oauth2.RetrieveError); ok {
				sendErrorMessage(w, http.StatusUnauthorized, string(oauthErr.Body))
				return
			}
			sendErrorMessage(w, http.StatusInternalServerError, err.Error())
			return
		}
		if token == nil {
			sendErrorMessage(w, http.StatusInternalServerError, "token is nil")
			return
		}

		if clientConf.CookieName != "" {
			referrer, err := url.Parse(r.Header.Get("referrer"))
			if err != nil {
				sendErrorMessage(w, http.StatusBadRequest, "invalid referror")
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
			return
		} else if clientConf.HeaderName != "" {
			w.Header().Set(clientConf.HeaderName, token)
			return
		}

		sendJSON(w, token)
	})
}
