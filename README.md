# GoTutorial

A project to learn Go, with a focus on concurrency and event-driven patterns.

## What's inside

- Practice files at the root: syntax, collections, structs, interfaces, errors, and concurrency (goroutines, channels, mutex, context cancellation).
- `cmd/mini-processor`: a small event-driven daemon. It takes snapshots of an order source, diffs them, and emits events (order added, removed, price changed). Events are broadcast to consumers (a stats counter and a MySQL mirror), the initial state is replayed as events on startup, and logs are structured JSON via slog. The mirror writes are idempotent upserts, so replayed events are safe. Database calls carry per-query context timeouts.
- `cmd/mini-mrp`: a work-in-progress rebuild of a production report processor. The goal: poll two trade sources with deliberately different schemas (MT4 and MT5 style), normalize them into one format, keep a rolling in-memory window, and feed consumers. So far: its own MySQL database and the first source table.
- Every topic is covered by table-driven tests.

## Run the tests

```
go test ./...
```

## Run the daemon

The mirror consumer needs MySQL. Copy `.env.example` to `.env`, pick a password, then start the databases with Docker (one container per daemon: `minimirror` on 3306, `minimrp` on 3307):

```
docker compose up -d
```

Set the same password in your shell and run the daemon:

```
$env:DB_PASSWORD="..."
go run -race ./cmd/mini-processor
```

A scripted mutator changes the orders every 300ms; stop with Ctrl-C to see each consumer's final result and check the mirror with `SELECT * FROM orders`.

## Why

- Learn Go from the ground up, one concept at a time.
- Build toward a real event-driven processor, like the ones used in trading systems.
