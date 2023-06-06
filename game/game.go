package game

import (
	"context"
	"database/sql"
	"fmt"
	"mariners/db"
	"mariners/player"
	"mariners/tee"
	"mariners/weather"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

type Game struct {
	Weather weather.WeatherHours
	Tee     tee.Tee
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	IsMatch bool   `json:"is_match"`
	Checkins
}

type Games []Game

type Checkin struct {
	PlayerID int64
	GameID   int64
	Date     string
}

type Checkins []Checkin

func (g *Game) AddGame() error {
	w, err := weather.AddWeather()
	if err != nil {
		return err
	}
	g.Weather = w

	for _, id := range g.Weather {
		log.Info().Msgf("Game has weather ID %d", id)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t := time.Now().In(loc)
	g.Date = t.Format("2006-01-02")

	query := fmt.Sprintf("INSERT INTO game (idgame, game_date, idninthtee, ismatch) VALUES (NULL, \"%s\", %d, %t);\n",
		g.Date,
		g.Tee.ID,
		g.IsMatch)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.Con.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	g.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) UpdateGame() error {
	query := fmt.Sprintf("UPDATE game set idninthtee = %d, ismatch=%t where idplayer=%d;", g.Tee.ID, g.IsMatch, g.ID)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.Con.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("UpdateGame: No game where idgame = %d", g.ID)
	}

	return nil
}

func (g *Game) GetGameByID(id int64) error {
	query := "SELECT idgame, game_date, idninthtee, ismatch FROM game WHERE idgame=?"

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.Date,
		&g.Tee.ID,
		&g.IsMatch)
	if err != nil {
		return err
	}

	return nil
}

func GetGameByDate(t time.Time) (Game, error) {
	g := Game{}
	d := fmt.Sprintf("%d-%02d-%02d 00:00", t.Year(), t.Month(), t.Day())
	e := fmt.Sprintf("%d-%02d-%02d 23:59", t.Year(), t.Month(), t.Day())
	query := fmt.Sprintf(
		"SELECT "+
			"idgame, "+
			"game_date, "+
			"idninthtee, "+
			"ismatch "+
			"FROM game WHERE "+
			"UNIXEPOCH(game_date) BETWEEN UNIXEPOCH('%s') AND UNIXEPOCH('%s')",
		d, e)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(
		&g.ID,
		&g.Date,
		&g.Tee.ID,
		&g.IsMatch)

	if err != nil {
		return g, err
	} else {
		w, err := weather.GetWeatherByDate(g.Date)
		if err != nil {
			return g, err
		}
		log.Info().Msgf("%#v", w)
		g.Weather = w

		err = g.Tee.GetTeeByID(g.Tee.ID)
		if err != nil {
			return g, err
		}

		err = g.GetCheckinsByDate(t)
		if err != nil && err != sql.ErrNoRows {
			return g, err
		}
	}

	return g, nil
}

func GetGames() (Games, error) {
	gs := make(Games, 0)

	query := "SELECT idgame, idweather, game_date, idninthtee, ismatch FROM game"

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.Con.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile("T")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var g Game
		err := rows.Scan(
			&g.ID,
			&g.Date,
			&g.Tee.ID,
			&g.IsMatch)
		if err != nil {
			return nil, err
		}

		var t time.Time
		awsdate := re.Match([]byte(g.Date))
		if err != nil {
			return nil, err
		}
		if awsdate {
			t, err = time.Parse("2006-01-02T15:04:00Z", g.Date)
			if err != nil {
				return nil, err
			}
		} else {
			t, err = time.Parse("2006-01-02 15:04:00", g.Date)
			if err != nil {
				return nil, err
			}
		}

		err = g.GetCheckinsByDate(t)
		if err != nil {
			return nil, err
		}

		g.Weather, err = weather.GetWeatherByDate(g.Date)
		if err != nil {
			return nil, err
		}

		gs = append(gs, g)
	}

	return gs, nil
}

func (g *Game) AddCheckin(p player.Player) error {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t := time.Now().In(loc)
	d := t.Format("2006-01-02 03:04")

	query := fmt.Sprintf("INSERT INTO checkins (%d, %d, %s)", p.ID, g.ID, d)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.Con.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) GetCheckinsByDate(t time.Time) error {
	g.Checkins = nil

	today := fmt.Sprintf("%d-%d-%d", t.Year(), t.Month(), t.Day())

	query := fmt.Sprintf("SELECT idgame FROM game WHERE game_date=%s", today)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(&g.ID)
	if err != nil {
		return err
	}

	err = g.GetGameByID(g.ID)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("SELECT idplayer, checkin_date FROM checkins where idgame=%d", g.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.Con.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	for rows.Next() {
		var ci Checkin
		err = rows.Scan(&ci.PlayerID, &ci.Date)
		if err != nil {
			return err
		}
		ci.GameID = g.ID
		g.Checkins = append(g.Checkins, ci)
	}

	return nil
}
