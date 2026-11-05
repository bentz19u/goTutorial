package database

import (
	"context"
	"fmt"
	"time"
)

type Mt5Deal struct {
	Deal       int
	PositionID int
	Entry      int
	Login      int
	Symbol     string
	Volume     int
	Time       time.Time
	Price      float64
	Profit     float64
}

func (s *Store) AddMt5Deal(ctx context.Context, md Mt5Deal) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, `
		INSERT INTO mt5_deals 
		(Deal, PositionID, Entry, Login, Symbol, Volume, Time, Price, Profit)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		md.Deal, md.PositionID, md.Entry, md.Login, md.Symbol,
		md.Volume, md.Time, md.Price, md.Profit,
	)
	if err != nil {
		return fmt.Errorf("failed to insert mt5_deal %d: %w", md.Deal, err)
	}

	return nil
}
