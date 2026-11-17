package producer

import (
	"log/slog"
	"time"
)

type tradeKey struct {
	server Server
	ticket int
}

type ClosedTradeWindow struct {
	duration time.Duration
	window   map[tradeKey]ClosedTrade
}

func (w *ClosedTradeWindow) add(ct ClosedTrade) {
	key := tradeKey{server: ct.ServerId, ticket: ct.Ticket}
	w.window[key] = ct
}

func (w *ClosedTradeWindow) removeExpired(fromTime time.Time) int {
	beforeTime := fromTime.Add(-w.duration)
	removed := 0
	for key, ct := range w.window {
		if ct.CloseTime.Before(beforeTime) {
			slog.Info("closed trade has been removed from window", "closed_trade", ct)
			delete(w.window, key)
			removed += 1
		}
	}
	return removed
}
