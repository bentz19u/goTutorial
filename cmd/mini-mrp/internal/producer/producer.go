package producer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
)

func StartProducer(ctx context.Context, store *database.Store, wg *sync.WaitGroup) {
	wg.Add(1)

	// on boot, for now, watermark start at nil
	var watermarkMt4 *time.Time
	var watermarkMt5 *int

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				newWatermarkMt4 := processMt4(ctx, watermarkMt4, store)
				if newWatermarkMt4 != nil {
					watermarkMt4 = newWatermarkMt4
				}
				newWatermarkMt5 := processMt5(ctx, watermarkMt5, store)
				if newWatermarkMt5 != nil {
					watermarkMt5 = newWatermarkMt5
				}
			case <-ctx.Done():
				slog.Info("producer shutting down")
				return
			}
		}
	}()
}

func processMt4(ctx context.Context, watermark *time.Time, store *database.Store) *time.Time {
	trades, err := store.GetMt4TradesFromCloseTime(ctx, watermark)
	if err != nil {
		slog.Error("failed to get new mt4 trades", "error", err)
		return nil
	}

	if len(trades) == 0 {
		return nil
	}

	slog.Info("new trades found", "trades", trades)

	// simulating processing them

	// updating watermark
	last := trades[len(trades)-1]
	watermark = &last.CloseTime.Time

	slog.Info("mt4 watermark updated", "watermark", watermark)
	return watermark
}

func processMt5(ctx context.Context, watermark *int, store *database.Store) *int {
	deals, err := store.GetClosedMt5DealsFromDeal(ctx, watermark)
	if err != nil {
		slog.Error("failed to get new mt5 deals", "error", err)
		return nil
	}

	if len(deals) == 0 {
		return nil
	}

	slog.Info("new deals found", "deals", deals)

	// simulating processing them

	// updating watermark
	last := deals[len(deals)-1]
	watermark = &last.Deal

	slog.Info("mt5 watermark updated", "watermark", watermark)
	return watermark
}
