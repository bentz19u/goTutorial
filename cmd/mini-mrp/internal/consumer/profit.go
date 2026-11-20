package consumer

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

func StartProfitConsumer(ctx context.Context, window Window, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				closedTrades := window.GetAllClosedTradesSorted()
				profits := sumProfit(closedTrades)
				slog.Info("window sum of profit", "profits", profits)
			case <-ctx.Done():
				slog.Info("profit consumer shutting down")
				return
			}
		}
	}()
}

func sumProfit(cts []model.ClosedTrade) float64 {
	var sum int64
	for _, ct := range cts {
		sum += ct.Profit
	}
	return float64(sum) / 100
}

func StartProfitPerAccountConsumer(ctx context.Context, store *database.Store, window Window, wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				closedTrades := window.GetAllClosedTradesSorted()
				processedTrades, err := processNewTrades(ctx, store, closedTrades)
				if err != nil {
					slog.Error("profit per account error", "error", err)
				} else {
					slog.Info("profit per account", "trades_count", len(processedTrades))
				}
			case <-ctx.Done():
				slog.Info("profit per account consumer shutting down")
				return
			}
		}
	}()
}

func processNewTrades(ctx context.Context, store *database.Store, cts []model.ClosedTrade) ([]database.ProcessedTrade, error) {
	// first get the processed trades already, it would not scale at it is because we never remove expired trades
	processedTrades, err := store.GetAllProcessedTrades(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching processed trades: %w", err)
	}

	// extract new information
	ptToInsert, apToInsert := extractNewTrades(processedTrades, cts)

	// then save new trades in DB
	err = saveBatch(ctx, store, ptToInsert, apToInsert)
	if err != nil {
		return nil, fmt.Errorf("error saving processed trades: %w", err)
	}

	return ptToInsert, nil
}

func extractNewTrades(processedTrades map[database.ProcessedTrade]bool, cts []model.ClosedTrade) ([]database.ProcessedTrade, []database.AccountProfit) {
	// process new trades and prepare the element to add/update in DB
	var ptToInsert []database.ProcessedTrade
	totals := make(map[database.AccountProfitKey]int64)
	for _, ct := range cts {
		if _, ok := processedTrades[database.ProcessedTrade{Server: ct.ServerId, Ticket: ct.Ticket}]; !ok {
			ptToInsert = append(ptToInsert, database.ProcessedTrade{Server: ct.ServerId, Ticket: ct.Ticket})
			totals[database.AccountProfitKey{Server: ct.ServerId, Login: ct.Account}] += ct.Profit
		}
	}

	// create the list of account profits to update
	apToInsert := make([]database.AccountProfit, 0, len(totals))
	for key, cents := range totals {
		apToInsert = append(apToInsert, database.AccountProfit{Server: key.Server, Login: key.Login, ProfitCents: cents})
	}

	return ptToInsert, apToInsert
}

func saveBatch(ctx context.Context, store *database.Store, pts []database.ProcessedTrade, aps []database.AccountProfit) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return store.WithTx(ctx, func(tx *sql.Tx) error {
		if err := store.InsertProcessedTrades(ctx, tx, pts); err != nil {
			return err
		}
		return store.UpsertAccountProfits(ctx, tx, aps)
	})
}
