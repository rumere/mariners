package main

// apiserver runs an http server and handles incoming requests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"mariners/game"
	"mariners/player"

	"github.com/gorilla/mux"
)

func RespondWithPlayer(w http.ResponseWriter, p player.Player) error {
	j, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
	fmt.Fprintf(w, "\n")

	return nil
}

func RespondWithPlayers(w http.ResponseWriter, p []player.Player) error {
	j, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
	fmt.Fprintf(w, "\n")

	return nil
}

func RespondWithGame(w http.ResponseWriter, g game.Game) error {
	j, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
	fmt.Fprintf(w, "\n")

	return nil
}

func AddPlayerHandler(w http.ResponseWriter, r *http.Request) {
	p := player.Player{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = player.AddPlayer(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = RespondWithPlayer(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func UpdatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	p := player.Player{}

	err = json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = player.UpdatePlayer(int64(id), &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = RespondWithPlayer(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func DeletePlayerHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = player.DeletePlayer(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Player %d deleted.\n", id)
}

func GetPlayerHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}

	err = player.GetPlayer(int64(id), &p)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithPlayer(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetPlayersHandler(w http.ResponseWriter, r *http.Request) {
	p, err := player.GetPlayers()
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithPlayers(w, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AddGameHandler(w http.ResponseWriter, r *http.Request) {
	g := game.Game{}

	err := game.AddGame(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = RespondWithGame(w, g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/player", AddPlayerHandler).Methods("POST")
	r.HandleFunc("/players", GetPlayersHandler).Methods("GET")
	r.HandleFunc("/player/{id}", GetPlayerHandler).Methods("GET")
	r.HandleFunc("/player/{id}", UpdatePlayerHandler).Methods("PUT")
	r.HandleFunc("/player/{id}", DeletePlayerHandler).Methods("DELETE")

	r.HandleFunc("/game", AddGameHandler).Methods("POST")
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
