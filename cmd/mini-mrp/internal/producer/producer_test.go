package producer

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
)

func TestNormalizeMt4(t *testing.T) {
	closeTime := time.Now()
	openTime := closeTime.Add(-10 * time.Second)

	tests := map[string]struct {
		trade database.Mt4Trade
		want  ClosedTrade
	}{"Base": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.54342, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: 2.56, Valid: true},
		},
		ClosedTrade{ServerId: Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 154342, ClosePrice: 155123, Profit: 256,
		},
	}, "Narrow decimal": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.00001, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: 2.56, Valid: true},
		},
		ClosedTrade{ServerId: Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 100001, ClosePrice: 155123, Profit: 256,
		},
	}, "Negative profit": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 1.54342, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: -0.5359, Valid: true},
		},
		ClosedTrade{ServerId: Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
			OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 154342, ClosePrice: 155123, Profit: -54,
		},
	}, "decimal open price": {
		database.Mt4Trade{Ticket: 1, Account: 1, Symbol: "USDJPY", Volume: 1000,
			OpenTime: openTime, CloseTime: sql.NullTime{Time: closeTime, Valid: true},
			OpenPrice: 0.29, ClosePrice: sql.NullFloat64{Float64: 1.55123, Valid: true},
			Profit: sql.NullFloat64{Float64: -0.5359, Valid: true},
		},
		ClosedTrade{ServerId: Mt4Live1, Ticket: 1, Account: 1, Symbol: "USDJPY", Lot: 10,
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
		want ClosedTrade
	}{"Base": {
		database.ClosedMt5Deal{OpenDeal: 1, CloseDeal: 2, PositionID: 3,
			Login: 1, Symbol: "USDJPY",
			Volume: 1000000, OpenTime: openTime, CloseTime: closeTime,
			OpenPrice: 1.54342, ClosePrice: 1.55123,
			Profit: 2.56,
		},
		ClosedTrade{ServerId: Mt5Live1, Ticket: 2, Account: 1, Symbol: "USDJPY", Lot: 100,
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
