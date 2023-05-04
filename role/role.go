package role

import (
	"context"
	"fmt"
	"mariners/db"
	"time"
)

type Roles map[int64]string

func AddRole(n string) (Roles, error) {
	r := make(Roles)

	query := fmt.Sprintf("INSERT INTO roles VALUES (null, \"%s\");", n)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.Con.ExecContext(ctx, query)
	if err != nil {
		return r, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return r, err
	}
	if rows == 0 {
		return r, fmt.Errorf("no role added")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return r, err
	}

	r[id] = n

	return r, nil
}

func GetRoleByID(id int64) (Roles, error) {
	r := make(Roles)
	var name string

	query := fmt.Sprintf("SELECT name FROM role WHERE idrole=%d", id)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(&name)
	if err != nil {
		return r, err
	}
	r[id] = name

	return r, nil
}

func GetRoleIDByName(name string) (int64, error) {
	var id int64

	query := fmt.Sprintf("SELECT idrole FROM role WHERE name=\"%s\"", name)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func GetRoles() (Roles, error) {
	r := make(Roles)

	query := "SELECT idrole, name FROM role"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.Con.QueryContext(ctx, query)
	if err != nil {
		return r, err
	}

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return r, err
		}
		r[id] = name
	}

	return r, nil
}

func GetRolesByPlayerID(id int64) (Roles, error) {
	r := make(Roles)

	query := fmt.Sprintf("SELECT role.idrole, role.name FROM role INNER JOIN role_members ON role.idrole=role_members.idrole WHERE role_members.idplayer=%d", id)
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.Con.QueryContext(ctx, query)
	if err != nil {
		return r, err
	}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return r, err
		}
		r[id] = name
	}

	return r, nil
}
