package hydrautil

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

func makeState(r *http.Request) string {
	return makeHash(r.RemoteAddr + r.UserAgent())
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
