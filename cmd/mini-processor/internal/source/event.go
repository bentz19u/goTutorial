package source

type EventStatus int

const (
	addedOrder   EventStatus = iota // 0
	removedOrder                    // 1
	priceChanged                    // 2
)

type Event struct {
	orderID int
	status  EventStatus
}

func diff(prev, curr map[int]Order) []Event {
	var events []Event
	for pID, p := range prev {
		c, ok := curr[pID]
		if !ok {
			events = append(events, Event{orderID: pID, status: removedOrder})
		} else if p.price != c.price {
			events = append(events, Event{orderID: pID, status: priceChanged})
		}
	}

	for cID := range curr {
		_, ok := prev[cID]
		if !ok {
			events = append(events, Event{orderID: cID, status: addedOrder})
		}
	}

	return events
}
