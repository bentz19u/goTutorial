package source

import (
	"errors"
	"testing"
)

func TestFakeSource(t *testing.T) {
	originSource := fakeSource{
		orders: map[int]Order{
			1: {1, "Pizza", 9.99},
			2: {2, "Burger", 5.45},
			3: {3, "Salad", 7.22},
		},
	}

	tests := map[string]struct {
		ID      int
		price   float32
		wantErr error
	}{"Price changed": {1, 19.49, nil}, "Order doesn't exist": {5, 19.49, ErrOrderNotExists}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			source := originSource.clone()

			snap1 := source.snapshot()

			// changing price, err means the order doesn't exist
			err := source.updatePrice(test.ID, test.price)
			if !errors.Is(err, test.wantErr) {
				t.Errorf("error doesn't match, got %v, want %v", err, test.wantErr)
			}

			if err == nil {
				// checking second snap has the right price
				snap2 := source.snapshot()
				if snap2[test.ID].Price != test.price {
					t.Errorf("second snap price doesn't match, got %f, want %f", snap2[test.ID].Price, test.price)
				}

				// snap1 should still have original value
				if snap1[test.ID].Price != 9.99 {
					t.Errorf("first snap price doesn't match source, got %f, want %f", snap1[test.ID].Price, 9.99)
				}
			}
		})
	}
}
