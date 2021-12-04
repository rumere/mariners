package game

import (
	"context"
	"database/sql"
	"fmt"
	"mariners/db"
	"time"
)

type Game struct {
	ID        int64  `json:"id"`
	WeatherID int64  `json:"weather_id"`
	Date      string `json:"date"`
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

	query = fmt.Sprintf("INSERT INTO game (idgame, idweather, date) VALUES (NULL, %d, \"%s\");\n",
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
