package source

import (
	"slices"
	"testing"
)

func TestDiff(t *testing.T) {
	originSource := FakeSource{
		orders: map[int]Order{
			1: {1, "Pizza", 9.99},
			2: {2, "Burger", 5.45},
			3: {3, "Salad", 7.22},
		},
	}

	tests := map[string]struct {
		changes []func(s *FakeSource) error
		want    []Event
	}{
		"add": {
			changes: []func(s *FakeSource) error{
				func(s *FakeSource) error {
					order := Order{ID: 4, Name: "Icecream", Price: 3.99}
					return s.Add(order)
				},
			},
			want: []Event{{OrderID: 4, Order: Order{ID: 4, Name: "Icecream", Price: 3.99}, Status: AddedOrder}},
		},
		"remove": {
			changes: []func(s *FakeSource) error{
				func(s *FakeSource) error { return s.Remove(2) },
			},
			want: []Event{{OrderID: 2, Status: RemovedOrder}},
		},
		"price change": {
			changes: []func(s *FakeSource) error{
				func(s *FakeSource) error { return s.UpdatePrice(1, 10.99) },
			},
			want: []Event{{OrderID: 1, Order: Order{ID: 1, Name: "Pizza", Price: 10.99}, Status: PriceChanged}},
		},
		"multi event": {
			changes: []func(s *FakeSource) error{
				func(s *FakeSource) error { return s.UpdatePrice(1, 10.99) },
				func(s *FakeSource) error { return s.Remove(2) },
				func(s *FakeSource) error {
					order := Order{ID: 4, Name: "Icecream", Price: 3.99}
					return s.Add(order)
				},
			},
			want: []Event{
				{OrderID: 1, Order: Order{ID: 1, Name: "Pizza", Price: 10.99}, Status: PriceChanged},
				{OrderID: 2, Status: RemovedOrder},
				{OrderID: 4, Order: Order{ID: 4, Name: "Icecream", Price: 3.99}, Status: AddedOrder},
			},
		},
		"no change": {
			changes: []func(s *FakeSource) error{
				func(s *FakeSource) error { return nil }},
			want: []Event{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			source := originSource.clone()
			snap1 := source.Snapshot()

			for _, change := range test.changes {
				err := change(source)
				if err != nil {
					t.Fatalf("change order failed, got %v, want nil", err)
				}
			}

			snap2 := source.Snapshot()
			events := Diff(snap1, snap2)
			wantLen := len(test.want)
			if len(events) != wantLen {
				t.Fatalf("length of diff doesn't match, got %d, want %d", len(events), wantLen)
			}

			if wantLen == 0 {
				return
			}

			if !slices.Equal(events, test.want) {
				t.Errorf("received events don't match, got %#v, want %#v", events, test.want)
			}
		})
	}
}
