package game

import (
	"context"
	"fmt"
	"log"
	"mariners/db"
	"mariners/tee"
	"mariners/weather"
	"time"
)

type Game struct {
	Weather weather.Weather
	Tee     tee.Tee
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	IsMatch bool   `json:"is_match"`
}

type Games []Game

func (g *Game) AddGame() error {
	db, err := db.DBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	g.Weather.AddWeather()

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return err
	}
	t := time.Now().In(loc)
	g.Date = t.Format("2006-01-02 03:04")

	fmt.Printf("%#v\n", g)

	query := fmt.Sprintf("INSERT INTO game (idgame, idweather, date, idninthtee, ismatch) VALUES (NULL, %d, \"%s\", %d, %t);\n",
		g.Weather.ID,
		g.Date,
		g.Tee.ID,
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

	return nil
}

func (g *Game) UpdateGame() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE game set idninthtee = %d, ismatch=%t where idplayer=%d;", g.Tee.ID, g.IsMatch, g.ID)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("UpdateGame: No game by where idgame = %d", g.ID)
	}

	return nil
}

func (g *Game) GetGameByID(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "SELECT idgame, idweather, date, idninthtee, ismatch FROM game WHERE idgame=?"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.Weather.ID,
		&g.Date,
		&g.Tee.ID,
		&g.IsMatch)
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
			"idninthtee, "+
			"ismatch "+
			"FROM game WHERE "+
			"date>=%s AND date<=%s",
		s, f)

	fmt.Printf("%s", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
		&g.ID,
		&g.Weather.ID,
		&g.Date,
		&g.Tee.ID,
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

	query := "SELECT idgame, idweather, date, idninthtee, ismatch FROM game"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	var game Game
	for rows.Next() {
		err := rows.Scan(
			&game.ID,
			&game.Weather.ID,
			&game.Date,
			&game.Tee.ID,
			&game.IsMatch)
		if err != nil {
			return err
		}
		gs = append(gs, game)
	}

	return nil
}
