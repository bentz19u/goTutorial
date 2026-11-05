package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/database"

	_ "github.com/go-sql-driver/mysql"
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
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
}
