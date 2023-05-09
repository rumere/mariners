package team

import (
	"context"
	"fmt"
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
	err := writeTeam(gid, t)
	if err != nil {
		return err
	}

	return nil
}

func GetTeam(id int64, t *Team) error {
	err := getTeam(id, t)
	if err != nil {
		return err
	}

	return nil
}

func getTeam(id int64, t *Team) error {
	query := "SELECT idteam, idgame FROM team WHERE idteam=?"

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.GameID)
	if err != nil {
		return err
	}

	return nil
}

func writeTeam(gid int64, t *Team) error {
	query := fmt.Sprintf("INSERT INTO team (idteam, idgame) VALUES (NULL, %d);\n", gid)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.Con.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	t.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}
