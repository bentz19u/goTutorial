package consumer

import (
	"context"
	"log/slog"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type Fanout struct {
	Consumers []PushConsumer
}

func (f *Fanout) Dispatch(ctx context.Context, trades []model.ClosedTrade) {
	for _, c := range f.Consumers {
		err := c.OnEvent(ctx, trades)
		if err != nil {
			slog.Error("dispatching trades to consumer failed", "consumer", c.GetName(), "error", err)
		}
	}
}
