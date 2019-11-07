package hydrautil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

func makeState(r *http.Request) string {
	state := makeHash(r.RemoteAddr + r.UserAgent())
	debugf("state from %s and %s = %s", r.RemoteAddr, r.UserAgent(), state)
	return state
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
