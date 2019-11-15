package hydrautil

import (
	"context"
	"fmt"
	"net/http"

	hydraRuntime "github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	hydraPublic "github.com/ory/hydra/sdk/go/hydra/client/public"
	"github.com/ory/hydra/sdk/go/hydra/models"
)

type contextKey string

// ContextKeyUserInfo is the context key for the user info
var ContextKeyUserInfo contextKey = "userinfo"

// UserInfo is the user info
type UserInfo models.SwaggeruserinfoResponsePayload

// ErrNoUserInfo is the error returned by UserInfoFromContext when the
// user info is missing from the context
var ErrNoUserInfo = fmt.Errorf("missing user info")

// UserInfoFromContext returns the userinfo on the context
func UserInfoFromContext(ctx context.Context) (UserInfo, error) {
	val := ctx.Value(ContextKeyUserInfo)
	if ui, ok := val.(UserInfo); ok {
		return ui, nil
	}
	return UserInfo{}, ErrNoUserInfo
}

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

		cookie, err := r.Cookie(conf.CookieName)
		if err != nil {
			w.WriteHeader(conf.MissingCookieStatus)
			return
		}

		userParams := hydraPublic.NewUserinfoParamsWithContext(r.Context())
		userAuthFunc := func(hydraReq hydraRuntime.ClientRequest, _ strfmt.Registry) error {
			if cookie.Value == "" {
				return fmt.Errorf("no value in cookie")
			}
			hydraReq.SetHeaderParam("authorization", fmt.Sprintf(`Bearer %s`, cookie.Value))
			return nil
		}

		userInfo, err := conf.Hydra.Public.Userinfo(userParams, hydraRuntime.ClientAuthInfoWriterFunc(userAuthFunc))
		if err != nil {

			switch hydraErr := err.(type) {

			case *hydraPublic.UserinfoUnauthorized:
				errData := hydraErr.GetPayload()
				err = fmt.Errorf("[%d] %s - %s (%s)", errData.Code, *errData.Name, errData.Description, errData.Debug)
				w.WriteHeader(int(errData.Code))

			case *hydraPublic.UserinfoInternalServerError:
				errData := hydraErr.GetPayload()
				err = fmt.Errorf("[%d] %s - %s (%s)", errData.Code, *errData.Name, errData.Description, errData.Debug)
				w.WriteHeader(int(errData.Code))

			case *hydraRuntime.APIError:
				if runtimeResp, ok := hydraErr.Response.(hydraRuntime.ClientResponse); ok {
					err = fmt.Errorf("unknown error getting userinfo: [%d] %s", hydraErr.Code, runtimeResp.Message())
				} else {
					err = fmt.Errorf("unknown error getting userinfo: [%d] %#v", hydraErr.Code, hydraErr.Response)
				}
				// w.WriteHeader(hydraErr.Code)
				w.WriteHeader(http.StatusInternalServerError)
			}

			debug("got error getting userinfo: %s", err)
			w.Write([]byte(err.Error()))
			return
		}

		if userInfo == nil || userInfo.GetPayload() == nil {
			debug("userinfo or it's payload is nil")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("userInfo is nil"))
			return
		}

		data := userInfo.GetPayload()
		ctx := context.WithValue(r.Context(), ContextKeyUserInfo, UserInfo(*data))

		debugf("got user info: %#v\n", userInfo.GetPayload())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
