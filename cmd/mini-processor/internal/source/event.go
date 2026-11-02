package source

import (
	"cmp"
	"fmt"
	"log/slog"
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
	Order   Order
	Status  EventStatus
}

func (es EventStatus) String() string {
	switch es {
	case AddedOrder:
		return "add"
	case RemovedOrder:
		return "remove"
	case PriceChanged:
		return "price update"
	default:
		return "unknown"
	}
}

func (e Event) String() string {
	return fmt.Sprintf("Order %d: %v", e.OrderID, e.Status)
}

func (e Event) LogValue() slog.Value {
	return slog.GroupValue(slog.Int("orderID", e.OrderID), slog.String("status", e.Status.String()), slog.Any("price", e.Order.Price))
}

func Diff(prev, curr map[int]Order) []Event {
	var events []Event
	for pID, p := range prev {
		c, ok := curr[pID]
		if !ok {
			events = append(events, Event{OrderID: pID, Status: RemovedOrder})
		} else if p.Price != c.Price {
			events = append(events, Event{OrderID: pID, Order: c, Status: PriceChanged})
		}
	}

	for cID, c := range curr {
		_, ok := prev[cID]
		if !ok {
			events = append(events, Event{OrderID: cID, Order: c, Status: AddedOrder})
		}
	}

	slices.SortFunc(events, func(a, b Event) int {
		return cmp.Compare(a.OrderID, b.OrderID)
	})

	return events
}
