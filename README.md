# GoTutorial

A project to learn Go, with a focus on concurrency and event-driven patterns.

## What's inside

- Practice files at the root: syntax, collections, structs, interfaces, errors, and concurrency (goroutines, channels, mutex, context cancellation).
- `cmd/mini-processor`: a small event-driven app in progress. It takes snapshots of an order source, diffs them, and emits events (order added, removed, price changed).
- Every topic is covered by table-driven tests.

## Run the tests

```
go test ./...
```

## Why

- Learn Go from the ground up, one concept at a time.
- Build toward a real event-driven processor, like the ones used in trading systems.
