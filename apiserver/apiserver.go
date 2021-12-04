package main

// apiserver runs an http server and handles incoming requests

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"src/github.com/rumere/mariners/player"

	"github.com/gorilla/mux"
)

func PlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["category"])
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/player", player.AddPlayer).Methods("PUT")
	r.HandleFunc("/player/{id}", PlayerHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
