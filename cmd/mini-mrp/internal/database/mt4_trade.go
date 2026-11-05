package database

import (
	"context"
	"fmt"
	"time"
)

type Mt4Trade struct {
	Ticket     int
	Account    int
	Symbol     string
	Volume     int
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	Profit     float64
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
