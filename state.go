package hydrautil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

func makeState(r *http.Request) string {
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
