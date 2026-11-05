package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Store struct {
	Db *sql.DB
}

func new(db *sql.DB) *Store {
	return &Store{Db: db}
}

func InitDatabase(ctx context.Context) (*Store, error) {
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
	return new(db), nil
}

func initMt4Table(ctx context.Context, s *Store) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// for MT4 the same open and close information are in the same ticket (UPDATE)
	// 1 lot = 100 volume
	_, err := s.Db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mt4_trades (
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

func initMt5Table(ctx context.Context, s *Store) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// one trade has two deals, one common positionID
	// one deal for the open ticket, one deal for the close ticket
	// entry 0 = open ticket, entry 1 = close ticket
	// 1 lot = 10000 volume
	_, err := s.Db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mt5_deals (
		Deal BIGINT PRIMARY KEY,
		PositionID INT NOT NULL,
		Entry INT NOT NULL,
		Login INT NOT NULL,
		Symbol VARCHAR(16) NOT NULL,
		Volume BIGINT NOT NULL,
		Time DATETIME NOT NULL,
		Price DOUBLE NOT NULL,
		Profit DOUBLE 
	)`)
	if err != nil {
		return err
	}
	return nil
}

func InitTables(ctx context.Context, s *Store) error {
	err := initMt4Table(ctx, s)
	if err != nil {
		return err
	}

	err = initMt5Table(ctx, s)
	if err != nil {
		return err
	}

	return nil
}

func PurgeTables(ctx context.Context, s *Store) error {
	err := s.purgeMt4Trades(ctx)
	if err != nil {
		return err
	}

	err = s.purgeMt5Deals(ctx)
	if err != nil {
		return err
	}

	return nil
}
