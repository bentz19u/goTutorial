package source

import (
	"cmp"
	"fmt"
	"slices"
)

type EventStatus int

const (
	AddedOrder   EventStatus = iota // 0
	RemovedOrder                    // 1
	PriceChanged                    // 2
)

type Event struct {
	OrderID int
	Status  EventStatus
}

func (e Event) String() string {
	var status string

	switch e.Status {
	case AddedOrder:
		status = "add"
	case RemovedOrder:
		status = "remove"
	case PriceChanged:
		status = "price update"
	}

	return fmt.Sprintf("Order %d: %v", e.OrderID, status)
}

func Diff(prev, curr map[int]Order) []Event {
	var events []Event
	for pID, p := range prev {
		c, ok := curr[pID]
		if !ok {
			events = append(events, Event{OrderID: pID, Status: RemovedOrder})
		} else if p.Price != c.Price {
			events = append(events, Event{OrderID: pID, Status: PriceChanged})
		}
	}

	for cID := range curr {
		_, ok := prev[cID]
		if !ok {
			events = append(events, Event{OrderID: cID, Status: AddedOrder})
		}
	}

	slices.SortFunc(events, func(a, b Event) int {
		return cmp.Compare(a.OrderID, b.OrderID)
	})

	return events
}
