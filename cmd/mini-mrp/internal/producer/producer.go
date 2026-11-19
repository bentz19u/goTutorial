package producer

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
)

type Server string

const (
	Mt4Live1 Server = "mt4_live1"
	Mt5Live1 Server = "mt5_live1"
)

const (
	volumeToLotMt4 = 100
	volumeToLotMt5 = 10000
	priceDecimal   = 100000
	profitDecimal  = 100
)

type ClosedTrade struct {
	ServerId   Server
	Ticket     int
	Account    int
	Symbol     string
	Lot        float64
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  int64
	ClosePrice int64
	Profit     int64
}

func StartProducer(ctx context.Context, window *ClosedTradeWindow, store *database.Store, wg *sync.WaitGroup) {
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

		for {
			select {
			case <-ticker.C:
				newWatermarkMt4 := processMt4(ctx, watermarkMt4, store, window)
				if newWatermarkMt4 != nil {
					watermarkMt4 = newWatermarkMt4
				}
				newWatermarkMt5 := processMt5(ctx, watermarkMt5, store, window)
				if newWatermarkMt5 != nil {
					watermarkMt5 = newWatermarkMt5
				}
			case <-tickerCleaningWindow.C:
				count := window.removeExpired(time.Now())
				slog.Info("removing expired trades from windows done", "trade_removed", count)
			case <-ctx.Done():
				slog.Info("producer shutting down")
				return
			}
		}
	}()
}

func processMt4(ctx context.Context, watermark *time.Time, store *database.Store, window *ClosedTradeWindow) *time.Time {
	trades, err := store.GetMt4TradesFromCloseTime(ctx, watermark)
	if err != nil {
		slog.Error("failed to get new mt4 trades", "error", err)
		return nil
	}

	if len(trades) == 0 {
		return nil
	}

	slog.Info("new trades found", "trades", trades)

	for _, t := range trades {
		normalizedTrade := normalizeMt4(t)

		window.add(normalizedTrade)

		slog.Info("trade normalized and added to rolling window", "closed_trade", normalizedTrade)
	}

	// updating watermark
	last := trades[len(trades)-1]
	watermark = &last.CloseTime.Time

	slog.Info("mt4 watermark updated", "watermark", watermark)
	return watermark
}

func normalizeMt4(t database.Mt4Trade) ClosedTrade {
	var ct ClosedTrade
	ct.ServerId = Mt4Live1
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

func processMt5(ctx context.Context, watermark *int, store *database.Store, window *ClosedTradeWindow) *int {
	deals, err := store.GetClosedMt5DealsFromDeal(ctx, watermark)
	if err != nil {
		slog.Error("failed to get new mt5 deals", "error", err)
		return nil
	}

	if len(deals) == 0 {
		return nil
	}

	slog.Info("new deals found", "deals", deals)

	for _, d := range deals {
		normalizedDeal := normalizeMt5(d)

		window.add(normalizedDeal)

		slog.Info("deal normalized and added to rolling window", "closed_trade", normalizedDeal)
	}

	// updating watermark
	last := deals[len(deals)-1]
	watermark = &last.CloseDeal

	slog.Info("mt5 watermark updated", "watermark", watermark)
	return watermark
}

func normalizeMt5(d database.ClosedMt5Deal) ClosedTrade {
	var ct ClosedTrade
	ct.ServerId = Mt5Live1
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
