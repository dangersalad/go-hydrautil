package hydrautil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
)

// GetStateFunc is a function that takes in a request and generates
// state for the oauth exchange
type GetStateFunc func(r *http.Request) string

// ValidateStateFunc is a function that takes in a state and an
// expected state and validates them
type ValidateStateFunc func(state, expectedState string) bool

func getStateFunc(conf ClientConfig) GetStateFunc {
	if conf.GetState == nil {
		if conf.StateKey == "" {
			panic(fmt.Errorf("state key is missing from client configuration"))
		}
		return defaultGetState(conf.StateKey)
	}
	return conf.GetState
}

func validateStateFunc(conf ClientConfig) ValidateStateFunc {
	if conf.ValidateState == nil {
		if conf.StateKey == "" {
			panic(fmt.Errorf("state key is missing from client configuration"))
		}
		return defaultValidateState(conf.StateKey)
	}
	return conf.ValidateState

}

func defaultGetState(stateKey string) GetStateFunc {
	return func(r *http.Request) string {
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
		state = signData(stateKey, state)
		debugf("state from %s = %s\n", hashData, state)
		return state
	}
}

func defaultValidateState(stateKey string) ValidateStateFunc {
	return func(state, expectedState string) bool {
		return stateValid(stateKey, state, expectedState)
	}
}

func makeHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func signData(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func stateValid(key, data, signed string) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	macMessage, err := hex.DecodeString(signed)
	if err != nil {
		logf("error decoding passed state %s: %s", signed, err)
		return false
	}
	return hmac.Equal(mac.Sum(nil), macMessage)
}
