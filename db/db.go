package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func dsn(dbName string) string {
	username := getEnv("dbuser", "admin")
	password := getEnv("dbpassword", "")
	hostname := getEnv("dbhost", "db.mplinksters.club:3306")

	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
}

func DBConnection() (*sql.DB, error) {
	dbname := getEnv("dbname", "mplinksters")

	db, err := sql.Open("mysql", dsn(dbname))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
