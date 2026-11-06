package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"
	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/generator"

	_ "github.com/go-sql-driver/mysql"
)

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

	var wg sync.WaitGroup

	// generator, it will regularly add new orders to one of the DB, or close one existing one
	generator.StartGenerator(ctx, store, &wg)

	wg.Wait()
}

func initLogger() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
