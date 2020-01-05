package hydrautil

import (
	"context"
	"errors"
	"net/http"
)

// CheckAuthHandler returns an http.Handler that will check the
// cookies for the access token and then verify it
func CheckAuthHandler(conf ClientConfig, next http.Handler) http.Handler {
	if conf.MissingCookieStatus == 0 {
		conf.MissingCookieStatus = http.StatusUnauthorized
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range conf.Bypasses {
			if b.canBypass(r) {
				next.ServeHTTP(w, r)
				return
			}
		}

		var token string
		if conf.CookieName != "" {
			cookie, err := r.Cookie(conf.CookieName)
			if err != nil {
				w.WriteHeader(conf.MissingCookieStatus)
				return
			}
			token = cookie.Value
		} else if conf.HeaderName != "" {
			token = r.Header.Get(conf.HeaderName)
		} else {
			sendErrorMessage(w, 500, "no cookie or header name configured")
		}

		if token == "" {
			w.WriteHeader(conf.MissingCookieStatus)
			return
		}

		ui, err := getUserInfo(conf, token)
		if err != nil {
			uiErr := userInfoError{}

			if errors.As(err, &uiErr) {
				debugf("user info error: %#v\n", uiErr)
				w.WriteHeader(uiErr.code)
				w.Write(uiErr.body)
			} else {
				debugf("user info error: %s\n", uiErr)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}

			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserInfo, ui)

		debugf("got user info: %#v\n", ui)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
