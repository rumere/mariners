package player

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mariners/db"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Player struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	GhinNumber    string `json:"ghin_number"`
	Role          string `json:"role"`
}

type Players []Player

func AddPlayer(p *Player) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = writePlayer(db, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Player) GetPlayerByID(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = getPlayer(db, id, p)
	if err != nil {
		return err
	}

	return nil
}

func GetPlayers() (Players, error) {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	p, err := getPlayers(db)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Player) GetPlayerByToken(token string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number, role FROM player WHERE token=?"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query, token).Scan(&p.ID, &p.Name, &p.PreferredName, &p.Phone, &p.Email, &p.GhinNumber, &p.Role)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", p)

	return nil
}

func (p *Player) WriteToken(token string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE player set token = \"%s\" WHERE idplayer = %d;\n", token, p.ID)

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
	fmt.Printf("Rows affected by update: %d\n", rows)

	return nil
}

func UpdatePlayer(id int64, p *Player) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = updatePlayer(db, id, p)
	if err != nil {
		return err
	}

	return nil
}

func DeletePlayer(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = deletePlayer(db, id)
	if err != nil {
		return err
	}

	return nil
}

func getPlayer(db *sql.DB, id int64, p *Player) error {
	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number, role FROM player WHERE idplayer=?"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.PreferredName, &p.Phone, &p.Email, &p.GhinNumber, &p.Role)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", p)

	return nil
}

func getPlayers(db *sql.DB) (Players, error) {
	p := make(Players, 0)

	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number FROM player ORDER BY preferred_name"

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return p, err
	}

	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Name, &player.PreferredName, &player.Phone, &player.Email, &player.GhinNumber); err != nil {
			return p, err
		}
		p = append(p, player)
	}

	return p, nil
}

func writePlayer(db *sql.DB, p *Player) error {
	query := fmt.Sprintf("INSERT INTO player (idplayer, name, preferred_name, phone, email, ghin_number) VALUES (NULL, \"%s\", \"%s\", \"%s\", \"%s\", \"%s\");\n",
		p.Name,
		p.PreferredName,
		p.Phone,
		p.Email,
		p.GhinNumber)

	fmt.Printf("\n\nQUERY: \n%s\n\n", query)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	p.ID, err = res.LastInsertId()
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

func updatePlayer(db *sql.DB, id int64, p *Player) error {
	query := fmt.Sprintf("UPDATE player set name = \"%s\", preferred_name = \"%s\", phone = \"%s\", email = \"%s\", ghin_number = \"%s\" WHERE idplayer = %d;\n",
		p.Name,
		p.PreferredName,
		p.Phone,
		p.Email,
		p.GhinNumber,
		id)

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
	fmt.Printf("Rows affected by update: %d\n", rows)

	p.ID = id

	return nil
}

func deletePlayer(db *sql.DB, id int64) error {
	query := fmt.Sprintf("DELETE FROM player WHERE idplayer=%d", id)

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

	fmt.Printf("Rows affected by delete: %d\n", rows)

	return nil
}
