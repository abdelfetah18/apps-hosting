package database

import (
	"database/sql"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func NewDatabase() *bun.DB {
	dsn := os.Getenv("DATABASE_CONNECTION_STRING")
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	sqldb := sql.OpenDB(stdlib.GetConnector(*config))

	return bun.NewDB(sqldb, pgdialect.New())
}
