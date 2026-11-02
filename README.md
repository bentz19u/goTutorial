# GoTutorial

A project to learn Go, with a focus on concurrency and event-driven patterns.

## What's inside

- Practice files at the root: syntax, collections, structs, interfaces, errors, and concurrency (goroutines, channels, mutex, context cancellation).
- `cmd/mini-processor`: a small event-driven daemon. It takes snapshots of an order source, diffs them, and emits events (order added, removed, price changed). Events are broadcast to consumers (a stats counter and an in-memory mirror), the initial state is replayed as events on startup, and logs are structured JSON via slog.
- Every topic is covered by table-driven tests.

## Run the tests

```
go test ./...
```

## Run the daemon

```
go run -race ./cmd/mini-processor
```

A scripted mutator changes the orders every 300ms; stop with Ctrl-C to see each consumer's final result.

## Why

- Learn Go from the ground up, one concept at a time.
- Build toward a real event-driven processor, like the ones used in trading systems.
