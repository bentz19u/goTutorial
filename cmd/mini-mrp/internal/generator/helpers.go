package generator

import "math/rand/v2"

func randPick[T any](items []T) T {
	return items[rand.IntN(len(items))]
}

func randPrice() float64 {
	return 1 + rand.Float64()*9
}
