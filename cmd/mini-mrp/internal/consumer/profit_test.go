package consumer

import (
	"testing"

	"github.com/bentz19u/gotutorial/cmd/mini-mrp/internal/producer"
)

func TestSumProfit(t *testing.T) {
	var cts []producer.ClosedTrade
	cts = append(cts, producer.ClosedTrade{Profit: 125})
	cts = append(cts, producer.ClosedTrade{Profit: -201})
	cts = append(cts, producer.ClosedTrade{Profit: -405})
	cts = append(cts, producer.ClosedTrade{Profit: -291})
	got := sumProfit(cts)

	if got != -7.72 {
		t.Errorf("Wrong sum profit: got %f, want %f", got, -7.72)
	}
}
