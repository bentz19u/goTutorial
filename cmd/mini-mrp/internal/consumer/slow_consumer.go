package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"
)

type SlowConsumer struct {
	channel chan []model.ClosedTrade
}

func (sc *SlowConsumer) GetName() string {
	return "Slow Consumer"
}

func (sc *SlowConsumer) OnEvent(ctx context.Context, trades []model.ClosedTrade) error {
	select {
	case sc.channel <- trades:
		return nil
	default:
		return fmt.Errorf("slow consumer queue is full, %d trades are gone", len(trades))
	}
}

func (sc *SlowConsumer) Start(ctx context.Context, wg *sync.WaitGroup) {
	sc.channel = make(chan []model.ClosedTrade, 2)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case trades := <-sc.channel:
				// simulate slow word
				select {
				case <-time.After(3 * time.Second):
				case <-ctx.Done():
					return
				}

				slog.Info("slow consumer processed new trades", "count", len(trades))
			case <-ctx.Done():
				slog.Info("slow consumer shutting down")
				return
			}
		}
	}()
}
