# Go Trade Processor

A small event-driven trade-processing daemon in Go. It is a from-scratch rebuild of the core of a production report processor, written to understand every design decision in it.

The daemon polls two trade sources with different schemas, converts the trades to one format, keeps the recent ones in memory, and feeds several consumers.

## How it works

```
MT4 table ─┐                                    ┌─ pull consumers (read the window on a timer)
           ├─ producer ─ normalize ─ window ────┤
MT5 table ─┘      │                             └─ push consumers (receive each new batch)
                  └─ window check, heartbeat
```

- **Two sources, two schemas.** An MT4-style table where closing a trade updates its row, and an MT5-style table where closing a position appends a second deal. A generator opens and closes random trades in both, to stand in for real trading servers.
- **Delta polling with watermarks.** Every second the producer asks each source only for what is new. MT4 uses the close time, which can return the same trade twice at the boundary. MT5 uses the deal id as an exact cursor.
- **One common format.** Both kinds of trade are normalized into one struct. A trade is identified by server and ticket. Prices and profit are stored as integers with a fixed scale, rounded, never truncated.
- **Rolling in-memory window.** Closed trades are kept in a map keyed by server and ticket, behind a mutex, and removed on a timer once they are older than the window. A replayed trade overwrites the stored one instead of duplicating it.
- **Pull consumers.** They read the window through a read-only interface on their own timer. One logs the total profit in the window. The other keeps a profit total per account in MySQL.
- **Push consumers.** After each poll the producer hands the newly closed trades to a fanout, which calls each registered consumer in turn. A consumer's error is logged and does not stop the others or the producer. Only trades that were not already in the window are pushed.
- **A slow consumer that does not slow the producer.** One push consumer simulates slow work. It places each batch on a bounded queue and returns at once, and its own goroutine does the work. When the queue is full the batch is dropped and reported, because blocking would stall the producer.

## Design points

- **No double counting.** The per-account consumer writes the profit totals and the list of trades it has counted in one transaction, so either both are saved or neither. The primary key on that list rejects a trade that is already recorded, which rolls back the whole batch.
- **At-least-once delivery.** A trade can reach a consumer more than once, at the MT4 boundary or after a restart. Consumers that write must therefore be safe to repeat.
- **The producer checks its own work.** On a timer it compares the window against the sources over a time band and logs any trade that is missing from the window, or present in it without a source row.
- **Heartbeat.** After a poll cycle where both sources were read successfully, the producer writes a timestamp to a `system_states` table, at most once per interval. It is a separate statement, never part of a data transaction. An external watchdog could use it to tell that the daemon is progressing, not only running.
- **Timeouts and shutdown.** Every database call has a context timeout. Ctrl-C cancels a shared context and every goroutine stops, tracked by a wait group.
- **Structured logs.** JSON through `slog`.

## Run it

The daemon needs MySQL. Copy `.env.example` to `.env`, pick a password, and start the database with Docker (container `minimrp`, port 3307):

```
docker compose up -d
```

Set the same password in your shell and run the daemon:

```
$env:DB_PASSWORD="..."
go run -race ./cmd/mini-mrp
```

Timings: the generator acts every 300ms, the producer polls every second, trades stay in the window for one minute, the pull consumers run every 10 seconds, the window check every 15 seconds, and the heartbeat is written at most every 10 seconds.

Things to look at while it runs:

```
SELECT * FROM account_profits;
SELECT * FROM system_states;
```

In the log, the slow consumer processes one batch every 3 seconds and reports the batches it drops.

## Tests

```
go test ./...
```

The tests cover the pure logic: trade normalization, the rolling window, the window-versus-source comparison, the profit sum, and the selection of new trades for the per-account totals. Code that talks to MySQL has no unit tests. It is checked by running the daemon against the real database, where the generator compares the table counts to its own counters.

## Limits

This is a learning rebuild, not a production system:

- The sources are simulated by the generator, inside the same process and database.
- All tables are emptied at startup, and the watermarks live in memory, so every run starts from zero.
- The window check only reports differences. It does not repair them.
- Batches dropped by the slow consumer are not recovered.
- The list of counted trades is never cleaned up.

## Also in this repository

- `cmd/mini-processor`: an earlier, smaller daemon. It takes snapshots of an order source, diffs them, and broadcasts the changes as events to consumers over channels, including a MySQL mirror with idempotent upserts. It uses the second container, `minimirror` on port 3306, and runs with `go run -race ./cmd/mini-processor`.
- `practice`: the first exercises, one file per topic: collections, structs, interfaces, errors, and concurrency. Run them with `go run ./practice`.
