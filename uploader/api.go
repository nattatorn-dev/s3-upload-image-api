package uploader

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

//	api error wrapper
func Error(w http.ResponseWriter, error string, code int) {
	res := Response{
		Success: false,
		Error:   error,
	}

	//	prepare json response
	js, err := json.Marshal(res)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"success":false,"error":"failed to marshal json"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

//	api success wrapper
func Success(w http.ResponseWriter, data interface{}) {
	res := Response{
		Success: true,
		Data:    data,
	}

	//	prepare json response
	js, err := json.Marshal(res)
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(js)
}
