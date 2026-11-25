package producer

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/consumer"
	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

const (
	volumeToLotMt4    = 100
	volumeToLotMt5    = 10000
	priceDecimal      = 100000
	profitDecimal     = 100
	heartBeatInterval = 10 * time.Second
)

func StartProducer(ctx context.Context, window *ClosedTradeWindow, fanout *consumer.Fanout, store *database.Store, wg *sync.WaitGroup) {
	wg.Add(1)

	// on boot, for now, watermark start at nil
	var watermarkMt4 *time.Time
	var watermarkMt5 *int

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		tickerCleaningWindow := time.NewTicker(20 * time.Second)
		defer tickerCleaningWindow.Stop()

		tickerReconcile := time.NewTicker(15 * time.Second)
		defer tickerReconcile.Stop()

		var lastHeartBeat time.Time

		for {
			select {
			case <-ticker.C:
				mt4Trades, newWatermarkMt4, errMt4 := processMt4(ctx, watermarkMt4, store, window)
				if errMt4 != nil {
					slog.Error("failed process mt4", "error", errMt4)
				}
				if newWatermarkMt4 != nil {
					watermarkMt4 = newWatermarkMt4
				}

				mt5Trades, newWatermarkMt5, errMt5 := processMt5(ctx, watermarkMt5, store, window)
				if errMt5 != nil {
					slog.Error("failed process mt5", "error", errMt5)
				}
				if newWatermarkMt5 != nil {
					watermarkMt5 = newWatermarkMt5
				}

				if len(mt4Trades) > 0 || len(mt5Trades) > 0 {
					merged := append(mt4Trades, mt5Trades...)
					fanout.Dispatch(ctx, merged)
				}

				if errMt4 == nil && errMt5 == nil {
					lastHeartBeat = heartbeat(ctx, store, lastHeartBeat)
				}
			case <-tickerCleaningWindow.C:
				count := window.removeExpired(time.Now())
				slog.Info("removing expired trades from windows done", "trade_removed", count)
			case <-tickerReconcile.C:
				missing, extra, err := processReconcile(ctx, window, store)
				if err != nil {
					slog.Error("reconcile raised an error", "error", err)
					continue
				}

				if len(missing) > 0 || len(extra) > 0 {
					slog.Warn("reconcile ran and found inconsistency", "missing", missing, "extra", extra)
				} else {
					slog.Info("reconcile found nothing")
				}
			case <-ctx.Done():
				slog.Info("producer shutting down")
				return
			}
		}
	}()
}

func processMt4(ctx context.Context, watermark *time.Time, store *database.Store, window *ClosedTradeWindow) ([]model.ClosedTrade, *time.Time, error) {
	trades, err := store.GetMt4TradesFromPeriod(ctx, watermark, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get new mt4 trades, error: %w", err)
	}

	if len(trades) == 0 {
		return nil, nil, nil
	}

	slog.Debug("new trades found", "trades", trades)

	var newTrades []model.ClosedTrade
	for _, t := range trades {
		normalizedTrade := normalizeMt4(t)

		isNew := window.add(normalizedTrade)
		if isNew {
			newTrades = append(newTrades, normalizedTrade)

			slog.Debug("trade normalized and added to rolling window", "closed_trade", normalizedTrade)
		}
	}

	// updating watermark
	last := trades[len(trades)-1]
	watermark = &last.CloseTime.Time

	slog.Debug("mt4 watermark updated", "watermark", watermark)
	return newTrades, watermark, nil
}

func normalizeMt4(t database.Mt4Trade) model.ClosedTrade {
	var ct model.ClosedTrade
	ct.ServerId = model.Mt4Live1
	ct.Ticket = t.Ticket
	ct.Account = t.Account
	ct.Symbol = t.Symbol
	ct.Lot = float64(t.Volume) / volumeToLotMt4
	ct.OpenTime = t.OpenTime
	ct.CloseTime = t.CloseTime.Time
	ct.OpenPrice = int64(math.Round(t.OpenPrice * priceDecimal))
	ct.ClosePrice = int64(math.Round(t.ClosePrice.Float64 * priceDecimal))
	ct.Profit = int64(math.Round(t.Profit.Float64 * profitDecimal))

	return ct
}

func processMt5(ctx context.Context, watermark *int, store *database.Store, window *ClosedTradeWindow) ([]model.ClosedTrade, *int, error) {
	deals, err := store.GetClosedMt5DealsFromDeal(ctx, watermark)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get new mt5 deals, error: %w", err)
	}

	if len(deals) == 0 {
		return nil, nil, nil
	}

	slog.Debug("new deals found", "deals", deals)

	var newTrades []model.ClosedTrade
	for _, d := range deals {
		normalizedDeal := normalizeMt5(d)

		isNew := window.add(normalizedDeal)
		if isNew {
			newTrades = append(newTrades, normalizedDeal)

			slog.Debug("deal normalized and added to rolling window", "closed_trade", normalizedDeal)
		}
	}

	// updating watermark
	last := deals[len(deals)-1]
	watermark = &last.CloseDeal

	slog.Debug("mt5 watermark updated", "watermark", watermark)
	return newTrades, watermark, nil
}

func normalizeMt5(d database.ClosedMt5Deal) model.ClosedTrade {
	var ct model.ClosedTrade
	ct.ServerId = model.Mt5Live1
	ct.Ticket = d.CloseDeal
	ct.Account = d.Login
	ct.Symbol = d.Symbol
	ct.Lot = float64(d.Volume) / volumeToLotMt5
	ct.OpenTime = d.OpenTime
	ct.CloseTime = d.CloseTime
	ct.OpenPrice = int64(math.Round(d.OpenPrice * priceDecimal))
	ct.ClosePrice = int64(math.Round(d.ClosePrice * priceDecimal))
	ct.Profit = int64(math.Round(d.Profit * profitDecimal))

	return ct
}

func processReconcile(ctx context.Context, window *ClosedTradeWindow, store *database.Store) ([]tradeKey, []tradeKey, error) {
	now := time.Now()
	toTime := now.Add(-2 * time.Second)
	fromTime := now.Add(-window.duration).Add(10 * time.Second)
	mt4ClosedTrades, err := store.GetMt4TradesFromPeriod(ctx, &fromTime, &toTime)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting the mt4 closed trades, err: %w", err)
	}

	mt5ClosedTrades, err := store.GetClosedMt5DealsFromPeriod(ctx, &fromTime, &toTime)
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting the mt5 closed deals, err: %w", err)
	}

	windowSnapshot := window.Snapshot(&fromTime, &toTime)

	missing, extra := verifyWindow(mt4ClosedTrades, mt5ClosedTrades, windowSnapshot)

	return missing, extra, nil
}

func verifyWindow(mt4Trades []database.Mt4Trade, mt5Deals []database.ClosedMt5Deal, snap map[tradeKey]bool) ([]tradeKey, []tradeKey) {
	var missing, extra []tradeKey
	setMt4Trades := make(map[tradeKey]bool)
	setMt5Deals := make(map[tradeKey]bool)

	for _, mt4T := range mt4Trades {
		tk := tradeKey{server: model.Mt4Live1, ticket: mt4T.Ticket}
		if _, ok := snap[tk]; !ok {
			missing = append(missing, tk)
		}
		setMt4Trades[tk] = true
	}

	for _, mt5D := range mt5Deals {
		tk := tradeKey{server: model.Mt5Live1, ticket: mt5D.CloseDeal}
		if _, ok := snap[tk]; !ok {
			missing = append(missing, tk)
		}
		setMt5Deals[tk] = true
	}

	for key := range snap {
		if _, ok := setMt4Trades[key]; ok {
			continue
		}
		if _, ok := setMt5Deals[key]; ok {
			continue
		}
		extra = append(extra, key)
	}

	return missing, extra
}

func heartbeat(ctx context.Context, store *database.Store, lastHeartBeat time.Time) time.Time {
	now := time.Now()
	if now.Sub(lastHeartBeat) > heartBeatInterval {
		err := store.UpsertSystemState(ctx, database.Heartbeat, now.Format(time.RFC3339))
		if err != nil {
			slog.Error("heartbeat failed to update", "error", err)
			return lastHeartBeat
		}
		slog.Info("heartbeat state updated")
		return now
	}
	return lastHeartBeat
}
