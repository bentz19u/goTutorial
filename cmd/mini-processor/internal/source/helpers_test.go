package source

import (
	"maps"
)

func (s fakeSource) clone() fakeSource {
	s.orders = maps.Clone(s.orders)
	return s
}
