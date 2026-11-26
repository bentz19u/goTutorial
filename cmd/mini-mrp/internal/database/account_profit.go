package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bentz19u/go-trade-processor/cmd/mini-mrp/internal/model"
)

type AccountProfitKey struct {
	Server model.Server
	Login  int
}

type AccountProfit struct {
	Server      model.Server
	Login       int
	ProfitCents int64
}

func (s *Store) UpsertAccountProfits(ctx context.Context, ex execer, aps []AccountProfit) error {
	if len(aps) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var sb strings.Builder
	sb.WriteString(`INSERT INTO account_profits (server, login, profitCents) VALUES `)

	args := make([]any, 0, len(aps)*3)
	for i, ap := range aps {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?,?,?)")
		args = append(args, ap.Server, ap.Login, ap.ProfitCents)
	}

	sb.WriteString(` AS new ON DUPLICATE KEY UPDATE profitCents = account_profits.profitCents + new.profitCents`)

	if _, err := ex.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("insert %d account profits: %w", len(aps), err)
	}
	return nil
}

func (s *Store) purgeAccountProfits(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, "DELETE FROM account_profits")
	if err != nil {
		return fmt.Errorf("failed to purge account_profits: %w", err)
	}

	return nil
}
