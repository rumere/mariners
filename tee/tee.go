package tee

import (
	"context"
	"fmt"
	"mariners/db"
	"time"
)

type Tee struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Tees []Tee

func (t *Tee) AddTee() error {
	query := fmt.Sprintf("INSERT INTO ninthtee VALUES (null, \"%s\");", t.Name)

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

	return nil
}

func (t *Tee) GetTeeByID(id int64) error {
	query := fmt.Sprintf("SELECT idninthtee, name FROM ninthtee WHERE idninthtee=%d", id)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(&t.ID, &t.Name)
	if err != nil {
		return err
	}

	return nil
}

func (t *Tee) GetTeeByName(name string) error {
	query := fmt.Sprintf("SELECT idninthtee, name FROM ninthtee WHERE name=\"%s\"", name)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err := db.Con.QueryRowContext(ctx, query).Scan(&t.ID, &t.Name)
	if err != nil {
		return err
	}

	return nil
}

func GetTees() (Tees, error) {
	ts := make(Tees, 0)

	query := "SELECT idninthtee, name FROM ninthtee"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	rows, err := db.Con.QueryContext(ctx, query)
	if err != nil {
		return ts, err
	}

	for rows.Next() {
		var t Tee
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return ts, err
		}
		ts = append(ts, t)
	}

	return ts, nil
}
