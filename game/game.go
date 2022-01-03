package game

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mariners/db"
	"time"
)

type Game struct {
	ID        int64  `json:"id"`
	WeatherID int64  `json:"weather_id"`
	Date      string `json:"date"`
	NinthTee  string `json:"ninth_tee"`
}

type Games []Game

func AddGame(g *Game) error {
	db, err := db.DBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	err = writeGame(db, g)
	if err != nil {
		return err
	}

	return nil
}

func GetGame(id int64, g *Game) error {
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

func GetGameByDate(d string, g *Game) error {
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
			"ninth_tee"+
			"FROM game WHERE "+
			"date>=%s AND date<=%s",
		s, f)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query).Scan(
		&g.ID,
		&g.Date,
		&g.WeatherID,
		&g.NinthTee)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", g)

	return nil
}

func GetGames() (Games, error) {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	g, err := getGames(db)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%#v", g)

	return g, nil
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

	query := "SELECT idgame, idweather, date FROM weather"

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
	g.Date = fmt.Sprintf("%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day())

	query := fmt.Sprintf("SELECT idweather FROM weather WHERE date >= \"%s 00:00:00\" AND date <= \"%s 23:59:59\"", g.Date, g.Date)
	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.QueryRowContext(ctx, query).Scan(&g.WeatherID)
	if err != nil {
		return err
	}

	fmt.Printf("%#v\n", g)

	query = fmt.Sprintf("INSERT INTO game (idgame, idweather, date, ninth_tee) VALUES (NULL, %d, \"%s\");\n",
		g.WeatherID,
		g.Date)

	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
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

	fmt.Printf("Rows affected by insert: %d\n", rows)

	return nil
}
