package main

import (
	"context"
	"database/sql"
	"log/slog"
	"maps"
	"math/rand/v2"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"

	_ "github.com/go-sql-driver/mysql"
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

type tradesIds struct {
	ticket   int
	deal     int
	position int
}

// the idea will be to either generate a trade, or close one randomly
func generator(ctx context.Context, s *database.Store, mt4TradeToClose map[int]database.Mt4Trade, mt5DealToClose map[int]database.Mt5Deal, ids *tradesIds) {

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD"}
	volumes := []int{100, 200, 300, 400, 500, 1000, 2000, 3000, 4000, 5000}
	mt4Accounts := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	mt5Accounts := []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	for {
		switch rand.IntN(4) {
		case 0:
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
			return
		case 1:
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
			return
		case 2:
			// nothing to update, we reloop
			if len(mt4TradeToClose) == 0 {
				break
			}

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
			return
		case 3:
			// nothing to update, we reloop
			if len(mt5DealToClose) == 0 {
				break
			}

			// we random pick a deal
			deal := randPick(slices.Collect(maps.Keys(mt5DealToClose)))
			openDeal := mt5DealToClose[deal]

			// close a pending deal
			closeDeal := database.Mt5Deal{
				Deal:       ids.deal,
				PositionID: openDeal.PositionID,
				Entry:      1,
				Login:      openDeal.Login,
				Symbol:     openDeal.Symbol,
				Volume:     openDeal.Volume,
				Time:       time.Now(),
				Price:      randPrice(),
			}
			closeDeal.Profit = closeDeal.Price - openDeal.Price

			err := s.AddMt5Deal(ctx, closeDeal)
			if err != nil {
				slog.Error("failed to insert closing mt5 deal", "open_deal", openDeal, "close_deal", closeDeal, "error", err)
				return
			}

			delete(mt5DealToClose, deal)
			ids.deal++

			slog.Info("updated mt5 deal", "openDeal", openDeal, "close_deal", closeDeal)
			return
		}
	}
}

func main() {
	initLogger()

	ctx := context.Background()

	store, errDb := database.InitDatabase(ctx)
	if errDb != nil {
		slog.Error("connection to database failed", "error", errDb)
		os.Exit(1)
	}
	defer store.Db.Close()

	errTables := database.InitTables(ctx, store)
	if errTables != nil {
		slog.Error("initTables failed", "error", errTables)
		os.Exit(1)
	}

	errPurge := database.PurgeTables(ctx, store)
	if errPurge != nil {
		slog.Error("purge table failed", "error", errPurge)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// generator, it will regularly add new orders to one of the DB, or close one existing one
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		// TODO: get the trades/deals to close from DB instead on boot
		// or we can decide to dump the tables content to start from scratch
		mt4TradeToClose := make(map[int]database.Mt4Trade)
		mt5DealToClose := make(map[int]database.Mt5Deal)

		ids := tradesIds{}

		for {
			select {
			case <-ticker.C:
				generator(ctx, store, mt4TradeToClose, mt5DealToClose, &ids)
			case <-ctx.Done():
				slog.Info("generator closing down")
				return
			}
		}
	}()

	time.Sleep(100 * time.Second)
}
