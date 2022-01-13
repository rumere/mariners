package game

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mariners/db"
	"mariners/weather"
	"time"
)

type Game struct {
	ID        int64  `json:"id"`
	WeatherID int64  `json:"weather_id"`
	Date      string `json:"date"`
	NinthTee  string `json:"ninth_tee"`
	IsMatch   bool   `json:"is_match"`
}

type Games []Game

func (g *Game) AddGame() error {
	db, err := db.DBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	w := weather.Weather{}
	w.AddWeather()

	fmt.Printf("%#v\n", &w)

	g.WeatherID = w.ID

	err = writeGame(db, g)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) GetGame(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getGame(db, id, g)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) GetGameByDate(d string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getGameByDate(db, d, g)
	if err != nil {
		return err
	}

	return nil
}

func getGameByDate(db *sql.DB, d string, g *Game) error {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return err
	}
	s := fmt.Sprintf("\"%d-%d-%d 00:00:00\"", t.Year(), t.Month(), t.Day())
	f := fmt.Sprintf("\"%d-%d-%d 23:59:59\"", t.Year(), t.Month(), t.Day())
	query := fmt.Sprintf(
		"SELECT "+
			"idgame, "+
			"idweather, "+
			"date, "+
			"ninth_tee, "+
			"ismatch "+
			"FROM game WHERE "+
			"date>=%s AND date<=%s",
		s, f)

	fmt.Printf("%s", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
		&g.ID,
		&g.WeatherID,
		&g.Date,
		&g.NinthTee,
		&g.IsMatch)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", g)

	return nil
}

func (gs Games) GetGames() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gs, err = getGames(db)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", gs)

	return nil
}

func getGame(db *sql.DB, id int64, g *Game) error {
	query := "SELECT idgame, idweather, date FROM game WHERE idgame=?"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.QueryRowContext(ctx, query, id).Scan(&g.ID, &g.WeatherID, &g.Date)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", g)

	return nil
}

func getGames(db *sql.DB) (Games, error) {
	g := make(Games, 0)

	query := "SELECT idgame, idweather, date FROM game"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return g, err
	}

	for rows.Next() {
		var game Game
		if err := rows.Scan(&game.ID, &game.WeatherID, &game.Date); err != nil {
			return g, err
		}
		g = append(g, game)
	}

	return g, nil
}

func writeGame(db *sql.DB, g *Game) error {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t := time.Now().In(loc)
	g.Date = t.Format("2006-01-02 03:04")

	fmt.Printf("%#v\n", g)

	query := fmt.Sprintf("INSERT INTO game (idgame, idweather, date, ninth_tee, ismatch) VALUES (NULL, %d, \"%s\", \"%s\", %t);\n",
		g.WeatherID,
		g.Date,
		g.NinthTee,
		g.IsMatch)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	g.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Printf("%#v", g)
	fmt.Printf("Rows affected by insert: %d\n", rows)

	return nil
}
