package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"mariners/db"
	"mariners/game"
	"mariners/mpevent"
	"mariners/player"
	"mariners/role"
	"mariners/scoring"
	"mariners/sms"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nyaruka/phonenumbers"
)

type Page struct {
	Title       string
	Players     player.Players
	Roles       role.Roles
	Events      mpevent.Events
	Scores      scoring.MPAverages
	User        player.Player
	FocusPlayer player.Player
	FocusEvent  mpevent.Event
	Game        game.Game
}

type MemberPage struct {
	Event mpevent.Event
	User  player.Player
}

func (p *Page) MessageTime() bool {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return false
	}
	t := time.Now().In(loc)
	if t.Hour() <= 8 && t.Hour() >= 20 {
		return false
	}
	return true
}

var pagedata Page

// Players
func playerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Players = pagedata.Players

	renderTemplate(w, "players", &p)
}

func playerinfoHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "playerinfo", &p)
}

func playereditHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("updateplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.FocusPlayer.GetPlayerByID(id)
	p.Players = pagedata.Players

	renderTemplate(w, "playeredit", &p)
}

func playerviewHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("playerviewHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.FocusPlayer.GetPlayerByID(id)
	p.Players = pagedata.Players

	renderTemplate(w, "playerview", &p)
}

func playeraddHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "playeradd", &p)
}

func putPlayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("putPlayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	strid := r.FormValue("id")
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("putPlayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	p.ID = int64(id)
	p.Name = r.FormValue("name")
	p.PreferredName = r.FormValue("preferred-name")
	p.Phone = r.FormValue("phone")
	p.Email = r.FormValue("email")
	p.GhinNumber = r.FormValue("ghin-number")
	p.Roles = make(role.Roles)
	p.TextPreference = r.FormValue("text-size")

	for _, strid := range r.Form["role"] {
		rid, err := strconv.Atoi(strid)
		if err != nil {
			log.Error().Msgf("putPlayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fr, err := role.GetRoleByID(int64(rid))
		if err != nil {
			log.Error().Msgf("putPlayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		if fr == nil {
			log.Printf("huh?")
		} else {
			p.Roles[int64(rid)] = fr[int64(rid)]
		}
	}

	err = p.UpdatePlayer()
	if err != nil {
		log.Error().Msgf("putPlayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("putPlayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func postPlayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := player.Player{}

	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("addplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p.Name = r.FormValue("name")
	p.PreferredName = r.FormValue("preferred-name")
	p.Phone = r.FormValue("phone")
	p.Email = r.FormValue("email")
	p.GhinNumber = r.FormValue("ghin-number")

	p.Roles = make(role.Roles)
	for _, strid := range r.Form["role"] {
		rid, err := strconv.Atoi(strid)
		if err != nil {
			log.Error().Msgf("addplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		fr, err := role.GetRoleByID(int64(rid))
		if err != nil {
			log.Error().Msgf("addplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		p.Roles[int64(rid)] = fr[int64(rid)]
	}

	err = player.AddPlayer(&p)
	if err != nil {
		log.Error().Msgf("addplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("addplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func delPlayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	p.ID = int64(id)
	err = p.DeletePlayer()
	if err != nil {
		log.Error().Msgf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

// Messages
func postMessageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	a, err := checkPerms(user, "Communications")
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusForbidden)
		return
	}
	if !a {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusForbidden)
		return
	}

	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("Message from %s: ", p.PreferredName)
	msg += r.FormValue("message")

	for _, p := range pagedata.Players {
		if p.HasRole("User") {
			num, err := phonenumbers.Parse(p.Phone, "US")
			if err != nil {
				log.Error().Msgf("sendmessageHandler: %s\n", err)
				continue
			}
			phone := phonenumbers.Format(num, phonenumbers.E164)
			sms.SendTextPhone(msg, phone)
			time.Sleep(time.Second)
		}
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func postTournamentMessageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	a, err := checkPerms(user, "Communications")
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusForbidden)
		return
	}
	if !a {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusForbidden)
		return
	}

	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		log.Error().Msgf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("Message from %s: ", p.PreferredName)
	msg += r.FormValue("message")

	for _, p := range pagedata.Players {
		if p.HasRole("User") && p.HasRole("Tournament") {
			num, err := phonenumbers.Parse(p.Phone, "US")
			if err != nil {
				log.Error().Msgf("sendmessageHandler: %s\n", err)
				continue
			}
			phone := phonenumbers.Format(num, phonenumbers.E164)
			sms.SendTextPhone(msg, phone)
			time.Sleep(time.Second)
		}
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func messageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Players = pagedata.Players

	renderTemplate(w, "message", &p)
}

func messageinfoHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "messageinfo", &p)
}

// Events
func eventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Players = pagedata.Players
	p.Events = pagedata.Events

	renderTemplate(w, "events", &p)
}

func eventinfoHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "eventinfo", &p)
}

func eventaddHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.Players = pagedata.Players
	p.User = user

	renderTemplate(w, "eventadd", &p)
}

func eventeditHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("eventeditHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	p.FocusEvent.GetEventByID(id)

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.FocusPlayer.GetPlayerByID(id)
	p.Players = pagedata.Players
	p.Events = pagedata.Events

	renderTemplate(w, "eventedit", &p)
}

func eventviewHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("eventviewHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	p.FocusEvent.GetEventByID(id)

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Players = pagedata.Players
	p.Events = pagedata.Events

	renderTemplate(w, "eventview", &p)
}

func postEventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	e := mpevent.Event{}

	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e.Name = r.FormValue("name")
	fd := r.FormValue("date")
	if fd == "" {
		e.Date = "0001-01-01T00:00"
	} else {
		loc, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			log.Error().Msgf("addeventHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		t, err := time.ParseInLocation("2006-01-02T15:04", fd, loc)
		if err != nil {
			log.Error().Msgf("addeventHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		e.Date = t.Format("2006-01-01T15:04")
	}
	e.Description = r.FormValue("desc")
	strid := r.FormValue("owner")
	if _, ok := r.Form["paid"]; ok {
		e.PaidEvent = true
		c, err := strconv.ParseFloat(r.Form["cost"][0], 32)
		if err != nil {
			log.Error().Msgf("addeventHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		e.Cost = c
	} else {
		e.PaidEvent = false
		e.Cost = 0
	}
	if _, ok := r.Form["invite"]; ok {
		e.InviteOnly = true
	} else {
		e.InviteOnly = false
	}
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	e.Owner.GetPlayerByID(int64(id))

	err = e.CreateEvent()
	if err != nil {
		log.Error().Msgf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func putEventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	strid := r.FormValue("id")
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Error().Msgf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	//2022-02-19T11:00 AM
	e.Name = r.FormValue("name")
	fd := r.FormValue("date")
	if fd == "" {
		e.Date = "0001-01-01T00:00"
	} else {
		e.Date = fd
	}
	e.Description = r.FormValue("description")
	stroid := r.FormValue("owner")
	oid, err := strconv.Atoi(stroid)
	if err != nil {
		log.Error().Msgf("editeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	e.Owner.GetPlayerByID(int64(oid))
	if _, ok := r.Form["paidevent"]; ok {
		e.PaidEvent = true
	} else {
		e.PaidEvent = false
	}
	c, err := strconv.ParseFloat(r.Form["cost"][0], 32)
	if err != nil {
		log.Error().Msgf("editeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	e.Cost = c
	if _, ok := r.Form["inviteonly"]; ok {
		e.InviteOnly = true
	} else {
		e.InviteOnly = false
	}

	err = e.UpdateEvent()
	if err != nil {
		log.Error().Msgf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func delEventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.DeleteEvent()
	if err != nil {
		log.Error().Msgf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func postEventMessageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	msg := r.FormValue("message")

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = e.SendEventMessage(msg, user.ID)
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func postMemberHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Error().Msgf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	mid, err := strconv.Atoi(r.FormValue("newmember"))
	if err != nil {
		log.Error().Msgf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.AddMember(int64(mid), false)
	if err != nil {
		log.Error().Msgf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func putMemberJoinHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	strpid := mux.Vars(r)["pid"]
	mid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Error().Msgf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.AddMember(int64(mid), false)
	if err != nil {
		log.Error().Msgf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func delMemberHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Error().Msgf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.DeleteMember(int64(pid))
	if err != nil {
		log.Error().Msgf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func putMemberPayHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.UpdateMember(int64(pid), true)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func putMemberUnpayHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.UpdateMember(int64(pid), false)
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Error().Msgf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

// Game
func gameinfoHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Game = pagedata.Game
	p.Players = pagedata.Players

	renderTemplate(w, "gameinfo", &p)
}

// Scoring
func scoresHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	ss, err := scoring.GetAverages()
	if err != nil {
		log.Error().Msgf("scoresHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	p.Title = title
	p.User = user
	p.Roles = pagedata.Roles
	p.Players = pagedata.Players
	p.Events = pagedata.Events
	p.Scores = ss

	renderTemplate(w, "scores", &p)
}

func scoresinfoHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "scoresinfo", &p)
}

// Game

func gameHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p = pagedata

	p.User = user
	p.Title = title

	renderTemplate(w, "game", &p)
}

func gamechangeHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p = pagedata

	p.User = user
	p.Title = title

	renderTemplate(w, "game", &p)
}

func gamecheckinHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p = pagedata

	p.User = user
	p.Title = title

	renderTemplate(w, "game", &p)
}

// Main
func indexHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	if title == "" {
		p.Title = "home"
	} else {
		p.Title = title
	}
	p.User = user
	p.Roles = pagedata.Roles
	p.Players = pagedata.Players
	p.Events = pagedata.Events

	renderTemplate(w, "index", &p)
}

func homeHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

	renderTemplate(w, "home", &p)
}

// Auth
func authHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}

	p.Players = pagedata.Players

	renderTemplate(w, "auth", &p)
}

func sendcodeHandler(w http.ResponseWriter, r *http.Request) {
	strid := r.FormValue("player")

	if strid == "Select Your Name" {
		http.Redirect(w, r, "/auth", http.StatusFound)
	}
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("sendcodeHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		log.Printf("sendcodeHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	num, err := phonenumbers.Parse(p.Phone, "US")
	if err != nil {
		log.Printf("sendcodeHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
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

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}

	p.Players = pagedata.Players

	renderTemplate(w, "verify", &p)
}

func maketokenHandler(w http.ResponseWriter, r *http.Request) {
	p := player.Player{}
	err := p.GetPlayerByToken(r.FormValue("code"))
	if err != nil {
		log.Printf("maketokenHandler: %s\n", err)
		http.Redirect(w, r, "/auth", http.StatusFound)
	}

	token := uuid.New().String()

	err = p.WriteToken(token)
	if err != nil {
		log.Printf("maketokenHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(id)
	if err != nil {
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = p.RemoveToken()
	if err != nil {
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: 0,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/auth", http.StatusFound)
}

// Utils
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles(
		"tmpl/header.html",
		"tmpl/menu.html",
		"tmpl/home.html",
		"tmpl/wait.html",
		"tmpl/profilemenu.html",
		"tmpl/error.html",
		"tmpl/playerdel.html",
		"tmpl/memberpay.html",
		"tmpl/memberunpay.html",
		"tmpl/memberdel.html",
		"tmpl/memberjoin.html",
		"tmpl/eventdel.html",
		"tmpl/"+tmpl+".html")
	if err != nil {
		log.Error().Msgf("renderTemplate: %s\n", err)
		return
	}
	err = t.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		log.Error().Msgf("renderTemplate: Execute: %s\n", err)
		return
	}
}

func checkPerms(p player.Player, n string) (bool, error) {
	for _, r := range p.Roles {
		if r == n {
			return true, nil
		}
	}

	return false, nil
}

var validPath = regexp.MustCompile("^/(ui|players|playeredit|playerview|updateplayer|addplayer|deleteplayer|events|editevent|addevent|delevent|addmember|addmemberedit|removemember|updatemember|games|auth|sendcode|verify|maketoken|message|sendmessage|addalluser|scores|scoresinfo|checkin|checkins)?")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, player.Player)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Error().Msgf("makeHandler: invalid path: %s", m[1])
			http.NotFound(w, r)
			return
		}

		token, err := r.Cookie("token")
		if err != nil || token == nil {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		user := player.Player{}
		err = user.GetPlayerByToken(token.Value)

		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		fn(w, r, m[1], user)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func redirectToHTTPS() {
	httpSrv := http.Server{
		Addr: ":8001",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := r.URL
			u.Host = net.JoinHostPort("www.mplinksters.club", "443")
			u.Scheme = "https"
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		}),
	}
	log.Fatal().Msgf("%s", httpSrv.ListenAndServe())
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	err := cacheData()
	if err != nil {
		log.Error().Msgf("cacheHandler: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	r.Body.Close()
}

func addAllUserHandler(w http.ResponseWriter, r *http.Request) {
	err := player.AddRoleAll(1)
	if err != nil {
		log.Error().Msgf("addAllUserHandler: %s", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func cacheData() error {
	log.Info().Msg("Refreshing data cache...")

	log.Info().Msg("Players...")

	ps, err := player.GetPlayers()
	if err != nil {
		return err
	}
	pagedata.Players = ps

	log.Info().Msg("Roles...")

	rs, err := role.GetRoles()
	if err != nil {
		return err
	}
	pagedata.Roles = rs

	log.Info().Msg("Events...")

	es, err := mpevent.GetEvents()
	if err != nil {
		return err
	}
	pagedata.Events = es

	log.Info().Msg("Game...")

	n := time.Now()
	g, err := game.GetGameByDate(n)
	if err != nil {
		log.Info().Err(err)
		if err == sql.ErrNoRows {
			log.Info().Msg("Creating game...")
			err = g.Tee.GetTeeByName("White")
			if err != nil {
				return err
			}
			err = g.AddGame()
			if err != nil {
				return err
			}
		} else {
			log.Error().Err(err)
			return err
		}
	}
	pagedata.Game = g

	log.Info().Msg("Done.")
	return nil
}

func errorHandlerStatus(w http.ResponseWriter, r *http.Request, e string, status int) {
	p := Page{}

	p.Title = e

	renderTemplate(w, "error", &p)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{}

	renderTemplate(w, "error", &p)
}

func main() {
	var err error
	db.Con, err = db.DBConnection()
	if err != nil {
		log.Fatal().Msgf("Could not connect to database:  %s", err)
	}
	defer db.Con.Close()

	listenport := getEnv("listenport", "8000")

	r := mux.NewRouter()

	r.HandleFunc("/", makeHandler(indexHandler))

	sr := r.PathPrefix("/render").Subrouter()
	fr := r.PathPrefix("/form").Subrouter()

	sr.HandleFunc("/home", makeHandler(homeHandler))

	sr.HandleFunc("/players", makeHandler(playerHandler))
	sr.HandleFunc("/playeradd", makeHandler(playeraddHandler))
	sr.HandleFunc("/playeredit/{id}", makeHandler(playereditHandler))
	sr.HandleFunc("/playerview/{id}", makeHandler(playerviewHandler))
	sr.HandleFunc("/playerinfo", makeHandler(playerinfoHandler))

	fr.HandleFunc("/postplayer", makeHandler(postPlayerHandler)).Methods("POST")
	fr.HandleFunc("/putplayer", makeHandler(putPlayerHandler)).Methods("PUT")
	fr.HandleFunc("/delplayer/{id}", makeHandler(delPlayerHandler)).Methods("DELETE")

	sr.HandleFunc("/message", makeHandler(messageHandler))
	sr.HandleFunc("/messageinfo", makeHandler(messageinfoHandler))

	fr.HandleFunc("/postmessage/{id}", makeHandler(postMessageHandler)).Methods("POST")
	fr.HandleFunc("/posttournymessage/{id}", makeHandler(postTournamentMessageHandler)).Methods("POST")

	sr.HandleFunc("/events", makeHandler(eventHandler))
	sr.HandleFunc("/eventadd", makeHandler(eventaddHandler))
	sr.HandleFunc("/eventedit/{id}", makeHandler(eventeditHandler))
	sr.HandleFunc("/eventview/{id}", makeHandler(eventviewHandler))
	sr.HandleFunc("/eventinfo", makeHandler(eventinfoHandler))

	fr.HandleFunc("/postevent", makeHandler(postEventHandler)).Methods("POST")
	fr.HandleFunc("/putevent/{id}", makeHandler(putEventHandler)).Methods("PUT")
	fr.HandleFunc("/delevent/{id}", makeHandler(delEventHandler)).Methods("DELETE")

	fr.HandleFunc("/postmember/{id}", makeHandler(postMemberHandler)).Methods("POST")
	fr.HandleFunc("/postmemberjoin/{id}/{pid}", makeHandler(putMemberJoinHandler)).Methods("POST")
	fr.HandleFunc("/putmemberpay/{id}/{pid}", makeHandler(putMemberPayHandler)).Methods("PUT")
	fr.HandleFunc("/putmemberunpay/{id}/{pid}", makeHandler(putMemberUnpayHandler)).Methods("PUT")
	fr.HandleFunc("/delmember/{id}/{pid}", makeHandler(delMemberHandler)).Methods("DELETE")

	sr.HandleFunc("/game", makeHandler(gameHandler))
	sr.HandleFunc("/gamechange", makeHandler(gamechangeHandler))
	fr.HandleFunc("/gameCheckin", makeHandler(gamecheckinHandler))
	sr.HandleFunc("/gameinfo", makeHandler(gameinfoHandler))

	fr.HandleFunc("/posteventmessage/{id}", makeHandler(postEventMessageHandler)).Methods("POST")

	sr.HandleFunc("/scores", makeHandler(scoresHandler))
	sr.HandleFunc("/scoresinfo", makeHandler(scoresinfoHandler))

	r.HandleFunc("/auth", authHandler)
	r.HandleFunc("/sendcode", sendcodeHandler)
	r.HandleFunc("/verify", verifyHandler)
	r.HandleFunc("/maketoken", maketokenHandler)
	r.HandleFunc("/logout/{id}", logoutHandler)

	r.HandleFunc("/error", errorHandler)
	r.HandleFunc("/cache", cacheHandler)
	r.HandleFunc("/addalluser", addAllUserHandler)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", r)

	go redirectToHTTPS()

	err = cacheData()
	if err != nil {
		log.Panic().Err(err)
	}

	log.Info().Msg("Starting HTTP Server...")

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenport),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal().Err(srv.ListenAndServe())
}
