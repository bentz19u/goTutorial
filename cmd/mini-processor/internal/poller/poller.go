package poller

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/source"
)

type Source interface {
	Snapshot() map[int]source.Order
}

func Run(ctx context.Context, src Source, interval time.Duration, out chan<- source.Event) {
	var lock sync.Mutex
	var inflight sync.WaitGroup

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(out)

	var prev map[int]source.Order

	for {
		select {
		case <-ticker.C:
			inflight.Add(1)
			// has to be in a goroutine so that it doesn't block the main loop
			// defer lock.Unlock() will also trigger when the goroutine is finished (and it acquired the lock)
			go func() {
				defer inflight.Done()
				if lock.TryLock() {
					defer lock.Unlock()
					curr := src.Snapshot()

					diff := source.Diff(prev, curr)
					for _, event := range diff {
						out <- event
					}

					prev = curr
				} else {
					slog.Warn("poller run lock still locked, skipping tick")
				}
			}()
		case <-ctx.Done():
			// we wait for all tick goroutines to finish to cleanly exit
			inflight.Wait()
			return
		}
	}
}
