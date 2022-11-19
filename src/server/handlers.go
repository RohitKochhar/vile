package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	// server needs core access to write and read from store
	// for incoming HTTP requests
	"rohitsingh/vile/core"

	// server needs transaction_logs access to record
	// HTTP request history in the transaction log
	"rohitsingh/vile/transaction_logs"

	"github.com/gorilla/mux"
)

var transact transaction_logs.TransactionLogger

func NewMux() http.Handler {
	r := mux.NewRouter()
	// Root path can be used as a liveness check
	r.HandleFunc("/", rootHandler).Methods(http.MethodGet)
	// Long-form path requests
	r.HandleFunc("/v1/key/{key}", putHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/key/{key}", getHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/key/{key}", delHandler).Methods(http.MethodDelete)
	// Short-form path requests
	r.HandleFunc("/{key}", putHandler).Methods(http.MethodPut)
	r.HandleFunc("/{key}", getHandler).Methods(http.MethodGet)
	r.HandleFunc("/{key}", delHandler).Methods(http.MethodDelete)
	return r
}

// StartServer creates a mux.NewRouter and attaches handlers to it
func StartServer() {
	// Initialize the logger
	txFilepath := "./transaction.log"
	var err error
	transact, err = transaction_logs.InitializeTransactionLog(txFilepath)
	if err != nil {
		panic(err)
	}
	log.Printf("Using transaction log located at %s", txFilepath)
	// Initialize the server
	port := 8080
	r := NewMux()
	log.Printf("Ready to accept connections on vile server at https://localhost:%d\n\n", port)
	err = http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "./keys/localhost.crt", "./keys/localhost.key", r)
	if err != nil {
		fmt.Printf("error while listening and serving: %q", err)
	}
}

// replyTextContent wraps text content in a HTTP response and sends it
func replyTextContent(w http.ResponseWriter, r *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	w.Write([]byte(content + "\n"))
}

// replyError wraps text content in an HTTP error response and sends it
func replyError(w http.ResponseWriter, r *http.Request, status int, message string) {
	log.Printf("%s %s: Error: %d %s", r.URL, r.Method, status, message)
	http.Error(w, http.StatusText(status), status)
}

// rootHandler handles requests sent to the root (duh)
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	content := "Check now hey! This is vile, man!"
	replyTextContent(w, r, http.StatusOK, content)
}

// putHandler expects to be called with a PUT request
// for the "/v1/key/{}" resource
func putHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received PUT request")
	// Get the variables from the path
	key := mux.Vars(r)["key"]
	// The request body has our value
	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		replyError(w, r, http.StatusInternalServerError, "Could not ready request body")
		return
	}
	if err = core.Put(key, string(value)); err != nil {
		replyError(w, r, http.StatusInternalServerError, "Could not store value in vile")
		return
	}
	// Record the event with the transaction logger
	transact.WritePut(key, string(value))
	msg := fmt.Sprintf("Successfully stored %s:%s", key, value)
	replyTextContent(w, r, http.StatusCreated, msg)
}

// getHandler returns the value stored at the key localted at /v1/key/{}
func getHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET request")
	key := mux.Vars(r)["key"] // The name of the key we are getting the value of
	value, err := core.Get(key)
	if errors.Is(err, core.ErrNoSuchKey) {
		msg := fmt.Sprintf("Could not find %s", key)
		replyError(w, r, http.StatusNotFound, msg)
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Error while getting %s", key)
		replyError(w, r, http.StatusInternalServerError, msg)
		return
	}
	replyTextContent(w, r, http.StatusOK, value)
}

// delHandler removes the value of the key provided in the path
func delHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received DELETE request")
	key := mux.Vars(r)["key"]
	if err := core.Delete(key); err != nil {
		if errors.Is(err, core.ErrNoSuchKey) {
			replyError(w, r, http.StatusNotFound, "The requested key could not be found")
			return
		}
		replyError(w, r, http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	// Record the event with the transaction logger
	transact.WriteDelete(key)
	msg := fmt.Sprintf("Successfully deleted entry %s", key)
	replyTextContent(w, r, http.StatusOK, msg)
}
