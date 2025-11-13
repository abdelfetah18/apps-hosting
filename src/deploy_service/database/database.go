package database

import (
	"database/sql"
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
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName("deploy_service_database")))
	return db
}
