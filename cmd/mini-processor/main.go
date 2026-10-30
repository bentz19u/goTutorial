package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/poller"
	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/source"
)

type step struct {
	orderType source.EventStatus
	order     source.Order
}

// the goal is to test the poller and also be able to interrupt it
func main() {
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
			if id == 0 {
				consumerStat(consumerCh)
			} else {
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
		fmt.Printf("event received, consumer id %d, got %v\n", id, event)
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

	fmt.Printf("consumerStat final result, added %d, removed %d, updated %d\n", added, removed, updated)
}
