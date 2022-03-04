package mpevent

import (
	"context"
	"fmt"
	"log"
	"mariners/db"
	"mariners/player"
	"mariners/sms"
	"regexp"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
)

type Event struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Date        string  `json:"date"`
	PaidEvent   bool    `json:"paid_event"`
	Cost        float64 `json:"cost"`
	TopicArn    string  `json:"topic_arn"`
	Description string  `json:"description"`
	Owner       player.Player
	InviteOnly  bool `json:"invite_only"`
	Members     EventMembers
	Messages    EventMessages
}

type EventMember struct {
	Player          player.Player
	Paid            bool   `json:"paid"`
	SubscriptionArn string `json:"subscription_arn"`
}

type EventMessage struct {
	Player    player.Player
	Message   string
	MessageID string
	Date      string
}

type Events []Event

type EventMembers []EventMember

type EventMessages []EventMessage

func (e *Event) CreateEvent() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	topicName := strings.Replace(e.Name, " ", "-", -1)

	topicARN, err := sms.CreateTopic(topicName)
	if err != nil {
		return err
	}
	e.TopicArn = topicARN

	query := fmt.Sprintf("INSERT INTO event (name, date, paid_event, topic_arn, description, ownerid, invite_only, cost) VALUES (\"%s\", \"%s\", %t, \"%s\", \"%s\", %d, %t, %f)",
		e.Name,
		e.Date,
		e.PaidEvent,
		e.TopicArn,
		e.Description,
		e.Owner.ID,
		e.InviteOnly,
		e.Cost)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Println("error is from createevent query")
		return err
	}

	e.ID, err = res.LastInsertId()
	if err != nil {
		log.Println("error is from lastinsertid")
		return err
	}

	err = e.AddMember(e.Owner.ID, true)
	if err != nil {
		log.Println("error is from addmember")
		return err
	}

	return nil
}

func (e *Event) UpdateEvent() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE event set date=\"%s\", paid_event=%t, description=\"%s\", ownerid=%d, invite_only=%t, cost=%f WHERE idevent=%d",
		e.Date,
		e.PaidEvent,
		e.Description,
		e.Owner.ID,
		e.InviteOnly,
		e.Cost,
		e.ID)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) AddMember(id int64, paid bool) error {
	p := player.Player{}
	err := p.GetPlayerByID(id)
	if err != nil {
		return err
	}

	num, err := phonenumbers.Parse(p.Phone, "US")
	if err != nil {
		return err
	}
	phone := phonenumbers.Format(num, phonenumbers.E164)

	subARN, err := sms.SubscribeUser(phone, e.TopicArn)
	if err != nil {
		return err
	}

	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO event_members (idevent, idplayer, paid, subscription_arn) VALUES (%d, %d, %t, \"%s\")",
		e.ID,
		id,
		paid,
		subARN)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// TODO: sms stuff should be handled by ui.go
	if (paid) || (e.Cost == 0) {
		msg := fmt.Sprintf("You have been added to the event \"%s\".", e.Name)
		_, err := sms.SendTextPhone(msg, phone)
		if err != nil {
			return err
		}
	} else {
		msg := fmt.Sprintf("You have been added to \"%s\".  The cost is $%.2f.  Please see %s to pay!", e.Name, e.Cost, e.Owner.PreferredName)
		_, err := sms.SendTextPhone(msg, phone)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Event) UpdateMember(id int64, paid bool) error {
	p := player.Player{}
	err := p.GetPlayerByID(id)
	if err != nil {
		return err
	}

	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE event_members set paid=%t WHERE idevent=%d and idplayer=%d",
		paid,
		e.ID,
		id)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	// TODO: sms stuff should be handled by ui.go
	num, err := phonenumbers.Parse(p.Phone, "US")
	if err != nil {
		return err
	}
	phone := phonenumbers.Format(num, phonenumbers.E164)

	if (paid) || (e.Cost == 0) {
		msg := fmt.Sprintf("You have been marked as paid for %s.", e.Name)
		_, err = sms.SendTextPhone(msg, phone)
		if err != nil {
			return err
		}
	} else {
		msg := fmt.Sprintf("You have been marked as NOT paid for %s.", e.Name)
		_, err = sms.SendTextPhone(msg, phone)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Event) DeleteMember(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, m := range e.Members {
		if m.Player.ID == id {
			err = sms.RemoveSubscriber(m.SubscriptionArn)
			if err != nil {
				return err
			}
			num, err := phonenumbers.Parse(m.Player.Phone, "US")
			if err != nil {
				return err
			}
			phone := phonenumbers.Format(num, phonenumbers.E164)
			msg := fmt.Sprintf("You have been removed from event %s.", e.Name)
			_, err = sms.SendTextPhone(msg, phone)
			if err != nil {
				return err
			}
			break
		}
	}
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM event_members WHERE idevent=%d and idplayer=%d",
		e.ID,
		id)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) SendEventMessage(msg string, sid int64) error {
	p := player.Player{}
	p.GetPlayerByID(sid)
	text := fmt.Sprintf("Message from %s: %s", p.PreferredName, msg)
	mid, err := sms.SendTextTopic(text, e.TopicArn)
	if err != nil {
		return err
	}

	m := EventMessage{}
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t := time.Now().In(loc)
	m.Date = t.Format("2006-01-02T15:04")
	m.Message = text
	m.Player = p
	m.MessageID = mid
	e.Messages = append(e.Messages, m)

	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO event_messages (idevent, idsender, message, date, idmessage) VALUES (%d, %d, \"%s\", \"%s\", \"%s\")",
		e.ID,
		m.Player.ID,
		m.Message,
		m.Date,
		m.MessageID)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) GetEventByID(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT idevent, name, date, paid_event, topic_arn, description, ownerid, invite_only, cost FROM event WHERE idevent=%d", id)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
		&e.ID,
		&e.Name,
		&e.Date,
		&e.PaidEvent,
		&e.TopicArn,
		&e.Description,
		&e.Owner.ID,
		&e.InviteOnly,
		&e.Cost)
	if err != nil {
		return err
	}

	t := time.Time{}
	awsdate, err := regexp.Match(`T`, []byte(e.Date))
	if err != nil {
		return err
	}
	if awsdate {
		t, err = time.Parse("2006-01-02T15:04:00Z", e.Date)
		if err != nil {
			return err
		}
	} else {
		t, err = time.Parse("2006-01-02 15:04:00", e.Date)
		if err != nil {
			return err
		}
	}
	e.Date = t.Format("2006-01-02T15:04")

	e.Owner.GetPlayerByID(e.Owner.ID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("SELECT idplayer, paid, subscription_arn FROM event_members WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m EventMember
		if err := rows.Scan(&m.Player.ID, &m.Paid, &m.SubscriptionArn); err != nil {
			return err
		}
		m.Player.GetPlayerByID(m.Player.ID)
		e.Members = append(e.Members, m)
	}

	query = fmt.Sprintf("SELECT idsender, message, date FROM event_messages WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err = db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m EventMessage
		if err := rows.Scan(&m.Player.ID, &m.Message, &m.Date); err != nil {
			return err
		}
		m.Player.GetPlayerByID(m.Player.ID)
		e.Messages = append(e.Messages, m)
	}

	return nil
}

func (e *Event) GetEventByName(name string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT idevent, name, date, paid_event, topic_arn, description, ownerid, invite_only, cost FROM event WHERE name=%s", name)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
		&e.ID,
		&e.Name,
		&e.Date,
		&e.PaidEvent,
		&e.TopicArn,
		&e.Description,
		&e.Owner.ID,
		&e.InviteOnly,
		&e.Cost)
	if err != nil {
		return err
	}

	t := time.Time{}
	awsdate, err := regexp.Match(`T`, []byte(e.Date))
	if err != nil {
		return err
	}
	if awsdate {
		t, err = time.Parse("2006-01-02T15:04:00Z", e.Date)
		if err != nil {
			return err
		}
	} else {
		t, err = time.Parse("2006-01-02 15:04:00", e.Date)
		if err != nil {
			return err
		}
	}
	e.Date = t.Format("2006-01-02T15:04")

	e.Owner.GetPlayerByID(e.Owner.ID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("SELECT idplayer, paid, subscription_arn FROM event_members WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	for rows.Next() {
		var m EventMember
		if err := rows.Scan(&m.Player.ID, &m.Paid, &m.SubscriptionArn); err != nil {
			return err
		}
		m.Player.GetPlayerByID(m.Player.ID)
		e.Members = append(e.Members, m)
	}

	query = fmt.Sprintf("SELECT idsender, message, date FROM event_messages WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err = db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	for rows.Next() {
		var m EventMessage
		if err := rows.Scan(&m.Player.ID, &m.Message, &m.Date); err != nil {
			return err
		}
		m.Player.GetPlayerByID(m.Player.ID)
		e.Messages = append(e.Messages, m)
	}

	return nil
}

func GetEvents() (Events, error) {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	es := make(Events, 0)

	query := "SELECT idevent, name, date, paid_event, topic_arn, description, ownerid, invite_only, cost FROM event"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return es, err
	}

	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Name, &e.Date, &e.PaidEvent, &e.TopicArn, &e.Description, &e.Owner.ID, &e.InviteOnly, &e.Cost); err != nil {
			return es, err
		}

		t := time.Time{}
		awsdate := strings.Contains(e.Date, "T")
		if err != nil {
			return es, err
		}
		if awsdate {
			t, err = time.Parse("2006-01-02T15:04:00Z", e.Date)
			if err != nil {
				return es, err
			}
		} else {
			t, err = time.Parse("2006-01-02 15:04:00", e.Date)
			if err != nil {
				return es, err
			}
		}
		e.Date = t.Format("2006-01-02T15:04")

		e.Owner.GetPlayerByID(e.Owner.ID)
		if err != nil {
			return es, err
		}
		es = append(es, e)
	}

	for i, e := range es {
		query := fmt.Sprintf("SELECT idplayer, paid, subscription_arn FROM event_members WHERE idevent=%d", e.ID)
		ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelfunc()
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return es, err
		}

		for rows.Next() {
			var m EventMember
			if err := rows.Scan(&m.Player.ID, &m.Paid, &m.SubscriptionArn); err != nil {
				return es, err
			}
			m.Player.GetPlayerByID(m.Player.ID)
			es[i].Members = append(es[i].Members, m)
		}

		query = fmt.Sprintf("SELECT idsender, message, date FROM event_messages WHERE idevent=%d", e.ID)
		ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelfunc()
		rows, err = db.QueryContext(ctx, query)
		if err != nil {
			return es, err
		}
		for rows.Next() {
			var m EventMessage
			if err := rows.Scan(&m.Player.ID, &m.Message, &m.Date); err != nil {
				return es, err
			}
			m.Player.GetPlayerByID(m.Player.ID)
			e.Messages = append(e.Messages, m)
		}
	}

	return es, nil
}

func (e *Event) HasMember(p player.Player) bool {
	hm := false

	for _, m := range e.Members {
		if m.Player.ID == p.ID {
			hm = true
		}
	}

	return hm
}

func (e *Event) DeleteEvent() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, m := range e.Members {
		num, err := phonenumbers.Parse(m.Player.Phone, "US")
		if err != nil {
			return err
		}
		phone := phonenumbers.Format(num, phonenumbers.E164)
		msg := fmt.Sprintf("You have been removed from event %s. The event is being deleted.", e.Name)
		_, err = sms.SendTextPhone(msg, phone)
		if err != nil {
			return err
		}
		err = sms.RemoveSubscriber(m.SubscriptionArn)
		if err != nil {
			return err
		}
	}

	err = sms.DeleteTopic(e.TopicArn)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM event_members WHERE idevent=%d", e.ID)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM event_messages WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM event WHERE idevent=%d", e.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no event deleted")
	}

	return nil
}
