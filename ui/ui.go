package main

import (
	"fmt"
	"html/template"
	"log"
	"mariners/mpevent"
	"mariners/player"
	"mariners/role"
	"mariners/sms"
	"math/rand"
	"net"
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
	Players     player.Players
	Roles       role.Roles
	Events      mpevent.Events
	User        player.Player
	FocusPlayer player.Player
	FocusEvent  mpevent.Event
}

type MemberPage struct {
	Event mpevent.Event
	User  player.Player
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

func playereditHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Printf("updateplayerHandler: %s\n", err)
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
		log.Printf("playerviewHandler: %s\n", err)
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

func updateplayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("updateplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	strid := r.FormValue("id")
	log.Printf("updateplayerHandler: FormData: %#v\n", r.Form)
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Printf("updateplayerHandler: %s\n", err)
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

	for _, strid := range r.Form["role"] {
		rid, err := strconv.Atoi(strid)
		if err != nil {
			log.Printf("updateplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fr, err := role.GetRoleByID(int64(rid))
		if err != nil {
			log.Printf("updateplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		if fr == nil {
			log.Println("huh?")
		} else {
			p.Roles[int64(rid)] = fr[int64(rid)]
		}
	}

	log.Printf("AddPlayer form values: %#v", p)

	err = p.UpdatePlayer()
	if err != nil {
		log.Printf("updateplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("updateplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addplayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := player.Player{}

	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("addplayerHandler: %s\n", err)
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
			log.Printf("addplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		fr, err := role.GetRoleByID(int64(rid))
		if err != nil {
			log.Printf("addplayerHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		p.Roles[int64(rid)] = fr[int64(rid)]
	}

	log.Printf("AddPlayer form values: %#v", p)

	err = player.AddPlayer(&p)
	if err != nil {
		log.Printf("addplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("addplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func deleteplayerHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	p.ID = int64(id)
	err = p.DeletePlayer()
	if err != nil {
		log.Printf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("deleteplayerHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// Messages
func sendmessageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	a, err := checkPerms(user, "Communications")
	if err != nil {
		log.Printf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	if !a {
		log.Printf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusForbidden)
		return
	}

	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(int64(id))
	if err != nil {
		log.Printf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	msg := fmt.Sprintf("Message from %s: ", p.PreferredName)
	msg += r.FormValue("message")

	_, err = sms.SendTextTopic(msg, "arn:aws:sns:us-east-1:939932615330:MPLINKSTERS")
	if err != nil {
		log.Printf("sendmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func messageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user
	p.Players = pagedata.Players

	renderTemplate(w, "message", &p)
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

func eventaddHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}
	p.Title = title
	p.Roles = pagedata.Roles
	p.Players = pagedata.Players
	p.User = user

	renderTemplate(w, "eventadd", &p)
}

func addeventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	e := mpevent.Event{}

	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("addeventHandler: %s\n", err)
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
			log.Printf("addeventHandler: %s\n", err)
			errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		t, err := time.ParseInLocation("2006-01-02 03:04 PM", fd, loc)
		if err != nil {
			log.Printf("addeventHandler: %s\n", err)
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
			log.Printf("addeventHandler: %s\n", err)
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
		log.Printf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	e.Owner.GetPlayerByID(int64(id))

	err = e.CreateEvent()
	if err != nil {
		log.Printf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("addeventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func eventeditHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	strid := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Printf("eventeditHandler: %s\n", err)
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
		log.Printf("eventviewHandler: %s\n", err)
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

func eventupdateHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	strid := r.FormValue("id")
	id, err := strconv.ParseInt(strid, 10, 64)
	if err != nil {
		log.Printf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("eventupdateHandler: %s\n", err)
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
		log.Printf("editeventHandler: %s\n", err)
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
		log.Printf("editeventHandler: %s\n", err)
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
		log.Printf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("eventupdateHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func eventmessageHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	msg := r.FormValue("message")

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = e.SendEventMessage(msg, user.ID)
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addmemberHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		log.Printf("eventmessageHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	mid, err := strconv.Atoi(r.FormValue("newmember"))
	if err != nil {
		log.Printf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.AddMember(int64(mid), false)
	if err != nil {
		log.Printf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("addmemberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func eventjoinHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	strpid := mux.Vars(r)["pid"]
	mid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Printf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.AddMember(int64(mid), false)
	if err != nil {
		log.Printf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("eventjoinHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func removememberHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Printf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.DeleteMember(int64(pid))
	if err != nil {
		log.Printf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("removememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func memberpayHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.UpdateMember(int64(pid), true)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func memberunpayHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	strpid := mux.Vars(r)["pid"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	pid, err := strconv.Atoi(strpid)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.UpdateMember(int64(pid), false)
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("updatememberHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}

func deleventHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	strid := mux.Vars(r)["id"]
	id, err := strconv.Atoi(strid)
	if err != nil {
		log.Printf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	e := mpevent.Event{}
	err = e.GetEventByID(int64(id))
	if err != nil {
		log.Printf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	err = e.DeleteEvent()
	if err != nil {
		log.Printf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	err = cacheData()
	if err != nil {
		log.Printf("deleventHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/events", http.StatusFound)
}

//Main
func indexHandler(w http.ResponseWriter, r *http.Request, title string, user player.Player) {
	p := Page{}

	p.Title = title
	p.Roles = pagedata.Roles
	p.User = user

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
		log.Printf("logoutHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	p := player.Player{}
	err = p.GetPlayerByID(id)
	if err != nil {
		log.Printf("logoutHandler: %s\n", err)
		errorHandlerStatus(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("logoutHandler: Removing token for %s\n", p.Name)
	err = p.RemoveToken()
	if err != nil {
		log.Printf("logoutHandler: %s\n", err)
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
		"tmpl/profilemenu.html",
		"tmpl/error.html",
		"tmpl/deleteplayer.html",
		"tmpl/memberpay.html",
		"tmpl/memberunpay.html",
		"tmpl/memberdel.html",
		"tmpl/eventjoin.html",
		"tmpl/eventdel.html",
		"tmpl/"+tmpl+".html")
	if err != nil {
		log.Printf("renderTemplate: %s\n", err)
		return
	}
	err = t.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		log.Printf("renderTemplate: %s\n", err)
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

var validPath = regexp.MustCompile("^/(ui|players|playeredit|playerview|updateplayer|addplayer|deleteplayer|events|editevent|addevent|delevent|addmember|addmemberedit|removemember|updatemember|games|auth|sendcode|verify|maketoken|message|sendmessage)?")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, player.Player)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Println("makeHandler: invalid path")
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
			log.Println(u.String())
			http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
		}),
	}
	log.Println(httpSrv.ListenAndServe())
}

func cacheData() error {
	log.Printf("Refreshing data cache...")

	ps, err := player.GetPlayers()
	if err != nil {
		return err
	}
	pagedata.Players = ps

	rs, err := role.GetRoles()
	if err != nil {
		return err
	}
	pagedata.Roles = rs

	es, err := mpevent.GetEvents()
	if err != nil {
		return err
	}
	pagedata.Events = es

	log.Printf("Done.\n")
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
	listenport := getEnv("listenport", "8000")

	r := mux.NewRouter()

	r.HandleFunc("/", makeHandler(indexHandler))
	r.HandleFunc("/home", makeHandler(homeHandler))

	r.HandleFunc("/players", makeHandler(playerHandler))
	r.HandleFunc("/updateplayer", makeHandler(updateplayerHandler))
	r.HandleFunc("/playeradd", makeHandler(playeraddHandler))
	r.HandleFunc("/playeredit/{id}", makeHandler(playereditHandler))
	r.HandleFunc("/playerview/{id}", makeHandler(playerviewHandler))
	r.HandleFunc("/deleteplayer/{id}", makeHandler(deleteplayerHandler))
	r.HandleFunc("/addplayer", makeHandler(addplayerHandler))

	r.HandleFunc("/message", makeHandler(messageHandler))
	r.HandleFunc("/sendmessage/{id}", makeHandler(sendmessageHandler))

	r.HandleFunc("/events", makeHandler(eventHandler))
	r.HandleFunc("/addevent", makeHandler(addeventHandler))
	r.HandleFunc("/eventedit/{id}", makeHandler(eventeditHandler)).Methods("GET")
	r.HandleFunc("/eventmessage/{id}", makeHandler(eventmessageHandler))
	r.HandleFunc("/eventupdate", makeHandler(eventupdateHandler))
	r.HandleFunc("/eventview/{id}", makeHandler(eventviewHandler))
	r.HandleFunc("/eventadd", makeHandler(eventaddHandler))
	r.HandleFunc("/delevent/{id}", makeHandler(deleventHandler))
	r.HandleFunc("/addmember/{id}", makeHandler(addmemberHandler))
	r.HandleFunc("/memberpay/{id}/{pid}", makeHandler(memberpayHandler))
	r.HandleFunc("/memberunpay/{id}/{pid}", makeHandler(memberunpayHandler))
	r.HandleFunc("/eventjoin/{id}/{pid}", makeHandler(eventjoinHandler))
	r.HandleFunc("/removemember/{id}/{pid}", makeHandler(removememberHandler))

	r.HandleFunc("/auth", authHandler)
	r.HandleFunc("/sendcode", sendcodeHandler)
	r.HandleFunc("/verify", verifyHandler)
	r.HandleFunc("/maketoken", maketokenHandler)
	r.HandleFunc("/logout/{id}", logoutHandler)

	r.HandleFunc("/error", errorHandler)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.Handle("/", r)

	go redirectToHTTPS()

	err := cacheData()
	if err != nil {
		log.Panic(err)
	}
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenport),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
