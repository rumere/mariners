package main

// apiserver runs an http server and handles incoming requests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mariners/team"
	"net/http"
	"os"
	"strconv"
	"time"

	"mariners/game"
	"mariners/player"
	"mariners/weather"

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

func RespondWithWeather(w http.ResponseWriter, wt weather.Weather) error {
	j, err := json.MarshalIndent(wt, "", "  ")
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

func RespondWithTeam(w http.ResponseWriter, t team.Team) error {
	j, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
	fmt.Fprintf(w, "\n")

	return nil
}

func RespondWithGames(w http.ResponseWriter, g []game.Game) error {
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

func AddWeatherHandler(w http.ResponseWriter, r *http.Request) {
	wt := weather.Weather{}

	err := wt.AddWeather()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = RespondWithWeather(w, wt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetWeatherHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wt := weather.Weather{}
	err = wt.GetWeather(int64(id))
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithWeather(w, wt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetWeatherByDateHandler(w http.ResponseWriter, r *http.Request) {
	sdate := mux.Vars(r)["date"]

	wt := weather.Weather{}
	err := wt.GetWeatherByDate(sdate)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithWeather(w, wt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetGameByDateHandler(w http.ResponseWriter, r *http.Request) {
	sdate := mux.Vars(r)["date"]

	g := game.Game{}
	err := g.GetGameByDate(sdate)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithGame(w, g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AddGameHandler(w http.ResponseWriter, r *http.Request) {
	g := game.Game{}

	err := json.NewDecoder(r.Body).Decode(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = g.AddGame()
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

func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g := game.Game{}

	err = g.GetGame(int64(id))
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		err = RespondWithGame(w, g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AddTeamHandler(w http.ResponseWriter, r *http.Request) {
	strgid := mux.Vars(r)["id"]
	gid, err := strconv.Atoi(strgid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t := team.Team{}
	err = team.AddTeam(int64(gid), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = RespondWithTeam(w, t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	listenport := getEnv("listenport", "8080")

	r := mux.NewRouter()

	r.HandleFunc("/player", AddPlayerHandler).Methods("POST")
	r.HandleFunc("/player", GetPlayersHandler).Methods("GET")
	r.HandleFunc("/player/{id}", GetPlayerHandler).Methods("GET")
	r.HandleFunc("/player/{id}", UpdatePlayerHandler).Methods("PUT")
	r.HandleFunc("/player/{id}", DeletePlayerHandler).Methods("DELETE")

	r.HandleFunc("/game", AddGameHandler).Methods("POST")
	r.HandleFunc("/game/{id}", GetGameHandler).Methods("GET")
	r.HandleFunc("/game/bydate/{date}", GetGameByDateHandler).Methods("GET")

	r.HandleFunc("/weather", AddWeatherHandler).Methods("POST")
	r.HandleFunc("/weather/{id}", GetWeatherHandler).Methods("GET")
	r.HandleFunc("/weather/bydate/{date}", GetWeatherByDateHandler).Methods("GET")

	r.HandleFunc("/team/{gameid}", AddTeamHandler).Methods("POST")

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenport),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
