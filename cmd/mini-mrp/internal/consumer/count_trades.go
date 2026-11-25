package consumer

import (
	"context"
	"log/slog"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type CountTradeConsumer struct{}

func (ct *CountTradeConsumer) GetName() string {
	return "Count Trade Consumer"
}

func (ct *CountTradeConsumer) OnEvent(ctx context.Context, trades []model.ClosedTrade) error {
	slog.Info("Count Trade consumer received new trades", "count", len(trades))
	return nil
}
