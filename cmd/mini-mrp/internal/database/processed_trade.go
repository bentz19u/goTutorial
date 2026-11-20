package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type ProcessedTrade struct {
	Server model.Server
	Ticket int
}

func (s *Store) GetAllProcessedTrades(ctx context.Context) (map[ProcessedTrade]bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `SELECT server, ticket
		FROM processed_trades`

	rows, err := s.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all processed_trades: %w", err)
	}
	defer rows.Close()

	result := make(map[ProcessedTrade]bool)
	for rows.Next() {
		var pt ProcessedTrade
		if err := rows.Scan(&pt.Server, &pt.Ticket); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[pt] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return result, nil
}

func (s *Store) InsertProcessedTrades(ctx context.Context, ex execer, pts []ProcessedTrade) error {
	if len(pts) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var sb strings.Builder
	sb.WriteString(`INSERT INTO processed_trades (server, ticket) VALUES `)

	args := make([]any, 0, len(pts)*2)
	for i, pt := range pts {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?,?)")
		args = append(args, pt.Server, pt.Ticket)
	}

	if _, err := ex.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert %d processed trades: %w", len(pts), err)
	}
	return nil
}

func (s *Store) purgeProcessedTrades(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, "DELETE FROM processed_trades")
	if err != nil {
		return fmt.Errorf("failed to purge processed_trades: %w", err)
	}

	return nil
}
