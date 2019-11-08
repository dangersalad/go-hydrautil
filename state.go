package hydrautil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

// GetStateFunc is a function that takes in a request and generates
// state for the oauth exchange
type GetStateFunc func(r *http.Request) string

func getStateFunc(conf ClientConfig) GetStateFunc {
	if conf.GetState == nil {
		return defaultGetState
	}
	return conf.GetState
}

func defaultGetState(r *http.Request) string {
	forwardedFor := r.Header.Get("x-forwarded-for")
	if forwardedFor == "" {
		debug("no value for x-forwarded-for, checking x-real-ip")
		forwardedFor = r.Header.Get("x-real-ip")
	}
	if forwardedFor == "" {
		debug("no value for x-real-ip, using http.Request.RemoteAddr")
		forwardedFor = r.RemoteAddr
	}
	hashData := forwardedFor + r.UserAgent()
	state := makeHash(hashData)
	debugf("state from %s = %s\n", hashData, state)
	return state
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
