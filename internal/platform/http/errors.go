package http

import stdhttp "net/http"

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w stdhttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
