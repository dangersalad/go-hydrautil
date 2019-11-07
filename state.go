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

func makeState(r *http.Request) (string, error) {
	hashData := r.UserAgent()
	for _, header := range stateHeaders {
		val := r.Header.Get(header)
		if val == "" {
			return "", fmt.Errorf("missing %s header", header)
		}
		hashData += val
	}
	state := makeHash(hashData)
	debugf("state from %s = %s", hashData, state)
	return state, nil
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
