package poller

import (
	"context"
	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/source"
	"time"
)

type Source interface {
	Snapshot() map[int]source.Order
}

func Run(ctx context.Context, src Source, interval time.Duration, out chan<- source.Event) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(out)

	prev := src.Snapshot()

	for {
		select {
		case <-ticker.C:
			curr := src.Snapshot()

			diff := source.Diff(prev, curr)
			for _, event := range diff {
				out <- event
			}

			prev = curr
		case <-ctx.Done():
			return
		}
	}
}
