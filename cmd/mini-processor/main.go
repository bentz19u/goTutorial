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

// the goal is to test the poller and also be able to interrupt it
func main() {
	originalSource := source.NewFakeSource()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

	out := make(chan source.Event)

	// so event can be generated
	toAdd := []source.Order{
		{ID: 3, Name: "Pizza1"},
		{ID: 4, Name: "Pizza2"},
		{ID: 5, Name: "Pizza3"},
		{ID: 6, Name: "Pizza4"},
		{ID: 7, Name: "Pizza5"},
		{ID: 8, Name: "Pizza6"},
		{ID: 9, Name: "Pizza7"},
		{ID: 10, Name: "Pizza8"},
	}

	go func(toAdd []source.Order) {
		for _, o := range toAdd {
			time.Sleep(300 * time.Millisecond)
			originalSource.Add(o)
		}
	}(toAdd)

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
			consumer(id, consumerCh)
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

func consumer(id int, in <-chan source.Event) {
	for event := range in {
		fmt.Printf("event received, consumer id %d, got %v\n", id, event)
	}
}
