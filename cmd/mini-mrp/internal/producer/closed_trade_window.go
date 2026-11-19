package producer

import (
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"
)

type tradeKey struct {
	server Server
	ticket int
}

type ClosedTradeWindow struct {
	duration time.Duration
	mu       sync.Mutex
	window   map[tradeKey]ClosedTrade
}

func NewClosedTradeWindow(d time.Duration) *ClosedTradeWindow {
	return &ClosedTradeWindow{
		duration: d,
		window:   make(map[tradeKey]ClosedTrade),
	}
}

func (w *ClosedTradeWindow) add(ct ClosedTrade) {
	key := tradeKey{server: ct.ServerId, ticket: ct.Ticket}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.window[key] = ct
}

func (w *ClosedTradeWindow) removeExpired(fromTime time.Time) int {
	beforeTime := fromTime.Add(-w.duration)
	removed := 0

	w.mu.Lock()
	defer w.mu.Unlock()

	for key, ct := range w.window {
		if ct.CloseTime.Before(beforeTime) {
			slog.Info("closed trade has been removed from window", "closed_trade", ct)
			delete(w.window, key)
			removed += 1
		}
	}

	return removed
}

func (w *ClosedTradeWindow) GetAllClosedTradesSorted() []ClosedTrade {
	w.mu.Lock()
	defer w.mu.Unlock()

	trades := slices.Collect(maps.Values(w.window))
	slices.SortFunc(trades, func(a, b ClosedTrade) int {
		return b.CloseTime.Compare(a.CloseTime)
	})
	return trades
}
