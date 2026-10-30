package poller

import (
	"context"
	"testing"
	"time"

	"github.com/bentz19u/gotutorial/cmd/mini-processor/internal/source"
)

type fakeSource struct {
	idx   int
	snaps []map[int]source.Order
}

func (f *fakeSource) Snapshot() map[int]source.Order {
	if f.idx >= len(f.snaps)-1 {
		return f.snaps[f.idx]
	}

	snap := f.snaps[f.idx]
	f.idx++
	return snap
}

func TestRun(t *testing.T) {
	fakeS := fakeSource{
		idx: 0,
		snaps: []map[int]source.Order{
			{
				1: {ID: 1, Name: "Pizza", Price: 9.99},
				2: {ID: 2, Name: "Burger", Price: 5.45},
			},
			{ // add an element
				1: {ID: 1, Name: "Pizza", Price: 9.99},
				2: {ID: 2, Name: "Burger", Price: 5.45},
				3: {ID: 3, Name: "Salad", Price: 7.22},
			},
			{ // ID 2 price changed
				1: {ID: 1, Name: "Pizza", Price: 9.99},
				2: {ID: 2, Name: "Burger", Price: 6.45},
				3: {ID: 3, Name: "Salad", Price: 7.22},
			},
			{ // ID 3 removed
				1: {ID: 1, Name: "Pizza", Price: 9.99},
				2: {ID: 2, Name: "Burger", Price: 6.45},
			},
			{ // ID 4 added and ID 2 price changed again
				1: {ID: 1, Name: "Pizza", Price: 9.99},
				2: {ID: 2, Name: "Burger", Price: 5.45},
				4: {ID: 4, Name: "Icecream", Price: 8.26},
			},
		},
	}

	want := []source.Event{
		{OrderID: 3, Order: source.Order{ID: 3, Name: "Salad", Price: 7.22}, Status: source.AddedOrder},
		{OrderID: 2, Order: source.Order{ID: 2, Name: "Burger", Price: 6.45}, Status: source.PriceChanged},
		{OrderID: 3, Status: source.RemovedOrder},
		{OrderID: 2, Order: source.Order{ID: 2, Name: "Burger", Price: 5.45}, Status: source.PriceChanged},
		{OrderID: 4, Order: source.Order{ID: 4, Name: "Icecream", Price: 8.26}, Status: source.AddedOrder},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := make(chan source.Event)

	go func() {
		Run(ctx, &fakeS, time.Duration(1)*time.Millisecond, out)
	}()

	wantIdx := 0
	for got := range out {
		if got != want[wantIdx] {
			t.Errorf("event doesn't match, got %v, want %v", got, want[wantIdx])
		}

		if wantIdx < len(want)-1 {
			wantIdx++
		} else {
			cancel()
		}
	}
}
