package hydrautil

import (
	"fmt"
	"net/http"

	hydraRuntime "github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	hydraPublic "github.com/ory/hydra/sdk/go/hydra/client/public"
)

// CheckAuthHandler returns an http.Handler that will check the
// cookies for the access token and then verify it
func CheckAuthHandler(next http.Handler, conf ClientConfig) http.Handler {
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

			w.Write([]byte(err.Error()))
			return
		}

		if userInfo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("userInfo is nil"))
			return
		}

		debugf("got user info: %#v\n", userInfo.GetPayload())

		next.ServeHTTP(w, r)
	})
}
