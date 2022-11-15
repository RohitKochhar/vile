package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// NewMux creates a mux.NewRouter and attaches handlers to it
func NewMux() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", rootHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/key/{key}", putHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/key/{key}", getHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/key/{key}", delHandler).Methods(http.MethodDelete)
	return r
}

// replyTextContent wraps text content in a HTTP response and sends it
func replyTextContent(w http.ResponseWriter, r *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content))
}

// replyError wraps text content in an HTTP error response and sends it
func replyError(w http.ResponseWriter, r *http.Request, status int, message string) {
	log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, message)
	http.Error(w, http.StatusText(status), status)
}
