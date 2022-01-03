package team

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mariners/db"
	"time"
)

type Team struct {
	ID     int64 `json:"id"`
	GameID int64 `json:"game_id"`
}
type TeamMember struct {
	TeamID       int64 `json:"team_id"`
	PlayerID     int64 `json:"player_id"`
	Ghost        bool  `json:"ghost"`
	NinthDropped bool  `json:"ninth_dropped"`
}

func AddTeam(gid int64, t *Team) error {
	db, err := db.DBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	err = writeTeam(db, gid, t)
	if err != nil {
		return err
	}

	return nil
}

func GetTeam(id int64, t *Team) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getTeam(db, id, t)
	if err != nil {
		return err
	}

	return nil
}

func getTeam(db *sql.DB, id int64, t *Team) error {
	query := "SELECT idteam, idgame FROM team WHERE idteam=?"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.GameID)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", t)

	return nil
}

func writeTeam(db *sql.DB, gid int64, t *Team) error {
	query := fmt.Sprintf("INSERT INTO team (idteam, idgame) VALUES (NULL, %d);\n", gid)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	t.ID, err = res.LastInsertId()
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
