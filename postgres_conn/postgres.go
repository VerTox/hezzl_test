package postgres_conn

import (
	"context"
	"github.com/jackc/pgx/v4"
	"os"
)

const defaultDBUrl = "postgres://postgres:mysecretpassword@localhost:5432/postgres"

const initTable = `
		CREATE TABLE IF NOT EXISTS users
		(
			id serial primary key,
			name varchar(255)
		)
		`

func NewPostgresConn() (*pgx.Conn, error) {
	baseUrl := os.Getenv("DB_URL")
	if baseUrl == "" {
		baseUrl = defaultDBUrl
	}
	conn, err := pgx.Connect(context.Background(), baseUrl)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(context.Background(), initTable)
	if err != nil {
		return nil, err
	}

	return conn, err
}
