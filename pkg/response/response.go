package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error": "Failed to marshal JSON"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func Error(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]string{"error": message})
}

func ValidationErrors(w http.ResponseWriter, code int, errors map[string]string) {
	JSON(w, code, map[string]interface{}{"errors": errors})
}
