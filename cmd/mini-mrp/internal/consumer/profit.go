package consumer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/producer"
)

func StartProfitConsumer(ctx context.Context, window Window, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				closedTrades := window.GetAllClosedTradesSorted()
				profits := sumProfit(closedTrades)
				slog.Info("window sum of profit", "profits", profits)
			case <-ctx.Done():
				slog.Info("profit consumer shutting down")
				return
			}
		}
	}()
}

func sumProfit(cts []producer.ClosedTrade) float64 {
	var sum int64
	for _, ct := range cts {
		sum += ct.Profit
	}
	return float64(sum) / 100
}
