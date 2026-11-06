package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Mt4Trade struct {
	Ticket     int
	Account    int
	Symbol     string
	Volume     int
	OpenTime   time.Time
	CloseTime  sql.NullTime
	OpenPrice  float64
	ClosePrice sql.NullFloat64
	Profit     sql.NullFloat64
}

func (s *Store) UpsertMt4Trade(ctx context.Context, mt Mt4Trade) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, `
		INSERT INTO mt4_trades 
		(ticket, account, symbol, volume, open_time, close_time, open_price, close_price, profit)
		VALUES (?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE close_time = ?, close_price = ?, profit = ?`,
		mt.Ticket, mt.Account, mt.Symbol, mt.Volume, mt.OpenTime,
		mt.CloseTime, mt.OpenPrice, mt.ClosePrice, mt.Profit,
		mt.CloseTime, mt.ClosePrice, mt.Profit,
	)
	if err != nil {
		return fmt.Errorf("failed to insert mt4_trade %d: %w", mt.Ticket, err)
	}

	return nil
}

func (s *Store) purgeMt4Trades(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, "DELETE FROM mt4_trades")
	if err != nil {
		return fmt.Errorf("failed to purge mt4_trades: %w", err)
	}

	return nil
}

func (s *Store) CountMt4Trades(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var result int
	err := s.Db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mt4_trades").Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to count mt4_trades: %w", err)
	}

	return result, nil
}
