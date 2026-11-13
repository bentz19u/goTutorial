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

func (s *Store) GetClosedMt5DealsFromDeal(ctx context.Context, fromDeal *int) ([]Mt5Deal, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `SELECT Deal, PositionID, Entry, Login, Symbol,
		Volume, Time, Price, Profit 
		FROM mt5_deals
		WHERE Entry = 1`
	var args []any
	if fromDeal != nil {
		query += " AND Deal > ?"
		args = append(args, fromDeal)
	}
	query += " ORDER BY Deal ASC"

	rows, err := s.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all mt5_deals, from deal %v: %w", fromDeal, err)
	}
	defer rows.Close()

	var result []Mt5Deal
	for rows.Next() {
		var d Mt5Deal
		if err := rows.Scan(&d.Deal, &d.PositionID, &d.Entry, &d.Login, &d.Symbol,
			&d.Volume, &d.Time, &d.Price, &d.Profit); err != nil {
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
