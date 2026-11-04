package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func initDatabase(ctx context.Context) (*sql.DB, error) {
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, errors.New("DB_PASSWORD not configured")
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3307)/minimrp", dbPassword))
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	errPing := db.PingContext(ctx)
	if errPing != nil {
		return nil, fmt.Errorf("pinging database: %w", errPing)
	}

	slog.Info("ping to database successful")
	return db, nil
}

func initMt4Table(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mt4_trades (
		ticket INT PRIMARY KEY,
		account INT NOT NULL,
		symbol VARCHAR(16) NOT NULL,
		volume INT NOT NULL,
		open_time DATETIME NOT NULL,
		close_time DATETIME,
		open_price DOUBLE NOT NULL,
		close_price DOUBLE,
		profit DOUBLE 
	)`)
	if err != nil {
		return err
	}
	return nil
}

func initTables(ctx context.Context, db *sql.DB) error {
	err := initMt4Table(ctx, db)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	initLogger()

	ctx := context.Background()

	db, errDb := initDatabase(ctx)
	if errDb != nil {
		slog.Error("connection to database failed", "error", errDb)
		os.Exit(1)
	}
	defer db.Close()

	errTables := initTables(ctx, db)
	if errTables != nil {
		slog.Error("initTables failed", "error", errTables)
		os.Exit(1)
	}
}
