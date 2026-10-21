package source

import (
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
		change func(s *fakeSource) error
		want   []Event
	}{
		"add": {
			change: func(s *fakeSource) error {
				order := Order{ID: 4, name: "Icecream", price: 3.99}
				return s.add(order)
			},
			want: []Event{{orderID: 4, status: addedOrder}},
		},
		"remove": {
			change: func(s *fakeSource) error { return s.remove(2) },
			want:   []Event{{orderID: 2, status: removedOrder}},
		},
		"price change": {
			change: func(s *fakeSource) error { return s.updatePrice(1, 10.99) },
			want:   []Event{{orderID: 1, status: priceChanged}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			source := originSource.clone()
			snap1 := source.snapshot()

			err := test.change(&source)
			if err != nil {
				t.Fatalf("adding order failed, got %v, want nil", err)
			}

			snap2 := source.snapshot()
			events := diff(snap1, snap2)
			if len(events) != 1 {
				t.Fatalf("length of diff doesn't match, got %d, want %d", len(events), 1)
			}

			// only one event in the list, we can get the first item here
			e := events[0]
			if e.orderID != test.want[0].orderID || e.status != test.want[0].status {
				t.Errorf("wrong event received, got id %d event %v, want id %d event %v", e.orderID, e.status, test.want[0].orderID, test.want[0].status)
			}
		})
	}
}
