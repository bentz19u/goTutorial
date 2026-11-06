package generator

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
)

var (
	symbols     = []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD"}
	volumes     = []int{100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000}
	mt4Accounts = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	mt5Accounts = []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
)

type tradesIds struct {
	ticket   int
	deal     int
	position int
}

func StartGenerator(ctx context.Context, store *database.Store, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		// TODO: get the trades/deals to close from DB instead on purging them on boot
		mt4TradeToClose := make(map[int]database.Mt4Trade)
		mt5DealToClose := make(map[int]database.Mt5Deal)
		ids := tradesIds{}

		for {
			select {
			case <-ticker.C:
				generate(ctx, store, mt4TradeToClose, mt5DealToClose, &ids)
			case <-ctx.Done():
				slog.Info("generator shutting down")
				return
			}
		}
	}()
}

// the idea will be to either generate a trade, or close one randomly
func generate(ctx context.Context, store *database.Store, mt4TradeToClose map[int]database.Mt4Trade,
	mt5DealToClose map[int]database.Mt5Deal, ids *tradesIds) {

	for {
		switch rand.IntN(4) {
		case 0:
			addTrade(ctx, store, mt4TradeToClose, ids)
			return
		case 1:
			addDeal(ctx, store, mt5DealToClose, ids)
			return
		case 2:
			// nothing to update, we reloop
			if len(mt4TradeToClose) == 0 {
				break
			}
			closeTrade(ctx, store, mt4TradeToClose)
			return
		case 3:
			// nothing to update, we reloop
			if len(mt5DealToClose) == 0 {
				break
			}
			closeDeal(ctx, store, mt5DealToClose, ids)
			return
		}
	}
}

func addTrade(ctx context.Context, s *database.Store, mt4TradeToClose map[int]database.Mt4Trade, ids *tradesIds) {
	// add a mt4 trade
	mt := database.Mt4Trade{
		Ticket:    ids.ticket,
		Account:   randPick(mt4Accounts),
		Symbol:    randPick(symbols),
		Volume:    randPick(volumes),
		OpenTime:  time.Now(),
		OpenPrice: randPrice(),
	}

	err := s.UpsertMt4Trade(ctx, mt)
	if err != nil {
		slog.Error("failed to insert mt4 trades", "mt4_trade", mt, "error", err)
		return
	}

	// we keep it in memory to close it later
	mt4TradeToClose[mt.Ticket] = mt
	ids.ticket++

	slog.Info("inserted to mt4 trades", "mt4_trade", mt)
}

func addDeal(ctx context.Context, s *database.Store, mt5DealToClose map[int]database.Mt5Deal, ids *tradesIds) {
	// add a mt5 deal, volume is *100 to match the table lot volume ratio
	md := database.Mt5Deal{
		Deal:       ids.deal,
		PositionID: ids.position,
		Entry:      0,
		Login:      randPick(mt5Accounts),
		Symbol:     randPick(symbols),
		Volume:     randPick(volumes) * 100,
		Time:       time.Now(),
		Price:      randPrice(),
	}

	err := s.AddMt5Deal(ctx, md)
	if err != nil {
		slog.Error("failed to insert mt5 deal", "mt5_deal", md, "error", err)
		return
	}

	// we keep it in memory to close it later
	mt5DealToClose[md.Deal] = md
	ids.deal++
	ids.position++

	slog.Info("inserted to mt5 deals", "mt5_deal", md)
}

func closeTrade(ctx context.Context, s *database.Store, mt4TradeToClose map[int]database.Mt4Trade) {
	// we random pick a trade
	ticket := randPick(slices.Collect(maps.Keys(mt4TradeToClose)))
	trade := mt4TradeToClose[ticket]

	// close a pending trade
	trade.CloseTime = sql.NullTime{Time: time.Now(), Valid: true}
	closePrice := randPrice()
	trade.ClosePrice = sql.NullFloat64{Float64: closePrice, Valid: true}
	trade.Profit = sql.NullFloat64{Float64: closePrice - trade.OpenPrice, Valid: true}

	err := s.UpsertMt4Trade(ctx, trade)
	if err != nil {
		slog.Error("failed to close mt4 trades", "mt4_trade", trade, "error", err)
		return
	}

	delete(mt4TradeToClose, ticket)

	slog.Info("updated mt4 trade", "mt4_trade", trade)
}

func closeDeal(ctx context.Context, s *database.Store, mt5DealToClose map[int]database.Mt5Deal, ids *tradesIds) {
	// we random pick a deal
	deal := randPick(slices.Collect(maps.Keys(mt5DealToClose)))
	openDeal := mt5DealToClose[deal]

	// close a pending deal
	toInsert := database.Mt5Deal{
		Deal:       ids.deal,
		PositionID: openDeal.PositionID,
		Entry:      1,
		Login:      openDeal.Login,
		Symbol:     openDeal.Symbol,
		Volume:     openDeal.Volume,
		Time:       time.Now(),
		Price:      randPrice(),
	}
	toInsert.Profit = toInsert.Price - openDeal.Price

	err := s.AddMt5Deal(ctx, toInsert)
	if err != nil {
		slog.Error("failed to insert closing mt5 deal", "open_deal", openDeal, "close_deal", toInsert, "error", err)
		return
	}

	delete(mt5DealToClose, deal)
	ids.deal++

	slog.Info("updated mt5 deal", "open_deal", openDeal, "close_deal", toInsert)
}
