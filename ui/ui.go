package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"mariners/game"
	"mariners/player"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Page struct {
	Title   string
	Body    []byte
	Players player.Players
	Games   game.Game
}

func loadPage(title string) (*Page, error) {
	filename := "tmpl/" + title + ".html"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func playerHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	p.Title = title
	players, err := player.GetPlayers()
	p.Players = players
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		renderTemplate(w, "players", p)
	}
}

func gamesHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	p.Title = title

	renderTemplate(w, "games", p)
}

func updateplayerHandler(w http.ResponseWriter, r *http.Request, title string) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	p.ID = int64(id)
	p.Name = r.FormValue("name")
	p.PreferredName = r.FormValue("preferred-name")
	p.Phone = r.FormValue("phone")
	p.Email = r.FormValue("email")
	p.GhinNumber = r.FormValue("ghin-number")

	err = player.UpdatePlayer(p.ID, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/players", http.StatusFound)
}

func addplayerHandler(w http.ResponseWriter, r *http.Request, title string) {
	p := player.Player{}

	p.Name = r.FormValue("name")
	p.PreferredName = r.FormValue("preferred-name")
	p.Phone = r.FormValue("phone")
	p.Email = r.FormValue("email")
	p.GhinNumber = r.FormValue("ghin-number")

	err := player.AddPlayer(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/players", http.StatusFound)
}

func deleteplayerHandler(w http.ResponseWriter, r *http.Request, title string) {
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

	http.Redirect(w, r, "/players", http.StatusFound)
}

func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
	body, err := os.ReadFile("tmpl/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p := &Page{Title: title, Body: body}
	renderTemplate(w, "index", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles("tmpl/" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(ui|players|updateplayer|addplayer|deleteplayer|games)?")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fmt.Printf("%s\n", m[1])
		fn(w, r, m[1])
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	listenport := getEnv("listenport", "8000")

	r := mux.NewRouter()

	r.HandleFunc("/", makeHandler(indexHandler))
	r.HandleFunc("/players", makeHandler(playerHandler))
	r.HandleFunc("/updateplayer/{id}", makeHandler(updateplayerHandler))
	r.HandleFunc("/deleteplayer/{id}", makeHandler(deleteplayerHandler))
	r.HandleFunc("/addplayer", makeHandler(addplayerHandler))
	r.HandleFunc("/games", makeHandler(gamesHandler))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenport),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
