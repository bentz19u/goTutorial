package source

import (
	"errors"
	"maps"
)

var ErrAlreadyExists = errors.New("order already exists")
var ErrOrderNotExists = errors.New("order does not exist")

type Order struct {
	ID    int
	Name  string
	Price float32
}

type fakeSource struct {
	orders map[int]Order
}

func (f *fakeSource) snapshot() map[int]Order {
	return maps.Clone(f.orders)
}

func (f *fakeSource) add(o Order) error {
	_, ok := f.orders[o.ID]
	if ok {
		return ErrAlreadyExists
	}
	f.orders[o.ID] = o
	return nil
}

func (f *fakeSource) remove(orderID int) error {
	_, ok := f.orders[orderID]
	if !ok {
		return ErrOrderNotExists
	}
	delete(f.orders, orderID)
	return nil
}

func (f *fakeSource) updatePrice(orderID int, price float32) error {
	o, ok := f.orders[orderID]
	if !ok {
		return ErrOrderNotExists
	}

	o.Price = price
	f.orders[orderID] = o
	return nil
}
