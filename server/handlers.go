package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"rohitsingh/vile/core"

	"github.com/gorilla/mux"
)

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
	msg := fmt.Sprintf("Successfully stored %s:%s", key, value)
	replyTextContent(w, r, http.StatusCreated, msg)
}

// getHandler returns the value stored at the key localted at /v1/key/{}
func getHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"] // The name of the key we are getting the value of
	value, err := core.Get(key)
	if errors.Is(err, core.ErrNoSuchKey) {
		msg := fmt.Sprintf("Could not find key=%s", key)
		replyError(w, r, http.StatusNotFound, msg)
		return
	}
	if err != nil {
		msg := fmt.Sprintf("Error while getting key=%s", key)
		replyError(w, r, http.StatusInternalServerError, msg)
		return
	}
	replyTextContent(w, r, http.StatusOK, value)
}

// delHandler removes the value of the key provided in the path
func delHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	if err := core.Delete(key); err != nil {
		if errors.Is(err, core.ErrNoSuchKey) {
			replyError(w, r, http.StatusNotFound, "The requested key could not be found")
			return
		}
		replyError(w, r, http.StatusInternalServerError, "Something went wrong :(")
		return
	}
	msg := fmt.Sprintf("Successfully deleted entry %s", key)
	replyTextContent(w, r, http.StatusOK, msg)
}
