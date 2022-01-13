package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"mariners/game"
	"mariners/player"
	"mariners/sms"
	"mariners/weather"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nyaruka/phonenumbers"
)

type Page struct {
	Title       string
	Body        []byte
	Players     player.Players
	Player      player.Player
	Games       game.Game
	Weather     weather.Weather
	IsGameToday bool
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

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	strdate := time.Now().In(loc).Format("2006-01-02")
	err = p.Games.GetGameByDate(strdate)
	switch {
	case err == sql.ErrNoRows:
		p.IsGameToday = false
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	default:
		p.IsGameToday = true
		p.Weather.ID = p.Games.WeatherID
		p.Weather.GetWeather()
		fmt.Printf("Weather: %#v\n", p.Weather)
	}

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

func addgameHandler(w http.ResponseWriter, r *http.Request, title string) {
	g := game.Game{}

	g.NinthTee = r.FormValue("ninth-tee")
	if r.FormValue("ismatch") != "" {
		im, err := strconv.ParseBool(r.FormValue("ismatch"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.IsMatch = im
	} else {
		g.IsMatch = false
	}
	err := g.AddGame()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/games", http.StatusFound)
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

func authHandler(w http.ResponseWriter, r *http.Request) {
	body, err := os.ReadFile("tmpl/auth.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p := &Page{Body: body}
	players, err := player.GetPlayers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	p.Players = players
	renderTemplate(w, "auth", p)
}

func sendcodeHandler(w http.ResponseWriter, r *http.Request) {
	strid := r.FormValue("player")
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	num, err := phonenumbers.Parse(p.Phone, "US")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	phone := phonenumbers.Format(num, phonenumbers.E164)

	var code string
	msg := "Your MPLINKSTERS Login Code is: "
	for i := 0; i < 5; i++ {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		code += strconv.Itoa(r1.Intn(9))
	}

	p.WriteToken(code)

	sms.SendTextPhone(msg+code, phone)

	http.Redirect(w, r, "/verify", http.StatusFound)
}

func sendmessageHandler(w http.ResponseWriter, r *http.Request, title string) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("Message from %s: ", p.PreferredName)
	msg += r.FormValue("message")

	sms.SendTextTopic(msg, "arn:aws:sns:us-east-1:939932615330:MPLINKSTERS")

	http.Redirect(w, r, "/", http.StatusFound)
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := os.ReadFile("tmpl/verify.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p := &Page{Body: body}
	players, err := player.GetPlayers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	p.Players = players
	renderTemplate(w, "verify", p)
}

func messageHandler(w http.ResponseWriter, r *http.Request, title string) {
	body, err := os.ReadFile("tmpl/message.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := r.Cookie("token")
	if err != nil || token == nil {
		fmt.Printf("%s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := &Page{Title: title, Body: body}
	user := player.Player{}
	err = user.GetPlayerByToken(token.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Player = user
	renderTemplate(w, "message", p)
}

func maketokenHandler(w http.ResponseWriter, r *http.Request) {
	p := player.Player{}
	err := p.GetPlayerByToken(r.FormValue("code"))
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		http.Redirect(w, r, "/auth", http.StatusFound)
	}

	token := uuid.New().String()

	err = p.WriteToken(token)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().AddDate(0, 1, 0),
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles("tmpl/menu.html", "tmpl/"+tmpl+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(ui|players|updateplayer|addplayer|deleteplayer|games|auth|sendcode|verify|maketoken|message|sendmessage)?")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		fmt.Printf("%s\n", m[1])
		if m == nil {
			http.NotFound(w, r)
			return
		}

		token, err := r.Cookie("token")
		if err != nil || token == nil {
			fmt.Printf("%s\n", err.Error())
			fmt.Printf("%s\n", m[1])
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		user := player.Player{}
		err = user.GetPlayerByToken(token.Value)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			http.Redirect(w, r, "/auth", http.StatusFound)
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
	r.HandleFunc("/addgame", makeHandler(addgameHandler))
	r.HandleFunc("/message", makeHandler(messageHandler))
	r.HandleFunc("/sendmessage/{id}", makeHandler(sendmessageHandler))
	r.HandleFunc("/auth", authHandler)
	r.HandleFunc("/sendcode", sendcodeHandler)
	r.HandleFunc("/verify", verifyHandler)
	r.HandleFunc("/maketoken", maketokenHandler)
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
