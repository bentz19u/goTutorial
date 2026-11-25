package consumer

import (
	"context"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type Window interface {
	GetAllClosedTradesSorted() []model.ClosedTrade
}

type PushConsumer interface {
	GetName() string
	OnEvent(ctx context.Context, cts []model.ClosedTrade) error
}
