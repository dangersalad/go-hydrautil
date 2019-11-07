package hydrautil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

var stateHeaders = []string{
	"x-forwarded-for",
	"user-agent",
}

func makeState(r *http.Request) string {
	forwardedFor := r.Header.Get("x-forwarded-for")
	if forwardedFor == "" {
		forwardedFor = r.Header.Get("x-real-ip")
	}
	if forwardedFor == "" {
		forwardedFor = r.RemoteAddr
	}
	hashData := forwardedFor + r.UserAgent()
	state := makeHash(hashData)
	debugf("state from %s = %s", hashData, state)
	return state
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
