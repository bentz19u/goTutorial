package source

import (
	"slices"
	"testing"
)

func TestDiff(t *testing.T) {
	originSource := fakeSource{
		orders: map[int]Order{
			1: {1, "Pizza", 9.99},
			2: {2, "Burger", 5.45},
			3: {3, "Salad", 7.22},
		},
	}

	tests := map[string]struct {
		changes []func(s *fakeSource) error
		want    []Event
	}{
		"add": {
			changes: []func(s *fakeSource) error{
				func(s *fakeSource) error {
					order := Order{ID: 4, name: "Icecream", price: 3.99}
					return s.add(order)
				},
			},
			want: []Event{{orderID: 4, status: addedOrder}},
		},
		"remove": {
			changes: []func(s *fakeSource) error{
				func(s *fakeSource) error { return s.remove(2) },
			},
			want: []Event{{orderID: 2, status: removedOrder}},
		},
		"price change": {
			changes: []func(s *fakeSource) error{
				func(s *fakeSource) error { return s.updatePrice(1, 10.99) },
			},
			want: []Event{{orderID: 1, status: priceChanged}},
		},
		"multi event": {
			changes: []func(s *fakeSource) error{
				func(s *fakeSource) error { return s.updatePrice(1, 10.99) },
				func(s *fakeSource) error { return s.remove(2) },
				func(s *fakeSource) error {
					order := Order{ID: 4, name: "Icecream", price: 3.99}
					return s.add(order)
				},
			},
			want: []Event{
				{orderID: 1, status: priceChanged},
				{orderID: 2, status: removedOrder},
				{orderID: 4, status: addedOrder},
			},
		},
		"no change": {
			changes: []func(s *fakeSource) error{
				func(s *fakeSource) error { return nil }},
			want: []Event{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			source := originSource.clone()
			snap1 := source.snapshot()

			for _, change := range test.changes {
				err := change(&source)
				if err != nil {
					t.Fatalf("change order failed, got %v, want nil", err)
				}
			}

			snap2 := source.snapshot()
			events := diff(snap1, snap2)
			wantLen := len(test.want)
			if len(events) != wantLen {
				t.Fatalf("length of diff doesn't match, got %d, want %d", len(events), wantLen)
			}

			if wantLen == 0 {
				return
			}

			if !slices.Equal(events, test.want) {
				t.Errorf("received events don't match, got %v, want %v", events, test.want)
			}
		})
	}
}
