package api

import (
	"encoding/json"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{
		"error": message,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
