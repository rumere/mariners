package player

import (
	"context"
	"fmt"
	"log"
	"mariners/db"
	"mariners/role"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Player struct {
	Roles         role.Roles
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	GhinNumber    string `json:"ghin_number"`
}

type Players []Player

func AddPlayer(p *Player) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO player (idplayer, name, preferred_name, phone, email, ghin_number) VALUES (NULL, \"%s\", \"%s\", \"%s\", \"%s\", \"%s\");\n",
		p.Name,
		p.PreferredName,
		p.Phone,
		p.Email,
		p.GhinNumber)
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
	if rows == 0 {
		return fmt.Errorf("no player added")
	}

	for rk := range p.Roles {
		query = fmt.Sprintf("INSERT INTO role_members (idrole, idplayer) VALUES (%d, %d)",
			rk,
			p.ID)
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
		if rows == 0 {
			return fmt.Errorf("no role added for player %d", p.ID)
		}
	}

	return nil
}

func (p *Player) GetPlayerByID(id int64) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number FROM player WHERE idplayer=?"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.PreferredName, &p.Phone, &p.Email, &p.GhinNumber)
	if err != nil {
		return err
	}

	p.Roles, err = role.GetRolesByPlayerID(p.ID)
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

	p := make(Players, 0)

	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number FROM player ORDER BY preferred_name"
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
		player.Roles, err = role.GetRolesByPlayerID(player.ID)
		if err != nil {
			return p, err
		}
		p = append(p, player)
	}

	return p, nil
}

func (p *Player) GetPlayerByToken(token string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := "SELECT idplayer, name, preferred_name, phone, email, ghin_number FROM player WHERE token=?"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.QueryRowContext(ctx, query, token).Scan(&p.ID, &p.Name, &p.PreferredName, &p.Phone, &p.Email, &p.GhinNumber)
	if err != nil {
		return err
	}

	p.Roles, err = role.GetRolesByPlayerID(p.ID)
	if err != nil {
		return err
	}

	return nil
}

func (p *Player) WriteToken(token string) error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE player set token = \"%s\" WHERE idplayer = %d;\n", token, p.ID)
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
	if rows == 0 {
		return fmt.Errorf("no token added")
	}

	return nil
}

func (p *Player) RemoveToken() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE player set token = null WHERE idplayer = %d;\n", p.ID)
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
	if rows == 0 {
		return fmt.Errorf("no token removed")
	}

	return nil
}

func (p *Player) UpdatePlayer() error {
	db, err := db.DBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("UPDATE player set name = \"%s\", preferred_name = \"%s\", phone = \"%s\", email = \"%s\", ghin_number = \"%s\" WHERE idplayer = %d;\n",
		p.Name,
		p.PreferredName,
		p.Phone,
		p.Email,
		p.GhinNumber,
		p.ID)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM role_members WHERE idplayer = %d",
		p.ID)
	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	for rk := range p.Roles {
		query = fmt.Sprintf("INSERT INTO role_members (idrole, idplayer) VALUES (%d, %d)",
			rk,
			p.ID)
		ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelfunc()
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Player) DeletePlayer() error {
	db, err := db.DBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("DELETE FROM role_members WHERE idplayer=%d", p.ID)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = fmt.Sprintf("DELETE FROM player WHERE idplayer=%d", p.ID)

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
		return fmt.Errorf("no player deleted")
	}

	return nil
}

func (p *Player) HasRole(rolename string) bool {
	hr := false

	for _, r := range p.Roles {
		if r == rolename {
			hr = true
		}
	}

	return hr
}
