package source

import (
	"errors"
	"maps"
	"sync"
)

var ErrAlreadyExists = errors.New("order already exists")
var ErrOrderNotExists = errors.New("order does not exist")

type Order struct {
	ID    int
	Name  string
	Price float32
}
type FakeSource struct {
	mu     sync.Mutex
	orders map[int]Order
}

func (f *FakeSource) Snapshot() map[int]Order {
	f.mu.Lock()
	defer f.mu.Unlock()

	return maps.Clone(f.orders)
}

func (f *FakeSource) Add(o Order) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, ok := f.orders[o.ID]
	if ok {
		return ErrAlreadyExists
	}
	f.orders[o.ID] = o

	return nil
}

func (f *FakeSource) Remove(orderID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, ok := f.orders[orderID]
	if !ok {
		return ErrOrderNotExists
	}
	delete(f.orders, orderID)

	return nil
}

func (f *FakeSource) UpdatePrice(orderID int, price float32) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	o, ok := f.orders[orderID]
	if !ok {
		return ErrOrderNotExists
	}

	o.Price = price
	f.orders[orderID] = o

	return nil
}

func NewFakeSource() *FakeSource {
	fs := FakeSource{orders: map[int]Order{}}
	// we know it cannot fail, so we discard it (instead of ignoring it), Go good practice
	_ = fs.Add(Order{ID: 1, Name: "Pizza", Price: 9.99})
	_ = fs.Add(Order{ID: 2, Name: "Burger", Price: 5.45})
	return &fs
}
