package producer

import (
	"testing"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type dummyTrade struct {
	tk        tradeKey
	closeTime time.Time
}

func TestAdd(t *testing.T) {
	window := map[tradeKey]model.ClosedTrade{}
	ctw := ClosedTradeWindow{duration: 1 * time.Minute, window: window}

	ct := model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY",
		Lot: 1.00, OpenTime: time.Now(), CloseTime: time.Now(), OpenPrice: 1, ClosePrice: 1, Profit: 1}
	ctw.add(ct)

	if len(ctw.window) != 1 {
		t.Errorf("Wrong amount of trades remaining in window: got %d, want %d", len(ctw.window), 1)
	}

	// testing idempotency, we should keep the same amount of trades
	ctw.add(ct)
	if len(ctw.window) != 1 {
		t.Errorf("Wrong amount of trades remaining in window: got %d, want %d", len(ctw.window), 1)
	}
}

func TestRemoveExpired(t *testing.T) {
	now := time.Now()
	defaultTrade := model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY",
		Lot: 1.00, OpenTime: now, CloseTime: now, OpenPrice: 1, ClosePrice: 1, Profit: 1}
	tradesKeys := []dummyTrade{
		{tk: tradeKey{server: model.Mt4Live1, ticket: 1}, closeTime: now.Add(-90 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 2}, closeTime: now.Add(-80 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 3}, closeTime: now.Add(-70 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 4}, closeTime: now.Add(-60 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 5}, closeTime: now.Add(-50 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 6}, closeTime: now.Add(-40 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 7}, closeTime: now.Add(-30 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 8}, closeTime: now.Add(-20 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 9}, closeTime: now.Add(-10 * time.Second)},
		{tk: tradeKey{server: model.Mt4Live1, ticket: 10}, closeTime: now},
	}

	window := map[tradeKey]model.ClosedTrade{}
	for _, tk := range tradesKeys {
		trade := defaultTrade
		trade.ServerId = tk.tk.server
		trade.Ticket = tk.tk.ticket
		trade.CloseTime = tk.closeTime
		window[tk.tk] = trade
	}
	ctw := ClosedTradeWindow{duration: 1 * time.Minute, window: window}

	removed := ctw.removeExpired(now)

	if removed != 3 {
		t.Errorf("Wrong amount of removed trades: got %d, want %d", removed, 3)
	}

	if len(ctw.window) != 7 {
		t.Errorf("Wrong amount of trades remaining in window: got %d, want %d", len(ctw.window), 7)
	}

	if _, ok := ctw.window[tradeKey{server: model.Mt4Live1, ticket: 3}]; ok {
		t.Errorf("Trade should have been removed: server %v, ticket %v", model.Mt4Live1, 3)
	}

	if _, ok := ctw.window[tradeKey{server: model.Mt4Live1, ticket: 4}]; !ok {
		t.Errorf("Trade should still be in the window: server %v, ticket %v", model.Mt4Live1, 4)
	}
}
