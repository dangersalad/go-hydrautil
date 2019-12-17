package hydrautil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func sendErrorMessage(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(bytes)
}
