package main

import (
	"fmt"
	"net/http"
	"os"
	"rohitsingh/vile/server"
	"time"
)

func main() {
	// ToDo: Add custom host port specification
	host := string("localhost")
	port := 8080
	s := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      server.NewMux(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
