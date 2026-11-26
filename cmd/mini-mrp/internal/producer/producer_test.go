package producer

import (
	"cmp"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/bentz19u/go-trade-processor/cmd/mini-mrp/internal/database"
	"github.com/bentz19u/go-trade-processor/cmd/mini-mrp/internal/model"
)

func TestNormalizeMt4(t *testing.T) {
	closeTime := time.Now()
	openTime := closeTime.Add(-10 * time.Second)

	tests := map[string]struct {
		trade database.Mt4Trade
		want  model.ClosedTrade
	}{"Base": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.54342, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: 2.56, Valid: true},
		},
		model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 154342, ClosePrice: 155123, Profit: 256,
		},
	}, "Narrow decimal": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.00001, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: 2.56, Valid: true},
		},
		model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 100001, ClosePrice: 155123, Profit: 256,
		},
	}, "Negative profit": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.54342, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: -0.5359, Valid: true},
		},
		model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 154342, ClosePrice: 155123, Profit: -54,
		},
	}, "decimal open price": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 0.29, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: -0.5359, Valid: true},
		},
		model.ClosedTrade{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 29000, ClosePrice: 155123, Profit: -54,
		},
	}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := normalizeMt4(test.trade)

			if test.want != got {
				t.Errorf("Failed closed trade: got %v, want %v", got, test.want)
			}
		})
	}
}

func TestNormalizeMt5(t *testing.T) {
	closeTime := time.Now()
	openTime := closeTime.Add(-10 * time.Second)

	tests := map[string]struct {
		deal database.ClosedMt5Deal
		want model.ClosedTrade
	}{"Base": {
		database.ClosedMt5Deal{OpenDeal: 1, CloseDeal: 2, PositionID: 3,
			Login: 1, Symbol: "USDJPY",
			Volume: 1000000, OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 1.54342, ClosePrice: 1.55123,
			Profit: 2.56,
		},
		model.ClosedTrade{ServerId: model.Mt5Live1, Ticket: 2, Account: 1, Symbol: "USDJPY", Lot: 100,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 154342, ClosePrice: 155123, Profit: 256,
		},
	}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := normalizeMt5(test.deal)

			if test.want != got {
				t.Errorf("Failed closed trade: got %v, want %v", got, test.want)
			}
		})
	}
}

func TestVerifyWindow(t *testing.T) {
	tests := map[string]struct {
		mt4Trades   []database.Mt4Trade
		mt5Deals    []database.ClosedMt5Deal
		snapshot    map[tradeKey]bool
		wantMissing []tradeKey
		wantExtra   []tradeKey
	}{
		"All good, nothing to report": {
			mt4Trades: []database.Mt4Trade{{Ticket: 1}, {Ticket: 2}},
			mt5Deals:  []database.ClosedMt5Deal{{CloseDeal: 1}, {CloseDeal: 2}},
			snapshot: map[tradeKey]bool{
				{server: model.Mt4Live1, ticket: 1}: true,
				{server: model.Mt4Live1, ticket: 2}: true,
				{server: model.Mt5Live1, ticket: 1}: true,
				{server: model.Mt5Live1, ticket: 2}: true,
			},
		},
		"One mt4 and one mt5 are missing": {
			mt4Trades: []database.Mt4Trade{{Ticket: 1}, {Ticket: 2}},
			mt5Deals:  []database.ClosedMt5Deal{{CloseDeal: 1}, {CloseDeal: 2}},
			snapshot: map[tradeKey]bool{
				{server: model.Mt4Live1, ticket: 1}: true,
				{server: model.Mt5Live1, ticket: 1}: true,
			},
			wantMissing: []tradeKey{
				{server: model.Mt4Live1, ticket: 2},
				{server: model.Mt5Live1, ticket: 2},
			},
		},
		"Two trades are in the window but not in the source": {
			mt4Trades: []database.Mt4Trade{{Ticket: 1}},
			mt5Deals:  []database.ClosedMt5Deal{{CloseDeal: 1}},
			snapshot: map[tradeKey]bool{
				{server: model.Mt4Live1, ticket: 1}: true,
				{server: model.Mt4Live1, ticket: 2}: true,
				{server: model.Mt5Live1, ticket: 1}: true,
				{server: model.Mt5Live1, ticket: 2}: true,
			},
			wantExtra: []tradeKey{
				{server: model.Mt4Live1, ticket: 2},
				{server: model.Mt5Live1, ticket: 2},
			},
		},
	}

	// extra is built by ranging over a map, so its order changes between runs:
	// sort both results before comparing (the want literals above are written in the same order)
	sortKeys := func(keys []tradeKey) {
		slices.SortFunc(keys, func(a, b tradeKey) int {
			return cmp.Or(cmp.Compare(a.server, b.server), cmp.Compare(a.ticket, b.ticket))
		})
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotMissing, gotExtra := verifyWindow(test.mt4Trades, test.mt5Deals, test.snapshot)
			sortKeys(gotMissing)
			sortKeys(gotExtra)

			if !slices.Equal(gotMissing, test.wantMissing) {
				t.Errorf("Wrong missing trades: got %v, want %v", gotMissing, test.wantMissing)
			}

			if !slices.Equal(gotExtra, test.wantExtra) {
				t.Errorf("Wrong extra trades: got %v, want %v", gotExtra, test.wantExtra)
			}
		})
	}
}
