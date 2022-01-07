package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var pool *pgxpool.Pool

func initDB(ctx context.Context) {
	db := config.Database
	connStr := ""
	if config.Database.Port == "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&connect_timeout=3", db.User, db.Password, db.Host, db.Name)
	} else {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&connect_timeout=3", db.User, db.Password, db.Host, db.Port, db.Name)
	}
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		panic(err)
	}
	config.MinConns = 4
	config.MaxConns = 256

	pool, err = pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
}

func RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return tx.Rollback(ctx)
	}
	return tx.Commit(ctx)
}
