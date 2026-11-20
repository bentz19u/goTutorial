package consumer

import (
	"cmp"
	"slices"
	"testing"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

func TestSumProfit(t *testing.T) {
	var cts []model.ClosedTrade
	cts = append(cts, model.ClosedTrade{Profit: 125})
	cts = append(cts, model.ClosedTrade{Profit: -201})
	cts = append(cts, model.ClosedTrade{Profit: -405})
	cts = append(cts, model.ClosedTrade{Profit: -291})
	got := sumProfit(cts)

	if got != -7.72 {
		t.Errorf("Wrong sum profit: got %f, want %f", got, -7.72)
	}
}

func TestExtractNewTrades(t *testing.T) {
	tests := map[string]struct {
		processedTrades    map[database.ProcessedTrade]bool
		closedTrades       []model.ClosedTrade
		wantTrades         []database.ProcessedTrade
		wantAccountProfits []database.AccountProfit
	}{
		"All new": {
			processedTrades: map[database.ProcessedTrade]bool{},
			closedTrades: []model.ClosedTrade{
				{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Profit: 100},
				{ServerId: model.Mt4Live1, Ticket: 2, Account: 2, Profit: -50},
			},
			wantTrades: []database.ProcessedTrade{
				{Server: model.Mt4Live1, Ticket: 1},
				{Server: model.Mt4Live1, Ticket: 2},
			},
			wantAccountProfits: []database.AccountProfit{
				{Server: model.Mt4Live1, Login: 1, ProfitCents: 100},
				{Server: model.Mt4Live1, Login: 2, ProfitCents: -50},
			},
		},
		"Some already processed": {
			processedTrades: map[database.ProcessedTrade]bool{
				{Server: model.Mt4Live1, Ticket: 1}: true,
			},
			closedTrades: []model.ClosedTrade{
				{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Profit: 100},
				{ServerId: model.Mt4Live1, Ticket: 2, Account: 1, Profit: 250},
			},
			wantTrades: []database.ProcessedTrade{
				{Server: model.Mt4Live1, Ticket: 2},
			},
			wantAccountProfits: []database.AccountProfit{
				{Server: model.Mt4Live1, Login: 1, ProfitCents: 250},
			},
		},
		"All already processed": {
			processedTrades: map[database.ProcessedTrade]bool{
				{Server: model.Mt4Live1, Ticket: 1}: true,
				{Server: model.Mt5Live1, Ticket: 1}: true,
			},
			closedTrades: []model.ClosedTrade{
				{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Profit: 100},
				{ServerId: model.Mt5Live1, Ticket: 1, Account: 1, Profit: 200},
			},
		},
		"Same account is summed into one row": {
			processedTrades: map[database.ProcessedTrade]bool{},
			closedTrades: []model.ClosedTrade{
				{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Profit: 100},
				{ServerId: model.Mt4Live1, Ticket: 2, Account: 1, Profit: -30},
				{ServerId: model.Mt4Live1, Ticket: 3, Account: 1, Profit: 5},
			},
			wantTrades: []database.ProcessedTrade{
				{Server: model.Mt4Live1, Ticket: 1},
				{Server: model.Mt4Live1, Ticket: 2},
				{Server: model.Mt4Live1, Ticket: 3},
			},
			wantAccountProfits: []database.AccountProfit{
				{Server: model.Mt4Live1, Login: 1, ProfitCents: 75},
			},
		},
		"Same account and ticket on two servers stay separate": {
			processedTrades: map[database.ProcessedTrade]bool{},
			closedTrades: []model.ClosedTrade{
				{ServerId: model.Mt4Live1, Ticket: 1, Account: 1, Profit: 100},
				{ServerId: model.Mt5Live1, Ticket: 1, Account: 1, Profit: 200},
			},
			wantTrades: []database.ProcessedTrade{
				{Server: model.Mt4Live1, Ticket: 1},
				{Server: model.Mt5Live1, Ticket: 1},
			},
			wantAccountProfits: []database.AccountProfit{
				{Server: model.Mt4Live1, Login: 1, ProfitCents: 100},
				{Server: model.Mt5Live1, Login: 1, ProfitCents: 200},
			},
		},
		"Empty window": {
			processedTrades: map[database.ProcessedTrade]bool{
				{Server: model.Mt4Live1, Ticket: 1}: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotTrades, gotAccountProfits := extractNewTrades(test.processedTrades, test.closedTrades)

			// account profits are built from a map, so their order changes between runs:
			// sort them before comparing (the want literals above are written in the same order)
			slices.SortFunc(gotAccountProfits, func(a, b database.AccountProfit) int {
				return cmp.Or(cmp.Compare(a.Server, b.Server), cmp.Compare(a.Login, b.Login))
			})

			if !slices.Equal(gotTrades, test.wantTrades) {
				t.Errorf("Wrong trades to mark as processed: got %v, want %v", gotTrades, test.wantTrades)
			}

			if !slices.Equal(gotAccountProfits, test.wantAccountProfits) {
				t.Errorf("Wrong account profits: got %v, want %v", gotAccountProfits, test.wantAccountProfits)
			}
		})
	}
}
