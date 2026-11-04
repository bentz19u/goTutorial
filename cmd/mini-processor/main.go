package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/poller"
	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/source"

	_ "github.com/go-sql-driver/mysql"
)

type step struct {
	orderType source.EventStatus
	order     source.Order
}

func initLogger() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func initDatabase(ctx context.Context) (*sql.DB, error) {
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, errors.New("DB_PASSWORD not configured")
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/minimirror", dbPassword))
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	errPing := db.PingContext(ctx)
	if errPing != nil {
		return nil, fmt.Errorf("pinging database: %w", errPing)
	}

	slog.Info("ping to database successful")
	return db, nil
}

func initTable(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS orders (
		id INT PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		price DECIMAL(10, 2) NOT NULL
	)`)
	if err != nil {
		return err
	}
	return nil
}

// the goal is to test the poller and also be able to interrupt it
func main() {
	initLogger()

	ctx := context.Background()

	db, errDb := initDatabase(ctx)
	if errDb != nil {
		slog.Error("connection to database failure", "error", errDb)
		os.Exit(1)
	}
	defer db.Close()

	err := initTable(ctx, db)
	if err != nil {
		slog.Error("initTable failed", "error", err)
		os.Exit(1)
	}

	originalSource := source.NewFakeSource()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

	out := make(chan source.Event)

	// so event can be generated
	toExecute := []step{
		{source.PriceChanged, source.Order{ID: 1, Price: 2.99}},
		{source.AddedOrder, source.Order{ID: 3, Name: "Pizza1"}},
		{source.AddedOrder, source.Order{ID: 4, Name: "Pizza2"}},
		{source.PriceChanged, source.Order{ID: 4, Price: 3.99}},
		{source.RemovedOrder, source.Order{ID: 4}},
		{source.AddedOrder, source.Order{ID: 5, Name: "Pizza3"}},
		{source.AddedOrder, source.Order{ID: 6, Name: "Pizza4"}},
		{source.AddedOrder, source.Order{ID: 7, Name: "Pizza5"}},
		{source.AddedOrder, source.Order{ID: 8, Name: "Pizza6"}},
		{source.RemovedOrder, source.Order{ID: 6}},
		{source.AddedOrder, source.Order{ID: 9, Name: "Pizza7"}},
		{source.AddedOrder, source.Order{ID: 10, Name: "Pizza8"}},
		{source.RemovedOrder, source.Order{ID: 3}},
	}

	go func(toExecute []step) {
		for _, o := range toExecute {
			time.Sleep(300 * time.Millisecond)

			switch o.orderType {
			case source.AddedOrder:
				originalSource.Add(o.order)
			case source.RemovedOrder:
				originalSource.Remove(o.order.ID)
			case source.PriceChanged:
				originalSource.UpdatePrice(o.order.ID, o.order.Price)
			}
		}
	}(toExecute)

	go func() {
		poller.Run(ctx, originalSource, time.Duration(1)*time.Millisecond, out)
	}()

	// main act a supervisor, it will generate goroutines for the consumers AND the dispatcher
	var wg sync.WaitGroup
	var consumerChans []chan source.Event
	for id := range 2 {
		consumerCh := make(chan source.Event)
		consumerChans = append(consumerChans, consumerCh)

		wg.Add(1)
		go func() {
			defer wg.Done()
			switch id {
			case 0:
				consumerStat(consumerCh)
			case 1:
				consumerMirror(id, db, consumerCh)
			default:
				consumer(id, consumerCh)
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatcher(out, consumerChans)
	}()

	wg.Wait()
}

func dispatcher(in <-chan source.Event, outs []chan source.Event) {
	for event := range in {
		for _, out := range outs {
			out <- event
		}
	}

	for _, out := range outs {
		close(out)
	}
}

// first iteration of consumer
func consumer(id int, in <-chan source.Event) {
	for event := range in {
		slog.Info("event received", "consumerID", id, "event", event)
	}
}

// second iteration, the consumer stat keep stat and print them when it closes
func consumerStat(in <-chan source.Event) {
	added, removed, updated := 0, 0, 0

	for event := range in {
		switch event.Status {
		case source.AddedOrder:
			added++
		case source.RemovedOrder:
			removed++
		case source.PriceChanged:
			updated++
		}
	}

	slog.Info("consumerStat final result", "added", added, "removed", removed, "updated", updated)
}

// third iteration, it keeps its own order list based on the events
func consumerMirror(id int, db *sql.DB, in <-chan source.Event) {
	for event := range in {
		if event.Status == source.AddedOrder || event.Status == source.PriceChanged {
			upsertOrder(db, event)
		} else {
			deleteOrder(db, event)
		}

		slog.Info("event received", "consumerID", id, "event", event)
	}

	mirror := selectOrderAll(db)

	slog.Info("consumerMirror final result", "mirror", mirror)
}

func upsertOrder(db *sql.DB, event source.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO orders VALUES (?,?,?) ON DUPLICATE KEY UPDATE name = ?, price = ?", event.OrderID, event.Order.Name, event.Order.Price, event.Order.Name, event.Order.Price)
	if err != nil {
		slog.Error("failed to upsert order", "event", event, "error", err)
	}
}

func deleteOrder(db *sql.DB, event source.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "DELETE FROM orders WHERE ID = ?", event.OrderID)
	if err != nil {
		slog.Error("failed to delete order", "event", event, "error", err)
	}
}

func selectOrderAll(db *sql.DB) map[int]source.Order {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT * FROM orders")
	if err != nil {
		slog.Error("querying mirror failed", "error", err)
		return nil
	}
	defer rows.Close()

	mirror := make(map[int]source.Order)
	for rows.Next() {
		var o source.Order
		if err := rows.Scan(&o.ID, &o.Name, &o.Price); err != nil {
			slog.Error("failed to scan mirror row", "error", err)
			return nil
		}
		mirror[o.ID] = o
	}

	if err := rows.Err(); err != nil {
		slog.Error("error during mirror rows iteration", "error", err)
		return nil
	}

	return mirror
}
