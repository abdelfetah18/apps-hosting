package database

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bunotel"
)

func NewDatabase() *bun.DB {
	dsn := os.Getenv("DATABASE_CONNECTION_STRING")
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}

	sqldb := sql.OpenDB(stdlib.GetConnector(*config))

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName(fmt.Sprintf("%s_database", os.Getenv("SERVICE_NAME")))))
	return db
}
