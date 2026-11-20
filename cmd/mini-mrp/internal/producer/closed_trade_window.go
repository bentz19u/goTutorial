package producer

import (
	"encoding/json"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type tradeKey struct {
	server model.Server
	ticket int
}

func (k tradeKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Server model.Server `json:"server"`
		Ticket int          `json:"ticket"`
	}{k.server, k.ticket})
}

type ClosedTradeWindow struct {
	duration time.Duration
	mu       sync.Mutex
	window   map[tradeKey]model.ClosedTrade
}

func NewClosedTradeWindow(d time.Duration) *ClosedTradeWindow {
	return &ClosedTradeWindow{
		duration: d,
		window:   make(map[tradeKey]model.ClosedTrade),
	}
}

func (w *ClosedTradeWindow) add(ct model.ClosedTrade) {
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
			slog.Debug("closed trade has been removed from window", "closed_trade", ct)
			delete(w.window, key)
			removed += 1
		}
	}

	return removed
}

func (w *ClosedTradeWindow) GetAllClosedTradesSorted() []model.ClosedTrade {
	w.mu.Lock()
	defer w.mu.Unlock()

	trades := slices.Collect(maps.Values(w.window))
	slices.SortFunc(trades, func(a, b model.ClosedTrade) int {
		return b.CloseTime.Compare(a.CloseTime)
	})
	return trades
}

func (w *ClosedTradeWindow) Snapshot(fromDate *time.Time, toDate *time.Time) map[tradeKey]bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	keys := make(map[tradeKey]bool, len(w.window))
	for k, ct := range w.window {
		if ct.CloseTime.Before(*fromDate) {
			continue
		}
		if !ct.CloseTime.Before(*toDate) {
			continue
		}
		keys[k] = true
	}
	return keys
}
