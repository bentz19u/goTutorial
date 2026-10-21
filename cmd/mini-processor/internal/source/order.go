package source

import (
	"errors"
	"maps"
)

var ErrOrderNotExists = errors.New("order does not exist")

type order struct {
	ID    int
	name  string
	price float32
}

type fakeSource struct {
	orders map[int]order
}

func (f *fakeSource) snapshot() map[int]order {
	return maps.Clone(f.orders)
}

func (f *fakeSource) updatePrice(orderID int, price float32) error {
	o, ok := f.orders[orderID]
	if !ok {
		return ErrOrderNotExists
	}

	o.price = price
	f.orders[orderID] = o
	return nil
}
