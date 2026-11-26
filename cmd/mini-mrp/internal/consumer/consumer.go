package consumer

import (
	"context"
	"sync"

	"github.com/bentz19u/go-trade-processor/cmd/mini-mrp/internal/model"
)

type Window interface {
	GetAllClosedTradesSorted() []model.ClosedTrade
}

type PushConsumer interface {
	GetName() string
	OnEvent(ctx context.Context, cts []model.ClosedTrade) error
	Start(ctx context.Context, wg *sync.WaitGroup)
}
