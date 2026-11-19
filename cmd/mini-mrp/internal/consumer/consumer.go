package consumer

import "github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/producer"

type Window interface {
	GetAllClosedTradesSorted() []producer.ClosedTrade
}
