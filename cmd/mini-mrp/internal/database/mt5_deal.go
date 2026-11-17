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

type ClosedMt5Deal struct {
	OpenDeal   int
	CloseDeal  int
	PositionID int
	Login      int
	Symbol     string
	Volume     int
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	Profit     float64
}

func (s *Store) GetClosedMt5DealsFromDeal(ctx context.Context, fromDeal *int) ([]ClosedMt5Deal, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `SELECT open.Deal as openDeal, close.Deal as closeDeal, 
		close.PositionID, close.Login, close.Symbol,
		close.Volume, open.Time as openTime, close.Time as closeTime, 
		open.Price as openPrice, close.Price as closePrice, close.Profit 
		FROM mt5_deals AS close
		INNER JOIN mt5_deals AS open ON close.PositionID = open.PositionId and open.Entry = 0
		WHERE close.Entry = 1`
	var args []any
	if fromDeal != nil {
		query += " AND close.Deal > ?"
		args = append(args, fromDeal)
	}
	query += " ORDER BY close.Deal ASC"

	rows, err := s.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all mt5_deals, from deal %v: %w", fromDeal, err)
	}
	defer rows.Close()

	var result []ClosedMt5Deal
	for rows.Next() {
		var d ClosedMt5Deal
		if err := rows.Scan(&d.OpenDeal, &d.CloseDeal, &d.PositionID,
			&d.Login, &d.Symbol, &d.Volume,
			&d.OpenTime, &d.CloseTime, &d.OpenPrice, &d.ClosePrice, &d.Profit); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result = append(result, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return result, nil
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

func (s *Store) purgeMt5Deals(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, "DELETE FROM mt5_deals")
	if err != nil {
		return fmt.Errorf("failed to purge mt5_deals: %w", err)
	}

	return nil
}

func (s *Store) CountMt5Deals(ctx context.Context) (deals int, positions int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = s.Db.QueryRowContext(ctx, "SELECT COUNT(Deal), COUNT(DISTINCT PositionID) FROM mt5_deals").Scan(&deals, &positions)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count mt5_deals: %w", err)
	}

	return deals, positions, nil
}
