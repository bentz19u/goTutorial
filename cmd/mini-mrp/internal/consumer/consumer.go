package consumer

import "github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/model"

type Window interface {
	GetAllClosedTradesSorted() []model.ClosedTrade
}
