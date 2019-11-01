package hydrautil

import (
	"fmt"
	"io/ioutil"
	"net/http"

	hydraRuntime "github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	hydra "github.com/ory/hydra/sdk/go/hydra/client"
	hydraPublic "github.com/ory/hydra/sdk/go/hydra/client/public"
)

// ClientConfig configures the client connection
type ClientConfig struct {
	CookieName          string
	Bypasses            []*Bypass
	MissingCookieStatus int
	Hydra               *hydra.OryHydra
}

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
					respData, readErr := ioutil.ReadAll(runtimeResp.Body())
					if readErr != nil {
						err = fmt.Errorf("[%d] unknown error trying %s: error reading body: %w", hydraErr.Code, hydraErr.OperationName, readErr)
					} else {
						err = fmt.Errorf("[%d] unknown error trying %s: %s", hydraErr.Code, hydraErr.OperationName, respData)
					}
				} else {
					err = fmt.Errorf("[%d] unknown error trying %s: %#v", hydraErr.Code, hydraErr.OperationName, hydraErr.Response)
				}
				w.WriteHeader(hydraErr.Code)
			}
			w.Write([]byte(err.Error()))
			return
		}

		if userInfo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("userInfo is nil"))
			return
		}
		fmt.Printf("user info: %#v\n", userInfo.GetPayload())
		next.ServeHTTP(w, r)
	})
}
