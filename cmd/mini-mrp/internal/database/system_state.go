package database

import (
	"context"
	"fmt"
	"time"
)

const (
	Heartbeat = "daemon_heartbeat"
)

func (s *Store) UpsertSystemState(ctx context.Context, key string, value string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, `
		INSERT INTO system_states
		(state_key, value)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE value = ?`,
		key, value, value,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert heartbeat: %w", err)
	}

	return nil
}

func (s *Store) purgeSystemStates(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := s.Db.ExecContext(ctx, "DELETE FROM system_states")
	if err != nil {
		return fmt.Errorf("failed to purge system_states: %w", err)
	}

	return nil
}
